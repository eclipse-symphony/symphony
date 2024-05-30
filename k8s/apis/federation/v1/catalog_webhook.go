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
	"gopls-workspace/apis/metrics/v1"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var cataloglog = logf.Log.WithName("catalog-resource")
var myCatalogClient client.Client
var catalogWebhookValidationMetrics *metrics.Metrics

func (r *Catalog) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Catalog{}, ".metadata.name", func(rawObj client.Object) []string {
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
func (r *Catalog) ValidateCreate() (admission.Warnings, error) {
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

	return nil, validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
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

	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateDelete() (admission.Warnings, error) {
	cataloglog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func (r *Catalog) validateCreateCatalog() error {
	return r.checkSchema()
}

func (r *Catalog) checkSchema() error {

	if r.Spec.Metadata != nil {
		if schemaName, ok := r.Spec.Metadata["schema"]; ok {
			cataloglog.Info("Find schema name", "name", schemaName)
			var catalogs CatalogList
			err := myCatalogClient.List(context.Background(), &catalogs, client.InNamespace(r.ObjectMeta.Namespace), client.MatchingFields{".metadata.name": schemaName})
			if err != nil || len(catalogs.Items) == 0 {
				cataloglog.Error(err, "Could not find the required schema.", "name", schemaName)
				return apierrors.NewBadRequest(fmt.Sprintf("Could not find the required schema, %s.", schemaName))
			}

			jData, _ := json.Marshal(catalogs.Items[0].Spec.Properties)
			var properties map[string]interface{}
			err = json.Unmarshal(jData, &properties)
			if err != nil {
				cataloglog.Error(err, "Invalid schema.", "name", schemaName)
				return apierrors.NewBadRequest(fmt.Sprintf("Invalid schema, %s.", schemaName))
			}
			if spec, ok := properties["spec"]; ok {
				var schemaObj utils.Schema
				jData, _ := json.Marshal(spec)
				err := json.Unmarshal(jData, &schemaObj)
				if err != nil {
					cataloglog.Error(err, "Invalid schema.", "name", schemaName)
					return apierrors.NewBadRequest(fmt.Sprintf("Invalid schema, %s.", schemaName))
				}
				jData, _ = json.Marshal(r.Spec.Properties)
				var properties map[string]interface{}
				err = json.Unmarshal(jData, &properties)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return apierrors.NewBadRequest("Invalid properties of the catalog.")
				}
				result, err := schemaObj.CheckProperties(properties, nil)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return apierrors.NewBadRequest("Validate failed for the catalog.")
				}
				if !result.Valid {
					cataloglog.Error(err, "Validating failed.")
					return apierrors.NewBadRequest("This is not a valid catalog according to the schema.")
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
