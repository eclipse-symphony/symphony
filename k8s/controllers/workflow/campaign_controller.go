/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package workflow

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workflowv1 "gopls-workspace/apis/workflow/v1"
)

// CampaignReconciler reconciles a Campaign object
type CampaignReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CampaignReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("Campaign").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&workflowv1.Campaign{}).
		Complete(r)
}
