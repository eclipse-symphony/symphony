/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package fabric

import (
	"context"
	"fmt"
	"os"
	"time"

	symphonyv1 "gopls-workspace/apis/fabric/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/utils/diagnostic"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
		if apierrors.IsNotFound(err) {
			diagnostic.InfoWithCtx(log, ctx, "Skipping this reconcile, since this CR has been deleted")
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
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, target, false, log, targetOperationStartTimeKey, operationName)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		diagnostic.InfoWithCtx(log, ctx, "Reconcile removing target:  "+req.Name+" in namespace "+req.Namespace)
		reconciliationType = metrics.DeleteOperationType
		if utils.ContainsString(target.GetFinalizers(), os.Getenv(constants.DeploymentFinalizer)) {
			// set finalizer to nil
			diagnostic.InfoWithCtx(log, ctx, "Reconcile removing target finalizer:  "+req.Name+" in namespace "+req.Namespace)
			patch := client.MergeFrom(target.DeepCopy())
			target.SetFinalizers([]string{})
			if err = r.Patch(ctx, target, patch); err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "Failed to patch target finalizers")
				resultType = metrics.ReconcileFailedResult
			} else {
				resultType = metrics.ReconcileSuccessResult
			}
		} else {
			deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, target, true, log, targetOperationStartTimeKey, operationName)
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
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("TargetQueueing").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		WithEventFilter(predicate.Or(genChangePredicate, operationIdPredicate)).
		For(&symphonyv1.Target{}).
		Complete(r)
}
