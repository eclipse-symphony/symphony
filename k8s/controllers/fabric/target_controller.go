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
	"gopls-workspace/reconcilers"
	"gopls-workspace/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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

	// PollingConcurrentReconciles defines the number of concurrent reconciles
	PollingConcurrentReconciles int

	// Controller Metrics
	m *metrics.Metrics

	dr reconcilers.Reconciler

	// DeleteSyncDelay defines the delay of waiting for status sync back in delete operations
	DeleteSyncDelay time.Duration
}

const (
	targetFinalizerName         = "target.fabric." + constants.FinalizerPostfix
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
// func (r *TargetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	log := ctrllog.FromContext(ctx)
// 	log.Info("shouldn't be called here")
// 	return ctrl.Result{}, nil
// }

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
		deployment, err := utils.CreateSymphonyDeploymentFromTarget(ctx, *target, object.GetNamespace())
		deployment.JobID = object.GetAnnotations()[constants.SummaryJobIdKey]
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
			return api_utils.GetTargetRuntimeKey(api_utils.ConstructSummaryId(target.GetName(), target.GetAnnotations()[api_constants.GuidKey]))
		}),
		reconcilers.WithDeploymentNameResolver(func(target reconcilers.Reconcilable) string {
			return api_utils.GetTargetRuntimeKey(target.GetName())
		}),
	)
}
