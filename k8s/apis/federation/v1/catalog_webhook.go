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
	commoncontainer "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/utils"
	"time"

	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/k8s/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

// log is for logging in this package.
var cataloglog = logf.Log.WithName("catalog-resource")
var myCatalogReaderClient client.Reader
var catalogWebhookValidationMetrics *metrics.Metrics

func (r *Catalog) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogReaderClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Catalog{}, ".metadata.name", func(rawObj client.Object) []string {
		catalog := rawObj.(*Catalog)
		return []string{catalog.Name}
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

	if r.Spec.RootResource != "" {
		var catalogContainer CatalogContainer
		err := myCatalogReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &catalogContainer)
		if err != nil {
			cataloglog.Error(err, "failed to get catalog container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(), //catalogContainer.APIVersion
				Kind:       "CatalogContainer",    //catalogContainer.Kind
				Name:       catalogContainer.Name,
				UID:        catalogContainer.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}

			if r.Labels == nil {
				r.Labels = make(map[string]string)
			}
			r.Labels["rootResource"] = r.Spec.RootResource
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update,versions=v1,name=vcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Catalog{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateCreate() (admission.Warnings, error) {
	cataloglog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being created on namespace %s", r.Name, r.Namespace)

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

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being updated on namespace %s", r.Name, r.Namespace)

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

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, constants.ActivityCategory_Activity, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, nil
}

func (r *Catalog) validateCreateCatalog() error {
	var allErrs field.ErrorList

	if err := r.checkSchema(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateRootResource(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "Catalog"}, r.Name, allErrs)
}

func (r *Catalog) checkSchema() *field.Error {
	if r.Spec.Metadata != nil {
		if schemaName, ok := r.Spec.Metadata["schema"]; ok {
			schemaName = utils.ReplaceLastSeperator(schemaName, ":", constants.ResourceSeperator)
			cataloglog.Info("Find schema name", "name", schemaName)
			var catalogs CatalogList
			err := myCatalogReaderClient.List(context.Background(), &catalogs, client.InNamespace(r.ObjectMeta.Namespace), client.MatchingFields{"metadata.name": schemaName}, client.Limit(1))
			if err != nil || len(catalogs.Items) == 0 {
				cataloglog.Error(err, "Could not find the required schema.", "name", schemaName)
				return field.Invalid(field.NewPath("spec").Child("Metadata"), schemaName, "could not find the required schema")
			}

			jData, _ := json.Marshal(catalogs.Items[0].Spec.Properties)
			var properties map[string]interface{}
			err = json.Unmarshal(jData, &properties)
			if err != nil {
				cataloglog.Error(err, "Invalid schema.", "name", schemaName)
				return field.Invalid(field.NewPath("spec").Child("properties"), schemaName, "invalid catalog properties")
			}
			if spec, ok := properties["spec"]; ok {
				var schemaObj api_utils.Schema
				jData, _ := json.Marshal(spec)
				err := json.Unmarshal(jData, &schemaObj)
				if err != nil {
					cataloglog.Error(err, "Invalid schema.", "name", schemaName)
					return field.Invalid(field.NewPath("spec").Child("properties"), schemaName, "invalid schema")
				}
				jData, _ = json.Marshal(r.Spec.Properties)
				var properties map[string]interface{}
				err = json.Unmarshal(jData, &properties)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return field.Invalid(field.NewPath("spec").Child("Properties"), schemaName, "unable to unmarshall properties of the catalog")
				}
				result, err := schemaObj.CheckProperties(properties, nil)
				if err != nil {
					cataloglog.Error(err, "Validating failed.")
					return field.Invalid(field.NewPath("spec").Child("Properties"), schemaName, "invalid properties of the catalog schema")
				}
				if !result.Valid {
					cataloglog.Error(err, "Validating failed.")
					return field.Invalid(field.NewPath("spec").Child("Properties"), schemaName, "invalid schema result")
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
	var allErrs field.ErrorList

	if err := r.checkSchema(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Solution"}, r.Name, allErrs)
}

func (r *Catalog) validateNameOnCreate() *field.Error {
	return configutils.ValidateObjectName(r.ObjectMeta.Name, r.Spec.RootResource)
}

func (r *Catalog) validateRootResource() *field.Error {
	var catalogContainer CatalogContainer
	err := myCatalogReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &catalogContainer)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("rootResource"), r.Spec.RootResource, "rootResource must be a valid catalog container")
	}

	if len(r.ObjectMeta.OwnerReferences) == 0 {
		return field.Invalid(field.NewPath("metadata").Child("ownerReference"), len(r.ObjectMeta.OwnerReferences), "ownerReference must be set")
	}

	return nil
}

func (r *CatalogContainer) Default() {
	commoncontainer.DefaultImpl(cataloglog, r)
}

func (r *CatalogContainer) ValidateCreate() (admission.Warnings, error) {
	return commoncontainer.ValidateCreateImpl(cataloglog, r)
}
func (r *CatalogContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return commoncontainer.ValidateUpdateImpl(cataloglog, r, old)
}

func (r *CatalogContainer) ValidateDelete() (admission.Warnings, error) {
	cataloglog.Info("validate delete catalog container", "name", r.Name)
	getSubResourceNums := func() (int, error) {
		var catalogList CatalogList
		err := myCatalogReaderClient.List(context.Background(), &catalogList, client.InNamespace(r.Namespace), client.MatchingLabels{"rootResource": r.Name}, client.Limit(1))
		if err != nil {
			return 0, err
		} else {
			return len(catalogList.Items), nil
		}
	}
	return commoncontainer.ValidateDeleteImpl(cataloglog, r, getSubResourceNums)
}
