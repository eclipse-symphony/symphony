/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package federation

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	federationv1 "gopls-workspace/apis/federation/v1"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
)

// CatalogReconciler reconciles a Site object
type CatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient
}

//+kubebuilder:rbac:groups=federation.symphony,resources=catalogs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=federation.symphony,resources=catalogs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=federation.symphony,resources=catalogs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Device object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *CatalogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	catalog := &federationv1.Catalog{}
	if err := r.Client.Get(ctx, req.NamespacedName, catalog); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if catalog.ObjectMeta.DeletionTimestamp.IsZero() { // update
		jData, _ := json.Marshal(catalog)
		err := r.ApiClient.CatalogHook(ctx, jData, "", "")
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// CatalogReconciler sets up the controller with the Manager.
func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&federationv1.Catalog{}).
		Complete(r)
}
