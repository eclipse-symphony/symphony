/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type CampaignState struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Spec       *CampaignSpec   `json:"spec,omitempty"`
	Status     *CampaignStatus `json:"status,omitempty"`
}

type CampaignSpec struct {
}

type CampaignStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CampaignSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c CampaignState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CampaignState)
	if !ok {
		return false, errors.New("parameter is not a CampaignState type")
	}

	equal, err := c.ObjectMeta.DeepEquals(otherC.ObjectMeta)
	if err != nil || !equal {
		return equal, err
	}

	equal, err = c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}

	return true, nil
}
