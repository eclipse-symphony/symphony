/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var solutionlog = logf.Log.WithName("solution-resource")
var mySolutionClient client.Client

func (r *Solution) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySolutionClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Solution{}, ".spec.displayName", func(rawObj client.Object) []string {
		target := rawObj.(*Solution)
		return []string{target.Spec.DisplayName}
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-solution,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update,versions=v1,name=msolution.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Solution{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Solution) Default() {
	solutionlog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solution,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update,versions=v1,name=vsolution.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Solution{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateCreate() (admission.Warnings, error) {
	solutionlog.Info("validate create", "name", r.Name)

	return nil, r.validateCreateSolution()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	solutionlog.Info("validate update", "name", r.Name)

	return nil, r.validateUpdateSolution()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Solution) ValidateDelete() (admission.Warnings, error) {
	solutionlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func (r *Solution) validateCreateSolution() error {
	var solutions SolutionList
	mySolutionClient.List(context.Background(), &solutions, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(solutions.Items) != 0 {
		return apierrors.NewBadRequest(fmt.Sprintf("solution display name '%s' is already taken", r.Spec.DisplayName))
	}
	return nil
}

func (r *Solution) validateUpdateSolution() error {
	var solutions SolutionList
	err := mySolutionClient.List(context.Background(), &solutions, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return apierrors.NewInternalError(err)
	}
	if !(len(solutions.Items) == 0 || len(solutions.Items) == 1 && solutions.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return apierrors.NewBadRequest(fmt.Sprintf("solution display name '%s' is already taken", r.Spec.DisplayName))
	}
	return nil
}
