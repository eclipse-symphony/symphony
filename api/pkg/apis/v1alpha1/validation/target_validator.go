/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
)

type TargetValidator struct {
	// Check Instance associated with the Solution
	TargetInstanceLookupFunc   LinkedObjectLookupFunc
	UniqueNameTargetLookupFunc ObjectLookupFunc
}

func NewTargetValidator(targetInstanceLookupFunc LinkedObjectLookupFunc, uniqueNameTargetLookupFunc ObjectLookupFunc) TargetValidator {
	return TargetValidator{
		TargetInstanceLookupFunc:   targetInstanceLookupFunc,
		UniqueNameTargetLookupFunc: uniqueNameTargetLookupFunc,
	}
}

// Validate Target creation or update
// 1. DisplayName is unique
func (t *TargetValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := t.ConvertInterfaceToTarget(newRef)
	old := t.ConvertInterfaceToTarget(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := t.ValidateTargetUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef != nil && (old.Spec.Scope != new.Spec.Scope) {
		errorFields = append(errorFields, ErrorField{
			FieldPath:       "spec.Scope",
			Value:           new.Spec.Scope,
			DetailedMessage: "The target is already created. Cannot change Scope of the target.",
		})
	}
	if oldRef != nil && (old.Spec.SolutionScope != new.Spec.SolutionScope) && t.TargetInstanceLookupFunc != nil {
		if found, err := t.TargetInstanceLookupFunc(ctx, new.ObjectMeta.Name, new.ObjectMeta.Namespace); err != nil || found {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.SolutionScope",
				Value:           new.Spec.SolutionScope,
				DetailedMessage: "Target has one or more associated instances. Cannot change SolutionScope of the target.",
			})
		}
	}
	if oldRef != nil && !old.Spec.IsDryRun && new.Spec.IsDryRun {
		errorFields = append(errorFields, ErrorField{
			FieldPath:       "spec.isDryRun",
			Value:           new.Spec.IsDryRun,
			DetailedMessage: "The target is already deployed. Cannot change isDryRun from false to true.",
		})
	}
	return errorFields
}

// Validate Target deletion
// 1. No associated instances
func (t *TargetValidator) ValidateDelete(ctx context.Context, newRef interface{}) []ErrorField {
	target := t.ConvertInterfaceToTarget(newRef)
	errorFields := []ErrorField{}
	if err := t.ValidateNoInstanceForTarget(ctx, target); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate DisplayName is unique, i.e. No existing target with the same DisplayName
// UniqueNameTargetLookupFunc will lookup targets with labels {"displayName": t.Spec.DisplayName}
func (t *TargetValidator) ValidateTargetUniqueName(ctx context.Context, target model.TargetState) *ErrorField {
	if t.UniqueNameTargetLookupFunc == nil {
		return nil
	}
	_, err := t.UniqueNameTargetLookupFunc(ctx, target.Spec.DisplayName, target.ObjectMeta.Namespace)
	if err == nil || !utils.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           target.Spec.DisplayName,
			DetailedMessage: "target displayName must be unique",
		}
	}
	return nil
}

// Validate No associated instances for the target
// TargetInstanceLookupFunc will lookup instances with labels {"target": t.ObjectMeta.Name}
func (t *TargetValidator) ValidateNoInstanceForTarget(ctx context.Context, target model.TargetState) *ErrorField {
	if t.TargetInstanceLookupFunc == nil {
		return nil
	}
	if found, err := t.TargetInstanceLookupFunc(ctx, target.ObjectMeta.Name, target.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           target.ObjectMeta.Name,
			DetailedMessage: "Target has one or more associated instances. Deletion is not allowed.",
		}
	}
	return nil
}

func (t *TargetValidator) ConvertInterfaceToTarget(ref interface{}) model.TargetState {
	if ref == nil {
		return model.TargetState{
			Spec: &model.TargetSpec{},
		}
	}
	if state, ok := ref.(model.TargetState); ok {
		if state.Spec == nil {
			state.Spec = &model.TargetSpec{}
		}
		return state
	} else {
		return model.TargetState{
			Spec: &model.TargetSpec{},
		}
	}
}
