/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	workflowv1 "gopls-workspace/apis/workflow/v1"
	"gopls-workspace/constants"
	"gopls-workspace/utils"
)

// CampaignReconciler reconciles a Campaign object
type CampaignReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient
}

const (
	campaignFinalizerName = "campaign.workflow." + constants.FinalizerPostfix
)

//+kubebuilder:rbac:groups=workflow.symphony,resources=campaigns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workflow.symphony,resources=campaigns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workflow.symphony,resources=campaigns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Campaign object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CampaignReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Campaign")

	// Get instance
	campaign := &workflowv1.Campaign{}
	if err := r.Client.Get(ctx, req.NamespacedName, campaign); err != nil {
		log.Error(err, "unable to fetch campaign object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	version := campaign.Spec.Version
	name := campaign.Spec.RootResource
	campaignName := name + ":" + version
	jData, _ := json.Marshal(campaign)
	log.Info(fmt.Sprintf("Reconcile campaign: %v %v", campaignName, version))
	log.Info(fmt.Sprintf("Reconcile campaign jdata: %v", campaign))

	log.Info(fmt.Sprintf("campaign.Labels: %v", campaign.Labels["version"]))

	if campaign.ObjectMeta.DeletionTimestamp.IsZero() { // update
		if !controllerutil.ContainsFinalizer(campaign, campaignFinalizerName) {
			log.Info("Add campaign finalizer")
			controllerutil.AddFinalizer(campaign, campaignFinalizerName)
			if err := r.Client.Update(ctx, campaign); err != nil {
				return ctrl.Result{}, err
			}
		}

		log.Info("campaign update")
		_, exists := campaign.Labels["version"]
		log.Info(fmt.Sprintf("campaign update: exists version tag, %v", exists))
		if !exists && version != "" && name != "" {
			log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>> Call API to upsert campaign")
			err := r.ApiClient.CreateCampaign(ctx, campaignName, jData, req.Namespace, "", "")
			if err != nil {
				log.Error(err, "Upsert campaign failed")
				return ctrl.Result{}, err
			}
		}
	} else { // delete
		value, exists := campaign.Labels["tag"]
		log.Info(fmt.Sprintf("campaign update: %v, %v", value, exists))

		if exists && value == "latest" {
			log.Info(">>>>>>>>>>>>>>>>>>> Call API to delete campaign")
			err := r.ApiClient.DeleteCampaign(ctx, campaignName, req.Namespace, "", "")
			if err != nil {
				log.Error(err, "Delete campaign failed")
				return ctrl.Result{}, err
			}
		}

		log.Info("Remove finalizer")
		controllerutil.RemoveFinalizer(campaign, campaignFinalizerName)
		if err := r.Client.Update(ctx, campaign); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CampaignReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		For(&workflowv1.Campaign{}).
		Complete(r)
}
