/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"strings"
	"time"

	configv1 "gopls-workspace/apis/config/v1"
	"gopls-workspace/apis/metrics/v1"
	v1 "gopls-workspace/apis/model/v1"
	configutils "gopls-workspace/configutils"

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
var targetlog = logf.Log.WithName("target-resource")
var myTargetClient client.Client
var targetValidationPolicies []configv1.ValidationPolicy
var targetWebhookValidationMetrics *metrics.Metrics

func (r *Target) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myTargetClient = mgr.GetClient()

	mgr.GetFieldIndexer().IndexField(context.Background(), &Target{}, "spec.displayName", func(rawObj client.Object) []string {
		target := rawObj.(*Target)
		return []string{target.Spec.DisplayName}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Target{}, ".spec.rootResource", func(rawObj client.Object) []string {
		target := rawObj.(*Target)
		return []string{target.Spec.RootResource}
	})

	dict, _ := configutils.GetValidationPoilicies()
	if v, ok := dict["target"]; ok {
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

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-fabric-symphony-v1-target,mutating=true,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targets,verbs=create;update,versions=v1,name=mtarget.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Target{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Target) Default() {
	targetlog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}

	if r.Spec.Scope == "" {
		r.Spec.Scope = "default"
	}

	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.State == "" {
		r.Spec.ReconciliationPolicy.State = v1.ReconciliationPolicy_Active
	}

	if r.Spec.RootResource != "" {
		var targetContainer TargetContainer
		err := myTargetClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &targetContainer)
		if err != nil {
			targetlog.Error(err, "failed to get target container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: targetContainer.APIVersion,
				Kind:       targetContainer.Kind,
				Name:       targetContainer.Name,
				UID:        targetContainer.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-fabric-symphony-v1-target,mutating=false,failurePolicy=fail,sideEffects=None,groups=fabric.symphony,resources=targets,verbs=create;update,versions=v1,name=vtarget.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Target{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateCreate() (admission.Warnings, error) {
	targetlog.Info("validate create", "name", r.Name)

	validateCreateTime := time.Now()
	validationError := r.validateCreateTarget()
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
	targetlog.Info("validate update", "name", r.Name)

	validateUpdateTime := time.Now()
	validationError := r.validateUpdateTarget()
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
	targetlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *Target) validateCreateTarget() error {
	var allErrs field.ErrorList

	if err := r.validateUniqueNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateValidationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateRootResource(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name, allErrs)
}

func (r *Target) validateUniqueNameOnCreate() *field.Error {
	var targets TargetList
	err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), client.MatchingFields{"spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return field.InternalError(&field.Path{}, err)
	}

	if len(targets.Items) != 0 {
		return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, "target display name is already taken")
	}

	return nil
}

func (r *Target) validateUniqueNameOnUpdate() *field.Error {
	var targets TargetList
	err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), client.MatchingFields{"spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return field.InternalError(&field.Path{}, err)
	}

	if !(len(targets.Items) == 0 || len(targets.Items) == 1 && targets.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, "target display name is already taken")
	}

	return nil
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

func (r *Target) validateNameOnCreate() *field.Error {
	return configutils.ValidateObjectName(r.ObjectMeta.Name, r.Spec.RootResource)
}

func (r *Target) validateRootResource() *field.Error {
	var targetContainer TargetContainer
	err := myTargetClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &targetContainer)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("rootResource"), r.Spec.RootResource, "rootResource must be a valid target container")
	}

	if len(r.ObjectMeta.OwnerReferences) == 0 {
		return field.Invalid(field.NewPath("metadata").Child("ownerReference"), len(r.ObjectMeta.OwnerReferences), "ownerReference must be set")
	}

	return nil
}

func (r *Target) validateUpdateTarget() error {
	var allErrs field.ErrorList
	if err := r.validateUniqueNameOnUpdate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateValidationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateReconciliationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "fabric.symphony", Kind: "Target"}, r.Name, allErrs)
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
