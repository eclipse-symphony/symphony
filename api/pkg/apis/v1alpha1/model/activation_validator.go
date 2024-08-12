/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"context"
	"encoding/json"
	"strings"
)

// Check Campaign existence
var CampaignLookupFunc ObjectLookupFunc

func (c ActivationState) ValidateCreateOrUpdate(ctx context.Context, old IValidation) []ErrorField {
	var oldActivation ActivationState
	if old != nil {
		var ok bool
		oldActivation, ok = old.(ActivationState)
		if !ok {
			old = nil
		}
	}
	if c.Spec == nil {
		c.Spec = &ActivationSpec{}
	}

	errorFields := []ErrorField{}
	if old == nil {
		if err := c.ValidateCampaignAndStage(ctx); err != nil {
			errorFields = append(errorFields, *err)
		}
	} else {
		// validate spec is immutable
		if equal, err := c.Spec.DeepEquals(*oldActivation.Spec); !equal {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec",
				Value:           c.Spec,
				DetailedMessage: "spec is immutable: " + err.Error(),
			})
		}
	}

	return errorFields
}

func (c ActivationState) ValidateDelete(ctx context.Context) []ErrorField {
	return []ErrorField{}
}

func (c ActivationState) ValidateCampaignAndStage(ctx context.Context) *ErrorField {
	if CampaignLookupFunc == nil {
		return nil
	}
	campaignName := ConvertReferenceToObjectName(c.Spec.Campaign)
	Campaign, err := CampaignLookupFunc(ctx, campaignName, c.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.campaign",
			Value:           campaignName,
			DetailedMessage: "campaign reference must be a valid Campaign object in the same namespace",
		}
	}
	if c.Spec.Stage == "" || strings.Contains(c.Spec.Stage, "$") {
		// Skip validation if stage is not provided or is an expression
		return nil
	}

	marshalResult, err := json.Marshal(Campaign)
	if err != nil {
		return nil
	}
	var campaign CampaignState
	err = json.Unmarshal(marshalResult, &campaign)
	if err != nil {
		return nil
	}
	for _, stage := range campaign.Spec.Stages {
		if stage.Name == c.Spec.Stage {
			return nil
		}
	}
	return &ErrorField{
		FieldPath:       "spec.stage",
		Value:           c.Spec.Stage,
		DetailedMessage: "spec.stage must be a valid stage in the campaign",
	}
}
