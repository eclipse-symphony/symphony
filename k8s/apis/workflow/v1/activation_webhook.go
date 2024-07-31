/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"
	"gopls-workspace/apis/metrics/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"time"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/k8s/utils"
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
var myActivationClient client.Client
var activationWebhookValidationMetrics *metrics.Metrics

func (r *Activation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myActivationClient = mgr.GetClient()
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

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-activation,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=mactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Activation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Activation) Default() {
	activationlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-activation,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=activations,verbs=create;update,versions=v1,name=mactivation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Activation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Activation) ValidateCreate() (admission.Warnings, error) {
	activationlog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.ActivationOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := context.TODO()
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, ctx, activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being created", r.Name)

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
	ctx := context.TODO()
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, ctx, activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being updated", r.Name)

	validateUpdateTime := time.Now()
	oldActivation, ok := old.(*Activation)
	if !ok {
		return nil, fmt.Errorf("expected an Activation object")
	}
	// Compare the Spec of the current and old Activation objects
	validationError := r.validateSpecOnUpdate(oldActivation)
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
	ctx := context.TODO()
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, ctx, activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being deleted", r.Name)

	return nil, nil
}

func (r *Activation) validateCreateActivation() error {
	var allErrs field.ErrorList

	if err := r.validateCampaignOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
}

func (r *Activation) validateCampaignOnCreate() *field.Error {
	if r.Spec.Campaign == "" {
		return field.Invalid(field.NewPath("spec").Child("campaign"), r.Spec.Campaign, "campaign must not be empty")
	}
	campaignName := utils.ReplaceLastSeperator(r.Spec.Campaign, ":", constants.ResourceSeperator)
	var campaign Campaign
	err := myActivationClient.Get(context.Background(), client.ObjectKey{Name: campaignName, Namespace: r.Namespace}, &campaign)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("campaign"), r.Spec.Campaign, "campaign doesn't exist")
	}
	return nil
}

func (r *Activation) validateSpecOnUpdate(oldActivation *Activation) error {
	var allErrs field.ErrorList
	if r.Spec.Campaign != oldActivation.Spec.Campaign {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("campaign"), r.Spec.Campaign, "updates to activation spec.Campaign are not allowed"))
	}
	if r.Spec.Stage != oldActivation.Spec.Stage {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("stage"), r.Spec.Stage, "updates to activation spec.Stage are not allowed"))
	}
	if r.Spec.Inputs.String() != oldActivation.Spec.Inputs.String() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("inputs"), r.Spec.Inputs, "updates to activation spec.Inputs are not allowed"))
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Activation"}, r.Name, allErrs)
}
