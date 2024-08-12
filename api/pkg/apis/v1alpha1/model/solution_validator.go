/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"context"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

// Check Instances associated with the Solution
var SolutionInstanceLookupFunc LinkedObjectLookupFunc
var SolutionContainerLookupFunc ObjectLookupFunc
var UniqueNameSolutionLookupFunc ObjectLookupFunc

func (s SolutionState) ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField {
	var oldSolution SolutionState
	if old != nil {
		var ok bool
		oldSolution, ok = old.(SolutionState)
		if !ok {
			old = nil
		}
	}
	if s.Spec == nil {
		s.Spec = &SolutionSpec{}
	}

	errorFields := []ErrorField{}
	if old == nil || s.Spec.DisplayName != oldSolution.Spec.DisplayName {
		if err := s.ValidateUniqueName(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if old == nil {
		// validate naming convension
		if err := ValidateObjectName(s.ObjectMeta.Name, s.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := s.ObjectMeta.ValidateRootResource(ctx, s.Spec.RootResource, SolutionContainerLookupFunc); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate rootResource is not changed
		if s.Spec.RootResource != oldSolution.Spec.RootResource {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.rootResource",
				Value:           s.Spec.RootResource,
				DetailedMessage: "rootResource is immutable",
			})
		}
	}

	return errorFields
}

func (s SolutionState) ValidateDelete(ctx context.Context) []ErrorField {
	errorFields := []ErrorField{}
	if err := s.ValidateNoInstance(ctx); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}
func (s *SolutionState) ValidateUniqueName(ctx context.Context) *ErrorField {
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
func (s *SolutionState) ValidateNoInstance(ctx context.Context) *ErrorField {
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
