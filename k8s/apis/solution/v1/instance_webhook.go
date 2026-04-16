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
	k8smodel "gopls-workspace/apis/model/v1"
	v1 "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"gopls-workspace/constants"
	"gopls-workspace/history"
	"gopls-workspace/utils/diagnostic"
	"time"

	fabric "gopls-workspace/apis/fabric/v1"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/google/uuid"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"

	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
var k8sClient client.Client
var instanceWebhookValidationMetrics *metrics.Metrics
var instanceProjectConfig *configv1.ProjectConfig
var instanceValidator validation.InstanceValidator
var instanceHistory history.InstanceHistory

func (r *Instance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myInstanceClient = mgr.GetAPIReader()
	k8sClient = mgr.GetClient()

	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.solution", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.Solution}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.target.name", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.Target.Name}
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
		return dynamicclient.GetObjectWithUniqueName(ctx, validation.Instance, displayName, namespace)
	}
	solutionLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(ctx, validation.Solution, name, namespace)
	}
	targetLookupFunc := func(ctx context.Context, name string, namespace string) (interface{}, error) {
		return dynamicclient.Get(ctx, validation.Target, name, namespace)
	}
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		instanceValidator = validation.NewInstanceValidator(uniqueNameInstanceLookupFunc, solutionLookupFunc, targetLookupFunc)
	} else {
		instanceValidator = validation.NewInstanceValidator(nil, solutionLookupFunc, targetLookupFunc)
	}

	saveInstanceHistoryFunc := func(ctx context.Context, objectName string, object interface{}) error {
		instance, ok := object.(*Instance)
		if !ok {
			err := fmt.Errorf("expected an Instance object")
			diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to convert old object to Instance", "name", r.Name, "namespace", r.Namespace)
			return err
		}
		currentTime := time.Now()
		diagnostic.InfoWithCtx(instancelog, ctx, "Saving old instance history", "Current time", currentTime, "name", instance.Name)

		// get solution spec
		var solutionSpec k8smodel.SolutionSpec
		res, err := dynamicclient.Get(ctx, validation.Solution, validation.ConvertReferenceToObjectName(instance.Spec.Solution), instance.Namespace)
		if err != nil {
			err := fmt.Errorf("failed to get solution, instance: %s, error: %v", instance.Name, err)
			diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get solution spec for instance", "name", r.Name, "namespace", r.Namespace)
		} else if res.Object != nil {
			jsonData, _ := json.Marshal(res.Object["spec"])
			err = utils.UnmarshalJson(jsonData, &solutionSpec)
			if err != nil {
				diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get solution spec for instance", "name", r.Name, "namespace", r.Namespace)
			}
		}

		// get target spec
		var targetSpec k8smodel.TargetSpec
		if instance.Spec.Target.Name != "" {
			res, err = dynamicclient.Get(ctx, validation.Target, validation.ConvertReferenceToObjectName(instance.Spec.Target.Name), instance.Namespace)
			if err != nil {
				err := fmt.Errorf("failed to get target, instance: %s, error: %v", instance.Name, err)
				diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get target spec for instance", "name", r.Name, "namespace", r.Namespace)
			}
			jsonData, _ := json.Marshal(res.Object["spec"])
			err = utils.UnmarshalJson(jsonData, &targetSpec)
			if err != nil {
				diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get target spec for instance", "name", r.Name, "namespace", r.Namespace)
			}
		}

		var history InstanceHistory
		history.ObjectMeta = metav1.ObjectMeta{
			Name:      instance.Name + constants.ResourceSeperator + currentTime.Format("20060102150405"),
			Namespace: instance.Namespace,
		}

		history.Spec = k8smodel.InstanceHistorySpec{
			DisplayName:          instance.Spec.DisplayName,
			Scope:                instance.Spec.Scope,
			Parameters:           instance.Spec.Parameters,
			Metadata:             instance.Spec.Metadata,
			Solution:             solutionSpec,
			SolutionId:           instance.Spec.Solution,
			Target:               targetSpec,
			TargetSelector:       instance.Spec.Target.Selector,
			TargetId:             instance.Spec.Target.Name,
			Topologies:           instance.Spec.Topologies,
			Pipelines:            instance.Spec.Pipelines,
			IsDryRun:             instance.Spec.IsDryRun,
			ReconciliationPolicy: instance.Spec.ReconciliationPolicy,
			RootResource:         instance.Name,
		}

		var result InstanceHistory
		err = k8sClient.Get(ctx, client.ObjectKey{Name: history.GetName(), Namespace: history.GetNamespace()}, &result)
		if err != nil && errors.IsNotFound(err) {
			// Resource does not exist, create it
			err = k8sClient.Create(ctx, &history)
			if err != nil {
				err := fmt.Errorf("upsert instance history failed, instance: %s, error: %v", instance.Name, err)
				diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to save instance history for instance", "name", r.Name, "namespace", r.Namespace)
				return err
			}
			// If the instance has a status, save it in the history
			if !instance.Status.LastModified.IsZero() {
				history.Status = instance.Status
				// Reset ProvisioningStatus to avoid saving it in the history
				history.Status.ProvisioningStatus = model.ProvisioningStatus{}
				err = k8sClient.Status().Update(ctx, &history)
				if err != nil {
					err := fmt.Errorf("upsert instance history status failed, instance: %s, history: %s, error: %v", instance.Name, history.GetName(), err)
					diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to save instance history for instance", "name", r.Name, "namespace", r.Namespace)
					return err
				}
			}
			diagnostic.InfoWithCtx(instancelog, ctx, "Saved instance history", "name", history.ObjectMeta.Name, "namespace", instance.Namespace)
		} else if err != nil {
			diagnostic.ErrorWithCtx(instancelog, ctx, err, "Unexpected error saving instance history", "name", r.Name, "namespace", r.Namespace)
		}

		return nil
	}
	instanceHistory = history.NewInstanceHistory(saveInstanceHistoryFunc)

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=minstance.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Instance{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Instance) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), instancelog)
	diagnostic.InfoWithCtx(instancelog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "status", r.Status)
	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.State == "" {
		r.Spec.ReconciliationPolicy.State = v1.ReconciliationPolicy_Active
	}

	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}
	if r.Annotations[api_constants.GuidKey] == "" {
		r.Annotations[api_constants.GuidKey] = uuid.New().String()
	}

	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	if instanceProjectConfig.UniqueDisplayNameForSolution {
		r.Labels[api_constants.DisplayName] = utils.ConvertStringToValidLabel(r.Spec.DisplayName)
	}

	// Remove api_constants.Solution and api_constants.Targetfrom r.Labels if it exists
	if _, exists := r.Labels[api_constants.Solution]; exists {
		delete(r.Labels, api_constants.Solution)
	}
	if _, exists := r.Labels[api_constants.Target]; exists {
		delete(r.Labels, api_constants.Target)
	}

	var solutionResult Solution
	err := k8sClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.Solution), Namespace: r.Namespace}, &solutionResult)
	if err != nil {
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get solution", "name", r.Name, "namespace", r.Namespace)
	}
	r.Labels[api_constants.SolutionUid] = string(solutionResult.UID)

	var targetResult fabric.Target
	err = k8sClient.Get(ctx, client.ObjectKey{Name: validation.ConvertReferenceToObjectName(r.Spec.Target.Name), Namespace: r.Namespace}, &targetResult)
	if err != nil {
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to get target", "name", r.Name, "namespace", r.Namespace)
	}
	r.Labels[api_constants.TargetUid] = string(targetResult.UID)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update;delete,versions=v1,name=vinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Instance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myInstanceClient, context.TODO(), instancelog)

	diagnostic.InfoWithCtx(instancelog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
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
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myInstanceClient, context.TODO(), instancelog)

	diagnostic.InfoWithCtx(instancelog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Instance %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldInstance, ok := old.(*Instance)
	if !ok {
		err := fmt.Errorf("expected an Instance object")
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to convert old object to Instance", "name", r.Name, "namespace", r.Namespace)
		return nil, err
	}

	// Save the old object
	if !r.Spec.DeepEquals(oldInstance.Spec) {
		diagnostic.InfoWithCtx(instancelog, ctx, "saving old instance history", "name", oldInstance.Name)
		err := instanceHistory.SaveInstanceHistoryFunc(ctx, oldInstance.Name, oldInstance)
		if err != nil {
			diagnostic.ErrorWithCtx(instancelog, ctx, err, "failed to save Instance history", "name", r.Name, "namespace", r.Namespace)
		}
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
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.InstanceOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myInstanceClient, context.TODO(), instancelog)

	diagnostic.InfoWithCtx(instancelog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Instance %s is being deleted on namespace %s", r.Name, r.Namespace)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *Instance) validateCreateInstance(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertInstanceState()
	if err != nil {
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "validate create instance - convert current", "name", r.Name, "namespace", r.Namespace)
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

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(instancelog, ctx, err, "validate create instance", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Instance) validateUpdateInstance(ctx context.Context, old *Instance) error {
	var allErrs field.ErrorList
	state, err := r.ConvertInstanceState()
	if err != nil {
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "validate update instance - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	oldState, err := old.ConvertInstanceState()
	if err != nil {
		diagnostic.ErrorWithCtx(instancelog, ctx, err, "validate update instance - convert old", "name", r.Name, "namespace", r.Namespace)
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

	err = apierrors.NewInvalid(schema.GroupKind{Group: "solution.symphony", Kind: "Instance"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(instancelog, ctx, err, "validate update instance", "name", r.Name, "namespace", r.Namespace)
	return err
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
