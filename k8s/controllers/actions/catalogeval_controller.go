/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package actions

import (
	"context"
	"encoding/json"
	"errors"

	federationv1 "gopls-workspace/apis/federation/v1"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// CatalogReconciler reconciles a Site object
type CatalogEvalReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient
}

//+kubebuilder:rbac:groups=federation.symphony,resources=catalogevalexpression,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=federation.symphony,resources=catalogevalexpression/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=federation.symphony,resources=catalogevalexpression/finalizers,verbs=update

func (r *CatalogEvalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ret ctrl.Result, reterr error) {
	log := ctrllog.FromContext(ctx)

	evalCR := &federationv1.CatalogEvalExpression{}

	err := r.Get(ctx, req.NamespacedName, evalCR)
	if err != nil {
		log.Error(err, "failed to parse CatalogEvalExpression")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Clean up the cr if it is older than 1 day
	// if clusterutils.CheckNeedtoDelete(evalCR.ObjectMeta) {
	// 	// Delete the CR
	// 	log.Info("Deleting the CatalogEvalExpression as it is older than 10 hrs", "CR", req.NamespacedName)
	// 	if err := r.Delete(ctx, evalCR); err != nil {
	// 		log.Error(err, "Error deleting the CatalogEvalExpression", "CR", req.NamespacedName)
	// 		return ctrl.Result{}, err
	// 	}
	// 	return ctrl.Result{}, nil
	// }

	if evalCR.DeletionTimestamp.IsZero() {
		// Get parent catalog
		catalog := &federationv1.Catalog{}
		if evalCR.Spec.ParentRef == nil || evalCR.Spec.ParentRef.Name == "" || evalCR.Spec.ParentRef.Namespace == "" {
			log.Error(errors.New("ParentRef is not set"), "ParentRef is not set")
			// Update status with results
			status := &federationv1.ActionStatusBase{}
			status.ActionStatus = federationv1.ActionResult{
				Error: &federationv1.ProvisioningError{
					Message: "ParentRef is not set",
					Code:    "ParentRefNotSet",
				},
				Status: federationv1.FailedActionState,
			}
			evalCR.Status = federationv1.CatalogEvalExpressionStatus{
				ActionStatusBase: *status,
			}
			err = r.Update(ctx, evalCR)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		if err := r.Get(ctx, *evalCR.Spec.ParentRef.GetNamespacedName(), catalog); err != nil {
			log.Error(err, "Error deleting the CatalogEvalExpression", "CR", req.NamespacedName)
			status := &federationv1.ActionStatusBase{}
			status.ActionStatus = federationv1.ActionResult{
				Error: &federationv1.ProvisioningError{
					Message: "ParentRef is not found",
					Code:    "ParentRefNotFound",
				},
				Status: federationv1.FailedActionState,
			}
			evalCR.Status = federationv1.CatalogEvalExpressionStatus{
				ActionStatusBase: *status,
			}
			err = r.Update(ctx, evalCR)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Send catalog spec properties to settings vendor
		properties, err := r.ApiClient.CatalogOnConfig(ctx, evalCR.Spec.ParentRef.Name, evalCR.Spec.ParentRef.Namespace, "", "")
		if err != nil {
		}

		propertiesJSON, err := json.Marshal(properties)
		if err != nil {
			log.Error(err, "Error marshalling properties to JSON")
			return ctrl.Result{}, nil
		}
		rawProperties := runtime.RawExtension{Raw: propertiesJSON}

		// Update status with results
		status := &federationv1.ActionStatusBase{}
		status.ActionStatus = federationv1.ActionResult{
			Output: rawProperties,
			Status: federationv1.SucceededActionState,
		}
		evalCR.Status = federationv1.CatalogEvalExpressionStatus{
			ActionStatusBase: *status,
		}
		err = r.Update(ctx, evalCR)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatalogEvalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&federationv1.CatalogEvalExpression{}).
		Complete(r)
}
