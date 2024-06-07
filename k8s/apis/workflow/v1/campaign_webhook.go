/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"gopls-workspace/apis/metrics/v1"
	"gopls-workspace/configutils"
	"time"

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
var myCampaignClient client.Client
var catalogWebhookValidationMetrics *metrics.Metrics

func (r *Campaign) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCampaignClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Campaign{}, ".metadata.name", func(rawObj client.Object) []string {
		campaign := rawObj.(*Campaign)
		return []string{campaign.Name}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Campaign{}, ".spec.rootResource", func(rawObj client.Object) []string {
		campaign := rawObj.(*Campaign)
		return []string{campaign.Spec.RootResource}
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

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaign,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update,versions=v1,name=mcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Campaign{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Campaign) Default() {
	campaignlog.Info("default", "name", r.Name)

	if r.Spec.RootResource != "" {
		var campaignContainer CampaignContainer
		err := myCampaignClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &campaignContainer)
		if err != nil {
			campaignlog.Error(err, "failed to get campaign container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: campaignContainer.APIVersion,
				Kind:       campaignContainer.Kind,
				Name:       campaignContainer.Name,
				UID:        campaignContainer.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaign,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update,versions=v1,name=vcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Campaign{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Campaign) ValidateCreate() (admission.Warnings, error) {
	campaignlog.Info("validate create", "name", r.Name)

	validateCreateTime := time.Now()
	validationError := r.validateCreateCampaign()
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

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Campaign) ValidateDelete() (admission.Warnings, error) {
	campaignlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func (r *Campaign) validateCreateCampaign() error {
	var allErrs field.ErrorList

	if err := r.validateNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateRootResource(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "workflow.symphony", Kind: "Campaign"}, r.Name, allErrs)
}

func (r *Campaign) validateNameOnCreate() *field.Error {
	return configutils.ValidateObjectName(r.ObjectMeta.Name, r.Spec.RootResource)
}

func (r *Campaign) validateRootResource() *field.Error {
	var campaignContainer CampaignContainer
	err := myCampaignClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &campaignContainer)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("rootResource"), r.Spec.RootResource, "rootResource must be a valid campaign container")
	}

	if len(r.ObjectMeta.OwnerReferences) == 0 {
		return field.Invalid(field.NewPath("metadata").Child("ownerReference"), len(r.ObjectMeta.OwnerReferences), "ownerReference must be set")
	}

	return nil
}
