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

var (
	solutionMaxNameLength = 61
	solutionMinNameLength = 7
)

type SolutionValidator struct {
	// Check Instances associated with the Solution
	SolutionInstanceLookupFunc   LinkedObjectLookupFunc
	SolutionContainerLookupFunc  ObjectLookupFunc
	UniqueNameSolutionLookupFunc ObjectLookupFunc
}

func NewSolutionValidator(solutionInstanceLookupFunc LinkedObjectLookupFunc, solutionContainerLookupFunc ObjectLookupFunc, uniqueNameSolutionLookupFunc ObjectLookupFunc) SolutionValidator {
	return SolutionValidator{
		SolutionInstanceLookupFunc:   solutionInstanceLookupFunc,
		SolutionContainerLookupFunc:  solutionContainerLookupFunc,
		UniqueNameSolutionLookupFunc: uniqueNameSolutionLookupFunc,
	}
}

// Validate Solution creation or update
// 1. DisplayName is unique
// 2. name and rootResource is valid. And rootResource is immutable for update
func (s *SolutionValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := s.ConvertInterfaceToSolution(newRef)
	old := s.ConvertInterfaceToSolution(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := s.ValidateSolutionUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil {
		// validate naming convension
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource, solutionMinNameLength, solutionMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, s.SolutionContainerLookupFunc); err != nil {
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
func (s *SolutionValidator) ValidateDelete(ctx context.Context, newRef interface{}) []ErrorField {
	new := s.ConvertInterfaceToSolution(newRef)
	errorFields := []ErrorField{}
	if err := s.ValidateNoInstanceForSolution(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate DisplayName is unique, i.e. No existing solution with the same DisplayName
// UniqueNameSolutionLookupFunc will lookup solutions with labels {"displayName": c.Spec.DisplayName}
func (s *SolutionValidator) ValidateSolutionUniqueName(ctx context.Context, solution model.SolutionState) *ErrorField {
	if s.UniqueNameSolutionLookupFunc == nil {
		return nil
	}
	_, err := s.UniqueNameSolutionLookupFunc(ctx, solution.Spec.DisplayName, solution.ObjectMeta.Namespace)
	if err == nil || !utils.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           solution.Spec.DisplayName,
			DetailedMessage: "solution displayName must be unique",
		}
	}
	return nil
}

// Validate no instance associated with the solution
// SolutionInstanceLookupFunc will lookup instances with labels {"solution": s.ObjectMeta.Name}
func (s *SolutionValidator) ValidateNoInstanceForSolution(ctx context.Context, solution model.SolutionState) *ErrorField {
	if s.SolutionInstanceLookupFunc == nil {
		return nil
	}
	if found, err := s.SolutionInstanceLookupFunc(ctx, solution.ObjectMeta.Name, solution.ObjectMeta.Namespace, string(solution.ObjectMeta.UID)); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           solution.ObjectMeta.Name,
			DetailedMessage: "Solution has one or more associated instances. Deletion is not allowed.",
		}
	}
	return nil
}

func (s *SolutionValidator) ConvertInterfaceToSolution(ref interface{}) model.SolutionState {
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
