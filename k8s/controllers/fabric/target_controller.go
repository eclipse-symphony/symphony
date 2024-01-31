/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package fabric

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"

	symphonyv1 "gopls-workspace/apis/fabric/v1"
	"gopls-workspace/constants"
	"gopls-workspace/utils"

	provisioningstates "gopls-workspace/utils/models"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TargetReconciler reconciles a Target object
type TargetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
	myFinalizerName := "target.fabric.symphony/finalizer"

	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Target")

	// Get target
	target := &symphonyv1.Target{}

	if err := r.Get(ctx, req.NamespacedName, target); err != nil {
		log.Error(err, "unable to fetch Target object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if target.Status.Properties == nil {
		target.Status.Properties = make(map[string]string)
	}

	if target.ObjectMeta.DeletionTimestamp.IsZero() { // update
		if !controllerutil.ContainsFinalizer(target, myFinalizerName) {
			controllerutil.AddFinalizer(target, myFinalizerName)
			if err := r.Update(ctx, target); err != nil {
				return ctrl.Result{}, err
			}
		}

		summary, err := api_utils.GetSummary(ctx, "http://symphony-service:8080/v1alpha2/", "admin", "", fmt.Sprintf("target-runtime-%s", target.ObjectMeta.Name), target.ObjectMeta.Namespace)
		if err != nil && !v1alpha2.IsNotFound(err) {
			uErr := r.updateTargetStatusToReconciling(target, err)
			if uErr != nil {
				log.Error(uErr, "failed to update target status to reconciling")
				return ctrl.Result{}, uErr
			}
			return ctrl.Result{}, err
		}

		generationMatch := true
		if v, err := strconv.ParseInt(summary.Generation, 10, 64); err == nil {
			generationMatch = v == target.GetGeneration()
		}

		if generationMatch && time.Since(summary.Time) <= time.Duration(60)*time.Second { //TODO: this is 60 second interval. Make if configurable?
			err = r.updateTargetStatus(target, summary.Summary)
			if err != nil {
				log.Error(err, "failed to update target status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		} else {
			// Queue a job every 60s or when the generation is changed
			err = api_utils.QueueJob(ctx, "http://symphony-service:8080/v1alpha2/", "admin", "", target.ObjectMeta.Name, target.ObjectMeta.Namespace, false, true)
			if err != nil {
				uErr := r.updateTargetStatusToReconciling(target, err)
				if uErr != nil {
					log.Error(uErr, "failed to update target status to reconciling")
					return ctrl.Result{}, uErr
				}
				return ctrl.Result{}, err
			}

			// Update status to Reconciling if there is a change on generation
			// If users uninstall a component manually without modifying manifest
			// files, jobs queued every 60s will catch the descrepdency and
			// re-deploy the uninstalled component. As users' behavior doesn't
			// trigger generation change, this behavior won't change the status
			// to reconciling.
			if !generationMatch {
				err = r.updateTargetStatusToReconciling(target, nil)
				if err != nil {
					log.Error(err, "failed to update target status to reconciling")
					return ctrl.Result{}, err
				}
			}

			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		}

	} else { // remove
		if controllerutil.ContainsFinalizer(target, myFinalizerName) {
			err := api_utils.QueueJob(ctx, "http://symphony-service:8080/v1alpha2/", "admin", "", target.ObjectMeta.Name, target.ObjectMeta.Namespace, true, true)

			if err != nil {
				uErr := r.updateTargetStatusToReconciling(target, err)
				if uErr != nil {
					log.Error(uErr, "failed to update target status to reconciling")
					return ctrl.Result{}, uErr
				}
				return ctrl.Result{}, err
			}
			timeout := time.After(5 * time.Minute)
			ticker := time.Tick(10 * time.Second) //TODO: configurable? adjust based on provider SLA?
		loop:
			for {
				select {
				case <-timeout:
					// Timeout exceeded, assume deletion failed and proceed with finalization
					break loop
				case <-ticker:
					summary, err := api_utils.GetSummary(ctx, "http://symphony-service:8080/v1alpha2/", "admin", "", fmt.Sprintf("target-runtime-%s", target.ObjectMeta.Name), target.ObjectMeta.Namespace)
					if err == nil && summary.Summary.IsRemoval == true && summary.Summary.SuccessCount == summary.Summary.TargetCount {
						break loop
					}
					if err != nil && !v1alpha2.IsNotFound(err) {
						log.Error(err, "failed to get target summary")
						break loop
					}
				}
			}
			// NOTE: we assume the message backend provides at-least-once delivery so that the removal event will be eventually handled.
			// Until the corresponding provider can successfully carry out the removal job, the job event will remain available for the
			// provider to pick up.
			controllerutil.RemoveFinalizer(target, myFinalizerName)
			if err := r.Update(ctx, target); err != nil {
				log.Error(err, "failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// updateTargetStatusToReconciling updates Target object to Reconciling (non-terminal) state
func (r *TargetReconciler) updateTargetStatusToReconciling(target *symphonyv1.Target, err error) error {
	if target.Status.Properties == nil {
		target.Status.Properties = make(map[string]string)
	}
	target.Status.Properties["status"] = provisioningstates.Reconciling
	target.Status.Properties["deployed"] = "pending"
	target.Status.Properties["targets"] = "pending"
	target.Status.Properties["status-details"] = ""
	if err != nil {
		target.Status.Properties["status-details"] = fmt.Sprintf("Reconciling due to %s", err.Error())
	}
	r.updateProvisioningStatusToReconciling(target, err)
	target.Status.LastModified = metav1.Now()
	return r.Status().Update(context.Background(), target)
}
func (r *TargetReconciler) updateTargetStatus(target *symphonyv1.Target, summary model.SummarySpec) error {
	if target.Status.Properties == nil {
		target.Status.Properties = make(map[string]string)
	}
	targetCount := strconv.Itoa(summary.TargetCount)
	successCount := strconv.Itoa(summary.SuccessCount)
	status := provisioningstates.Succeeded
	if successCount != targetCount {
		status = provisioningstates.Failed
	}
	target.Status.Properties["status"] = status
	target.Status.Properties["deployed"] = successCount
	target.Status.Properties["targets"] = targetCount
	target.Status.Properties["status-details"] = summary.SummaryMessage

	// If a component is ever deployed, it will always show in Status.Properties
	// If a component is not deleted, it will first be reset to Untouched and
	// then changed to corresponding status later
	for k, v := range target.Status.Properties {
		if utils.IsComponentKey(k) && v != v1alpha2.Deleted.String() {
			target.Status.Properties[k] = v1alpha2.Untouched.String()
		}
	}

	// Change to corresponding status
	for k, v := range summary.TargetResults {
		target.Status.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
		for kc, c := range v.ComponentResults {
			target.Status.Properties["targets."+k+"."+kc] = fmt.Sprintf("%s - %s", c.Status, c.Message)
		}
	}

	r.updateProvisioningStatus(target, status, summary)
	target.Status.LastModified = metav1.Now()
	return r.Status().Update(context.Background(), target)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TargetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	genChangePredicate := predicate.GenerationChangedPredicate{}
	annotationPredicate := predicate.AnnotationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.Or(genChangePredicate, annotationPredicate)).
		For(&symphonyv1.Target{}).
		Complete(r)
}

func (r *TargetReconciler) ensureOperationState(target *symphonyv1.Target, provisioningState string) {
	target.Status.ProvisioningStatus.Status = provisioningState
	target.Status.ProvisioningStatus.OperationID = target.ObjectMeta.Annotations[constants.AzureOperationKey]
}

func (r *TargetReconciler) updateProvisioningStatus(target *symphonyv1.Target, provisioningStatus string, summary model.SummarySpec) {
	r.ensureOperationState(target, provisioningStatus)
	// Start with a clean Error object and update all the fields
	target.Status.ProvisioningStatus.Error = apimodel.ErrorType{}
	// Output field is updated if status is Succeeded
	target.Status.ProvisioningStatus.Output = make(map[string]string)

	if provisioningStatus == provisioningstates.Failed {
		errorObj := &target.Status.ProvisioningStatus.Error

		// Fill error details into target
		errorObj.Code = "Symphony: [500]"
		errorObj.Message = "Deployment failed."
		errorObj.Target = "Symphony"
		errorObj.Details = make([]apimodel.TargetError, 0)
		for k, v := range summary.TargetResults {
			targetObject := apimodel.TargetError{
				Code:    v.Status,
				Message: v.Message,
				Target:  k,
				Details: make([]apimodel.ComponentError, 0),
			}
			for ck, cv := range v.ComponentResults {
				targetObject.Details = append(targetObject.Details, apimodel.ComponentError{
					Code:    cv.Status.String(),
					Message: cv.Message,
					Target:  ck,
				})
			}
			errorObj.Details = append(errorObj.Details, targetObject)
		}
	} else if provisioningStatus == provisioningstates.Succeeded {
		outputMap := target.Status.ProvisioningStatus.Output
		// Fill component details into output field
		for k, v := range summary.TargetResults {
			for ck, cv := range v.ComponentResults {
				outputMap[fmt.Sprintf("%s.%s", k, ck)] = cv.Status.String()
			}
		}
	}
}

// updateProvisioningStatusToReconciling updates ProvisioningStatus to Reconciling (non-terminal) state
func (r *TargetReconciler) updateProvisioningStatusToReconciling(target *symphonyv1.Target, err error) {
	provisioningStatus := provisioningstates.Reconciling
	if err != nil {
		provisioningStatus = fmt.Sprintf("%s: due to %s", provisioningstates.Reconciling, err.Error())
	}
	r.ensureOperationState(target, provisioningStatus)
	// Start with a clean Error object and update all the fields
	target.Status.ProvisioningStatus.Error = apimodel.ErrorType{}
}
