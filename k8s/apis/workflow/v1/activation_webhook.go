/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"gopls-workspace/apis/dynamicclient"
	"gopls-workspace/apis/metrics/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	"time"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var activationlog = logf.Log.WithName("activation-resource")
var myActivationClient client.Reader
var activationWebhookValidationMetrics *metrics.Metrics
var activationValidator validation.ActivationValidator

func (r *Activation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myActivationClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Activation{}, ".metadata.name", func(rawObj client.Object) []string {
		activation := rawObj.(*Activation)
		return []string{activation.Name}
	})

	// initialize the controller operation metrics
	if activationWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		activationWebhookValidationMetrics = metrics
	}

	activationValidator = validation.NewActivationValidator(func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(ctx, validation.Campaign, name, namespace)
	})

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-activation,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=mactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Activation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Activation) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), activationlog)
	diagnostic.InfoWithCtx(activationlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	if r.Spec.Campaign != "" {
		diagnostic.InfoWithCtx(activationlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec.campaign", r.Spec.Campaign)
		// Remove api_constants.Campaign from r.Labels if it exists
		if _, exists := r.Labels[api_constants.Campaign]; exists {
			delete(r.Labels, api_constants.Campaign)
		}
		var campaignResult Campaign
		err := myActivationClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.Campaign), Namespace: r.Namespace}, &campaignResult)
		if err != nil {
			diagnostic.ErrorWithCtx(activationlog, ctx, err, "failed to get campaign", "name", r.Name, "namespace", r.Namespace)
			return
		}
		r.Labels[api_constants.CampaignUid] = string(campaignResult.UID)

	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-activation,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=vactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Activation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Activation) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myActivationClient, context.TODO(), activationlog)
	diagnostic.InfoWithCtx(activationlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateActivation(ctx)
	if validationError != nil {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Activation) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myActivationClient, context.TODO(), activationlog)

	diagnostic.InfoWithCtx(activationlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldActivation, ok := old.(*Activation)
	if !ok {
		err := fmt.Errorf("expected an Activation object")
		diagnostic.ErrorWithCtx(activationlog, ctx, err, "failed to convert old object to Activation", "name", r.Name, "namespace", r.Namespace)
		return nil, err
	}
	// Compare the Spec of the current and old Activation objects
	validationError := r.validateUpdateActivation(ctx, oldActivation)
	if validationError != nil {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.InstanceResourceType,
		)
	} else {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.InstanceResourceType,
		)
	}
	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Activation) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myActivationClient, context.TODO(), activationlog)

	diagnostic.InfoWithCtx(activationlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, nil
}

func (r *Activation) validateCreateActivation(ctx context.Context) error {
	state, err := r.ConvertActivationState()
	if err != nil {
		diagnostic.ErrorWithCtx(activationlog, ctx, err, "validate create activation - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := activationValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}
	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(activationlog, ctx, err, "validate create activation", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Activation) validateUpdateActivation(ctx context.Context, oldActivation *Activation) error {
	state, err := r.ConvertActivationState()
	if err != nil {
		diagnostic.ErrorWithCtx(activationlog, ctx, err, "validate update activation - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	old, err := oldActivation.ConvertActivationState()
	if err != nil {
		diagnostic.ErrorWithCtx(activationlog, ctx, err, "validate update activation - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := activationValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(activationlog, ctx, err, "validate update activation", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Activation) ConvertActivationState() (model.ActivationState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to activation state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.ActivationState{}, retErr
	}
	var state model.ActivationState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.ActivationState{}, retErr
	}
	return state, nil
}
