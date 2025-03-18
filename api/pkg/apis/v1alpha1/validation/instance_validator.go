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

type InstanceValidator struct {
	UniqueNameInstanceLookupFunc ObjectLookupFunc
	SolutionLookupFunc           ObjectLookupFunc
	TargetLookupFunc             ObjectLookupFunc
}

func NewInstanceValidator(uniqueNameInstanceLookupFunc ObjectLookupFunc, solutionLookupFunc ObjectLookupFunc, targetLookupFunc ObjectLookupFunc) InstanceValidator {
	return InstanceValidator{
		UniqueNameInstanceLookupFunc: uniqueNameInstanceLookupFunc,
		SolutionLookupFunc:           solutionLookupFunc,
		TargetLookupFunc:             targetLookupFunc,
	}
}

// Validate Instance creation or update
// 1. DisplayName is unique
// 2. Solution exists
// 3. Target exists if provided by name rather than selector
// 4. Target is valid, i.e. either name or selector is provided
func (i *InstanceValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := i.ConvertInterfaceToInstance(newRef)
	old := i.ConvertInterfaceToInstance(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := i.ValidateUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil || new.Spec.Solution != old.Spec.Solution {
		if err := i.ValidateSolutionExist(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if (oldRef == nil || new.Spec.Target.Name != old.Spec.Target.Name) && new.Spec.Target.Name != "" {
		if err := i.ValidateTargetExist(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef != nil && (old.Spec.Scope != new.Spec.Scope) {
		errorFields = append(errorFields, ErrorField{
			FieldPath:       "spec.Scope",
			Value:           new.Spec.Scope,
			DetailedMessage: "The instance is already created. Cannot change Scope of the instance.",
		})
	}
	if oldRef != nil && !old.Spec.IsDryRun && new.Spec.IsDryRun {
		errorFields = append(errorFields, ErrorField{
			FieldPath:       "spec.isDryRun",
			Value:           new.Spec.IsDryRun,
			DetailedMessage: "The instance is already deployed. Cannot change isDryRun from false to true.",
		})
	}
	if err := i.ValidateTargetValid(new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

func (i *InstanceValidator) ValidateDelete(ctx context.Context, new interface{}) []ErrorField {
	return []ErrorField{}
}

// Validate DisplayName is unique, i.e. No existing instance with the same DisplayName
// UniqueNameInstanceLookupFunc will lookup instances with labels {"displayName": c.Spec.DisplayName}
func (i *InstanceValidator) ValidateUniqueName(ctx context.Context, c model.InstanceState) *ErrorField {
	if i.UniqueNameInstanceLookupFunc == nil {
		return nil
	}
	_, err := i.UniqueNameInstanceLookupFunc(ctx, c.Spec.DisplayName, c.ObjectMeta.Namespace)
	if err == nil || !utils.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           c.Spec.DisplayName,
			DetailedMessage: "instance displayName must be unique",
		}
	}
	return nil
}

// Validate Solution exists for instance
func (i *InstanceValidator) ValidateSolutionExist(ctx context.Context, c model.InstanceState) *ErrorField {
	if i.SolutionLookupFunc == nil {
		return nil
	}
	_, err := i.SolutionLookupFunc(ctx, ConvertReferenceToObjectName(c.Spec.Solution), c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.solution",
			Value:           c.Spec.Solution,
			DetailedMessage: "solution does not exist",
		}
	}
	return nil
}

// Validate Target exists for instance if target name is provided
func (i *InstanceValidator) ValidateTargetExist(ctx context.Context, c model.InstanceState) *ErrorField {
	if i.TargetLookupFunc == nil {
		return nil
	}
	_, err := i.TargetLookupFunc(ctx, ConvertReferenceToObjectName(c.Spec.Target.Name), c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.target.name",
			Value:           c.Spec.Target.Name,
			DetailedMessage: "target does not exist",
		}

	}
	return nil
}

// Validate Target is valid, i.e. either name or selector is provided
func (i *InstanceValidator) ValidateTargetValid(c model.InstanceState) *ErrorField {
	if c.Spec.Target.Name == "" && (c.Spec.Target.Selector == nil || len(c.Spec.Target.Selector) == 0) {
		return &ErrorField{
			FieldPath:       "spec.target",
			Value:           c.Spec.Target,
			DetailedMessage: "target must have either name or valid selector",
		}
	}
	return nil
}

func (i *InstanceValidator) ConvertInterfaceToInstance(ref interface{}) model.InstanceState {
	if ref == nil {
		return model.InstanceState{
			Spec: &model.InstanceSpec{},
		}
	}
	if state, ok := ref.(model.InstanceState); ok {
		if state.Spec == nil {
			state.Spec = &model.InstanceSpec{}
		}
		return state
	} else {
		return model.InstanceState{
			Spec: &model.InstanceSpec{},
		}
	}
}
