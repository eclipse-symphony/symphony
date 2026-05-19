/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"
	"encoding/json"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
)

type InstanceValidator struct {
	UniqueNameInstanceLookupFunc ObjectLookupFunc
	SolutionVersionLookupFunc           ObjectLookupFunc
	TargetLookupFunc             ObjectLookupFunc
}

var (
	instanceMaxNameLength = 61
	instanceMinNameLength = 1
)

func NewInstanceValidator(uniqueNameInstanceLookupFunc ObjectLookupFunc, solutionversionLookupFunc ObjectLookupFunc, targetLookupFunc ObjectLookupFunc) InstanceValidator {
	return InstanceValidator{
		UniqueNameInstanceLookupFunc: uniqueNameInstanceLookupFunc,
		SolutionVersionLookupFunc:           solutionversionLookupFunc,
		TargetLookupFunc:             targetLookupFunc,
	}
}

// Validate Instance creation or update
// 1. DisplayName is unique
// 2. SolutionVersion exists
// 3. Target exists if provided by name rather than selector
// 4. Target is valid, i.e. either name or selector is provided
func (i *InstanceValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := i.ConvertInterfaceToInstance(newRef)
	old := i.ConvertInterfaceToInstance(oldRef)

	errorFields := []ErrorField{}

	if oldRef == nil {
		if err := ValidateRootObjectName(new.ObjectMeta.Name, instanceMinNameLength, instanceMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}
	}

	if oldRef == nil || new.Spec.DisplayName != old.Spec.DisplayName {
		if err := i.ValidateUniqueName(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	if oldRef == nil || new.Spec.SolutionVersion != old.Spec.SolutionVersion {
		if err := i.ValidateSolutionVersionExist(ctx, new); err != nil {
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

// Validate SolutionVersion exists for instance
func (i *InstanceValidator) ValidateSolutionVersionExist(ctx context.Context, c model.InstanceState) *ErrorField {
	if i.SolutionVersionLookupFunc == nil {
		return nil
	}
	solutionversion, err := i.SolutionVersionLookupFunc(ctx, ConvertReferenceToObjectName(c.Spec.SolutionVersion), c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.solutionversion",
			Value:           c.Spec.SolutionVersion,
			DetailedMessage: "solutionversion does not exist",
		}
	}
	marshalResult, err := json.Marshal(solutionversion)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.solutionversion",
			Value:           c.Spec.SolutionVersion,
			DetailedMessage: "solutionversion can not be parsed.",
		}
	}
	var solutionversionState model.SolutionVersionState
	err = json.Unmarshal(marshalResult, &solutionversionState)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.solutionversion",
			Value:           c.Spec.SolutionVersion,
			DetailedMessage: "solutionversion can not be parsed.",
		}
	}

	if c.ObjectMeta.Labels[api_constants.SolutionVersionUid] != string(solutionversionState.ObjectMeta.UID) {
		return &ErrorField{
			FieldPath:       "metadata.labels.solutionversionUid",
			Value:           string(solutionversionState.ObjectMeta.UID),
			DetailedMessage: "metadata.labels.solutionversionUid is empty or doesn't match with the solutionversion.",
		}
	}

	return nil
}

// Validate Target exists for instance if target name is provided
func (i *InstanceValidator) ValidateTargetExist(ctx context.Context, c model.InstanceState) *ErrorField {
	if i.TargetLookupFunc == nil {
		return nil
	}
	target, err := i.TargetLookupFunc(ctx, ConvertReferenceToObjectName(c.Spec.Target.Name), c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.target.name",
			Value:           c.Spec.Target.Name,
			DetailedMessage: "target does not exist",
		}

	}

	marshalResult, err := json.Marshal(target)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.target",
			Value:           c.Spec.Target.Name,
			DetailedMessage: "target can not be parsed.",
		}
	}
	var targetState model.TargetState
	err = json.Unmarshal(marshalResult, &targetState)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.target",
			Value:           c.Spec.Target.Name,
			DetailedMessage: "target can not be parsed.",
		}
	}

	if c.ObjectMeta.Labels[api_constants.TargetUid] != string(targetState.ObjectMeta.UID) {
		return &ErrorField{
			FieldPath:       "metadata.labels.targetUid",
			Value:           string(targetState.ObjectMeta.UID),
			DetailedMessage: "metadata.labels.targetUid is empty or doesn't match with the target.",
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
