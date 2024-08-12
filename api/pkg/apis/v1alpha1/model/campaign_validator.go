/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"context"
	"fmt"
	"strings"
)

// Check Campaign Container existence
var CampaignContainerLookupFunc ObjectLookupFunc

// Check Activations associated with the Campaign
var CampaignActivationsLookupFunc LinkedObjectLookupFunc

func (c CampaignState) ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField {
	var oldCampaign CampaignState
	if old != nil {
		var ok bool
		oldCampaign, ok = old.(CampaignState)
		if !ok {
			old = nil
		}
	}

	if c.Spec == nil {
		c.Spec = &CampaignSpec{}
	}

	errorFields := []ErrorField{}
	// validate first stage if it is changed
	if old == nil || c.Spec.FirstStage != oldCampaign.Spec.FirstStage {
		if err := c.ValidateFirstStage(); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	// validate StageSelector
	if err := c.ValidateStages(); err != nil {
		errorFields = append(errorFields, *err)
	}
	// validate rootResource is not changed in update
	if old == nil {
		// validate create specific fields
		if err := ValidateObjectName(c.ObjectMeta.Name, c.Spec.RootResource); err != nil {
			errorFields = append(errorFields, *err)
		}
		// validate rootResource
		if err := c.ObjectMeta.ValidateRootResource(ctx, c.Spec.RootResource, CampaignContainerLookupFunc); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate update specific fields
		if c.Spec.RootResource != oldCampaign.Spec.RootResource {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec.rootResource",
				Value:           c.Spec.RootResource,
				DetailedMessage: "rootResource is immutable",
			})
		}
		if err := c.ValidateRunningActivation(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	}
	return errorFields
}

func (c CampaignState) ValidateDelete(ctx context.Context) []ErrorField {
	errorFields := []ErrorField{}
	// validate no running activations
	if err := c.ValidateRunningActivation(ctx); err != nil {
		errorFields = append(errorFields, *err)
	}
	return errorFields
}

func (c CampaignState) ValidateFirstStage() *ErrorField {
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

func (c CampaignState) ValidateStages() *ErrorField {
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

func (c CampaignState) ValidateRunningActivation(ctx context.Context) *ErrorField {
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
