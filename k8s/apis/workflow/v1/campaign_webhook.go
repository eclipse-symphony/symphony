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
var campaignlog = logf.Log.WithName("campaign-resource")
var myCampaignReaderClient client.Reader
var catalogWebhookValidationMetrics *metrics.Metrics
var campaignValidator validation.CampaignValidator

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
			return dynamicclient.Get(validation.CampaignContainer, name, namespace)
		},
		// Look up running activation
		func(ctx context.Context, campaign string, namespace string) (bool, error) {
			activationList, err := dynamicclient.ListWithLabels(validation.Activation, namespace, map[string]string{"campaign": campaign}, 1)
			if err != nil {
				return false, err
			}
			return len(activationList.Items) > 0, nil
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
	campaignlog.Info("default", "name", r.Name)

	if r.Spec.RootResource != "" {
		var campaignContainer CampaignContainer
		err := myCampaignReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &campaignContainer)
		if err != nil {
			campaignlog.Error(err, "failed to get campaign container", "name", r.Spec.RootResource)
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
			r.Labels[api_constants.RootResource] = r.Spec.RootResource
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaign,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update;delete,versions=v1,name=vcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Campaign{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Campaign) ValidateCreate() (admission.Warnings, error) {
	campaignlog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

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
	campaignlog.Info("validate update", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldCampaign, ok := old.(*Campaign)
	if !ok {
		return nil, fmt.Errorf("expected an Campaign object")
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
	campaignlog.Info("validate delete", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "Campaign %s is being deleted on namespace %s", r.Name, r.Namespace)

	validationError := r.validateDeleteCampaign(ctx)
	return nil, validationError
}

func (r *Campaign) validateCreateCampaign(ctx context.Context) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		return err
	}
	ErrorFields := campaignValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
}

func (r *Campaign) validateDeleteCampaign(ctx context.Context) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		return err
	}
	ErrorFields := campaignValidator.ValidateDelete(ctx, state)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
}

func (r *Campaign) validateUpdateCampaign(ctx context.Context, oldCampaign *Campaign) error {
	state, err := r.ConvertCampaignState()
	if err != nil {
		return err
	}
	old, err := oldCampaign.ConvertCampaignState()
	if err != nil {
		return err
	}
	ErrorFields := campaignValidator.ValidateCreateOrUpdate(ctx, state, old)
	allErrs := validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
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
	commoncontainer.DefaultImpl(campaignlog, r)
}

func (r *CampaignContainer) ValidateCreate() (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being created on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateCreateImpl(campaignlog, r)
}
func (r *CampaignContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being updated on namespace %s", r.Name, r.Namespace)

	return commoncontainer.ValidateUpdateImpl(campaignlog, r, old)
}

func (r *CampaignContainer) ValidateDelete() (admission.Warnings, error) {
	campaignlog.Info("validate delete campaign container", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.CampaignContainerOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), activationlog)

	observ_utils.EmitUserAuditsLogs(ctx, "CampaignContainer %s is being deleted on namespace %s", r.Name, r.Namespace)

	getSubResourceNums := func() (int, error) {
		var campaignList CampaignList
		err := myCampaignReaderClient.List(context.Background(), &campaignList, client.InNamespace(r.Namespace), client.MatchingLabels{api_constants.RootResource: r.Name}, client.Limit(1))
		if err != nil {
			return 0, err
		} else {
			return len(campaignList.Items), nil
		}
	}
	return commoncontainer.ValidateDeleteImpl(campaignlog, r, getSubResourceNums)
}
