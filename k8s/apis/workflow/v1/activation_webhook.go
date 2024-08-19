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

	"time"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
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
var activationWebhookValidationMetrics *metrics.Metrics

func (r *Activation) SetupWebhookWithManager(mgr ctrl.Manager) error {
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

	validation.CampaignLookupFunc = func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(validation.Campaign, name, namespace)
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-activation,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=mactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Activation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Activation) Default() {
	activationlog.Info("default", "name", r.Name, "spec", r.Spec, "status", r.Status)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	if r.Spec.Campaign != "" {
		activationlog.Info("default", "name", r.Name, "spec.campaign", r.Spec.Campaign)
		r.Labels["campaign"] = validation.ConvertReferenceToObjectName(r.Spec.Campaign)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-activation,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=mactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Activation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Activation) ValidateCreate() (admission.Warnings, error) {
	activationlog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateActivation()
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
	activationlog.Info("validate update", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldActivation, ok := old.(*Activation)
	if !ok {
		return nil, fmt.Errorf("expected an Activation object")
	}
	// Compare the Spec of the current and old Activation objects
	validationError := r.validateUpdateActivation(oldActivation)
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
	activationlog.Info("validate delete", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, nil
}

func (r *Activation) validateCreateActivation() error {
	state, err := r.ConvertActivationState()
	if err != nil {
		return err
	}
	ErrorFields := validation.ValidateCreateOrUpdate(context.TODO(), state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
}

func (r *Activation) validateUpdateActivation(oldActivation *Activation) error {
	state, err := r.ConvertActivationState()
	if err != nil {
		return err
	}
	old, err := oldActivation.ConvertActivationState()
	if err != nil {
		return err
	}
	ErrorFields := validation.ValidateCreateOrUpdate(context.TODO(), state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
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
