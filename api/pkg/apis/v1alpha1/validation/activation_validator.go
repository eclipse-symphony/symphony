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
	// Check CampaignVersion existence
	CampaignVersionLookupFunc ObjectLookupFunc
}

var (
	activationMaxNameLength = 61
	activationMinNameLength = 1
)

func NewActivationValidator(campaignversionLookupFunc ObjectLookupFunc) ActivationValidator {
	return ActivationValidator{
		CampaignVersionLookupFunc: campaignversionLookupFunc,
	}
}

// Validate Activation creation or update
// 1. CampaignVersion exists
// 2. If initial stage is provided in the activation spec, validate it is a valid stage in the campaignversion
// 3. Spec is immutable for update
func (a *ActivationValidator) ValidateCreateOrUpdate(ctx context.Context, newRef interface{}, oldRef interface{}) []ErrorField {
	new := a.ConvertInterfaceToActivation(newRef)
	old := a.ConvertInterfaceToActivation(oldRef)

	errorFields := []ErrorField{}
	if oldRef == nil {
		if err := a.ValidateCampaignVersionAndStage(ctx, new); err != nil {
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

// Validate CampaignVersion exists for the activation
// And if initial stage is provided in the activation spec, validate it is a valid stage in the campaignversion
func (a *ActivationValidator) ValidateCampaignVersionAndStage(ctx context.Context, new model.ActivationState) *ErrorField {
	if a.CampaignVersionLookupFunc == nil {
		return nil
	}
	campaignversionName := ConvertReferenceToObjectName(new.Spec.CampaignVersion)
	CampaignVersion, err := a.CampaignVersionLookupFunc(ctx, campaignversionName, new.ObjectMeta.Namespace)
	if err != nil {
		return &ErrorField{
			FieldPath:       "spec.campaignversion",
			Value:           campaignversionName,
			DetailedMessage: "campaignversion reference must be a valid CampaignVersion object in the same namespace",
		}
	}
	if new.Spec.Stage == "" || strings.Contains(new.Spec.Stage, "$") {
		// Skip validation if stage is not provided or is an expression
		return nil
	}

	marshalResult, err := json.Marshal(CampaignVersion)
	if err != nil {
		return nil
	}
	var campaignversion model.CampaignVersionState
	err = json.Unmarshal(marshalResult, &campaignversion)
	if err != nil {
		return nil
	}
	if new.ObjectMeta.Labels[api_constants.CampaignVersionUid] != string(campaignversion.ObjectMeta.UID) {
		return &ErrorField{
			FieldPath:       "metadata.labels.campaignversionUid",
			Value:           string(campaignversion.ObjectMeta.UID),
			DetailedMessage: "metadata.labels.campaignversionUid is empty or doesn't match with the campaignversion.",
		}
	}
	for _, stage := range campaignversion.Spec.Stages {
		if stage.Name == new.Spec.Stage {
			return nil
		}
	}
	return &ErrorField{
		FieldPath:       "spec.stage",
		Value:           new.Spec.Stage,
		DetailedMessage: "spec.stage must be a valid stage in the campaignversion",
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
