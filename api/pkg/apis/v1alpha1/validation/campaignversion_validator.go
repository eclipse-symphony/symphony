/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

var (
	campaignversionMaxNameLength = 61
	campaignversionMinNameLength = 1
)

type CampaignVersionValidator struct {
	// Check CampaignVersion Container existence
	CampaignLookupFunc ObjectLookupFunc

	// Check Activations associated with the CampaignVersion
	CampaignVersionActivationsLookupFunc LinkedObjectLookupFunc
}

func NewCampaignVersionValidator(campaignversionContainerLookupFunc ObjectLookupFunc, campaignversionActivationsLookupFunc LinkedObjectLookupFunc) CampaignVersionValidator {
	return CampaignVersionValidator{
		CampaignLookupFunc:   campaignversionContainerLookupFunc,
		CampaignVersionActivationsLookupFunc: campaignversionActivationsLookupFunc,
	}
}

// Validate CampaignVersion creation or update
// 1. First stage is valid
// 2. Stages in the list are
// 3. campaignversion name and rootResource is valid. And rootResource is immutable
// 4. Update is not allow when there are running activations
func (c *CampaignVersionValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := c.ConvertInterfaceToCampaignVersion(newRef)
	old := c.ConvertInterfaceToCampaignVersion(oldRef)

	errorFields := []ErrorField{}
	// validate first stage if it is changed
	if oldRef == nil || new.Spec.FirstStage != old.Spec.FirstStage {
		if err := c.ValidateFirstStage(new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	// validate StageSelector
	if err := c.ValidateStages(new); err != nil {
		errorFields = append(errorFields, *err)
	}
	if oldRef == nil {
		// validate create specific fields
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource, campaignversionMinNameLength, campaignversionMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, c.CampaignLookupFunc); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate update specific fields
		if new.Spec.RootResource != old.Spec.RootResource {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.rootResource",
				Value:           new.Spec.RootResource,
				DetailedMessage: "rootResource is immutable",
			})
		}
		if err := c.ValidateRunningActivation(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	return errorFields
}

// Validate campaignversion deletion
// 1. No running activations
func (c *CampaignVersionValidator) ValidateDelete(ctx context.Context, newRef interface{}) []ErrorField {
	new := c.ConvertInterfaceToCampaignVersion(newRef)
	errorFields := []ErrorField{}
	// validate no running activations
	if err := c.ValidateRunningActivation(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate First stage of the campaignversion
// 1. If stages is empty, firstStage must be empty
// 2. If stages is not empty, firstStage must be one of the stages in the list
func (c *CampaignVersionValidator) ValidateFirstStage(campaignversion model.CampaignVersionState) *ErrorField {
	isValid := false
	if campaignversion.Spec.FirstStage == "" {
		if campaignversion.Spec.Stages == nil || len(campaignversion.Spec.Stages) == 0 {
			isValid = true
		}
	}
	for _, stage := range campaignversion.Spec.Stages {
		if stage.Name == campaignversion.Spec.FirstStage {
			isValid = true
		}
	}
	if !isValid {
		return &ErrorField{
			FieldPath:       "spec.firstStage",
			Value:           campaignversion.Spec.FirstStage,
			DetailedMessage: "firstStage must be one of the stages in the stages list",
		}
	} else {
		return nil
	}
}

// Validate stageSelector of stages should always be one of the stages in the stages list
func (c *CampaignVersionValidator) ValidateStages(campaignversion model.CampaignVersionState) *ErrorField {
	stages := make(map[string]struct{}, 0)
	for _, stage := range campaignversion.Spec.Stages {
		stages[stage.Name] = struct{}{}
	}
	for _, stage := range campaignversion.Spec.Stages {
		if !strings.Contains(stage.StageSelector, "$") && stage.StageSelector != "" {
			if _, ok := stages[stage.StageSelector]; !ok {
				return &ErrorField{
					FieldPath:       fmt.Sprintf("spec.stages.%s.stageSelector", stage.Name),
					Value:           stage.StageSelector,
					DetailedMessage: "stageSelector must be one of the stages in the stages list",
				}
			}
		}
	}
	return nil
}

// Validate NO running activations
// CampaignVersionActivationsLookupFunc will look up activations with label {"campaignversion" : c.ObjectMeta.Name}
func (c *CampaignVersionValidator) ValidateRunningActivation(ctx context.Context, campaignversion model.CampaignVersionState) *ErrorField {
	if c.CampaignVersionActivationsLookupFunc == nil {
		return nil
	}
	if found, err := c.CampaignVersionActivationsLookupFunc(ctx, campaignversion.ObjectMeta.Name, campaignversion.ObjectMeta.Namespace, string(campaignversion.ObjectMeta.UID)); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           campaignversion.ObjectMeta.Name,
			DetailedMessage: "CampaignVersion has one or more running activations. Update or Deletion is not allowed",
		}
	}
	return nil
}
func (c *CampaignVersionValidator) ConvertInterfaceToCampaignVersion(ref interface{}) model.CampaignVersionState {
	if ref == nil {
		return model.CampaignVersionState{
			Spec: &model.CampaignVersionSpec{},
		}
	}
	if state, ok := ref.(model.CampaignVersionState); ok {
		if state.Spec == nil {
			state.Spec = &model.CampaignVersionSpec{}
		}
		return state
	} else {
		return model.CampaignVersionState{
			Spec: &model.CampaignVersionSpec{},
		}
	}
}
