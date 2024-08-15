/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"
	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/apis/metrics/v1"
	v1 "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"time"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var instancelog = logf.Log.WithName("instance-resource")
var myInstanceClient client.Reader
var instanceWebhookValidationMetrics *metrics.Metrics
var instanceProjectConfig *configv1.ProjectConfig

func (r *Instance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myInstanceClient = mgr.GetAPIReader()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.solution", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.Solution}
	})
	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	instanceProjectConfig = myConfig
	// initialize the controller operation metrics
	if instanceWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		instanceWebhookValidationMetrics = metrics
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=minstance.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Instance{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Instance) Default() {
	instancelog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		r.Labels["displayName"] = r.Spec.DisplayName
	}
	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.State == "" {
		r.Spec.ReconciliationPolicy.State = v1.ReconciliationPolicy_Active
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=vinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Instance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateCreate() (admission.Warnings, error) {
	instancelog.Info("validate create", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), instancelog)

	observ_utils.EmitUserAuditsLogs(ctx, "Instance %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateInstance()
	if validationError != nil {
		instanceWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.InstanceResourceType,
		)
	} else {
		instanceWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.InstanceResourceType,
		)
	}

	return nil, validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	instancelog.Info("validate update", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), instancelog)

	observ_utils.EmitUserAuditsLogs(ctx, "Instance %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	validationError := r.validateUpdateInstance()
	if validationError != nil {
		instanceWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.InstanceResourceType,
		)
	} else {
		instanceWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.InstanceResourceType,
		)
	}

	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateDelete() (admission.Warnings, error) {
	instancelog.Info("validate delete", "name", r.Name)

	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(resourceK8SId, r.Annotations, operationName, context.TODO(), instancelog)

	observ_utils.EmitUserAuditsLogs(ctx, "Instance %s is being deleted on namespace %s", r.Name, r.Namespace)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *Instance) validateCreateInstance() error {
	var allErrs field.ErrorList

	if err := r.validateUniqueNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
}

func (r *Instance) validateUpdateInstance() error {
	var allErrs field.ErrorList
	if err := r.validateUniqueNameOnUpdate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
}

func (r *Instance) validateUniqueNameOnCreate() *field.Error {
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		instancelog.Info("validate unique display name", "name", r.Name)
		var instances InstanceList
		err := myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingLabels{"displayName": r.Spec.DisplayName}, client.Limit(1))
		if err != nil {
			return field.InternalError(&field.Path{}, err)
		}

		if len(instances.Items) != 0 {
			return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, fmt.Sprintf("instance display name '%s' is already taken", r.Spec.DisplayName))
		}
	}
	return nil
}

func (r *Instance) validateUniqueNameOnUpdate() *field.Error {
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		instancelog.Info("validate unique display name", "name", r.Name)
		var instances InstanceList
		err := myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingLabels{"displayName": r.Spec.DisplayName}, client.Limit(2))
		if err != nil {
			return field.InternalError(&field.Path{}, err)
		}

		if !(len(instances.Items) == 0 || len(instances.Items) == 1 && instances.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
			return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, fmt.Sprintf("instance display name '%s' is already taken", r.Spec.DisplayName))
		}
	}
	return nil
}

func (r *Instance) validateReconciliationPolicy() *field.Error {
	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.Interval != nil {
		if duration, err := time.ParseDuration(*r.Spec.ReconciliationPolicy.Interval); err == nil {
			if duration != 0 && duration < 1*time.Minute {
				return field.Invalid(field.NewPath("spec").Child("reconciliationPolicy").Child("interval"), r.Spec.ReconciliationPolicy.Interval, "must be a non-negative value with a minimum of 1 minute, e.g. 1m")
			}
		} else {
			return field.Invalid(field.NewPath("spec").Child("reconciliationPolicy").Child("interval"), r.Spec.ReconciliationPolicy.Interval, "cannot be parsed as type of time.Duration")
		}
	}

	if r.Spec.ReconciliationPolicy != nil {
		if !r.Spec.ReconciliationPolicy.State.IsActive() && !r.Spec.ReconciliationPolicy.State.IsInActive() {
			return field.Invalid(field.NewPath("spec").Child("reconciliationPolicy").Child("state"), r.Spec.ReconciliationPolicy.State, "must be either 'active' or 'inactive'")
		}
	}

	return nil
}
