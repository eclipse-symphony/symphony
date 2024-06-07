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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var solutioncontainerlog = logf.Log.WithName("solutioncontainer-resource")
var mySolutionContainerClient client.Client
var solutionContainerWebhookValidationMetrics *metrics.Metrics

func (r *SolutionContainer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionContainerClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &SolutionContainer{}, ".metadata.name", func(rawObj client.Object) []string {
		solution := rawObj.(*SolutionContainer)
		return []string{solution.Name}
	})

	// initialize the controller operation metrics
	if solutionContainerWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		solutionContainerWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-solutioncontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutioncontainers,verbs=create;update,versions=v1,name=msolutioncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SolutionContainer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SolutionContainer) Default() {
	solutioncontainerlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solutioncontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutioncontainers,verbs=create;update;delete,versions=v1,name=vsolutioncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SolutionContainer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionContainer) ValidateCreate() (admission.Warnings, error) {
	solutioncontainerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	solutioncontainerlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SolutionContainer) ValidateDelete() (admission.Warnings, error) {
	solutioncontainerlog.Info("validate delete", "name", r.Name)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteSolutionContainer()
	if validationError != nil {
		solutionContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		solutionContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

func (r *SolutionContainer) validateDeleteSolutionContainer() error {
	return r.validateSolutions()
}

func (r *SolutionContainer) validateSolutions() error {
	var solution SolutionList
	err := mySolutionContainerClient.List(context.Background(), &solution, client.InNamespace(r.Namespace), client.MatchingFields{".spec.rootResource": r.Name})
	if err != nil {
		solutioncontainerlog.Error(err, "could not list solutions", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("could not list solutions for solution container %s.", r.Name))
	}

	if len(solution.Items) != 0 {
		solutioncontainerlog.Error(err, "solutions are not empty", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("solutions with root resource '%s' are not empty", r.Name))
	}

	return nil
}
