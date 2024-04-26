/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"encoding/json"
	"gopls-workspace/apis/metrics/v1"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var cataloglog = logf.Log.WithName("catalog-resource")
var myCatalogClient client.Client
var catalogWebhookValidationMetrics *metrics.Metrics

func (r *Catalog) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Catalog{}, ".spec.name", func(rawObj client.Object) []string {
		target := rawObj.(*Catalog)
		return []string{target.Name}
	})

	// initialize the controller operation metrics
	if catalogWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalog,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update,versions=v1,name=mcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Catalog{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Catalog) Default() {
	cataloglog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update,versions=v1,name=vcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Catalog{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateCreate() error {
	cataloglog.Info("validate create", "name", r.Name)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCatalog()
	if validationError != nil {
		catalogWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		catalogWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateUpdate(old runtime.Object) error {
	cataloglog.Info("validate update", "name", r.Name)

	validateUpdateTime := time.Now()
	validationError := r.validateUpdateCatalog()
	if validationError != nil {
		catalogWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		catalogWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateDelete() error {
	cataloglog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *Catalog) validateCreateCatalog() error {
	return r.checkSchema()
}

func (r *Catalog) checkSchema() error {

	if r.Spec.Metadata != nil {
		if schemaName, ok := r.Spec.Metadata["schema"]; ok {
			cataloglog.Info("Find schema name", "name", schemaName)
			var catalogs CatalogList
			err := myCatalogClient.List(context.Background(), &catalogs, client.InNamespace(r.ObjectMeta.Namespace), client.MatchingFields{".spec.name": schemaName})
			if err != nil || len(catalogs.Items) == 0 {
				cataloglog.Error(err, "Could not find the required schema.", "name", schemaName)
				return v1alpha2.NewCOAError(err, "schema not found", v1alpha2.NotFound)
			}

			jData, _ := json.Marshal(catalogs.Items[0].Spec.Properties)
			var properties map[string]interface{}
			err = json.Unmarshal(jData, &properties)
			if err != nil {
				cataloglog.Error(err, "Invalid schema.", "name", schemaName)
				return v1alpha2.NewCOAError(err, "invalid schema", v1alpha2.ValidateFailed)
			}
			if spec, ok := properties["spec"]; ok {
				var schemaObj utils.Schema
				jData, _ := json.Marshal(spec)
				err := json.Unmarshal(jData, &schemaObj)
				if err != nil {
					cataloglog.Error(err, "Invalid schema.", "name", schemaName)
					return v1alpha2.NewCOAError(err, "invalid schema", v1alpha2.ValidateFailed)
				}
				jData, _ = json.Marshal(r.Spec.Properties)
				var properties map[string]interface{}
				err = json.Unmarshal(jData, &properties)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return v1alpha2.NewCOAError(err, "invalid properties", v1alpha2.ValidateFailed)
				}
				result, err := schemaObj.CheckProperties(properties, nil)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return v1alpha2.NewCOAError(err, "invalid properties", v1alpha2.ValidateFailed)
				}
				if !result.Valid {
					cataloglog.Error(err, "Validating failed.")
					return v1alpha2.NewCOAError(err, "invalid properties", v1alpha2.ValidateFailed)
				}
			}
			cataloglog.Info("Validation finished.", "name", r.Name)
		}
	} else {
		cataloglog.Info("Catalog no meta.", "name", r.Name)
	}
	return nil
}
func (r *Catalog) validateUpdateCatalog() error {
	return r.checkSchema()
}
