/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package validation

import (
	"context"
	"encoding/json"
	"strings"

	api_constants "github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

type ActivationValidator struct {
	// Check Campaign existence
	CampaignLookupFunc ObjectLookupFunc
}

var (
	activationMaxNameLength = 61
	activationMinNameLength = 3
)

func NewActivationValidator(campaignLookupFunc ObjectLookupFunc) ActivationValidator {
	return ActivationValidator{
		CampaignLookupFunc: campaignLookupFunc,
	}
}

// Validate Activation creation or update
// 1. Campaign exists
// 2. If initial stage is provided in the activation spec, validate it is a valid stage in the campaign
// 3. Spec is immutable for update
func (a *ActivationValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := a.ConvertInterfaceToActivation(newRef)
	old := a.ConvertInterfaceToActivation(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil {
		if err := a.ValidateCampaignAndStage(ctx, new); err != nil {
			errorFields = append(errorFields, *err)
		}

		if err := ValidateRootObjectName(new.ObjectMeta.Name, activationMinNameLength, activationMaxNameLength); err != nil {
			errorFields = append(errorFields, *err)
		}

	} else {
		// validate spec is immutable
		if equal, err := new.Spec.DeepEquals(*old.Spec); !equal {
			errorFields = append(errorFields, ErrorField{
				FieldPath:       "spec",
				Value:           new.Spec,
				DetailedMessage: "spec is immutable: " + err.Error(),
			})
		}
	}

	return errorFields
}

func (a *ActivationValidator) ValidateDelete(ctx context.Context, activation interface{}) []ErrorField {
	return []ErrorField{}
}

// Validate Campaign exists for the activation
// And if initial stage is provided in the activation spec, validate it is a valid stage in the campaign
func (a *ActivationValidator) ValidateCampaignAndStage(ctx context.Context, new model.ActivationState) *ErrorField {
	if a.CampaignLookupFunc == nil {
		return nil
	}
	campaignName := ConvertReferenceToObjectName(new.Spec.Campaign)
	Campaign, err := a.CampaignLookupFunc(ctx, campaignName, new.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.campaign",
			Value:           campaignName,
			DetailedMessage: "campaign reference must be a valid Campaign object in the same namespace",
		}
	}
	if new.Spec.Stage == "" || strings.Contains(new.Spec.Stage, "$") {
		// Skip validation if stage is not provided or is an expression
		return nil
	}

	marshalResult, err := json.Marshal(Campaign)
	if err != nil {
		return nil
	}
	var campaign model.CampaignState
	err = json.Unmarshal(marshalResult, &campaign)
	if err != nil {
		return nil
	}
	if new.ObjectMeta.Labels[api_constants.CampaignUid] != string(campaign.ObjectMeta.UID) {
		return &ErrorField{
			FieldPath:       "metadata.labels.campaignUid",
			Value:           string(campaign.ObjectMeta.UID),
			DetailedMessage: "metadata.labels.campaignUid is empty or doesn't match with the campaign.",
		}
	}
	for _, stage := range campaign.Spec.Stages {
		if stage.Name == new.Spec.Stage {
			return nil
		}
	}
	return &ErrorField{
		FieldPath:       "spec.stage",
		Value:           new.Spec.Stage,
		DetailedMessage: "spec.stage must be a valid stage in the campaign",
	}
}

func (a *ActivationValidator) ConvertInterfaceToActivation(ref interface{}) model.ActivationState {
	if ref == nil {
		return model.ActivationState{
			Spec: &model.ActivationSpec{},
		}
	}
	if state, ok := ref.(model.ActivationState); ok {
		if state.Spec == nil {
			state.Spec = &model.ActivationSpec{}
		}
		return state
	} else {
		return model.ActivationState{
			Spec: &model.ActivationSpec{},
		}
	}
}
