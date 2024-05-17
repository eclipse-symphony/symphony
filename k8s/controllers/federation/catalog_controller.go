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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	federationv1 "gopls-workspace/apis/federation/v1"
	"gopls-workspace/constants"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
)

// CatalogReconciler reconciles a Site object
type CatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient
}

const (
	catalogFinalizerName = "catalog.federation." + constants.FinalizerPostfix
)

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
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Catalog " + req.Name + " in namespace " + req.Namespace)

	catalog := &federationv1.Catalog{}
	if err := r.Client.Get(ctx, req.NamespacedName, catalog); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	version := catalog.Spec.Version
	name := catalog.Spec.RootResource
	jData, _ := json.Marshal(catalog)
	catalogName := name + ":" + version
	log.Info(fmt.Sprintf("Reconcile Catalog: %v %v", catalogName, version))
	log.Info(fmt.Sprintf("Reconcile Catalog jdata: %v", catalog))
	log.Info(fmt.Sprintf("Catalog .Labels: %v", catalog.Labels["version"]))

	if catalog.ObjectMeta.DeletionTimestamp.IsZero() { // update
		if !controllerutil.ContainsFinalizer(catalog, catalogFinalizerName) {
			log.Info("Add catalog finalizer")
			controllerutil.AddFinalizer(catalog, catalogFinalizerName)
			if err := r.Client.Update(ctx, catalog); err != nil {
				return ctrl.Result{}, err
			}
		}

		_, exists := catalog.Labels["version"]
		log.Info(fmt.Sprintf("Catalog update: exists version tag, %v", exists))
		if !exists && version != "" && name != "" {
			log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>> Call API to upsert Catalog")
			err := r.ApiClient.UpsertCatalog(ctx, catalogName, jData, "", "")
			if err != nil {
				log.Error(err, "Upsert Catalog failed")
				return ctrl.Result{}, err
			}

			log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>> End API to upsert catalog update, fetch again")
			if err := r.Get(ctx, req.NamespacedName, catalog); err != nil {
				log.Error(err, "unable to fetch catalog object after catalog update")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}

		jData, _ := json.Marshal(catalog)
		err := r.ApiClient.CatalogHook(ctx, jData, "", "")
		if err != nil {
			return ctrl.Result{}, err
		}
	} else { // delete
		value, exists := catalog.Labels["tag"]
		log.Info(fmt.Sprintf("Solution update: %v, %v", value, exists))

		if exists && value == "latest" {
			log.Info(">>>>>>>>>>>>>>>>>>> Call API to delete solution")
			err := r.ApiClient.DeleteCatalog(ctx, catalogName, req.Namespace, "", "")
			if err != nil {
				log.Error(err, "Delete solution failed")
				return ctrl.Result{}, err
			}
		}

		log.Info("Remove finalizer")
		controllerutil.RemoveFinalizer(catalog, catalogFinalizerName)
		if err := r.Client.Update(ctx, catalog); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// CatalogReconciler sets up the controller with the Manager.
func (r *CatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		For(&federationv1.Catalog{}).
		Complete(r)
}
