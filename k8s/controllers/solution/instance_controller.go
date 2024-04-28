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
	"gopls-workspace/controllers/metrics"
	"gopls-workspace/predicates"
	"gopls-workspace/reconcilers"

	"gopls-workspace/constants"
	"gopls-workspace/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	// Controller metrics
	m *metrics.Metrics

	dr reconcilers.Reconciler

	// DeleteSyncDelay defines the delay of waiting for status sync back in delete operations
	DeleteSyncDelay time.Duration
}

const (
	instanceFinalizerName         = "instance.solution." + constants.FinalizerPostfix
	instanceOperationStartTimeKey = "instance.solution." + constants.OperationStartTimeKeyPostfix
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
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Instance " + req.Name + " in namespace " + req.Namespace)

	// Initialize reconcileTime for latency metrics
	reconcileTime := time.Now()

	// Get instance
	instance := &solution_v1.Instance{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		log.Error(err, "unable to fetch Instance object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	reconciliationType := metrics.CreateOperationType
	resultType := metrics.ReconcileSuccessResult
	reconcileResult := ctrl.Result{}
	deploymentOperationType := metrics.DeploymentQueued
	var err error

	if instance.ObjectMeta.DeletionTimestamp.IsZero() { // update
		reconciliationType = metrics.UpdateOperationType
		deploymentOperationType, reconcileResult, err = r.dr.AttemptUpdate(ctx, instance, log, instanceOperationStartTimeKey)
		if err != nil {
			resultType = metrics.ReconcileFailedResult
		}
	} else { // remove
		deploymentOperationType, reconcileResult, err = r.dr.AttemptRemove(ctx, instance, log, instanceOperationStartTimeKey)
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

func (r *InstanceReconciler) deploymentBuilder(ctx context.Context, object reconcilers.Reconcilable) (*model.DeploymentSpec, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Building deployment")
	var deployment model.DeploymentSpec
	instance, ok := object.(*solution_v1.Instance)
	if !ok {
		return nil, v1alpha2.NewCOAError(nil, "not able to convert object to instance", v1alpha2.ObjectInstanceCoversionFailed)
	}

	deploymentResources := &utils.DeploymentResources{
		Instance:         *instance,
		Solution:         solution_v1.Solution{},
		TargetList:       fabric_v1.TargetList{},
		TargetCandidates: []fabric_v1.Target{},
	}

	log.Info(fmt.Sprintf("Instance controller>>>>>>>>>>>>>>>>>>>>> v2v2: try to get solution %v", instance.Spec.Solution))

	// Get solution
	solution, err := r.ApiClient.GetSolution(ctx, instance.Spec.Solution, instance.Namespace)
	//api_utils.GetSolution(ctx, "http://symphony-service:8080/v1alpha2/", instance.Spec.Solution, "admin", "", instance.Namespace)
	if err != nil {
		log.Error(v1alpha2.NewCOAError(err, "failed to get solution from symphony", v1alpha2.SolutionGetFailed), "proceed with no solution found")
	}

	log.Info(fmt.Sprintf("Instance controller>>>>>>>>>>>>>>>>>>>>>>>: try to get solution response  %v", solution.ObjectMeta.Name))
	if err := r.Get(ctx, types.NamespacedName{Name: solution.ObjectMeta.Name, Namespace: instance.Namespace}, &deploymentResources.Solution); err != nil {
		log.Error(v1alpha2.NewCOAError(err, "failed to get solution", v1alpha2.SolutionGetFailed), "proceed with no solution found")
	}
	// Get targets
	if err := r.List(ctx, &deploymentResources.TargetList, client.InNamespace(instance.Namespace)); err != nil {
		log.Error(v1alpha2.NewCOAError(err, "failed to list targets", v1alpha2.TargetListGetFailed), "proceed with no targets found")
	}

	// Get target candidates
	deploymentResources.TargetCandidates = utils.MatchTargets(*instance, deploymentResources.TargetList)
	if len(deploymentResources.TargetCandidates) == 0 {
		log.Error(v1alpha2.NewCOAError(nil, "no target candidates found", v1alpha2.TargetCandidatesNotFound), "proceed with no target candidates found")
	}

	deployment, err = utils.CreateSymphonyDeployment(ctx, *instance, deploymentResources.Solution, deploymentResources.TargetCandidates, object.GetNamespace())

	if err != nil {
		return nil, err
	}
	return &deployment, nil
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

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		Watches(&source.Kind{Type: &solution_v1.Solution{}}, handler.EnqueueRequestsFromMapFunc(
			r.handleSolution)).
		Watches(&source.Kind{Type: &fabric_v1.Target{}}, handler.EnqueueRequestsFromMapFunc(
			r.handleTarget)).
		Complete(r)
}

func (r *InstanceReconciler) handleTarget(obj client.Object) []ctrl.Request {
	ret := make([]ctrl.Request, 0)
	tarObj := obj.(*fabric_v1.Target)
	var instances solution_v1.InstanceList

	options := []client.ListOption{client.InNamespace(tarObj.Namespace)}
	err := r.List(context.Background(), &instances, options...)
	if err != nil {
		log.Log.Error(err, "Failed to list instances")
		return ret
	}

	targetList := fabric_v1.TargetList{}
	targetList.Items = append(targetList.Items, *tarObj)

	updatedInstanceNames := make([]string, 0)
	for _, instance := range instances.Items {
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

	if len(ret) > 0 {
		log.Log.Info(fmt.Sprintf("Watched target %s under namespace %s is updated, needs to requeue instances related, count: %d, list: %s", tarObj.Name, tarObj.Namespace, len(ret), strings.Join(updatedInstanceNames, ",")))
	}

	return ret
}

func (r *InstanceReconciler) handleSolution(obj client.Object) []ctrl.Request {
	ret := make([]ctrl.Request, 0)
	solObj := obj.(*solution_v1.Solution)
	var instances solution_v1.InstanceList

	labels := solObj.ObjectMeta.Labels
	resourceName := labels["rootResource"]
	version := labels["version"]

	var solutionName string
	if resourceName == "" || version == "" {
		solutionName = solObj.Name
	} else {
		solutionName = resourceName + ":" + version
	}

	log.Log.Info(fmt.Sprintf("Instance handlesolution >>>>>>>> start %s", solutionName))

	options := []client.ListOption{
		client.InNamespace(solObj.Namespace),
		client.MatchingFields{"spec.solution": solutionName},
	}
	error := r.List(context.Background(), &instances, options...)
	if error != nil {
		log.Log.Error(error, "Failed to list instances")
		return ret
	}
	log.Log.Info(fmt.Sprintf("Instance handlesolution >>>>>>>> instances count %d", len(instances.Items)))
	log.Log.Info(fmt.Sprintf("Instance handlesolution >>>>>>>> label %s", labels["tag"]))

	if labels["tag"] == "latest" {
		var instancesWithLatest solution_v1.InstanceList
		solutionName = resourceName + ":" + "latest"
		options := []client.ListOption{
			client.InNamespace(solObj.Namespace),
			client.MatchingFields{"spec.solution": solutionName},
		}

		error := r.List(context.Background(), &instancesWithLatest, options...)
		if error != nil {
			log.Log.Error(error, "Failed to list instances")
			return ret
		}

		instances.Items = append(instances.Items, instancesWithLatest.Items...)
		log.Log.Info(fmt.Sprintf("Instance handlesolution >>>>>>>>222 instances count with latest %d", len(instances.Items)))
	}

	updatedInstanceNames := make([]string, 0)
	for _, instance := range instances.Items {
		var interval time.Duration = 30
		if instance.Spec.ReconciliationPolicy != nil && instance.Spec.ReconciliationPolicy.Interval != nil {
			parsedInterval, err := time.ParseDuration(*instance.Spec.ReconciliationPolicy.Interval)
			if err != nil {
				log.Log.Error(err, "Instance handlesolution parse interval >>>>>>>>  ")
				parsedInterval = 30
			}
			interval = parsedInterval
		}

		if instance.Spec.ReconciliationPolicy != nil && instance.Spec.ReconciliationPolicy.State.IsInActive() || interval == 0 {
			log.Log.Info(fmt.Sprintf("Instance handlesolution >>>>>>>> inactive no watch %s", instance.ObjectMeta.Name))
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

	if len(ret) > 0 {
		log.Log.Info(fmt.Sprintf("Watched solution %s under namespace %s is updated, needs to requeue instances related, count: %d, list: %s", solObj.Name, solObj.Namespace, len(ret), strings.Join(updatedInstanceNames, ",")))
	}

	return ret
}
