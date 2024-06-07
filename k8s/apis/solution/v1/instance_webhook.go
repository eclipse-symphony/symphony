/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"
	"gopls-workspace/apis/metrics/v1"
	v1 "gopls-workspace/apis/model/v1"
	"gopls-workspace/configutils"
	"time"

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
var myInstanceClient client.Client
var instanceWebhookValidationMetrics *metrics.Metrics

func (r *Instance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myInstanceClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.displayName", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.DisplayName}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, "spec.solution", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.Solution}
	})
	mgr.GetFieldIndexer().IndexField(context.Background(), &Instance{}, ".spec.rootResource", func(rawObj client.Object) []string {
		instance := rawObj.(*Instance)
		return []string{instance.Spec.RootResource}
	})

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

	if r.Spec.ReconciliationPolicy != nil && r.Spec.ReconciliationPolicy.State == "" {
		r.Spec.ReconciliationPolicy.State = v1.ReconciliationPolicy_Active
	}

	if r.Spec.RootResource != "" {
		var instanceContainer InstanceContainer
		err := myInstanceClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &instanceContainer)
		if err != nil {
			instancelog.Error(err, "failed to get instance container", "name", r.Spec.RootResource)
		} else {
			ownerReference := metav1.OwnerReference{
				APIVersion: instanceContainer.APIVersion,
				Kind:       instanceContainer.Kind,
				Name:       instanceContainer.Name,
				UID:        instanceContainer.UID,
			}

			if !configutils.CheckOwnerReferenceAlreadySet(r.OwnerReferences, ownerReference) {
				r.OwnerReferences = append(r.OwnerReferences, ownerReference)
			}
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=instances,verbs=create;update,versions=v1,name=vinstance.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Instance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Instance) ValidateCreate() (admission.Warnings, error) {
	instancelog.Info("validate create", "name", r.Name)

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
	if err := r.validateNameOnCreate(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateRootResource(); err != nil {
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
	var instances InstanceList
	err := myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingFields{"spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return field.InternalError(&field.Path{}, err)
	}

	if len(instances.Items) != 0 {
		return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, fmt.Sprintf("instance display name '%s' is already taken", r.Spec.DisplayName))
	}
	return nil
}

func (r *Instance) validateUniqueNameOnUpdate() *field.Error {
	var instances InstanceList
	err := myInstanceClient.List(context.Background(), &instances, client.InNamespace(r.Namespace), client.MatchingFields{"spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return field.InternalError(&field.Path{}, err)
	}

	if !(len(instances.Items) == 0 || len(instances.Items) == 1 && instances.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, fmt.Sprintf("instance display name '%s' is already taken", r.Spec.DisplayName))
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

func (r *Instance) validateNameOnCreate() *field.Error {
	return configutils.ValidateObjectName(r.ObjectMeta.Name, r.Spec.RootResource)
}

func (r *Instance) validateRootResource() *field.Error {
	var instanceContainer InstanceContainer
	err := myInstanceClient.Get(context.Background(), client.ObjectKey{Name: r.Spec.RootResource, Namespace: r.Namespace}, &instanceContainer)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("rootResource"), r.Spec.RootResource, "rootResource must be a valid instance container")
	}

	if len(r.ObjectMeta.OwnerReferences) == 0 {
		return field.Invalid(field.NewPath("metadata").Child("ownerReference"), len(r.ObjectMeta.OwnerReferences), "ownerReference must be set")
	}

	return nil
}
