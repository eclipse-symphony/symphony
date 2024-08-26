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
	"gopls-workspace/apis/dynamicclient"
	"gopls-workspace/apis/metrics/v1"
	commoncontainer "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"time"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
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

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

// log is for logging in this package.
var cataloglog = logf.Log.WithName("catalog-resource")
var myCatalogReaderClient client.Reader
var catalogWebhookValidationMetrics *metrics.Metrics
var catalogValidator validation.CatalogValidator

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

	catalogValidator = validation.NewCatalogValidator(
		// Look up catalog
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(validation.Catalog, name, namespace)
		},
		// Look up catalog container
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(validation.CatalogContainer, name, namespace)
		},
		// Look up child catalog
		func(ctx context.Context, name string, namespace string) (bool, error) {
			catalogList, err := dynamicclient.ListWithLabels(validation.Catalog, namespace, map[string]string{api_constants.ParentName: name}, 1)
			if err != nil {
				return false, err
			}
			return len(catalogList.Items) > 0, nil
		},
	)
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
			r.Labels[api_constants.RootResource] = utils.ConvertStringToValidLabel(r.Spec.RootResource)
			if r.Spec.ParentName != "" {
				r.Labels[api_constants.ParentName] = utils.ConvertStringToValidLabel(validation.ConvertReferenceToObjectName(r.Spec.ParentName))
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update;delete,versions=v1,name=vcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Catalog{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Catalog) ValidateCreate() (admission.Warnings, error) {
	cataloglog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCatalog(ctx)
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
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldCatalog, ok := old.(*Catalog)
	if !ok {
		return nil, fmt.Errorf("expected a Catalog object")
	}
	validationError := r.validateUpdateCatalog(ctx, oldCatalog)
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
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), cataloglog)

	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, r.validateDeleteCatalog(ctx)
}

func (r *Catalog) validateCreateCatalog(ctx context.Context) error {
	state, err := r.ConvertCatalogState()
	if err != nil {
		return err
	}
	ErrorFields := catalogValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "Catalog"}, r.Name, allErrs)
}

func (r *Catalog) validateUpdateCatalog(ctx context.Context, oldCatalog *Catalog) error {
	state, err := r.ConvertCatalogState()
	if err != nil {
		return err
	}
	old, err := oldCatalog.ConvertCatalogState()
	if err != nil {
		return err
	}
	ErrorFields := catalogValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "Catalog"}, r.Name, allErrs)
}

func (r *Catalog) validateDeleteCatalog(ctx context.Context) error {
	state, err := r.ConvertCatalogState()
	if err != nil {
		return err
	}

	ErrorFields := catalogValidator.ValidateDelete(ctx, state)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "Catalog"}, r.Name, allErrs)
}

func (r *Catalog) ConvertCatalogState() (model.CatalogState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "Catalog"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to catalog state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.CatalogState{}, retErr
	}
	var state model.CatalogState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.CatalogState{}, retErr
	}
	return state, nil
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
		err := myCatalogReaderClient.List(context.Background(), &catalogList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
		if err != nil {
			return 0, err
		} else {
			return len(catalogList.Items), nil
		}
	}
	return commoncontainer.ValidateDeleteImpl(cataloglog, r, getSubResourceNums)
}
