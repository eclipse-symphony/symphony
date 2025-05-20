/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type CampaignContainerState struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Spec       *CampaignContainerSpec   `json:"spec,omitempty"`
	Status     *CampaignContainerStatus `json:"status,omitempty"`
}

type CampaignContainerSpec struct {
}

type CampaignContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CampaignContainerSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c CampaignContainerState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CampaignContainerState)
	if !ok {
		return false, errors.New("parameter is not a CampaignContainerState type")
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
