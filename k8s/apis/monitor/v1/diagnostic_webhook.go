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
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var diagnosticlog = logf.Log.WithName("diagnostic-resource")
var myDiagnosticClient client.Reader
var diagnosticWebhookValidationMetrics *metrics.Metrics

func (r *Diagnostic) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myDiagnosticClient = mgr.GetAPIReader()

	// initialize the controller operation metrics
	if diagnosticWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		diagnosticWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-monitor-symphony-v1-diagnostic,mutating=true,failurePolicy=fail,sideEffects=None,groups=monitor.symphony,resources=diagnostics,verbs=create;update;delete,versions=v1,name=mdiagnostic.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Diagnostic{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Diagnostic) Default() {
	diagnosticlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-monitor-symphony-v1-diagnostic,mutating=false,failurePolicy=fail,sideEffects=None,groups=monitor.symphony,resources=diagnostics,verbs=create;update;delete,versions=v1,name=mdiagnostic.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Diagnostic{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Diagnostic) ValidateCreate() (admission.Warnings, error) {
	diagnosticlog.Info("validate create", "name", r.Name)

	// insert validation logic here
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	validatedCreateTime := time.Now()
	validatedError := r.validateCreateOrUpdateImpl(ctx)
	if validatedError != nil {
		diagnosticWebhookValidationMetrics.ControllerValidationLatency(
			validatedCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.DiagnosticResourceType,
		)
	} else {
		diagnosticWebhookValidationMetrics.ControllerValidationLatency(
			validatedCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.DiagnosticResourceType,
		)
	}

	return nil, validatedError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Diagnostic) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	diagnosticlog.Info("validate update", "name", r.Name)

	// insert validation logic here
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	validatedUpdateTime := time.Now()
	validatedError := r.validateCreateOrUpdateImpl(ctx)
	if validatedError != nil {
		diagnosticWebhookValidationMetrics.ControllerValidationLatency(
			validatedUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.DiagnosticResourceType,
		)
	} else {
		diagnosticWebhookValidationMetrics.ControllerValidationLatency(
			validatedUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.DiagnosticResourceType,
		)
	}
	return nil, validatedError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Diagnostic) ValidateDelete() (admission.Warnings, error) {
	diagnosticlog.Info("validate delete", "name", r.Name)

	// insert validation logic here
	return nil, nil
}

func (r *Diagnostic) validateCreateOrUpdateImpl(ctx context.Context) error {
	var allErrs field.ErrorList
	if err := r.validateUniqueInEdgeLocations(ctx); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "monitor.symphony", Kind: "Diagnostic"}, r.Name, allErrs)
}

// we need to ensure for one edge location, there's only one diagnostic resource in this namespace.
func (r *Diagnostic) validateUniqueInEdgeLocations(ctx context.Context) *field.Error {
	diagnosticlog.Info("validate create", "name", r.Name, "namespace", r.Namespace)

	edgeLocation := r.Annotations[constants.AzureEdgeLocationKey]
	if edgeLocation == "" {
		return field.Required(field.NewPath("metadata.annotations").Child(constants.AzureEdgeLocationKey), "Azure Edge Location is required")
	}

	existingResource, err := GetDiagnosticCustomResource(r.Namespace, edgeLocation, myDiagnosticClient, ctx, diagnosticlog)
	if err != nil {
		return field.InternalError(nil, v1alpha2.NewCOAError(err, fmt.Sprintf("Failed to check uniqueness of diagnostic resource on edge location: %s", edgeLocation), v1alpha2.InternalError))
	}
	if existingResource != nil {
		return field.Invalid(field.NewPath("metadata.annotations").Child(constants.AzureEdgeLocationKey), edgeLocation, fmt.Sprintf("Diagnostic resource already exists for edge location: %s", edgeLocation))
	}
	return nil
}
