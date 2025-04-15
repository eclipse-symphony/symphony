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
	"strings"
	"time"

	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/apis/dynamicclient"
	"gopls-workspace/apis/metrics/v1"
	v1 "gopls-workspace/apis/model/v1"
	"gopls-workspace/utils/diagnostic"

	configutils "gopls-workspace/configutils"
	"gopls-workspace/constants"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/google/uuid"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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
var targetlog = logf.Log.WithName("target-resource")
var myTargetClient client.Reader
var targetValidationPolicies []configv1.ValidationPolicy
var targetWebhookValidationMetrics *metrics.Metrics
var projectConfig *configv1.ProjectConfig
var targetValidator validation.TargetValidator

func (r *Target) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myTargetClient = mgr.GetAPIReader()
	myConfig, err := configutils.GetProjectConfig()
	if err != nil {
		return err
	}
	projectConfig = myConfig
	if v, ok := projectConfig.ValidationPolicies["target"]; ok {
		targetValidationPolicies = v
	}

	// initialize the controller operation metrics
	if targetWebhookValidationMetrics == nil {
		metrics, err := metrics.New()
		if err != nil {
			return err
		}
		targetWebhookValidationMetrics = metrics
	}

	// Set up the target validator
	uniqueNameTargetLookupFunc := func(ctx context.Context, displayName string, namespace string) (interface{}, error) {
		return dynamicclient.GetObjectWithUniqueName(ctx, validation.Target, displayName, namespace)
	}
	targetInstanceLookupFunc := func(ctx context.Context, targetName string, namespace string, targetUid string) (bool, error) {
		instanceList, err := dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.TargetUid: targetUid}, 1)
		if err != nil {
			return false, err
		}
		// use uid label first and then name label
		if len(instanceList.Items) > 0 {
			diagnostic.InfoWithCtx(targetlog, ctx, "target look up instance using UID", "name", r.Name, "namespace", r.Namespace)
			observ_utils.EmitUserAuditsLogs(ctx, "target (%s) in namespace (%s) look up instance using UID ", r.Name, r.Namespace)
			return len(instanceList.Items) > 0, nil
		}

		if len(targetName) < 64 {
			instanceList, err = dynamicclient.ListWithLabels(ctx, validation.Instance, namespace, map[string]string{api_constants.Target: targetName}, 1)
			if err != nil {
				return false, err
			}
			if len(instanceList.Items) > 0 {
				diagnostic.InfoWithCtx(targetlog, ctx, "target look up instance using NAME", "name", r.Name, "namespace", r.Namespace)
				observ_utils.EmitUserAuditsLogs(ctx, "target (%s) in namespace (%s) look up instance using NAME ", r.Name, r.Namespace)
				return len(instanceList.Items) > 0, nil
			}
		}
		return false, nil
	}
	if projectConfig.UniqueDisplayNameForSolution {
		targetValidator = validation.NewTargetValidator(targetInstanceLookupFunc, uniqueNameTargetLookupFunc)
	} else {
		targetValidator = validation.NewTargetValidator(targetInstanceLookupFunc, nil)
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-fabric-symphony-v1-target,mutating=true,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targets,verbs=create;update,versions=v1,name=mtarget.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Target{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Target) Default() {
	ctx := diagnostic.ConstructDiagnosticContextFromAnnotations(r.Annotations, context.TODO(), targetlog)
	diagnostic.InfoWithCtx(targetlog, ctx, "default", "name", r.Name, "namespace", r.Namespace, "spec", r.Spec, "status", r.Status, "annotation", r.Annotations)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.Scope == "" {
		r.Spec.Scope = "default"
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
	if projectConfig.UniqueDisplayNameForSolution {
		r.Labels[api_constants.DisplayName] = utils.ConvertStringToValidLabel(r.Spec.DisplayName)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-fabric-symphony-v1-target,mutating=false,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targets,verbs=create;update;delete,versions=v1,name=vtarget.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Target{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateCreate() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.TargetOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myTargetClient, context.TODO(), targetlog)

	// DO NOT REMOVE THIS COMMENT
	// gofail: var validateError error

	diagnostic.InfoWithCtx(targetlog, ctx, "validate create", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Target %s is being created on namespace %s", r.Name, r.Namespace)

	validateCreateTime := time.Now()
	validationError := r.validateCreateTarget(ctx)
	if validationError != nil {
		targetWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.InvalidResource,
			metrics.TargetResourceType,
		)
	} else {
		targetWebhookValidationMetrics.ControllerValidationLatency(
			validateCreateTime,
			metrics.CreateOperationType,
			metrics.ValidResource,
			metrics.TargetResourceType,
		)
	}

	return nil, validationError
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.TargetOperationNamePrefix, constants.ActivityOperation_Write)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myTargetClient, context.TODO(), targetlog)

	diagnostic.InfoWithCtx(targetlog, ctx, "validate update", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Target %s is being updated on namespace %s", r.Name, r.Namespace)

	validateUpdateTime := time.Now()
	oldTarget, ok := old.(*Target)
	if !ok {
		err := fmt.Errorf("expected a Target object")
		diagnostic.ErrorWithCtx(targetlog, ctx, err, "failed to convert old object to Target")
		return nil, err
	}
	validationError := r.validateUpdateTarget(ctx, oldTarget)
	if validationError != nil {
		targetWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.InvalidResource,
			metrics.TargetResourceType,
		)
	} else {
		targetWebhookValidationMetrics.ControllerValidationLatency(
			validateUpdateTime,
			metrics.UpdateOperationType,
			metrics.ValidResource,
			metrics.TargetResourceType,
		)
	}

	return nil, validationError
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateDelete() (admission.Warnings, error) {
	resourceK8SId := r.GetNamespace() + "/" + r.GetName()
	operationName := fmt.Sprintf("%s/%s", constants.TargetOperationNamePrefix, constants.ActivityOperation_Delete)
	ctx := configutils.PopulateActivityAndDiagnosticsContextFromAnnotations(r.GetNamespace(), resourceK8SId, r.Annotations, operationName, myTargetClient, context.TODO(), targetlog)

	diagnostic.InfoWithCtx(targetlog, ctx, "validate delete", "name", r.Name, "namespace", r.Namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "Target %s is being deleted on namespace %s", r.Name, r.Namespace)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, r.validateDeleteTarget(ctx)
}

func (r *Target) validateCreateTarget(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertTargetState()
	if err != nil {
		diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate create target - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := targetValidator.ValidateCreateOrUpdate(ctx, state, nil)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if err := r.validateValidationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate create target", "name", r.Name, "namespace", r.Namespace)
	return err
}

func (r *Target) validateValidationPolicy() *field.Error {
	var targets TargetList
	if len(targetValidationPolicies) > 0 {
		err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), &client.ListOptions{})
		if err != nil {
			return field.InternalError(&field.Path{}, err)
		}
		for _, p := range targetValidationPolicies {
			pack := extractTargetValidationPack(targets, p)
			ret, err := configutils.CheckValidationPack(r.ObjectMeta.Name, readTargetValidationTarget(r, p), p.ValidationType, pack)
			if err != nil {
				return field.InternalError(&field.Path{}, err)
			}
			if ret != "" {
				return field.Forbidden(&field.Path{}, strings.ReplaceAll(p.Message, "%s", ret))
			}
		}
	}
	return nil
}

func (r *Target) validateReconciliationPolicy() *field.Error {
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

func (r *Target) validateUpdateTarget(ctx context.Context, old *Target) error {
	var allErrs field.ErrorList
	state, err := r.ConvertTargetState()
	if err != nil {
		diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate update target - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	oldState, err := old.ConvertTargetState()
	if err != nil {
		diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate update target - convert old", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := targetValidator.ValidateCreateOrUpdate(ctx, state, oldState)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)
	if err := r.validateValidationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate update target", "name", r.Name, "namespace", r.Namespace)
	return err
}
func (r *Target) validateDeleteTarget(ctx context.Context) error {
	var allErrs field.ErrorList
	state, err := r.ConvertTargetState()
	if err != nil {
		diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate delete target - convert current", "name", r.Name, "namespace", r.Namespace)
		return err
	}
	ErrorFields := targetValidator.ValidateDelete(ctx, state)
	allErrs = validation.ConvertErrorFieldsToK8sError(ErrorFields)

	if len(allErrs) == 0 {
		return nil
	}

	err = apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name, allErrs)
	diagnostic.ErrorWithCtx(targetlog, ctx, err, "validate delete target", "name", r.Name, "namespace", r.Namespace)
	return err
}
func readTargetValidationTarget(target *Target, p configv1.ValidationPolicy) string {
	if p.SelectorType == "topologies.bindings" && p.SelectorKey == "provider" {
		for _, topology := range target.Spec.Topologies {
			for _, binding := range topology.Bindings {
				if binding.Provider == p.SelectorValue {
					if strings.HasPrefix(p.SpecField, "binding.config.") {
						dictKey := p.SpecField[15:]
						return binding.Config[dictKey]
					}
				}
			}
		}
	}
	return ""
}
func extractTargetValidationPack(list TargetList, p configv1.ValidationPolicy) []configv1.ValidationStruct {
	pack := make([]configv1.ValidationStruct, 0)
	for _, t := range list.Items {
		s := configv1.ValidationStruct{}
		if p.SelectorType == "topologies.bindings" && p.SelectorKey == "provider" {
			found := false
			for _, topology := range t.Spec.Topologies {
				for _, binding := range topology.Bindings {
					if binding.Provider == p.SelectorValue {
						if strings.HasPrefix(p.SpecField, "binding.config.") {
							dictKey := p.SpecField[15:]
							s.Field = binding.Config[dictKey]
							s.Name = t.ObjectMeta.Name
							pack = append(pack, s)
						}
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
	}
	return pack
}

func (r *Target) ConvertTargetState() (model.TargetState, error) {
	retErr := apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name,
		field.ErrorList{field.InternalError(nil, v1alpha2.NewCOAError(nil, "Unable to convert to target state", v1alpha2.BadRequest))})
	bytes, err := json.Marshal(r)
	if err != nil {
		return model.TargetState{}, retErr
	}
	var state model.TargetState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		return model.TargetState{}, retErr
	}
	return state, nil
}
