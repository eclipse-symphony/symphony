/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ai

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiv1 "gopls-workspace/apis/ai/v1"
)

// SkillPackageReconciler reconciles a SkillPackage object
type SkillPackageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ai.symphony,resources=skillpackages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ai.symphony,resources=skillpackages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ai.symphony,resources=skillpackages/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SkillPackage object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SkillPackageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkillPackageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("SkillPackage").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&aiv1.SkillPackage{}).
		Complete(r)
}
