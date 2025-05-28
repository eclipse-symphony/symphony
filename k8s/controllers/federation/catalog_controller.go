/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package federation

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	federationv1 "gopls-workspace/apis/federation/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	k8s_utils "gopls-workspace/utils"

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
	ctrlLog := log.FromContext(ctx)

	diagnostic.InfoWithCtx(ctrlLog, ctx, "Reconciling Catalog", "Name", req.Name, "Namespace", req.Namespace)
	catalog := &federationv1.Catalog{}
	resourceK8SId := catalog.GetNamespace() + "/" + catalog.GetName()
	operationName := constants.CatalogOperationNamePrefix
	if catalog.ObjectMeta.DeletionTimestamp.IsZero() {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Write)
	} else {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Delete)
	}
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(catalog.GetNamespace(), resourceK8SId, catalog.GetAnnotations(), operationName, r, ctx, ctrlLog)
	if err := r.Client.Get(ctx, req.NamespacedName, catalog); err != nil {
		diagnostic.ErrorWithCtx(ctrlLog, ctx, err, "unable to fetch Catalog")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if catalog.ObjectMeta.DeletionTimestamp.IsZero() { // update
		catalogState, err := k8s_utils.K8SCatalogToAPICatalogState(*catalog)
		if err != nil {
			diagnostic.ErrorWithCtx(ctrlLog, ctx, err, "unable to convert Catalog to API CatalogState")
			return ctrl.Result{}, err
		}
		jData, _ := json.Marshal(catalogState)
		err = r.ApiClient.CatalogHook(ctx, jData, "", "")
		if err != nil {
			diagnostic.ErrorWithCtx(ctrlLog, ctx, err, "unable to update Catalog when calling catalogHook")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// CatalogReconciler sets up the controller with the Manager.
func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("Catalog").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&federationv1.Catalog{}).
		Complete(r)
}
