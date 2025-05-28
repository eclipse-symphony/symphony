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

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	workflowv1 "gopls-workspace/apis/workflow/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
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
	diagnostic.InfoWithCtx(log, ctx, "Reconciling Activation", "Name", req.Name, "Namespace", req.Namespace)

	//get Activation
	activation := &workflowv1.Activation{}

	resourceK8SId := activation.GetNamespace() + "/" + activation.GetName()
	operationName := constants.ActivationOperationNamePrefix
	if activation.ObjectMeta.DeletionTimestamp.IsZero() {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Write)
	} else {
		operationName = fmt.Sprintf("%s/%s", operationName, constants.ActivityOperation_Delete)
	}
	ctx = configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(activation.GetNamespace(), resourceK8SId, activation.Annotations, operationName, r, ctx, log)

	if err := r.Get(ctx, req.NamespacedName, activation); err != nil {
		diagnostic.ErrorWithCtx(log, ctx, err, "unable to fetch Activation")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if strconv.FormatInt(activation.Generation, 10) == activation.Status.ActivationGeneration {
		return ctrl.Result{}, nil
	}

	if activation.ObjectMeta.DeletionTimestamp.IsZero() {
		diagnostic.InfoWithCtx(log, ctx, fmt.Sprintf("Activation status: %v", activation.Status.Status))
		if activation.Status.UpdateTime == "" && activation.ObjectMeta.Labels[api_constants.StatusMessage] == "" &&
			activation.Status.Status != v1alpha2.Paused && activation.Status.Status != v1alpha2.Done && activation.Status.ActivationGeneration == "" {
			diagnostic.InfoWithCtx(log, ctx, "Publishing activation event", "Name", activation.Name, "Namespace", activation.Namespace)
			err := r.ApiClient.PublishActivationEvent(ctx, v1alpha2.ActivationData{
				Campaign:             activation.Spec.Campaign,
				Activation:           activation.Name,
				ActivationGeneration: strconv.FormatInt(activation.Generation, 10),
				Stage:                activation.Spec.Stage,
				Inputs:               convertRawExtensionToMap(&activation.Spec.Inputs),
				Namespace:            activation.Namespace,
			}, "", "")
			if err != nil {
				diagnostic.ErrorWithCtx(log, ctx, err, "unable to publish activation event")
				return ctrl.Result{}, err
			}
		}
	} else {
		diagnostic.InfoWithCtx(log, ctx, "Deleting activation", "name", activation.ObjectMeta.Name, "namespace", activation.ObjectMeta.Namespace)
		observ_utils.EmitUserAuditsLogs(ctx, "Activation %s is being deleted on namespace %s", activation.ObjectMeta.Name, activation.ObjectMeta.Namespace)
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
	// We need to re-able recoverPanic once the behavior is tested #691
	recoverPanic := false
	return ctrl.NewControllerManagedBy(mgr).
		Named("Activation").
		WithOptions((controller.Options{RecoverPanic: &recoverPanic})).
		For(&workflowv1.Activation{}).
		Complete(r)
}
