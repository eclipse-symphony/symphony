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

// Check Campaign Container existence
var CampaignContainerLookupFunc ObjectLookupFunc

// Check Activations associated with the Campaign
var CampaignActivationsLookupFunc LinkedObjectLookupFunc

// Validate Campaign creation or update
// 1. First stage is valid
// 2. Stages in the list are
// 3. campaign name and rootResource is valid. And rootResource is immutable
// 4. Update is not allow when there are running activations
func ValidateCreateOrUpdateCampaign(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := ConvertInterfaceToCampaign(newRef)
	old := ConvertInterfaceToCampaign(oldRef)

	errorFields := []ErrorField{}
	// validate first stage if it is changed
	if oldRef == nil || new.Spec.FirstStage != old.Spec.FirstStage {
		if err := ValidateFirstStage(new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	// validate StageSelector
	if err := ValidateStages(new); err != nil {
		errorFields = append(errorFields, *err)
	}
	if oldRef == nil {
		// validate create specific fields
		if err := ValidateObjectName(new.ObjectMeta.Name, new.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := ValidateRootResource(ctx, new.ObjectMeta, new.Spec.RootResource, CampaignContainerLookupFunc); err != nil {
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
		if err := ValidateRunningActivation(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	return errorFields
}

// Validate campaign deletion
// 1. No running activations
func ValidateDeleteCampaign(ctx context.Context, newRef interface{}) []ErrorField {
	new := ConvertInterfaceToCampaign(newRef)
	errorFields := []ErrorField{}
	// validate no running activations
	if err := ValidateRunningActivation(ctx, new); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

// Validate First stage of the campaign
// 1. If stages is empty, firstStage must be empty
// 2. If stages is not empty, firstStage must be one of the stages in the list
func ValidateFirstStage(c model.CampaignState) *ErrorField {
	isValid := false
	if c.Spec.FirstStage == "" {
		if c.Spec.Stages == nil || len(c.Spec.Stages) == 0 {
			isValid = true
		}
	}
	for _, stage := range c.Spec.Stages {
		if stage.Name == c.Spec.FirstStage {
			isValid = true
		}
	}
	if !isValid {
		return &ErrorField{
			FieldPath:       "spec.firstStage",
			Value:           c.Spec.FirstStage,
			DetailedMessage: "firstStage must be one of the stages in the stages list",
		}
	} else {
		return nil
	}
}

// Validate stageSelector of stages should always be one of the stages in the stages list
func ValidateStages(c model.CampaignState) *ErrorField {
	stages := make(map[string]struct{}, 0)
	for _, stage := range c.Spec.Stages {
		stages[stage.Name] = struct{}{}
	}
	for _, stage := range c.Spec.Stages {
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
// CampaignActivationsLookupFunc will look up activations with label {"campaign" : c.ObjectMeta.Name}
func ValidateRunningActivation(ctx context.Context, c model.CampaignState) *ErrorField {
	if CampaignActivationsLookupFunc == nil {
		return nil
	}
	if found, err := CampaignActivationsLookupFunc(ctx, c.ObjectMeta.Name, c.ObjectMeta.Namespace); err != nil || found {
		return &ErrorField{
			FieldPath:       "metadata.name",
			Value:           c.ObjectMeta.Name,
			DetailedMessage: "Campaign has one or more running activations. Update or Deletion is not allowed",
		}
	}
	return nil
}
func ConvertInterfaceToCampaign(ref interface{}) model.CampaignState {
	if ref == nil {
		return model.CampaignState{
			Spec: &model.CampaignSpec{},
		}
	}
	if state, ok := ref.(model.CampaignState); ok {
		if state.Spec == nil {
			state.Spec = &model.CampaignSpec{}
		}
		return state
	} else {
		return model.CampaignState{
			Spec: &model.CampaignSpec{},
		}
	}
}
