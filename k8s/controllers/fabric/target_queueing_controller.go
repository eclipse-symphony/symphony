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
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/utils/diagnostic"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TargetReconciler reconciles a Target object
type TargetQueueingReconciler struct {
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
func (r *TargetQueueingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	diagnostic.InfoWithCtx(log, ctx, "Reconcile Queueing Target", "Name", req.Name, "Namespace", req.Namespace)
	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get target
	target := &symphonyv1.Target{}

	if err := r.Get(ctx, req.NamespacedName, target); err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "unable to fetch Target object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	if target.ObjectMeta.DeletionTimestamp.IsZero() { // update
		reconciliationType = metrics.UpdateOperationType
		operationName := fmt.Sprintf("%s/%s", constants.TargetOperationNamePrefix, constants.ActivityOperation_Write)
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, target, false, log, targetOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		reconciliationType = metrics.DeleteOperationType
		operationName := fmt.Sprintf("%s/%s", constants.TargetOperationNamePrefix, constants.ActivityOperation_Delete)
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, target, true, log, targetOperationStartTimeKey, operationName)
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
func (r *TargetQueueingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	metrics, err := metrics.New()
	if err != nil {
		return err
	}

	r.m = metrics
	genChangePredicate := predicate.GenerationChangedPredicate{}
	operationIdPredicate := predicates.OperationIdPredicate{}

	r.dr, err = r.buildDeploymentReconciler()
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.Or(genChangePredicate, operationIdPredicate)).
		For(&symphonyv1.Target{}).
		Complete(r)
}
