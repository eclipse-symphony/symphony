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

// Check Instance associated with the Solution
var TargetInstanceLookupFunc LinkedObjectLookupFunc

var UniqueNameTargetLookupFunc ObjectLookupFunc

func (t TargetState) ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField {
	var oldTarget TargetState
	if old != nil {
		var ok bool
		oldTarget, ok = old.(TargetState)
		if !ok {
			old = nil
		}
	}
	if t.Spec == nil {
		t.Spec = &TargetSpec{}
	}

	errorFields := []ErrorField{}
	if old == nil || t.Spec.DisplayName != oldTarget.Spec.DisplayName {
		if err := t.ValidateUniqueName(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	return errorFields
}

func (t TargetState) ValidateDelete(ctx context.Context) []ErrorField {
	errorFields := []ErrorField{}
	if err := t.ValidateNoInstance(ctx); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}
func (t *TargetState) ValidateUniqueName(ctx context.Context) *ErrorField {
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
func (t *TargetState) ValidateNoInstance(ctx context.Context) *ErrorField {
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
