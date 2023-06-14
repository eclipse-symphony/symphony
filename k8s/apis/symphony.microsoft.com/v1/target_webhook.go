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

package v1

import (
	"context"
	"strings"

	configv1 "gopls-workspace/apis/config/v1"
	configutils "gopls-workspace/configutils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var targetlog = logf.Log.WithName("target-resource")
var myTargetClient client.Client
var targetValidationPolicies []configv1.ValidationPolicy

func (r *Target) SetupWebhookWithManager(mgr ctrl.Manager) error {
	myTargetClient = mgr.GetClient()

	mgr.GetFieldIndexer().IndexField(context.Background(), &Target{}, ".spec.displayName", func(rawObj client.Object) []string {
		target := rawObj.(*Target)
		return []string{target.Spec.DisplayName}
	})

	dict, _ := configutils.GetValidationPoilicies()
	if v, ok := dict["target"]; ok {
		targetValidationPolicies = v
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-symphony-microsoft-com-v1-target,mutating=true,failurePolicy=fail,sideEffects=None,groups=symphony.microsoft.com,resources=targets,verbs=create;update,versions=v1,name=mtarget.kb.io,admissionReviewVersions=v1

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
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-symphony-microsoft-com-v1-target,mutating=false,failurePolicy=fail,sideEffects=None,groups=symphony.microsoft.com,resources=targets,verbs=create;update,versions=v1,name=vtarget.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Target{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateCreate() error {
	targetlog.Info("validate create", "name", r.Name)

	return r.validateCreateTarget()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateUpdate(old runtime.Object) error {
	targetlog.Info("validate update", "name", r.Name)

	return r.validateUpdateTarget()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Target) ValidateDelete() error {
	targetlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Target) validateCreateTarget() error {
	var allErrs field.ErrorList
	var targets TargetList
	err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		allErrs = append(allErrs, field.InternalError(&field.Path{}, err))
		return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
	}
	if len(targets.Items) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, "target display name is already taken"))
		return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
	}
	if len(targetValidationPolicies) > 0 {
		err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), &client.ListOptions{})
		if err != nil {
			allErrs = append(allErrs, field.InternalError(&field.Path{}, err))
			return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
		}
		for _, p := range targetValidationPolicies {
			pack := extractTargetValidationPack(targets, p)
			ret, err := configutils.CheckValidationPack(r.ObjectMeta.Name, readTargetValidationTarget(r, p), p.ValidationType, pack)
			if err != nil {
				return err
			}
			if ret != "" {
				allErrs = append(allErrs, field.Forbidden(&field.Path{}, strings.ReplaceAll(p.Message, "%s", ret)))
				return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
			}
		}
	}
	return nil
}

func (r *Target) validateUpdateTarget() error {
	var allErrs field.ErrorList
	var targets TargetList
	err := myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		allErrs = append(allErrs, field.InternalError(&field.Path{}, err))
		return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
	}
	if !(len(targets.Items) == 0 || len(targets.Items) == 1 && targets.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("displayName"), r.Spec.DisplayName, "target display name is already taken"))
		return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
	}
	if len(targetValidationPolicies) > 0 {
		err = myTargetClient.List(context.Background(), &targets, client.InNamespace(r.Namespace), &client.ListOptions{})
		if err != nil {
			allErrs = append(allErrs, field.InternalError(&field.Path{}, err))
			return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
		}
		for _, p := range targetValidationPolicies {
			pack := extractTargetValidationPack(targets, p)
			ret, err := configutils.CheckValidationPack(r.ObjectMeta.Name, readTargetValidationTarget(r, p), p.ValidationType, pack)
			if err != nil {
				return err
			}
			if ret != "" {
				allErrs = append(allErrs, field.Forbidden(&field.Path{}, strings.ReplaceAll(p.Message, "%s", ret)))
				return apierrors.NewInvalid(schema.GroupKind{Group: "symphony.microsoft.com", Kind: "Target"}, r.Name, allErrs)
			}
		}
	}
	return nil
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
