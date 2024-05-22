/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package fabric

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	symphonyv1 "gopls-workspace/apis/fabric/v1"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/reconcilers"
	"gopls-workspace/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

// TargetReconciler reconciles a Target object
type TargetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// ApiClient is the client for Symphony API
	ApiClient utils.ApiClient

	// ReconciliationInterval defines the reconciliation interval
	ReconciliationInterval time.Duration

	// DeleteTimeOut defines the timeout for delete operations
	DeleteTimeOut time.Duration

	// PollInterval defines the poll interval
	PollInterval time.Duration

	// Controller Metrics
	m *metrics.Metrics

	dr reconcilers.Reconciler

	// DeleteSyncDelay defines the delay of waiting for status sync back in delete operations
	DeleteSyncDelay time.Duration
}

const (
	targetFinalizerName         = "target.fabric." + constants.FinalizerPostfix
	targetTagFinalizerName      = "targettag.fabric." + constants.FinalizerPostfix
	targetOperationStartTimeKey = "target.fabric." + constants.OperationStartTimeKeyPostfix
)

//+kubebuilder:rbac:groups=fabric.symphony,resources=targets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fabric.symphony,resources=targets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fabric.symphony,resources=targets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Target object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *TargetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Target " + req.Name + " in namespace " + req.Namespace)

	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get target
	target := &symphonyv1.Target{}

	if err := r.Get(ctx, req.NamespacedName, target); err != nil {
		log.Error(err, "unable to fetch Target object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	version := target.Spec.Version
	name := target.Spec.RootResource
	targetName := name + ":" + version
	jData, _ := json.Marshal(target)

	if target.ObjectMeta.DeletionTimestamp.IsZero() { // update
		_, exists := target.Labels["version"]
		log.Info(fmt.Sprintf("Target update: version tag exists - %v", exists))
		if !exists && version != "" && name != "" {
			err := r.ApiClient.CreateTarget(ctx, targetName, jData, req.Namespace, "", "")
			if err != nil {
				log.Error(err, "upsert target failed")
				return ctrl.Result{}, err
			}

			if err := r.Get(ctx, req.NamespacedName, target); err != nil {
				log.Error(err, "unable to fetch Target object after target update")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}

		reconciliationType = metrics.UpdateOperationType
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, target, log, targetOperationStartTimeKey)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		value, exists := target.Labels["tag"]
		log.Info(fmt.Sprintf("Target remove update: latest tag - %v, %v", value, exists))

		if exists && value == "latest" {
			err := r.ApiClient.DeleteTarget(ctx, targetName, req.Namespace, "", "")
			if err != nil {
				log.Error(err, "failed to delete target latest tag")
				return ctrl.Result{}, err
			}

			if err := r.Get(ctx, req.NamespacedName, target); err != nil {
				log.Error(err, "unable to fetch Target object after target tag removal")
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}

		deploymentOperationType, reconcileResult, err = r.dr.AttemptRemove(ctx, target, log, targetOperationStartTimeKey)
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
func (r *TargetReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

func (r *TargetReconciler) populateProvisioningError(summaryResult *model.SummaryResult, err error, errorObj *apimodel.ErrorType) {
	errorObj.Code = "Symphony Orchestrator: [500]"
	if summaryResult != nil {
		summary := summaryResult.Summary

		// Additional message besides target level status(mostly, error message
		// but with lower priority than target level error message)
		if summary.IsRemoval {
			errorObj.Message = fmt.Sprintf("Uninstall failed. %s", summary.SummaryMessage)
		} else {
			errorObj.Message = fmt.Sprintf("Deployment failed. %s", summary.SummaryMessage)
		}

		// Fill error details into target
		// We assume there is one and only one target in summary spec. As opposed
		// to instance CR error object, target CR error object is one layer less. [TODO: We probably shouldn't do this]
		for k, v := range summary.TargetResults {
			// fill errorObj with target level status
			errorObj.Code = v.Status
			errorObj.Message = v.Message
			errorObj.Target = k
			errorObj.Details = make([]apimodel.TargetError, 0)
			// fill errorObj.Details with component level status
			for ck, cv := range v.ComponentResults {
				errorObj.Details = append(errorObj.Details, apimodel.TargetError{
					Code:    cv.Status.String(),
					Message: cv.Message,
					Target:  ck,
				})
			}
		}
	}
	if err != nil {
		errorObj.Message = fmt.Sprintf("%s, %s", err.Error(), errorObj.Message)
	}
}

func (r *TargetReconciler) deploymentBuilder(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
	if target, ok := object.(*symphonyv1.Target); ok {
		deployment, err := utils.CreateSymphonyDeploymentFromTarget(*target, object.GetNamespace())
		if err != nil {
			return nil, err
		}
		return &deployment, nil
	}
	return nil, fmt.Errorf("not able to convert object to target")
}

func (r *TargetReconciler) buildDeploymentReconciler() (reconcilers.Reconciler, error) {
	return reconcilers.NewDeploymentReconciler(
		reconcilers.WithApiClient(r.ApiClient),
		reconcilers.WithDeleteTimeOut(r.DeleteTimeOut),
		reconcilers.WithPollInterval(r.PollInterval),
		reconcilers.WithClient(r.Client),
		reconcilers.WithReconciliationInterval(r.ReconciliationInterval),
		reconcilers.WithFinalizerName(targetFinalizerName),
		reconcilers.WithDeploymentErrorBuilder(r.populateProvisioningError),
		reconcilers.WithDeploymentBuilder(r.deploymentBuilder),
		reconcilers.WithDeleteSyncDelay(r.DeleteSyncDelay),
		reconcilers.WithDeploymentKeyResolver(func(target reconcilers.Reconcilable) string {
			return fmt.Sprintf("target-runtime-%s", target.GetName())
		}),
	)
}
