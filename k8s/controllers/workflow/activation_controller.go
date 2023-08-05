/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workflow

import (
	"context"
	"encoding/json"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workflowv1 "gopls-workspace/apis/workflow/v1"

	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ActivationReconciler reconciles a Campaign object
type ActivationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
		log.Error(err, "unable to fetch Activation")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if activation.ObjectMeta.DeletionTimestamp.IsZero() {
		if !activation.Status.IsActive && activation.Status.ActivationGeneration != strconv.FormatInt(activation.Generation, 10) {
			err := api_utils.PublishActivationEvent("http://symphony-service:8080/v1alpha2/", "admin", "", v1alpha2.ActivationData{
				Campaign:             activation.Spec.Campaign,
				Activation:           activation.Name,
				ActivationGeneration: strconv.FormatInt(activation.Generation, 10),
				Stage:                "",
				Inputs:               convertRawExtensionToMap(&activation.Spec.Inputs),
			})
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
