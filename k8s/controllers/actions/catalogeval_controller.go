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
	"fmt"

	federationv1 "gopls-workspace/apis/federation/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"

	"gopls-workspace/utils/diagnostic"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	diagnostic.InfoWithCtx(log, ctx, "Entering the CatalogEvalExpression reconciler")

	evalCR := &federationv1.CatalogEvalExpression{}

	resourceK8SId := evalCR.GetNamespace() + "/" + evalCR.GetName()
	operationName := constants.CatalogEvalOperationNamePrefix
	if evalCR.DeletionTimestamp.IsZero() {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Write)
	} else {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Delete)
	}
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(evalCR.GetNamespace(), resourceK8SId, evalCR.GetAnnotations(), operationName, r, ctx, log)

	err := r.Get(ctx, req.NamespacedName, evalCR)
	if err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to parse CatalogEvalExpression")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	operationId := evalCR.GetOperationID()

	// Clean up the cr if it is older than 10 hrs
	if CheckNeedtoDelete(evalCR.ObjectMeta) {
		// Delete the CR
		diagnostic.InfoWithCtx(log, ctx, "Deleting the CatalogEvalExpression as it is older than 10 hrs", "CR", req.NamespacedName)
		if err := r.Delete(ctx, evalCR); err != nil {
			// If delete fails, log the error and return (it shall retry)
			diagnostic.ErrorWithCtx(log, ctx, err, "Error deleting the CatalogEvalExpression", "CR", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if evalCR.DeletionTimestamp.IsZero() {
		// Get parent catalog
		catalog := &federationv1.Catalog{}
		if evalCR.Spec.ResourceRef.Name == "" || evalCR.Spec.ResourceRef.Namespace == "" {
			diagnostic.ErrorWithCtx(log, ctx, errors.New("ParentRef is not set"), "ParentRef is not set")
			// Update status with results
			status := &federationv1.ActionStatusBase{}
			status.ActionStatus = federationv1.ActionResult{
				Error: &federationv1.ProvisioningError{
					Message: "ParentRef is not set",
					Code:    "ParentRefNotSet",
				},
				Status:      federationv1.FailedActionState,
				OperationID: operationId,
			}
			evalCR.Status = federationv1.CatalogEvalExpressionStatus{
				ActionStatusBase: *status,
			}
			r.Status().Update(ctx, evalCR)
			err = r.Status().Update(ctx, evalCR)
			if err != nil {
				// if update fail, log the error and return (it shall retry)
				diagnostic.ErrorWithCtx(log, ctx, err, "Update failed", "CR", req.NamespacedName)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		if err := r.Get(ctx, *evalCR.Spec.ResourceRef.GetNamespacedName(), catalog); err != nil {
			diagnostic.ErrorWithCtx(log, ctx, err, "Error getting the parent catalog resource", "CR", req.NamespacedName)
			status := &federationv1.ActionStatusBase{}
			status.ActionStatus = federationv1.ActionResult{
				Error: &federationv1.ProvisioningError{
					Message: "ParentRef is not found",
					Code:    "ParentRefNotFound",
				},
				Status:      federationv1.FailedActionState,
				OperationID: operationId,
			}
			evalCR.Status = federationv1.CatalogEvalExpressionStatus{
				ActionStatusBase: *status,
			}
			err = r.Status().Update(ctx, evalCR)
			if err != nil {
				// if update fail, log the error and return (it shall retry)
				diagnostic.ErrorWithCtx(log, ctx, err, "Update failed", "CR", req.NamespacedName)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// Send catalog spec properties to settings vendor
		properties, err := r.ApiClient.GetParsedCatalogProperties(ctx, evalCR.Spec.ResourceRef.Name, evalCR.Spec.ResourceRef.Namespace, "", "")
		if properties == nil {
			// log error parsing properties and reuturn (it shall retry)
			diagnostic.ErrorWithCtx(log, ctx, err, "Getting parsed properties failed.", "CR", evalCR.Spec.ResourceRef.Name, "Namespace", evalCR.Spec.ResourceRef.Namespace)
			return ctrl.Result{}, err
		}

		propertiesJSON, err := json.Marshal(properties)
		if err != nil {
			// log error marshalling properties and reuturn (it shall not retry)
			diagnostic.ErrorWithCtx(log, ctx, err, "Error marshalling properties to JSON")
			return ctrl.Result{}, nil
		}
		rawProperties := runtime.RawExtension{Raw: propertiesJSON}

		// Update status with results
		status := &federationv1.ActionStatusBase{}
		status.ActionStatus = federationv1.ActionResult{
			OperationID: operationId,
			Output:      rawProperties,
			Status:      federationv1.SucceededActionState,
		}
		evalCR.Status = federationv1.CatalogEvalExpressionStatus{
			ActionStatusBase: *status,
		}
		err = r.Status().Update(ctx, evalCR)
		if err != nil {
			// if update fail, log the error and return (it shall retry)
			diagnostic.ErrorWithCtx(log, ctx, err, "Update failed", "CR", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CatalogEvalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("CatalogEval").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&federationv1.CatalogEvalExpression{}).
		Complete(r)
}
