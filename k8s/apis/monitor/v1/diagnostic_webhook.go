/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"gopls-workspace/apis/metrics/v1"
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
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	diagnostic.InfoWithCtx(diagnosticlog, ctx, "default", "name", r.Name, "namespace", r.Namespace)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-monitor-symphony-v1-diagnostic,mutating=false,failurePolicy=fail,sideEffects=None,groups=monitor.symphony,resources=diagnostics,verbs=create;update;delete,versions=v1,name=mdiagnostic.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Diagnostic{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Diagnostic) ValidateCreate() (admission.Warnings, error) {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	diagnostic.InfoWithCtx(diagnosticlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)

	// insert validation logic here
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
	// insert validation logic here
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	diagnostic.InfoWithCtx(diagnosticlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
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
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), diagnosticlog)
	diagnostic.InfoWithCtx(diagnosticlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)

	// insert validation logic here

	// clear global diagnostic resource cache
	ClearDiagnosticResourceCache()
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
	err := apierrors.NewInvalid(schema.GroupKind{Group: "monitor.symphony", Kind: "Diagnostic"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(diagnosticlog, ctx, err, "Failed to validate create or update Diagnostic resource", "name", r.Name, "namespace", r.Namespace)
	return err
}

// we need to ensure for one edge location, there's only one diagnostic resource in this namespace.
func (r *Diagnostic) validateUniqueInEdgeLocations(ctx context.Context) *field.Error {
	diagnostic.InfoWithCtx(diagnosticlog, ctx, "validate unique in edge locations", "name", r.Name, "namespace", r.Namespace)

	filedErr := ValidateDiagnosticResourceAnnoations(r.Annotations)
	if filedErr != nil {
		return filedErr
	}

	existingResource, err := GetGlobalDiagnosticResourceInCluster(r.Annotations, myDiagnosticClient, ctx, diagnosticlog)
	if err != nil {
		return field.InternalError(nil, v1alpha2.NewCOAError(err, "Failed to check uniqueness of diagnostic resource in cluster", v1alpha2.InternalError))
	}
	// if existing resource is not nil and the name is different, then it's a conflict
	if existingResource != nil && (existingResource.Name != r.Name || existingResource.Namespace != r.Namespace) {
		return GenerateDiagnosticResourceUniquenessFieldError(r.Name, r.Namespace, existingResource)
	}
	return nil
}
