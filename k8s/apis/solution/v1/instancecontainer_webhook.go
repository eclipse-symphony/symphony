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
var instancecontainerlog = logf.Log.WithName("instancecontainer-resource")
var myInstanceContainerClient client.Client
var instanceContainerWebhookValidationMetrics *metrics.Metrics

func (r *InstanceContainer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myInstanceContainerClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &InstanceContainer{}, ".metadata.name", func(rawObj client.Object) []string {
		instance := rawObj.(*InstanceContainer)
		return []string{instance.Name}
	})

	// initialize the controller operation metrics
	if instanceContainerWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		instanceContainerWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-instancecontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instancecontainers,verbs=create;update,versions=v1,name=minstancecontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &InstanceContainer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *InstanceContainer) Default() {
	instancecontainerlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instancecontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instancecontainers,verbs=create;update;delete,versions=v1,name=vinstancecontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &InstanceContainer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceContainer) ValidateCreate() (admission.Warnings, error) {
	instancecontainerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	instancecontainerlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceContainer) ValidateDelete() (admission.Warnings, error) {
	instancecontainerlog.Info("validate delete", "name", r.Name)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteInstanceContainer()
	if validationError != nil {
		instanceContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		instanceContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

func (r *InstanceContainer) validateDeleteInstanceContainer() error {
	return r.validateInstances()
}

func (r *InstanceContainer) validateInstances() error {
	var instance InstanceList
	err := myInstanceContainerClient.List(context.Background(), &instance, client.InNamespace(r.Namespace), client.MatchingFields{".spec.rootResource": r.Name})
	if err != nil {
		instancecontainerlog.Error(err, "could not list instances", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("could not list instances for instance container %s.", r.Name))
	}

	if len(instance.Items) != 0 {
		instancecontainerlog.Error(err, "instances are not empty", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("instances with root resource '%s' are not empty", r.Name))
	}

	return nil
}
