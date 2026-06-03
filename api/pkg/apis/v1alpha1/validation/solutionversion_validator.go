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
	solutionversionMaxNameLength = 61
	solutionversionMinNameLength = 1
)

type SolutionVersionValidator struct {
	// Check Instances associated with the SolutionVersion
	SolutionVersionInstanceLookupFunc   LinkedObjectLookupFunc
	SolutionLookupFunc  ObjectLookupFunc
	UniqueNameSolutionVersionLookupFunc ObjectLookupFunc
}

func NewSolutionVersionValidator(solutionversionInstanceLookupFunc LinkedObjectLookupFunc, solutionversionContainerLookupFunc ObjectLookupFunc, uniqueNameSolutionVersionLookupFunc ObjectLookupFunc) SolutionVersionValidator {
	return SolutionVersionValidator{
		SolutionVersionInstanceLookupFunc:   solutionversionInstanceLookupFunc,
		SolutionLookupFunc:  solutionversionContainerLookupFunc,
		UniqueNameSolutionVersionLookupFunc: uniqueNameSolutionVersionLookupFunc,
	}
}

// Validate SolutionVersion creation or update
// 1. DisplayName is unique
// 2. name and rootResource is valid. And rootResource is immutable for update
func (s *SolutionVersionValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := s.ConvertInterfaceToSolutionVersion(newRef)
	old := s.ConvertInterfaceToSolutionVersion(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := s.ValidateSolutionVersionUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil {
		// validate naming convension
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource, solutionversionMinNameLength, solutionversionMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, s.SolutionLookupFunc); err != nil {
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

// Validate SolutionVersion deletion
// 1. No associated instances
func (s *SolutionVersionValidator) ValidateDelete(ctx context.Context, newRef interface{}) []ErrorField {
	new := s.ConvertInterfaceToSolutionVersion(newRef)
	errorFields := []ErrorField{}
	if err := s.ValidateNoInstanceForSolutionVersion(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate DisplayName is unique, i.e. No existing solutionversion with the same DisplayName
// UniqueNameSolutionVersionLookupFunc will lookup solutionversions with labels {"displayName": c.Spec.DisplayName}
func (s *SolutionVersionValidator) ValidateSolutionVersionUniqueName(ctx context.Context, solutionversion model.SolutionVersionState) *ErrorField {
	if s.UniqueNameSolutionVersionLookupFunc == nil {
		return nil
	}
	_, err := s.UniqueNameSolutionVersionLookupFunc(ctx, solutionversion.Spec.DisplayName, solutionversion.ObjectMeta.Namespace)
	if err == nil || !utils.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           solutionversion.Spec.DisplayName,
			DetailedMessage: "solutionversion displayName must be unique",
		}
	}
	return nil
}

// Validate no instance associated with the solutionversion
// SolutionVersionInstanceLookupFunc will lookup instances with labels {"solutionversion": s.ObjectMeta.Name}
func (s *SolutionVersionValidator) ValidateNoInstanceForSolutionVersion(ctx context.Context, solutionversion model.SolutionVersionState) *ErrorField {
	if s.SolutionVersionInstanceLookupFunc == nil {
		return nil
	}
	if found, err := s.SolutionVersionInstanceLookupFunc(ctx, solutionversion.ObjectMeta.Name, solutionversion.ObjectMeta.Namespace, string(solutionversion.ObjectMeta.UID)); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           solutionversion.ObjectMeta.Name,
			DetailedMessage: "SolutionVersion has one or more associated instances. Deletion is not allowed.",
		}
	}
	return nil
}

func (s *SolutionVersionValidator) ConvertInterfaceToSolutionVersion(ref interface{}) model.SolutionVersionState {
	if ref == nil {
		return model.SolutionVersionState{
			Spec: &model.SolutionVersionSpec{},
		}
	}
	if state, ok := ref.(model.SolutionVersionState); ok {
		if state.Spec == nil {
			state.Spec = &model.SolutionVersionSpec{}
		}
		return state
	} else {
		return model.SolutionVersionState{
			Spec: &model.SolutionVersionSpec{},
		}
	}
}
