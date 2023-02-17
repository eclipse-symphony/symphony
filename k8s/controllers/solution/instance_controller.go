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

package solution

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/log"

	fabricv1 "gopls-workspace/apis/fabric/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"
	utils "gopls-workspace/utils"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
	_ = log.FromContext(ctx)
	myFinalizerName := "instance.solution.symphony/finalizer"

	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Instance")

	solutionGone := false
	targetsGone := false

	// Get instance
	instance := &solutionv1.Instance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		log.Error(err, "unable to fetch Instance object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}
	// Get solution
	solution := &solutionv1.Solution{}
	if err := r.Get(ctx, types.NamespacedName{Name: instance.Spec.Solution, Namespace: req.Namespace}, solution); err != nil {
		log.Error(err, "unable to fetch Solution object")
		instance.Status.Properties["status"] = "Solution Missing"
		iErr := r.Status().Update(context.Background(), instance)
		if iErr != nil {
			return ctrl.Result{}, iErr
		}
		if !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		solutionGone = true
	}

	// Get targets
	targets := &fabricv1.TargetList{}
	if err := r.List(ctx, targets, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "unable to fetch Target objects")
		instance.Status.Properties["status"] = "No Targets"
		iErr := r.Status().Update(context.Background(), instance)
		if iErr != nil {
			return ctrl.Result{}, iErr
		}
		if !k8s_errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		targetsGone = true
	}

	// Get target candidates
	var targetCandidates []fabricv1.Target

	if !targetsGone {
		targetCandidates = utils.MatchTargets(*instance, *targets)
		if len(targetCandidates) == 0 {
			log.Info("no Targets are selected")
			instance.Status.Properties["status"] = "No Targets"
			iErr := r.Status().Update(context.Background(), instance)
			if iErr != nil {
				return ctrl.Result{}, iErr
			}
			// return ctrl.Result{}, nil
			// if instance.ObjectMeta.DeletionTimestamp.IsZero() {
			// 	return ctrl.Result{RequeueAfter: 180 * time.Second}, nil //keep checking matching targets
			// }
		}
	}

	// UPDATE extenral resources
	deployment := model.DeploymentSpec{}
	var err error
	if !solutionGone && !targetsGone {
		deployment, err = utils.CreateSymphonyDeployment(*instance, *solution, targetCandidates, nil)
		if err != nil {
			log.Error(err, "failed to generate Symphony deployment")
			instance.Status.Properties["status"] = "Creation failed"
			iErr := r.Status().Update(context.Background(), instance)
			if iErr != nil {
				return ctrl.Result{}, iErr
			}
			return ctrl.Result{}, err
		}
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() { // update
		if !controllerutil.ContainsFinalizer(instance, myFinalizerName) {
			controllerutil.AddFinalizer(instance, myFinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		if !solutionGone && !targetsGone && len(targetCandidates) > 0 {
			summary, err := utils.Deploy(deployment)
			if err != nil {
				return ctrl.Result{}, r.updateInstanceStatus(instance, "Failed", summary)
			}

			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, r.updateInstanceStatus(instance, "State Failed", summary)
			} else {
				err = r.updateInstanceStatus(instance, "OK", summary)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		return ctrl.Result{RequeueAfter: 180 * time.Second}, nil

	} else { // remove
		if controllerutil.ContainsFinalizer(instance, myFinalizerName) {
			summary := model.SummarySpec{}
			if !solutionGone && !targetsGone {
				summary, err = utils.Remove(deployment)
				if err != nil { // TODO: this could stop the CRD being removed if the underlying component is permanantly destroyed
					log.Error(err, "failed to delete components")
					return ctrl.Result{}, r.updateInstanceStatus(instance, "Remove Failed", summary)
				}
			}
			controllerutil.RemoveFinalizer(instance, myFinalizerName)
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, r.updateInstanceStatus(instance, "State Failed", summary)
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) updateInstanceStatus(instance *solutionv1.Instance, status string, summary model.SummarySpec) error {
	if instance.Status.Properties == nil {
		instance.Status.Properties = make(map[string]string)
	}
	instance.Status.Properties["status"] = status
	instance.Status.Properties["targets"] = strconv.Itoa(summary.TargetCount)
	instance.Status.Properties["deployed"] = strconv.Itoa(summary.SuccessCount)
	for k, v := range summary.TargetResults {
		instance.Status.Properties["targets."+k] = fmt.Sprintf("%s - %s", v.Status, v.Message)
	}
	return r.Status().Update(context.Background(), instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&solutionv1.Instance{}).
		Watches(&source.Kind{Type: &solutionv1.Solution{}}, handler.EnqueueRequestsFromMapFunc(
			func(obj client.Object) []ctrl.Request {
				ret := make([]ctrl.Request, 0)
				solObj := obj.(*solutionv1.Solution)
				var instances solutionv1.InstanceList
				mgr.GetClient().List(context.Background(), &instances, client.InNamespace(solObj.Namespace), client.MatchingFields{".spec.solution": solObj.Name})
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
