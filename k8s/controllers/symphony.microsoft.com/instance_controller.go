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
	"errors"
	"fmt"
	"strconv"

	symphonyv1 "gopls-workspace/apis/symphony.microsoft.com/v1"
	"gopls-workspace/constants"
	utils "gopls-workspace/utils"
	provisioningstates "gopls-workspace/utils/models"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	myFinalizerName := "instance.symphony.microsoft.com/finalizer"

	log := log.FromContext(ctx)
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

		solution, targets, errStatus, errDetails := r.prepareForUpdate(ctx, req, instance)

		if solution != nil && targets != nil && len(targets) > 0 {
			deployment, err := utils.CreateSymphonyDeployment(*instance, *solution, targets)
			if err == nil {
				summary, err := api_utils.Deploy("http://symphony-service:8080/v1alpha2/", "admin", "", deployment)
				if err != nil {
					return ctrl.Result{}, r.updateInstanceStatus(instance, "Failed", provisioningstates.Failed, summary, err)
				}

				if err := r.Update(ctx, instance); err != nil {
					return ctrl.Result{}, r.updateInstanceStatus(instance, "State Failed", provisioningstates.Failed, summary, err)
				} else {
					err = r.updateInstanceStatus(instance, "OK", provisioningstates.Succeeded, summary, nil)
					if err != nil {
						return ctrl.Result{}, err
					}
				}

			} else {
				if instance.Status.Properties == nil {
					instance.Status.Properties = make(map[string]string)
				}
				instance.Status.Properties["status"] = "Failed to create deployment"
				instance.Status.Properties["status-details"] = err.Error()
				r.ensureOperationState(instance, provisioningstates.Failed)
				instance.Status.ProvisioningStatus.Error.Code = "deploymentFailed"
				instance.Status.ProvisioningStatus.Error.Message = err.Error()
				instance.Status.LastModified = metav1.Now()
				iErr := r.Status().Update(context.Background(), instance)
				if iErr != nil {
					return ctrl.Result{}, iErr
				}
			}

		} else if errStatus != "" && errDetails != "" {
			if instance.Status.Properties == nil {
				instance.Status.Properties = make(map[string]string)
			}
			instance.Status.Properties["status"] = errStatus
			instance.Status.Properties["status-details"] = errDetails
			r.ensureOperationState(instance, provisioningstates.Reconciling)
			instance.Status.LastModified = metav1.Now()
			iErr := r.Status().Update(context.Background(), instance)
			if iErr != nil {
				return ctrl.Result{}, iErr
			}
		}

		return ctrl.Result{}, nil
	} else { // remove
		if controllerutil.ContainsFinalizer(instance, myFinalizerName) {
			//summary := model.SummarySpec{}
			solution, targets, errP, errDetails := r.prepareForUpdate(ctx, req, instance)
			if solution != nil && targets != nil && len(targets) > 0 {
				deployment, err := utils.CreateSymphonyDeployment(*instance, *solution, targets)
				if err == nil {
					_, err = api_utils.Remove("http://symphony-service:8080/v1alpha2/", "admin", "", deployment)
					if err != nil {
						log.Error(err, "failed to delete components")
						// Note: we only log errors and allow objects to be removed. Otherwise the instance
						// object may get stuck, which causes problems during system update/removal. The downside
						// of this is that external resources may get left behind
					}
				} else {
					log.Error(err, "failed to create deployment")
				}
			} else if errP != "" && errDetails != "" {
				log.Error(errors.New(errDetails), errP)
			}

			controllerutil.RemoveFinalizer(instance, myFinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) prepareForUpdate(ctx context.Context, req ctrl.Request, instance *symphonyv1.Instance) (*symphonyv1.Solution, []symphonyv1.Target, string, string) {
	// Get solution
	solution := &symphonyv1.Solution{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Spec.Solution, Namespace: req.Namespace}, solution); err != nil {
		return nil, nil, "Solution Missing", fmt.Sprintf("unable to fetch Solution object: %v", err)
	}

	// Get targets
	targets := &symphonyv1.TargetList{}
	if err := r.List(ctx, targets, client.InNamespace(req.Namespace)); err != nil {
		return nil, nil, "No Targets", fmt.Sprintf("unable to fetch Target objects: %v", err)
	}

	// Get target candidates
	targetCandidates := utils.MatchTargets(*instance, *targets)
	if len(targetCandidates) == 0 {
		return nil, nil, "No Matching Targets", "no Targets are selected"
	}

	return solution, targetCandidates, "", ""
}

func (r *InstanceReconciler) updateInstanceStatus(instance *symphonyv1.Instance, status string, provisioningStatus string, summary model.SummarySpec, provisioningError error) error {
	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}

	r.ensureOperationState(instance, provisioningStatus)
	instance.Status.Properties["status"] = status
	instance.Status.Properties["targets"] = strconv.Itoa(summary.TargetCount)
	instance.Status.Properties["deployed"] = strconv.Itoa(summary.SuccessCount)

	instance.Status.ProvisioningStatus.Error = symphonyv1.ErrorType{}
	if provisioningError != nil {
		instance.Status.ProvisioningStatus.Error.Code = "deploymentFailed"
		instance.Status.ProvisioningStatus.Error.Message = provisioningError.Error()
	}

	for k, v := range summary.TargetResults {
		instance.Status.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
	}
	instance.Status.LastModified = metav1.Now()
	return r.Status().Update(context.Background(), instance)
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

func (r *InstanceReconciler) ensureOperationState(instance *symphonyv1.Instance, provisioningState string) {
	instance.Status.ProvisioningStatus.Status = provisioningState
	instance.Status.ProvisioningStatus.OperationID = instance.ObjectMeta.Annotations[constants.AzureOperationKey]
}
