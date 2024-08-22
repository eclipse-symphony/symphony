/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

// Check Instances associated with the Solution
var SolutionInstanceLookupFunc LinkedObjectLookupFunc
var SolutionContainerLookupFunc ObjectLookupFunc
var UniqueNameSolutionLookupFunc ObjectLookupFunc

// Validate Solution creation or update
// 1. DisplayName is unique
// 2. name and rootResource is valid. And rootResource is immutable for update
func ValidateCreateOrUpdateSolution(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := ConvertInterfaceToSolution(newRef)
	old := ConvertInterfaceToSolution(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := ValidateSolutionUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil {
		// validate naming convension
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, SolutionContainerLookupFunc); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate rootResource is not changed
		if new.Spec.RootResource != old.Spec.RootResource {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.rootResource",
				Value:           new.Spec.RootResource,
				DetailedMessage: "rootResource is immutable",
			})
		}
	}

	return errorFields
}

// Validate Solution deletion
// 1. No associated instances
func ValidateDeleteSolution(ctx context.Context, newRef interface{}) []ErrorField {
	new := ConvertInterfaceToSolution(newRef)
	errorFields := []ErrorField{}
	if err := ValidateNoInstanceForSolution(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate DisplayName is unique, i.e. No existing solution with the same DisplayName
// UniqueNameSolutionLookupFunc will lookup solutions with labels {"displayName": c.Spec.DisplayName}
func ValidateSolutionUniqueName(ctx context.Context, s model.SolutionState) *ErrorField {
	if UniqueNameSolutionLookupFunc == nil {
		return nil
	}
	_, err := UniqueNameSolutionLookupFunc(ctx, s.Spec.DisplayName, s.ObjectMeta.Namespace)
	if err == nil || !v1alpha2.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           s.Spec.DisplayName,
			DetailedMessage: "solution displayName must be unique",
		}
	}
	return nil
}

// Validate no instance associated with the solution
// SolutionInstanceLookupFunc will lookup instances with labels {"solution": s.ObjectMeta.Name}
func ValidateNoInstanceForSolution(ctx context.Context, s model.SolutionState) *ErrorField {
	if SolutionInstanceLookupFunc == nil {
		return nil
	}
	if found, err := SolutionInstanceLookupFunc(ctx, s.ObjectMeta.Name, s.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           s.ObjectMeta.Name,
			DetailedMessage: "Solution has one or more associated instances. Deletion is not allowed.",
		}
	}
	return nil
}

func ConvertInterfaceToSolution(ref interface{}) model.SolutionState {
	if ref == nil {
		return model.SolutionState{
			Spec: &model.SolutionSpec{},
		}
	}
	if state, ok := ref.(model.SolutionState); ok {
		if state.Spec == nil {
			state.Spec = &model.SolutionSpec{}
		}
		return state
	} else {
		return model.SolutionState{
			Spec: &model.SolutionSpec{},
		}
	}
}
