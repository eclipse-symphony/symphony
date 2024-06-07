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
var targetcontainerlog = logf.Log.WithName("targetcontainer-resource")
var myTargetContainerClient client.Client
var targetContainerWebhookValidationMetrics *metrics.Metrics

func (r *TargetContainer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myTargetContainerClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &TargetContainer{}, ".metadata.name", func(rawObj client.Object) []string {
		target := rawObj.(*TargetContainer)
		return []string{target.Name}
	})

	// initialize the controller operation metrics
	if targetContainerWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		targetContainerWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-fabric-symphony-v1-targetcontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targetcontainers,verbs=create;update,versions=v1,name=mtargetcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &TargetContainer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *TargetContainer) Default() {
	targetcontainerlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-fabric-symphony-v1-targetcontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targetcontainers,verbs=create;update;delete,versions=v1,name=vtargetcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &TargetContainer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *TargetContainer) ValidateCreate() (admission.Warnings, error) {
	targetcontainerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *TargetContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	targetcontainerlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *TargetContainer) ValidateDelete() (admission.Warnings, error) {
	targetcontainerlog.Info("validate delete", "name", r.Name)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteTargetContainer()
	if validationError != nil {
		targetContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		targetContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

func (r *TargetContainer) validateDeleteTargetContainer() error {
	return r.validateTargets()
}

func (r *TargetContainer) validateTargets() error {
	var target TargetList
	err := myTargetContainerClient.List(context.Background(), &target, client.InNamespace(r.Namespace), client.MatchingFields{".spec.rootResource": r.Name})
	if err != nil {
		targetcontainerlog.Error(err, "could not list targets", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("could not list targets for target container %s.", r.Name))
	}

	if len(target.Items) != 0 {
		targetcontainerlog.Error(err, "targets are not empty", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("targets with root resource '%s' are not empty", r.Name))
	}

	return nil
}
