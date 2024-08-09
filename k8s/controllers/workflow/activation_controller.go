/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workflowv1 "gopls-workspace/apis/workflow/v1"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ActivationReconciler reconciles a Campaign object
type ActivationReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	ApiClient utils.ApiClient
}

//+kubebuilder:rbac:groups=workflow.symphony,resources=activations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workflow.symphony,resources=activations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workflow.symphony,resources=activations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Campaign object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ActivationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reconcile Activation")

	//get Activation
	activation := &workflowv1.Activation{}
	if err := r.Get(ctx, req.NamespacedName, activation); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if strconv.FormatInt(activation.Generation, 10) == activation.Status.ActivationGeneration {
		return ctrl.Result{}, nil
	}

	if activation.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info(fmt.Sprintf("Activation status: %v", activation.Status.Status))
		if activation.Status.UpdateTime == "" && activation.Status.Status != v1alpha2.Paused && activation.Status.Status != v1alpha2.Done && activation.Status.ActivationGeneration == "" {
			err := r.ApiClient.PublishActivationEvent(ctx, v1alpha2.ActivationData{
				Campaign:             activation.Spec.Campaign,
				Activation:           activation.Name,
				ActivationGeneration: strconv.FormatInt(activation.Generation, 10),
				Stage:                activation.Spec.Stage,
				Inputs:               convertRawExtensionToMap(&activation.Spec.Inputs),
				Namespace:            activation.Namespace,
			}, "", "")
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func convertRawExtensionToMap(raw *runtime.RawExtension) map[string]interface{} {
	if raw == nil {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw.Raw, &data); err != nil {
		return nil
	}
	return data
}

// SetupWithManager sets up the controller with the Manager.
func (r *ActivationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workflowv1.Activation{}).
		Complete(r)
}
