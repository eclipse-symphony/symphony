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
	campaignContainerMaxNameLength  = 61
	campaignContainerMinNameLength  = 3
	campaignlog                     = logf.Log.WithName("campaign-resource")
	myCampaignReaderClient          client.Reader
	catalogWebhookValidationMetrics *metrics.Metrics
	campaignValidator               validation.CampaignValidator
)

func (r *Campaign) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCampaignReaderClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Campaign{}, ".metadata.name", func(rawObj client.Object) []string {
		campaign := rawObj.(*Campaign)
		return []string{campaign.Name}
	})

	// initialize the controller operation metrics
	if catalogWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		catalogWebhookValidationMetrics = metrics
	}

	campaignValidator = validation.NewCampaignValidator(
		// look up campaign container
		func(ctx context.Context, name string, namespace string) (interface{}, error) {
			return dynamicclient.Get(ctx, validation.CampaignContainer, name, namespace)
		},
		// Look up running activation
		func(ctx context.Context, campaign string, namespace string, uid string) (bool, error) {
			// check if the campaign has running activations using the UID first
			activationList, err := dynamicclient.ListWithLabels(ctx, validation.Activation, namespace, map[string]string{api_constants.CampaignUid: uid, api_constants.StatusMessage: v1alpha2.Running.String()}, 1)
			if err != nil {
				return false, err
			}
			if len(activationList.Items) > 0 {
				diagnostic.InfoWithCtx(campaignlog, ctx, "campaign look up activation using UID", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "campaign (%s) in namespace (%s) look up activation using UID ", r.Name, r.Namespace)
				return true, nil
			}

			// if couldn't find any, then use the campaign name
			if len(campaign) < api_constants.LabelLengthUpperLimit {
				activationList, err = dynamicclient.ListWithLabels(ctx, validation.Activation, namespace, map[string]string{api_constants.Campaign: campaign, api_constants.StatusMessage: v1alpha2.Running.String()}, 1)
				if err != nil {
					return false, err
				}
				if len(activationList.Items) > 0 {
					diagnostic.InfoWithCtx(campaignlog, ctx, "campaign look up activation using NAME", "name", r.Name, "namespace", r.Namespace)
					observ_utils.EmitUserAuditsLogs(ctx, "campaign (%s) in namespace (%s) look up activation using NAME ", r.Name, r.Namespace)
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

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaign,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update,versions=v1,name=mcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Campaign{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Campaign) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), campaignlog)
	diagnostic.InfoWithCtx(campaignlog, ctx, "default", "name", r.Name, "namespace", r.Namespace)

	if r.Spec.RootResource != "" {
		var campaignContainer CampaignContainer
		err := myCampaignReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &campaignContainer)
		if err != nil {
			diagnostic.ErrorWithCtx(campaignlog, ctx, err, "failed to get campaign container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(), //campaignContainer.APIVersion
				Kind:       "CampaignContainer",   //campaignContainer.Kind
				Name:       campaignContainer.Name,
				UID:        campaignContainer.UID,
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
			var campaignContainer CampaignContainer
			err := myCampaignReaderClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.RootResource), Namespace: r.Namespace}, &campaignContainer)
			if err != nil {
				diagnostic.ErrorWithCtx(campaignlog, ctx, err, "failed to get campaigncontainer", "name", r.Name, "namespace", r.Namespace)
			}
			r.Labels[api_constants.RootResourceUid] = string(campaignContainer.UID)
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaign,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update;delete,versions=v1,name=vcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Campaign{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Campaign) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCampaign(ctx)
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
func (r *Campaign) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldCampaign, ok := old.(*Campaign)
	if !ok {
		err := fmt.Errorf("expected an Campaign object")
		diagnostic.ErrorWithCtx(campaignlog, ctx, err, "failed to convert old object to Campaign")
		return nil, err
	}
	validationError := r.validateUpdateCampaign(ctx, oldCampaign)
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
func (r *Campaign) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being deleted on namespace %s", r.Name, r.Namespace)

	validationError := r.validateDeleteCampaign(ctx)
	return nil, validationError
}

func (r *Campaign) validateCreateCampaign(ctx context.Context) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate create campaign - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate create campaign", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Campaign) validateDeleteCampaign(ctx context.Context) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate delete campaign - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignValidator.ValidateDelete(ctx, state)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)
	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate delete campaign", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Campaign) validateUpdateCampaign(ctx context.Context, oldCampaign *Campaign) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate update campaign - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	old, err := oldCampaign.ConvertCampaignState()
	if err != nil {
		diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate update campaign - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := campaignValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(campaignlog, ctx, err, "validate update campaign", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Campaign) ConvertCampaignState() (model.CampaignState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to campaign state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.CampaignState{}, retErr
	}
	var state model.CampaignState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.CampaignState{}, retErr
	}
	return state, nil
}

// CampaignContainer Webhook

func (r *CampaignContainer) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), campaignlog)
	commoncontainer.DefaultImpl(campaignlog, ctx, r)
}

func (r *CampaignContainer) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate create campaign container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(campaignlog, ctx, r, campaignContainerMinNameLength, campaignContainerMaxNameLength)
}
func (r *CampaignContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate update campaign container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(campaignlog, ctx, r, old)
}

func (r *CampaignContainer) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myCampaignReaderClient, context.TODO(), campaignlog)

	diagnostic.InfoWithCtx(campaignlog, ctx, "validate delete campaign container", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var campaignList CampaignList
		err := myCampaignReaderClient.List(context.Background(), &campaignList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResourceUid: string(r.UID)}, client.Limit(1))
		if err != nil {
			diagnostic.ErrorWithCtx(campaignlog, ctx, err, "failed to list campaigns", "name", r.Name, "namespace", r.Namespace)
			return 0, err
		}

		if len(campaignList.Items) > 0 {
			diagnostic.InfoWithCtx(campaignlog, ctx, "campaigncontainer look up campaign using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "campaigncontainer (%s) in namespace (%s) look up campaign using UID ", r.Name, r.Namespace)
			return len(campaignList.Items), nil
		}

		if len(r.Name) < api_constants.LabelLengthUpperLimit {
			err = myCampaignReaderClient.List(context.Background(), &campaignList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
			if err != nil {
				diagnostic.ErrorWithCtx(campaignlog, ctx, err, "failed to list campaigns", "name", r.Name, "namespace", r.Namespace)
				return 0, err
			}
			if len(campaignList.Items) > 0 {
				diagnostic.InfoWithCtx(campaignlog, ctx, "campaigncontainer look up campaign using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "campaigncontainer (%s) in namespace (%s) look up campaign using NAME ", r.Name, r.Namespace)
				return len(campaignList.Items), nil
			}
		}
		return 0, nil
	}
	return commoncontainer.ValidateDeleteImpl(campaignlog, ctx, r, getSubResourceNums)
}
