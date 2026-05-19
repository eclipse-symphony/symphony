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
	"gopls-workspace/utils/diagnostic"
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
var (
	catalogMaxNameLength   = 61
	catalogMinNameLength   = 1
	catalogversionlog                      = logf.Log.WithName("catalogversion-resource")
	myCatalogVersionReaderClient           client.Reader
	catalogversionWebhookValidationMetrics *metrics.Metrics
	catalogversionValidator                validation.CatalogVersionValidator
)

func (r *CatalogVersion) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCatalogVersionReaderClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &CatalogVersion{}, ".metadata.name", func(rawObj client.Object) []string {
		catalogversion := rawObj.(*CatalogVersion)
		return []string{catalogversion.Name}
	})

	// initialize the controller operation metrics
	if catalogversionWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogversionWebhookValidationMetrics = metrics
	}

	catalogversionValidator = validation.NewCatalogVersionValidator(
		// Look up catalogversion
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(ctx, validation.CatalogVersion, name, namespace)
		},
		// Look up catalogversion container
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(ctx, validation.Catalog, name, namespace)
		},
		// Look up child catalogversion
		func(ctx context.Context, name string, namespace string, uid string) (bool, error) {
			catalogversionList, err := dynamicclient.ListWithLabels(ctx, validation.CatalogVersion, namespace, map[string]string{api_constants.ParentName: name}, 1)
			if err != nil {
				return false, err
			}
			return len(catalogversionList.Items) > 0, nil
		},
	)
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalogversion,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogversions,verbs=create;update,versions=v1,name=mcatalogversion.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CatalogVersion{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CatalogVersion) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), catalogversionlog)
	diagnostic.InfoWithCtx(catalogversionlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status)

	if r.Spec.RootResource != "" {
		var catalog Catalog
		err := myCatalogVersionReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &catalog)
		if err != nil {
			diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "failed to get catalogversion container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(), //catalog.APIVersion
				Kind:       "Catalog",    //catalog.Kind
				Name:       catalog.Name,
				UID:        catalog.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}

			if r.Labels == nil {
				r.Labels = make(map[string]string)
			}

			// Remove api_constants.RootResource from r.Labels if it exists
			if _, exists := r.Labels[api_constants.RootResource]; exists {
				delete(r.Labels, api_constants.RootResource)
			}
			var catalog Catalog
			err := myCatalogVersionReaderClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.RootResource), Namespace: r.Namespace}, &catalog)
			if err != nil {
				diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "failed to get catalog", "name", r.Name, "namespace", r.Namespace)
			}
			r.Labels[api_constants.RootResourceUid] = string(catalog.UID)

			if r.Spec.ParentName != "" {
				r.Labels[api_constants.ParentName] = utils.ConvertStringToValidLabel(validation.ConvertReferenceToObjectName(r.Spec.ParentName))
			} else if r.Labels[api_constants.ParentName] != "" {
				// If the spec does not have parent name anymore, we should remove the outdated parent label.
				delete(r.Labels, api_constants.ParentName)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalogversion,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogversions,verbs=create;update;delete,versions=v1,name=vcatalogversion.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CatalogVersion{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersion) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CatalogVersion %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCatalogVersion(ctx)
	if validationError != nil {
		catalogversionWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogVersionResourceType)
	} else {
		catalogversionWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogVersionResourceType)
	}

	return nil, validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersion) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CatalogVersion %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldCatalogVersion, ok := old.(*CatalogVersion)
	if !ok {
		err := fmt.Errorf("expected a CatalogVersion object")
		diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "failed to convert to catalogversion object")
		return nil, err
	}
	validationError := r.validateUpdateCatalogVersion(ctx, oldCatalogVersion)
	if validationError != nil {
		catalogversionWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.CatalogVersionResourceType)
	} else {
		catalogversionWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.CatalogVersionResourceType)
	}

	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CatalogVersion) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CatalogVersion %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, r.validateDeleteCatalogVersion(ctx)
}

func (r *CatalogVersion) validateCreateCatalogVersion(ctx context.Context) error {
	state, err := r.ConvertCatalogVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate create catalogversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := catalogversionValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "CatalogVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate create catalogversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CatalogVersion) validateUpdateCatalogVersion(ctx context.Context, oldCatalogVersion *CatalogVersion) error {
	state, err := r.ConvertCatalogVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate update catalogversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	old, err := oldCatalogVersion.ConvertCatalogVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate update catalogversion - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := catalogversionValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "CatalogVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate update catalogversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CatalogVersion) validateDeleteCatalogVersion(ctx context.Context) error {
	state, err := r.ConvertCatalogVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate delete catalogversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}

	ErrorFields := catalogversionValidator.ValidateDelete(ctx, state)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "CatalogVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "validate delete catalogversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CatalogVersion) ConvertCatalogVersionState() (model.CatalogVersionState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "federation.symphony", Kind: "CatalogVersion"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to catalogversion state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.CatalogVersionState{}, retErr
	}
	var state model.CatalogVersionState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.CatalogVersionState{}, retErr
	}
	return state, nil
}

func (r *Catalog) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), catalogversionlog)
	commoncontainer.DefaultImpl(catalogversionlog, ctx, r)
}

func (r *Catalog) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate create catalogversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(catalogversionlog, ctx, r, catalogMinNameLength, catalogMaxNameLength)
}
func (r *Catalog) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate update catalogversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being updated on namespace %s", r.Name, r.Namespace)
	return commoncontainer.ValidateUpdateImpl(catalogversionlog, ctx, r, old)
}

func (r *Catalog) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CatalogVersionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCatalogVersionReaderClient, context.TODO(), catalogversionlog)

	diagnostic.InfoWithCtx(catalogversionlog, ctx, "validate delete catalogversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Catalog %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var catalogversionList CatalogVersionList
		err := myCatalogVersionReaderClient.List(context.Background(), &catalogversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResourceUid: string(r.UID)}, client.Limit(1))
		if err != nil {
			diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "failed to list catalogversions", "name", r.Name, "namespace", r.Namespace)
			return 0, err
		}

		if len(catalogversionList.Items) > 0 {
			diagnostic.InfoWithCtx(catalogversionlog, ctx, "catalog look up catalogversion using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "catalog (%s) in namespace (%s) look up catalogversion using UID ", r.Name, r.Namespace)
			return len(catalogversionList.Items), nil
		}
		if len(r.Name) < api_constants.LabelLengthUpperLimit {
			err = myCatalogVersionReaderClient.List(context.Background(), &catalogversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
			if err != nil {
				diagnostic.ErrorWithCtx(catalogversionlog, ctx, err, "failed to list catalogversions", "name", r.Name, "namespace", r.Namespace)
				return 0, err
			}
			if len(catalogversionList.Items) > 0 {
				diagnostic.InfoWithCtx(catalogversionlog, ctx, "catalog look up catalogversion using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "catalog (%s) in namespace (%s) look up catalogversion using NAME ", r.Name, r.Namespace)
				return len(catalogversionList.Items), nil
			}
		}
		return 0, nil
	}
	return commoncontainer.ValidateDeleteImpl(catalogversionlog, ctx, r, getSubResourceNums)
}
