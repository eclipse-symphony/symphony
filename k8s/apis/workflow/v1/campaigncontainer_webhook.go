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
var campaigncontainerlog = logf.Log.WithName("campaigncontainer-resource")
var myCampaignContainerClient client.Client
var campaignContainerWebhookValidationMetrics *metrics.Metrics

func (r *CampaignContainer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myCampaignContainerClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &CampaignContainer{}, ".metadata.name", func(rawObj client.Object) []string {
		campaign := rawObj.(*CampaignContainer)
		return []string{campaign.Name}
	})

	// initialize the controller operation metrics
	if campaignContainerWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		campaignContainerWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaigncontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigncontainers,verbs=create;update,versions=v1,name=mcampaigncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CampaignContainer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CampaignContainer) Default() {
	campaigncontainerlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaigncontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigncontainers,verbs=create;update;delete,versions=v1,name=vcampaigncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CampaignContainer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CampaignContainer) ValidateCreate() (admission.Warnings, error) {
	campaigncontainerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CampaignContainer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	campaigncontainerlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CampaignContainer) ValidateDelete() (admission.Warnings, error) {
	campaigncontainerlog.Info("validate delete", "name", r.Name)

	validateDeleteTime := time.Now()
	validationError := r.validateDeleteCampaignContainer()
	if validationError != nil {
		campaignContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.CatalogResourceType)
	} else {
		campaignContainerWebhookValidationMetrics.ControllerValidationLatency(
			validateDeleteTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.CatalogResourceType)
	}

	return nil, validationError
}

func (r *CampaignContainer) validateDeleteCampaignContainer() error {
	return r.validateCampaigns()
}

func (r *CampaignContainer) validateCampaigns() error {
	var campaign CampaignList
	err := myCampaignContainerClient.List(context.Background(), &campaign, client.InNamespace(r.Namespace), client.MatchingFields{".spec.rootResource": r.Name})
	if err != nil {
		campaigncontainerlog.Error(err, "could not list campaigns", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("could not list campaigns for campaign container %s.", r.Name))
	}

	if len(campaign.Items) != 0 {
		campaigncontainerlog.Error(err, "campaigns are not empty", "name", r.Name)
		return apierrors.NewBadRequest(fmt.Sprintf("campaigns with root resource '%s' are not empty", r.Name))
	}

	return nil
}
