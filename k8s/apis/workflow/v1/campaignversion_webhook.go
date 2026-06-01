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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
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
)

// log is for logging in this package.
var (
	campaignversionContainerMaxNameLength  = 61
	campaignversionContainerMinNameLength  = 1
	campaignversionlog                     = logf.Log.WithName("campaignversion-resource")
	myCampaignVersionReaderClient          client.Reader
	catalogversionWebhookValidationMetrics *metrics.Metrics
	campaignversionValidator               validation.CampaignVersionValidator
)

func (r *CampaignVersion) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCampaignVersionReaderClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &CampaignVersion{}, ".metadata.name", func(rawObj client.Object) []string {
		campaignversion := rawObj.(*CampaignVersion)
		return []string{campaignversion.Name}
	})

	// initialize the controller operation metrics
	if catalogversionWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogversionWebhookValidationMetrics = metrics
	}

	campaignversionValidator = validation.NewCampaignVersionValidator(
		// look up campaignversion container
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(ctx, validation.Campaign, name, namespace)
		},
		// Look up running activation
		func(ctx context.Context, campaignversion string, namespace string, uid string) (bool, error) {
			// check if the campaignversion has running activations using the UID first
			activationList, err := dynamicclient.ListWithLabels(ctx, validation.Activation, namespace, map[string]string{api_constants.CampaignVersionUid: uid, api_constants.StatusMessage: v1alpha2.Running.String()}, 1)
			if err != nil {
				return false, err
			}
			if len(activationList.Items) > 0 {
				diagnostic.InfoWithCtx(campaignversionlog, ctx, "campaignversion look up activation using UID", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "campaignversion (%s) in namespace (%s) look up activation using UID ", r.Name, r.Namespace)
				return true, nil
			}

			// if couldn't find any, then use the campaignversion name
			if len(campaignversion) < api_constants.LabelLengthUpperLimit {
				activationList, err = dynamicclient.ListWithLabels(ctx, validation.Activation, namespace, map[string]string{api_constants.CampaignVersion: campaignversion, api_constants.StatusMessage: v1alpha2.Running.String()}, 1)
				if err != nil {
					return false, err
				}
				if len(activationList.Items) > 0 {
					diagnostic.InfoWithCtx(campaignversionlog, ctx, "campaignversion look up activation using NAME", "name", r.Name, "namespace", r.Namespace)
					observ_utils.EmitUserAuditsLogs(ctx, "campaignversion (%s) in namespace (%s) look up activation using NAME ", r.Name, r.Namespace)
					return true, nil
				}
			}

			// if still finds nothing, we think there's no running activations
			return false, nil
		})

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaignversion,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaignversions,verbs=create;update,versions=v1,name=mcampaignversion.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CampaignVersion{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CampaignVersion) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), campaignversionlog)
	diagnostic.InfoWithCtx(campaignversionlog, ctx, "default", "name", r.Name, "namespace", r.Namespace)

	if r.Spec.RootResource != "" {
		var campaignversionContainer Campaign
		err := myCampaignVersionReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &campaignversionContainer)
		if err != nil {
			diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "failed to get campaignversion container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(), //campaignversionContainer.APIVersion
				Kind:       "Campaign",   //campaignversionContainer.Kind
				Name:       campaignversionContainer.Name,
				UID:        campaignversionContainer.UID,
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
			var campaignversionContainer Campaign
			err := myCampaignVersionReaderClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.RootResource), Namespace: r.Namespace}, &campaignversionContainer)
			if err != nil {
				diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "failed to get campaign", "name", r.Name, "namespace", r.Namespace)
			}
			r.Labels[api_constants.RootResourceUid] = string(campaignversionContainer.UID)
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaignversion,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaignversions,verbs=create;update;delete,versions=v1,name=vcampaignversion.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CampaignVersion{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CampaignVersion) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignVersion %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCampaignVersion(ctx)
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
func (r *CampaignVersion) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignVersionOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignVersion %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldCampaignVersion, ok := old.(*CampaignVersion)
	if !ok {
		err := fmt.Errorf("expected an CampaignVersion object")
		diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "failed to convert old object to CampaignVersion")
		return nil, err
	}
	validationError := r.validateUpdateCampaignVersion(ctx, oldCampaignVersion)
	if validationError != nil {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.InstanceResourceType,
		)
	} else {
		activationWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.InstanceResourceType,
		)
	}
	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CampaignVersion) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignVersionOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignVersion %s is being deleted on namespace %s", r.Name, r.Namespace)

	validationError := r.validateDeleteCampaignVersion(ctx)
	return nil, validationError
}

func (r *CampaignVersion) validateCreateCampaignVersion(ctx context.Context) error {
	state, err := r.ConvertCampaignVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate create campaignversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignversionValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "CampaignVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate create campaignversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CampaignVersion) validateDeleteCampaignVersion(ctx context.Context) error {
	state, err := r.ConvertCampaignVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate delete campaignversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignversionValidator.ValidateDelete(ctx, state)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)
	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "CampaignVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate delete campaignversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CampaignVersion) validateUpdateCampaignVersion(ctx context.Context, oldCampaignVersion *CampaignVersion) error {
	state, err := r.ConvertCampaignVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate update campaignversion - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	old, err := oldCampaignVersion.ConvertCampaignVersionState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate update campaignversion - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignversionValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "CampaignVersion"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "validate update campaignversion", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *CampaignVersion) ConvertCampaignVersionState() (model.CampaignVersionState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "CampaignVersion"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to campaignversion state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.CampaignVersionState{}, retErr
	}
	var state model.CampaignVersionState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.CampaignVersionState{}, retErr
	}
	return state, nil
}

// Campaign Webhook

func (r *Campaign) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), campaignversionlog)
	commoncontainer.DefaultImpl(campaignversionlog, ctx, r)
}

func (r *Campaign) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate create campaignversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(campaignversionlog, ctx, r, campaignversionContainerMinNameLength, campaignversionContainerMaxNameLength)
}
func (r *Campaign) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate update campaignversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(campaignversionlog, ctx, r, old)
}

func (r *Campaign) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignVersionReaderClient, context.TODO(), campaignversionlog)

	diagnostic.InfoWithCtx(campaignversionlog, ctx, "validate delete campaignversion container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var campaignversionList CampaignVersionList
		err := myCampaignVersionReaderClient.List(context.Background(), &campaignversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResourceUid: string(r.UID)}, client.Limit(1))
		if err != nil {
			diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "failed to list campaignversions", "name", r.Name, "namespace", r.Namespace)
			return 0, err
		}

		if len(campaignversionList.Items) > 0 {
			diagnostic.InfoWithCtx(campaignversionlog, ctx, "campaign look up campaignversion using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "campaign (%s) in namespace (%s) look up campaignversion using UID ", r.Name, r.Namespace)
			return len(campaignversionList.Items), nil
		}

		if len(r.Name) < api_constants.LabelLengthUpperLimit {
			err = myCampaignVersionReaderClient.List(context.Background(), &campaignversionList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
			if err != nil {
				diagnostic.ErrorWithCtx(campaignversionlog, ctx, err, "failed to list campaignversions", "name", r.Name, "namespace", r.Namespace)
				return 0, err
			}
			if len(campaignversionList.Items) > 0 {
				diagnostic.InfoWithCtx(campaignversionlog, ctx, "campaign look up campaignversion using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "campaign (%s) in namespace (%s) look up campaignversion using NAME ", r.Name, r.Namespace)
				return len(campaignversionList.Items), nil
			}
		}
		return 0, nil
	}
	return commoncontainer.ValidateDeleteImpl(campaignversionlog, ctx, r, getSubResourceNums)
}
