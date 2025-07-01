/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"fmt"
	"os"
	"time"

	fabric_v1 "gopls-workspace/apis/fabric/v1"
	solution_v1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/utils/diagnostic"
	utilsmodel "gopls-workspace/utils/model"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	diagnostic.InfoWithCtx(log, ctx, "Reconcile Queueing Instance "+req.Name+" in namespace "+req.Namespace)

	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get instance
	instance := &solution_v1.Instance{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			diagnostic.InfoWithCtx(log, ctx, "Skipping this reconcile, since this CR has been deleted")
			return ctrl.Result{}, nil
		} else {
			diagnostic.ErrorWithCtx(log, ctx, err, "unable to fetch Instance object")
			return ctrl.Result{}, err
		}
	}

	// reform context with annotations
	resourceK8SId := instance.GetNamespace() + "/" + instance.GetName()
	operationName := constants.InstanceOperationNamePrefix
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Write)
	} else {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Delete)
	}
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(instance.GetNamespace(), resourceK8SId, instance.Annotations, operationName, r, ctx, log)

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	if checkSkipReconcile(log, ctx, instance) {
		diagnostic.InfoWithCtx(log, ctx, "Skipping this reconcile, since this instance "+req.Name+" in namespace "+req.Namespace+" is inactive and already removed")
		return ctrl.Result{}, nil
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() { // update
		reconciliationType = metrics.UpdateOperationType
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, instance, false, log, instanceOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		reconciliationType = metrics.DeleteOperationType
		diagnostic.InfoWithCtx(log, ctx, "Reconcile removing instance: "+req.Name+" in namespace "+req.Namespace)
		// check the finalizer - uninstall finalizer if exists, set finalizer to nil
		if utils.ContainsString(instance.GetFinalizers(), os.Getenv(constants.DeploymentFinalizer)) {
			// set finalizer to nil
			diagnostic.InfoWithCtx(log, ctx, "Reconcile removing instance finalizer: "+req.Name+" in namespace "+req.Namespace)
			patch := client.MergeFrom(instance.DeepCopy())
			instance.SetFinalizers([]string{})
			if err := r.Patch(ctx, instance, patch); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "Failed to patch instance finalizers")
				resultType = metrics.ReconcileFailedResult
			} else {
				resultType = metrics.ReconcileSuccessResult
			}
		} else {
			deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, instance, true, log, instanceOperationStartTimeKey, operationName)
			if err != nil {
				resultType = metrics.ReconcileFailedResult
			}
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
	// We need to re-able recoverPanic once the behavior is tested
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("InstanceQueueing").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&solution_v1.Instance{}).
		WithEventFilter(predicate.Or(generationChange, operationIdPredicate)).
		Watches(new(solution_v1.Solution), handler.EnqueueRequestsFromMapFunc(
			r.handleSolution)).
		Watches(new(fabric_v1.Target), handler.EnqueueRequestsFromMapFunc(
			r.handleTarget)).
		Complete(r)
}

// We can only skip reconcile if
// 1. the deployment of instance is already removed when instance is inactive
// 2. the new instance spec is still inactive
// If the instance is deleted, we can directly remove the CR.
// What if the instance changes from inactive -> active (not summary reported) -> inactive
// "removed" property will be removed before making queuedeployment calls to symphony API server
// so that later inactive instance can be reconciled again.
func checkSkipReconcile(log logr.Logger, ctx context.Context, instance *solution_v1.Instance) bool {
	if instance.Spec.ActiveState != model.ActiveState_Inactive {
		return false
	}
	if instance.Status.Properties != nil {
		status, ok := instance.Status.Properties["status"]
		if !ok || status != string(utilsmodel.ProvisioningStatusSucceeded) {
			diagnostic.InfoWithCtx(log, ctx, "Instance "+instance.Name+" in namespace "+instance.Namespace+" has not reach succeeded status, do not skip reconcile")
			return false
		}
		removed, ok := instance.Status.Properties["removed"]
		if !ok || removed != "true" {
			diagnostic.InfoWithCtx(log, ctx, "Instance "+instance.Name+" in namespace "+instance.Namespace+" has not been removed, do not skip reconcile")
			return false
		}
		diagnostic.InfoWithCtx(log, ctx, "Instance "+instance.Name+" in namespace "+instance.Namespace+" is inactive and already removed, skip reconcile")
		return true
	}
	diagnostic.InfoWithCtx(log, ctx, "Instance "+instance.Name+" in namespace "+instance.Namespace+" status is nil, do not skip reconcile")
	return false
}
