/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE

*/

package symphonymicrosoftcom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	symphonyv1 "gopls-workspace/apis/symphony.microsoft.com/v1"
	"gopls-workspace/constants"
	"gopls-workspace/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	provisioningstates "gopls-workspace/utils/models"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
}

//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=instances/finalizers,verbs=update

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
	myFinalizerName := "instance.solution.symphony/finalizer"

	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Instance")

	// Get instance
	instance := &symphonyv1.Instance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		log.Error(err, "unable to fetch Instance object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.ensureOperationState(instance, provisioningstates.Reconciling)
	err := r.Status().Update(ctx, instance)
	if err != nil {
		log.Error(err, "unable to update Instance status")
		return ctrl.Result{}, err
	}

	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() { // update
		if !controllerutil.ContainsFinalizer(instance, myFinalizerName) {
			controllerutil.AddFinalizer(instance, myFinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		summary, err := api_utils.GetSummary("http://symphony-service:8080/v1alpha2/", "admin", "", instance.ObjectMeta.Name)
		if err != nil && !v1alpha2.IsNotFound(err) {
			uErr := r.updateInstanceStatusFromError(instance, err)
			if uErr != nil {
				return ctrl.Result{}, uErr
			}
			return ctrl.Result{}, err
		}

		generationMatch := true
		if v, err := strconv.ParseInt(summary.Generation, 10, 64); err == nil {
			generationMatch = v == instance.GetGeneration()
		}

		if generationMatch && time.Since(summary.Time) <= time.Duration(60)*time.Second { //TODO: this is 60 second interval. Make if configurable?
			err = r.updateInstanceStatus(instance, summary.Summary)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		} else {
			err = api_utils.QueueJob("http://symphony-service:8080/v1alpha2/", "admin", "", instance.ObjectMeta.Name, false, false)
			if err != nil {
				uErr := r.updateInstanceStatusFromError(instance, err)
				if uErr != nil {
					return ctrl.Result{}, uErr
				}
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		}

	} else { // remove
		if controllerutil.ContainsFinalizer(instance, myFinalizerName) {
			err = api_utils.QueueJob("http://symphony-service:8080/v1alpha2/", "admin", "", instance.ObjectMeta.Name, true, false)

			if err != nil {
				uErr := r.updateInstanceStatusFromError(instance, err)
				if uErr != nil {
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
					summary, err := api_utils.GetSummary("http://symphony-service:8080/v1alpha2/", "admin", "", instance.ObjectMeta.Name)
					if err == nil && summary.Summary.IsRemoval == true && summary.Summary.SuccessCount == summary.Summary.TargetCount {
						break loop
					}
				}
			}
			// NOTE: we assume the message backend provides at-least-once delivery so that the removal event will be eventually handled.
			// Until the corresponding provider can successfully carry out the removal job, the job event will remain available for the
			// provider to pick up.
			controllerutil.RemoveFinalizer(instance, myFinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}
func (r *InstanceReconciler) ensureOperationState(instance *symphonyv1.Instance, provisioningState string) {
	instance.Status.ProvisioningStatus.Status = provisioningState
	instance.Status.ProvisioningStatus.OperationID = instance.ObjectMeta.Annotations[constants.AzureOperationKey]
}
func (r *InstanceReconciler) updateInstanceStatusFromError(instance *symphonyv1.Instance, err error) error {
	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}
	instance.Status.Properties["status"] = provisioningstates.Failed
	instance.Status.Properties["deployed"] = "0"
	instance.Status.Properties["status-details"] = err.Error()
	r.updateProvisioningStatus(instance, provisioningstates.Failed, err)
	instance.Status.LastModified = metav1.Now()
	return r.Status().Update(context.Background(), instance)
}
func (r *InstanceReconciler) updateInstanceStatus(instance *symphonyv1.Instance, summary model.SummarySpec) error {
	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}
	targetCount := strconv.Itoa(summary.TargetCount)
	successCount := strconv.Itoa(summary.SuccessCount)
	status := provisioningstates.Succeeded
	if successCount != targetCount {
		status = provisioningstates.Failed
	}
	instance.Status.Properties["status"] = status
	instance.Status.Properties["targets"] = targetCount
	instance.Status.Properties["deployed"] = successCount

	for k, v := range summary.TargetResults {
		instance.Status.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
	}

	if status == provisioningstates.Failed {
		r.updateProvisioningStatus(instance, provisioningstates.Failed, fmt.Errorf("deployment failed: %s", summary.SummaryMessage))
	} else {
		r.updateProvisioningStatus(instance, provisioningstates.Succeeded, nil)
	}
	instance.Status.LastModified = metav1.Now()
	return r.Status().Update(context.Background(), instance)
}

func (r *InstanceReconciler) updateProvisioningStatus(instance *symphonyv1.Instance, provisioningStatus string, provisioningError error) {
	r.ensureOperationState(instance, provisioningStatus)
	// Start with a clean Error object and update all the fields
	instance.Status.ProvisioningStatus.Error = symphonyv1.ErrorType{}

	// Fill error details into instance
	errorObj := &instance.Status.ProvisioningStatus.Error
	if provisioningError != nil {
		parsedError, err := utils.ParseAsAPIError(provisioningError)
		if err != nil {
			errorObj.Code = "500"
			errorObj.Message = fmt.Sprintf("Deployment failed. %s", provisioningError.Error())
			return
		}
		errorObj.Code = parsedError.Code
		errorObj.Message = "Deployment failed."
		errorObj.Target = "Symphony"
		errorObj.Details = make([]symphonyv1.TargetError, 0)
		for k, v := range parsedError.Spec.TargetResults {
			errorObj.Details = append(errorObj.Details, symphonyv1.TargetError{
				Code:    v.Status,
				Message: v.Message,
				Target:  k,
			})
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	generationChange := predicate.GenerationChangedPredicate{}
	annotationChange := predicate.AnnotationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&symphonyv1.Instance{}).
		WithEventFilter(predicate.Or(generationChange, annotationChange)).
		Watches(&source.Kind{Type: &symphonyv1.Solution{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []ctrl.Request {
				ret := make([]ctrl.Request, 0)
				solObj := obj.(*symphonyv1.Solution)
				var instances symphonyv1.InstanceList
				options := []client.ListOption{
					client.InNamespace(solObj.Namespace),
					client.MatchingFields{"spec.solution": solObj.Name},
				}
				error := mgr.GetClient().List(context.Background(), &instances, options...)
				if error != nil {
					log.Log.Error(error, "Failed to list instances")
					return ret
				}

				for _, instance := range instances.Items {
					ret = append(ret, ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name:      instance.Name,
							Namespace: instance.Namespace,
						},
					})
				}
				return ret
			})).
		Complete(r)
}
