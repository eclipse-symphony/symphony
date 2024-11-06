/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"fmt"
	"time"

	fabric_v1 "gopls-workspace/apis/fabric/v1"
	solution_v1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// InstanceQueueingReconciler reconciles a Instance object
type InstanceQueueingReconciler struct {
	InstanceReconciler
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *InstanceQueueingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Queueing Instance " + req.Name + " in namespace " + req.Namespace)

	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get instance
	instance := &solution_v1.Instance{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Skipping this reconcile, since this CR has been deleted")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "unable to fetch Instance object")
			return ctrl.Result{}, err
		}
	}

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	if instance.ObjectMeta.DeletionTimestamp.IsZero() { // update
		reconciliationType = metrics.UpdateOperationType
		operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Write)
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, instance, false, log, instanceOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		reconciliationType = metrics.DeleteOperationType
		operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Delete)
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, instance, true, log, instanceOperationStartTimeKey, operationName)
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
func (r *InstanceQueueingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var err error
	if r.m, err = metrics.New(); err != nil {
		return err
	}

	if r.dr, err = r.buildDeploymentReconciler(); err != nil {
		return err
	}

	generationChange := predicate.GenerationChangedPredicate{}
	operationIdPredicate := predicates.OperationIdPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&solution_v1.Instance{}).
		WithEventFilter(predicate.Or(generationChange, operationIdPredicate)).
		Watches(new(solution_v1.Solution), handler.EnqueueRequestsFromMapFunc(
			r.handleSolution)).
		Watches(new(fabric_v1.Target), handler.EnqueueRequestsFromMapFunc(
			r.handleTarget)).
		Complete(r)
}
