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

	provisioningstates "gopls-workspace/utils/models"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
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

//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=targets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=targets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=symphony.microsoft.com,resources=targets/finalizers,verbs=update

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

	r.ensureOperationState(target, provisioningstates.Reconciling)
	err := r.Status().Update(ctx, target)
	if err != nil {
		log.Error(err, "unable to update Instance status")
		return ctrl.Result{}, err
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

		summary, err := api_utils.GetSummary("http://symphony-service:8080/v1alpha2/", "admin", "", fmt.Sprintf("target-runtime-%s", target.ObjectMeta.Name))
		if err != nil && !v1alpha2.IsNotFound(err) {
			uErr := r.updateTargetStatusFromError(target, err)
			if uErr != nil {
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
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		} else {
			err = api_utils.QueueJob("http://symphony-service:8080/v1alpha2/", "admin", "", target.ObjectMeta.Name, false, true)
			if err != nil {
				uErr := r.updateTargetStatusFromError(target, err)
				if uErr != nil {
					return ctrl.Result{}, uErr
				}
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
		}

	} else { // remove
		if controllerutil.ContainsFinalizer(target, myFinalizerName) {
			err = api_utils.QueueJob("http://symphony-service:8080/v1alpha2/", "admin", "", target.ObjectMeta.Name, true, true)

			if err != nil {
				uErr := r.updateTargetStatusFromError(target, err)
				if uErr != nil {
					return ctrl.Result{}, uErr
				}
				return ctrl.Result{}, err
			}

			// NOTE: we assume the message backend provides at-least-once delivery so that the removal event will be eventually handled.
			// Until the corresponding provider can successfully carry out the removal job, the job event will remain available for the
			// provider to pick up.
			controllerutil.RemoveFinalizer(target, myFinalizerName)
			if err := r.Update(ctx, target); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *TargetReconciler) updateTargetStatusFromError(target *symphonyv1.Target, err error) error {
	if target.Status.Properties == nil {
		target.Status.Properties = make(map[string]string)
	}
	target.Status.Properties["status"] = "Failed"
	target.Status.Properties["deployed"] = "0"
	target.Status.Properties["status-details"] = err.Error()
	r.updateProvisioningStatus(target, provisioningstates.Failed, err)
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
	target.Status.Properties["targets"] = targetCount
	target.Status.Properties["deployed"] = successCount
	for k, v := range summary.TargetResults {
		target.Status.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
		for kc, c := range v.ComponentResults {
			target.Status.Properties["targets."+k+"."+kc] = fmt.Sprintf("%s - %s", c.Status, c.Message)
		}
	}
	if status == provisioningstates.Failed {
		r.updateProvisioningStatus(target, provisioningstates.Failed, fmt.Errorf("deployment failed: %s", summary.SummaryMessage))
	} else {
		r.updateProvisioningStatus(target, provisioningstates.Succeeded, nil)
	}
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

func (r *TargetReconciler) updateProvisioningStatus(target *symphonyv1.Target, provisioningStatus string, provisioningError error) {
	r.ensureOperationState(target, provisioningStatus)
	// Start with a clean Error object and update all the fields
	target.Status.ProvisioningStatus.Error = symphonyv1.ErrorType{}

	// Fill error details into target
	errorObj := &target.Status.ProvisioningStatus.Error
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
