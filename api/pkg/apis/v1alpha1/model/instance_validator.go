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

var UniqueNameInstanceLookupFunc ObjectLookupFunc
var SolutionLookupFunc ObjectLookupFunc
var TargetLookupFunc ObjectLookupFunc

func (c InstanceState) ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField {
	var oldInstance InstanceState
	if old != nil {
		var ok bool
		oldInstance, ok = old.(InstanceState)
		if !ok {
			old = nil
		}
	}
	if c.Spec == nil {
		c.Spec = &InstanceSpec{}
	}

	errorFields := []ErrorField{}
	if old == nil || c.Spec.DisplayName != oldInstance.Spec.DisplayName {
		if err := c.ValidateUniqueName(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if old == nil || c.Spec.Solution != oldInstance.Spec.Solution {
		if err := c.ValidateSolutionExist(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if (old == nil || c.Spec.Target.Name != oldInstance.Spec.Target.Name) && c.Spec.Target.Name != "" {
		if err := c.ValidateTargetExist(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if err := c.ValidateTargetValid(); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

func (c InstanceState) ValidateDelete(ctx context.Context) []ErrorField {
	return []ErrorField{}
}

func (c InstanceState) ValidateUniqueName(ctx context.Context) *ErrorField {
	if UniqueNameInstanceLookupFunc == nil {
		return nil
	}
	_, err := UniqueNameInstanceLookupFunc(ctx, c.Spec.DisplayName, c.ObjectMeta.Namespace)
	if err == nil || !v1alpha2.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           c.Spec.DisplayName,
			DetailedMessage: "instance displayName must be unique",
		}
	}
	return nil
}

func (c InstanceState) ValidateSolutionExist(ctx context.Context) *ErrorField {
	if SolutionLookupFunc == nil {
		return nil
	}
	_, err := SolutionLookupFunc(ctx, ConvertReferenceToObjectName(c.Spec.Solution), c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.solution",
			Value:           c.Spec.Solution,
			DetailedMessage: "solution does not exist",
		}
	}
	return nil
}

func (c InstanceState) ValidateTargetExist(ctx context.Context) *ErrorField {
	if TargetLookupFunc == nil {
		return nil
	}
	_, err := TargetLookupFunc(ctx, c.Spec.Target.Name, c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.target.name",
			Value:           c.Spec.Target.Name,
			DetailedMessage: "target does not exist",
		}

	}
	return nil
}

func (c InstanceState) ValidateTargetValid() *ErrorField {
	if c.Spec.Target.Name == "" && (c.Spec.Target.Selector == nil || len(c.Spec.Target.Selector) == 0) {
		return &ErrorField{
			FieldPath:       "spec.target",
			Value:           c.Spec.Target,
			DetailedMessage: "target must have either name or valid selector",
		}
	}
	return nil
}
