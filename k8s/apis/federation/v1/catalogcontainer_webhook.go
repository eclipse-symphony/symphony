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
var catalogcontainerlog = logf.Log.WithName("catalogcontainer-resource")
var myCatalogContainerClient client.Client
var catalogContainerWebhookValidationMetrics *metrics.Metrics

func (r *CatalogContainer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogContainerClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &CatalogContainer{}, ".metadata.name", func(rawObj client.Object) []string {
		catalog := rawObj.(*CatalogContainer)
		return []string{catalog.Name}
	})

	// initialize the controller operation metrics
	if catalogContainerWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogContainerWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalogcontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogcontainers,verbs=create;update,versions=v1,name=mcatalogcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CatalogContainer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CatalogContainer) Default() {
	catalogcontainerlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalogcontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogcontainers,verbs=create;update;delete,versions=v1,name=vcatalogcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CatalogContainer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogContainer) ValidateCreate() (admission.Warnings, error) {
	catalogcontainerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	catalogcontainerlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogContainer) ValidateDelete() (admission.Warnings, error) {
	catalogcontainerlog.Info("validate delete", "name", r.Name)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteCatalogContainer()
	if validationError != nil {
		catalogContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		catalogContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

func (r *CatalogContainer) validateDeleteCatalogContainer() error {
	return r.validateCatalogs()
}

func (r *CatalogContainer) validateCatalogs() error {
	var catalog CatalogList
	err := myCatalogContainerClient.List(context.Background(), &catalog, client.InNamespace(r.Namespace), client.MatchingFields{".spec.rootResource": r.Name})
	if err != nil {
		catalogcontainerlog.Error(err, "could not list catalogs", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("could not list catalogs for catalog container %s.", r.Name))
	}

	if len(catalog.Items) != 0 {
		catalogcontainerlog.Error(err, "catalogs are not empty", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("catalogs with root resource '%s' are not empty", r.Name))
	}

	return nil
}
