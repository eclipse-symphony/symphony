/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/apis/dynamicclient"
	"gopls-workspace/apis/metrics/v1"
	v1 "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"time"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"

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
var instanceWebhookValidationMetrics *metrics.Metrics
var instanceProjectConfig *configv1.ProjectConfig
var instanceValidator validation.InstanceValidator

func (r *Instance) SetupWebhookWithManager(mgr ctrl.Manager) error {
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

	// Load validator functions
	uniqueNameInstanceLookupFunc := func(ctx context.Context, displayName string, namespace string) (interface{}, error) {
		return dynamicclient.GetObjectWithUniqueName(validation.Instance, displayName, namespace)
	}
	solutionLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(validation.Solution, name, namespace)
	}
	targetLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(validation.Target, name, namespace)
	}
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		instanceValidator = validation.NewInstanceValidator(uniqueNameInstanceLookupFunc, solutionLookupFunc, targetLookupFunc)
	} else {
		instanceValidator = validation.NewInstanceValidator(nil, solutionLookupFunc, targetLookupFunc)
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

	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.State == "" {
		r.Spec.ReconciliationPolicy.State = v1.ReconciliationPolicy_Active
	}

	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		r.Labels[api_constants.DisplayName] = utils.ConvertStringToValidLabel(r.Spec.DisplayName)
	}
	r.Labels[api_constants.Solution] = validation.ConvertReferenceToObjectName(r.Spec.Solution)
	r.Labels[api_constants.Target] = r.Spec.Target.Name
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
	validationError := r.validateCreateInstance(ctx)
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
	oldInstance, ok := old.(*Instance)
	if !ok {
		return nil, fmt.Errorf("expected an Instance object")
	}
	validationError := r.validateUpdateInstance(ctx, oldInstance)
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

func (r *Instance) validateCreateInstance(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertInstanceState()
	if err != nil {
		return err
	}
	// TODO: add proper context
	ErrorFields := instanceValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
}

func (r *Instance) validateUpdateInstance(ctx context.Context, old *Instance) error {
	var allErrs field.ErrorList
	state, err := r.ConvertInstanceState()
	if err != nil {
		return err
	}
	oldState, err := old.ConvertInstanceState()
	if err != nil {
		return err
	}
	// TODO: add proper context
	ErrorFields := instanceValidator.ValidateCreateOrUpdate(ctx, state, oldState)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
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

func (r *Instance) ConvertInstanceState() (model.InstanceState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to instance state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.InstanceState{}, retErr
	}
	var state model.InstanceState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.InstanceState{}, retErr
	}
	return state, nil
}
