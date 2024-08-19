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

// Check Instance associated with the Solution
var TargetInstanceLookupFunc LinkedObjectLookupFunc

var UniqueNameTargetLookupFunc ObjectLookupFunc

func ValidateCreateOrUpdateTarget(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := ConvertInterfaceToTarget(newRef)
	old := ConvertInterfaceToTarget(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := ValidateTargetUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	return errorFields
}

func ValidateDeleteTarget(ctx context.Context, newRef interface{}) []ErrorField {
	t := ConvertInterfaceToTarget(newRef)
	errorFields := []ErrorField{}
	if err := ValidateNoInstanceForTarget(ctx, t); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}
func ValidateTargetUniqueName(ctx context.Context, t model.TargetState) *ErrorField {
	if UniqueNameTargetLookupFunc == nil {
		return nil
	}
	_, err := UniqueNameTargetLookupFunc(ctx, t.Spec.DisplayName, t.ObjectMeta.Namespace)
	if err == nil || !v1alpha2.IsNotFound(err) {
		return &ErrorField{
			FieldPath:       "spec.displayName",
			Value:           t.Spec.DisplayName,
			DetailedMessage: "target displayName must be unique",
		}
	}
	return nil
}
func ValidateNoInstanceForTarget(ctx context.Context, t model.TargetState) *ErrorField {
	if TargetInstanceLookupFunc == nil {
		return nil
	}
	if found, err := TargetInstanceLookupFunc(ctx, t.ObjectMeta.Name, t.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           t.ObjectMeta.Name,
			DetailedMessage: "Target has one or more associated instances. Deletion is not allowed.",
		}
	}
	return nil
}

func ConvertInterfaceToTarget(ref interface{}) model.TargetState {
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
