/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/utils/diagnostic"
	"os"
	"strings"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var historyLog = logf.Log.WithName("instance-history-resource")

var historyReaderClient client.Reader

func (r *InstanceHistory) SetupWebhookWithManager(mgr ctrl.Manager) error {
	historyReaderClient = mgr.GetAPIReader()

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-instancehistory,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instancehistories,verbs=create,versions=v1,name=minstancehistory.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &InstanceHistory{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *InstanceHistory) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), historyLog)
	diagnostic.InfoWithCtx(historyLog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status)

	// Set owner reference for the instance history
	if r.Spec.RootResource != "" {
		var instance Instance
		err := historyReaderClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &instance)
		if err != nil {
			diagnostic.ErrorWithCtx(historyLog, ctx, err, "failed to get instance", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: GroupVersion.String(),
				Kind:       "Instance",
				Name:       instance.Name,
				UID:        instance.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}
			if r.Labels == nil {
				r.Labels = make(map[string]string)
			}
			r.Labels[api_constants.RootResource] = r.Spec.RootResource
		}
	}

	// Set annotation for the instance history
	annotations := r.ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations = utils.GenerateSystemDataAnnotations(annotations, r.Spec.SolutionId)
	annotation_name := os.Getenv("ANNOTATION_KEY")
	if annotation_name != "" {
		parts := strings.Split(r.Name, constants.ResourceSeperator)
		annotations[annotation_name] = parts[len(parts)-1]
	}
	r.ObjectMeta.SetAnnotations(annotations)
}

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instancehistory,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instancehistories,verbs=create;update;delete,versions=v1,name=vinstancehistory.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &InstanceHistory{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceHistory) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceHistoryOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, historyReaderClient, context.TODO(), historyLog)

	diagnostic.InfoWithCtx(historyLog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Instance history %s is being created on namespace %s", r.Name, r.Namespace)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceHistory) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceHistoryOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, historyReaderClient, context.TODO(), historyLog)

	diagnostic.InfoWithCtx(historyLog, ctx, "Instance history is readonly", "name", r.Name, "namespace", r.Namespace)
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *InstanceHistory) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceHistoryOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, historyReaderClient, context.TODO(), historyLog)

	diagnostic.InfoWithCtx(historyLog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Instance history %s is being deleted on namespace %s", r.Name, r.Namespace)

	return nil, nil
}
