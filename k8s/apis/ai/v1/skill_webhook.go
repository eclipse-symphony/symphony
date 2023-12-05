/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var skilllog = logf.Log.WithName("skill-resource")
var mySkillClient client.Client

func (r *Skill) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mySkillClient = mgr.GetClient()
	mgr.GetFieldIndexer().IndexField(context.Background(), &Skill{}, ".spec.displayName", func(rawObj client.Object) []string {
		skill := rawObj.(*Skill)
		return []string{skill.Spec.DisplayName}
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

var _ webhook.Defaulter = &Skill{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Skill) Default() {
	skilllog.Info("default", "name", r.Name)

	if r.Spec.DisplayName == "" {
		r.Spec.DisplayName = r.ObjectMeta.Name
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

var _ webhook.Validator = &Skill{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Skill) ValidateCreate() error {
	skilllog.Info("validate create", "name", r.Name)

	return r.validateCreateSkill()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Skill) ValidateUpdate(old runtime.Object) error {
	skilllog.Info("validate update", "name", r.Name)

	return r.validateUpdateSkill()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Skill) ValidateDelete() error {
	skilllog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Skill) validateCreateSkill() error {
	var skills SkillList
	mySkillClient.List(context.Background(), &skills, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if len(skills.Items) != 0 {
		return fmt.Errorf("skill display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}

func (r *Skill) validateUpdateSkill() error {
	var skills SkillList
	err := mySkillClient.List(context.Background(), &skills, client.InNamespace(r.Namespace), client.MatchingFields{".spec.displayName": r.Spec.DisplayName})
	if err != nil {
		return err
	}
	if !(len(skills.Items) == 0 || len(skills.Items) == 1 && skills.Items[0].ObjectMeta.Name == r.ObjectMeta.Name) {
		return fmt.Errorf("skill display name '%s' is already taken", r.Spec.DisplayName)
	}
	return nil
}
