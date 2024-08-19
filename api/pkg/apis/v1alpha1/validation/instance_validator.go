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

var UniqueNameInstanceLookupFunc ObjectLookupFunc
var SolutionLookupFunc ObjectLookupFunc
var TargetLookupFunc ObjectLookupFunc

func ValidateCreateOrUpdateInstance(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := ConvertInterfaceToInstance(newRef)
	old := ConvertInterfaceToInstance(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := ValidateUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil || new.Spec.Solution != old.Spec.Solution {
		if err := ValidateSolutionExist(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if (oldRef == nil || new.Spec.Target.Name != old.Spec.Target.Name) && new.Spec.Target.Name != "" {
		if err := ValidateTargetExist(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if err := ValidateTargetValid(new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

func ValidateDeleteInstance(ctx context.Context, new interface{}) []ErrorField {
	return []ErrorField{}
}

func ValidateUniqueName(ctx context.Context, c model.InstanceState) *ErrorField {
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

func ValidateSolutionExist(ctx context.Context, c model.InstanceState) *ErrorField {
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

func ValidateTargetExist(ctx context.Context, c model.InstanceState) *ErrorField {
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

func ValidateTargetValid(c model.InstanceState) *ErrorField {
	if c.Spec.Target.Name == "" && (c.Spec.Target.Selector == nil || len(c.Spec.Target.Selector) == 0) {
		return &ErrorField{
			FieldPath:       "spec.target",
			Value:           c.Spec.Target,
			DetailedMessage: "target must have either name or valid selector",
		}
	}
	return nil
}

func ConvertInterfaceToInstance(ref interface{}) model.InstanceState {
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
