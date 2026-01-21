/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"fmt"
	"strings"
	"time"

	fabric_v1 "gopls-workspace/apis/fabric/v1"
	solution_v1 "gopls-workspace/apis/solution/v1"
	"gopls-workspace/constants"
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/reconcilers"
	"gopls-workspace/utils"
	"gopls-workspace/utils/diagnostic"

	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
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

	// Controller metrics
	m *metrics.Metrics

	dr reconcilers.Reconciler

	// DeleteSyncDelay defines the delay of waiting for status sync back in delete operations
	DeleteSyncDelay time.Duration
}

const (
	instanceFinalizerName               = "instance.solution." + constants.FinalizerPostfix
	instanceOperationStartTimeKey       = "instance.solution." + constants.OperationStartTimeKeyPostfix
	instanceDeleteOperationStartTimeKey = "instance.solution." + constants.DeleteOperationStartTimeKeyPostfix
)

//+kubebuilder:rbac:groups=solution.symphony,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=solution.symphony,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=solution.symphony,resources=instances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
// func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	log := ctrllog.FromContext(ctx)
// 	log.Info("shouldn't be called here")
// 	return ctrl.Result{}, nil
// }

func (r *InstanceReconciler) deploymentBuilder(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
	log := ctrllog.FromContext(ctx)
	diagnostic.InfoWithCtx(log, ctx, "Building deployment")
	var deployment model.DeploymentSpec
	instance, ok := object.(*solution_v1.Instance)
	if !ok {
		err := v1alpha2.NewCOAError(nil, "not able to convert object to instance", v1alpha2.ObjectInstanceCoversionFailed)
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to convert object to instance when building deployment")
		return nil, err
	}

	deploymentResources := &utils.DeploymentResources{
		Instance:         *instance,
		Solution:         solution_v1.Solution{},
		TargetList:       fabric_v1.TargetList{},
		TargetCandidates: []fabric_v1.Target{},
	}

	solutionName := api_utils.ConvertReferenceToObjectName(instance.Spec.Solution)
	if err := r.Get(ctx, types.NamespacedName{Name: solutionName, Namespace: instance.Namespace}, &deploymentResources.Solution); err != nil {
		err = v1alpha2.NewCOAError(err, "failed to get solution", v1alpha2.SolutionGetFailed)
		diagnostic.ErrorWithCtx(log, ctx, err, "proceed with no solution found")
	}
	// Get targets
	if err := r.List(ctx, &deploymentResources.TargetList, client.InNamespace(instance.Namespace)); err != nil {
		err = v1alpha2.NewCOAError(err, "failed to list targets", v1alpha2.TargetListGetFailed)
		diagnostic.ErrorWithCtx(log, ctx, err, "proceed with no targets found")
	}

	// Get target candidates
	deploymentResources.TargetCandidates = r.MatchTargetsForInstance(ctx, instance)
	if len(deploymentResources.TargetCandidates) == 0 {
		err := v1alpha2.NewCOAError(nil, "no target candidates found", v1alpha2.TargetCandidatesNotFound)
		diagnostic.ErrorWithCtx(log, ctx, err, "proceed with no target candidates found")
	}

	deployment, err := utils.CreateSymphonyDeployment(ctx, *instance, deploymentResources.Solution, deploymentResources.TargetCandidates, object.GetNamespace())
	deployment.JobID = object.GetAnnotations()[constants.SummaryJobIdKey]
	if err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "failed to create symphony deployment")
		return nil, err
	}
	return &deployment, nil
}

func (r *InstanceReconciler) MatchTargetsForInstance(ctx context.Context, instance *solution_v1.Instance) []fabric_v1.Target {
	if instance.Spec.Target.Name != "" {
		// First try to get target directly with the full name
		target := fabric_v1.Target{}
		targetName := api_utils.ConvertReferenceToObjectName(instance.Spec.Target.Name)
		err := r.Get(ctx, types.NamespacedName{Name: targetName, Namespace: instance.Namespace}, &target)
		if err == nil {
			// Found directly
			return []fabric_v1.Target{target}
		}

		// No target found
		return []fabric_v1.Target{}
	}

	if len(instance.Spec.Target.Selector) > 0 {
		// List all targets in namespace
		var targets fabric_v1.TargetList
		if err := r.List(ctx, &targets, client.InNamespace(instance.Namespace)); err != nil {
			return []fabric_v1.Target{}
		}

		// Filter targets based on selector
		matchedTargets := []fabric_v1.Target{}
		for _, t := range targets.Items {
			fullMatch := true
			for k, v := range instance.Spec.Target.Selector {
				if tv, ok := t.Spec.Properties[k]; !ok || v != tv {
					fullMatch = false
					break
				}
			}
			if fullMatch {
				matchedTargets = append(matchedTargets, t)
			}
		}

		return matchedTargets
	}

	return []fabric_v1.Target{}
}

func (r *InstanceReconciler) buildDeploymentReconciler() (reconcilers.Reconciler, error) {
	return reconcilers.NewDeploymentReconciler(
		reconcilers.WithApiClient(r.ApiClient),
		reconcilers.WithDeleteTimeOut(r.DeleteTimeOut),
		reconcilers.WithPollInterval(r.PollInterval),
		reconcilers.WithClient(r.Client),
		reconcilers.WithReconciliationInterval(r.ReconciliationInterval),
		reconcilers.WithDeleteSyncDelay(r.DeleteSyncDelay),
		reconcilers.WithFinalizerName(instanceFinalizerName),
		reconcilers.WithDeploymentBuilder(r.deploymentBuilder),
	)
}

func (r *InstanceReconciler) handleTarget(ctx context.Context, obj client.Object) []ctrl.Request {
	ret := make([]ctrl.Request, 0)
	tarObj := obj.(*fabric_v1.Target)
	var instances solution_v1.InstanceList

	// Check if target has AzureResourceIdKey annotation for different processing logic
	updatedInstanceNames := make([]string, 0)
	if azureResourceId, exists := tarObj.Annotations[constants.AzureResourceIdKey]; exists && azureResourceId != "" {
		// Use field index to query instances by Azure resource ID
		options := []client.ListOption{
			client.InNamespace(tarObj.Namespace),
			client.MatchingFields{"spec.target.name": azureResourceId},
		}
		err := r.List(ctx, &instances, options...)
		if err != nil {
			diagnostic.ErrorWithCtx(log.Log, ctx, err, "Failed to list instances by Azure resource ID index")
			return ret
		}

		// For Azure targets, directly add all matching instances to result
		for _, instance := range instances.Items {
			if !utils.NeedWatchInstance(instance) {
				continue
			}
			ret = append(ret, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      instance.Name,
					Namespace: instance.Namespace,
				},
			})
			updatedInstanceNames = append(updatedInstanceNames, instance.Name)
		}
	} else {
		// Use  full namespace listing and MatchTargets
		options := []client.ListOption{client.InNamespace(tarObj.Namespace)}
		err := r.List(ctx, &instances, options...)
		if err != nil {
			diagnostic.ErrorWithCtx(log.Log, ctx, err, "Failed to list instances")
			return ret
		}

		targetList := fabric_v1.TargetList{}
		targetList.Items = append(targetList.Items, *tarObj)

		for _, instance := range instances.Items {
			if !utils.NeedWatchInstance(instance) {
				continue
			}

			targetCandidates := utils.MatchTargets(instance, targetList)
			if len(targetCandidates) > 0 {
				ret = append(ret, ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      instance.Name,
						Namespace: instance.Namespace,
					},
				})
				updatedInstanceNames = append(updatedInstanceNames, instance.Name)
			}
		}
	}

	if len(ret) > 0 {
		diagnostic.InfoWithCtx(log.Log, ctx, fmt.Sprintf("Watched target %s under namespace %s is updated, needs to requeue instances related, count: %d, list: %s", tarObj.Name, tarObj.Namespace, len(ret), strings.Join(updatedInstanceNames, ",")))
	}

	return ret
}

func (r *InstanceReconciler) findRelatedInstances(ctx context.Context, solutionRef string, solutionRefNamespace string, updatedInstanceNames []string, requests []ctrl.Request) ([]ctrl.Request, []string) {
	var instances solution_v1.InstanceList
	ret := make([]ctrl.Request, 0)
	if requests != nil {
		ret = requests
	}
	if updatedInstanceNames == nil {
		updatedInstanceNames = make([]string, 0)
	}
	options := []client.ListOption{
		client.InNamespace(solutionRefNamespace),
		client.MatchingFields{"spec.solution": solutionRef},
	}
	error := r.List(ctx, &instances, options...)
	if error != nil {
		diagnostic.ErrorWithCtx(log.Log, ctx, error, "Failed to list instances")
		return ret, updatedInstanceNames
	}
	for _, instance := range instances.Items {
		if !utils.NeedWatchInstance(instance) {
			continue
		}

		if !utils.ContainsString(updatedInstanceNames, instance.Name) {
			ret = append(ret, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      instance.Name,
					Namespace: instance.Namespace,
				},
			})
			updatedInstanceNames = append(updatedInstanceNames, instance.Name)
		}
	}

	if len(ret) > 0 {
		diagnostic.InfoWithCtx(log.Log, ctx, fmt.Sprintf("Watched solution %s under namespace %s is updated, needs to requeue instances related, count: %d, list: %s", solutionRef, solutionRefNamespace, len(ret), strings.Join(updatedInstanceNames, ",")))
	}
	return ret, updatedInstanceNames
}

func (r *InstanceReconciler) handleSolution(ctx context.Context, obj client.Object) []ctrl.Request {
	solObj := obj.(*solution_v1.Solution)
	updatedInstanceNames := make([]string, 0)
	ret := make([]ctrl.Request, 0)

	// oss reference
	solutionName := api_utils.ConvertObjectNameToReference(solObj.Name)
	ret, updatedInstanceNames = r.findRelatedInstances(ctx, solutionName, solObj.Namespace, updatedInstanceNames, ret)

	// azure reference
	azureSolutionName := strings.ToLower(solObj.Annotations[constants.AzureResourceIdKey])
	if azureSolutionName != "" {
		ret, updatedInstanceNames = r.findRelatedInstances(ctx, azureSolutionName, solObj.Namespace, updatedInstanceNames, ret)
	}

	if len(ret) > 0 {
		diagnostic.InfoWithCtx(log.Log, ctx, fmt.Sprintf("In total, needs to requeue instances related, count: %d, list: %s", len(ret), strings.Join(updatedInstanceNames, ",")))
	}

	return ret
}
