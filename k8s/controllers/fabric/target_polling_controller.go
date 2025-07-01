/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package fabric

import (
	"context"
	"fmt"
	"time"

	symphonyv1 "gopls-workspace/apis/fabric/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/utils/diagnostic"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// TargetReconciler reconciles a Target object
type TargetPollingReconciler struct {
	TargetReconciler
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Target object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TargetPollingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	diagnostic.InfoWithCtx(log, ctx, "Reconcile Polling Target", "Name", req.Name, "Namespace", req.Namespace)

	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get target
	target := &symphonyv1.Target{}

	if err := r.Get(ctx, req.NamespacedName, target); err != nil {
		if apierrors.IsNotFound(err) {
			diagnostic.InfoWithCtx(log, ctx, "Skipping this reconcile, since the CR has been deleted")
			return ctrl.Result{}, nil
		} else {
			diagnostic.ErrorWithCtx(log, ctx, err, "unable to fetch Target object")
			return ctrl.Result{}, err
		}
	}

	// reform context with annotations
	resourceK8SId := target.GetNamespace() + "/" + target.GetName()
	operationName := constants.TargetOperationNamePrefix
	if target.ObjectMeta.DeletionTimestamp.IsZero() {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Write)
	} else {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Delete)
	}
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(target.GetNamespace(), resourceK8SId, target.Annotations, operationName, r, ctx, log)

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	if target.ObjectMeta.DeletionTimestamp.IsZero() { // update
		reconciliationType = metrics.UpdateOperationType
		deploymentOperationType, reconcileResult, err = r.dr.PollingResult(ctx, target, false, log, targetOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		reconciliationType = metrics.DeleteOperationType
		deploymentOperationType, reconcileResult, err = r.dr.PollingResult(ctx, target, true, log, targetOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	}

	r.m.ControllerReconcileLatency(
		reconcileTime,
		reconciliationType,
		resultType,
		metrics.InstanceResourceType,
		deploymentOperationType,
	)

	return reconcileResult, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *TargetPollingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	metrics, err := metrics.New()
	if err != nil {
		return err
	}

	r.m = metrics
	jobIDPredicate := predicates.JobIDPredicate{}

	r.dr, err = r.buildDeploymentReconciler()
	if err != nil {
		return err
	}
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("TargetPolling").
		WithOptions(controller.Options{RecoverPanic: &recoverPanic, MaxConcurrentReconciles: r.PollingConcurrentReconciles}).
		WithEventFilter(jobIDPredicate).
		For(&symphonyv1.Target{}).
		Complete(r)
}
