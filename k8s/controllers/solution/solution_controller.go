/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	solutionv1 "gopls-workspace/apis/solution/v1"
)

// SolutionReconciler reconciles a Solution object
type SolutionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=solution.symphony,resources=solutions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=solution.symphony,resources=solutions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=solution.symphony,resources=solutions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Solution object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SolutionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Solution " + req.Name + " in namespace " + req.Namespace)

	// TODO(user): your logic here

	solution := &solutionv1.Solution{}
	if err := r.Client.Get(ctx, req.NamespacedName, solution); err != nil {
		log.Error(err, "unable to fetch Solution object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !solution.ObjectMeta.DeletionTimestamp.IsZero() {
		finalizers := solution.GetFinalizers()
		remaining := []string{}
		if solution.Finalizers == nil || len(solution.Finalizers) == 0 {
			return ctrl.Result{}, nil
		} else {
			for _, finalizer := range finalizers {
				if strings.HasPrefix(finalizer, "instance.") {
					instanceName := strings.TrimPrefix(finalizer, "instance.")
					instance := &solutionv1.Instance{}
					if err := r.Client.Get(ctx, client.ObjectKey{Namespace: solution.Namespace, Name: instanceName}, instance); err == nil {
						if instance.Spec.Solution == solution.Name {
							log.Error(err, "Instance "+instanceName+" still refers to solution "+solution.Name)
							remaining = append(remaining, finalizer)
						}
					} else if !apierrors.IsNotFound(err) {
						log.Error(err, "Instance "+instanceName+" still exists or unable to determine existance")
						remaining = append(remaining, finalizer)
					} else {
						log.Info("Instance " + instanceName + " does not exist, remove finalizer")
					}
				}
			}
		}
		solution.ObjectMeta.SetFinalizers(remaining)
		r.Client.Update(ctx, solution)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SolutionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&solutionv1.Solution{}).
		Complete(r)
}
