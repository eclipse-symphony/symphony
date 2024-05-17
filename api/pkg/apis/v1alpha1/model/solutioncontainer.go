/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type SolutionContainerState struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Spec       *SolutionContainerSpec   `json:"spec,omitempty"`
	Status     *SolutionContainerStatus `json:"status,omitempty"`
}

type SolutionContainerSpec struct {
}

type SolutionContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c SolutionContainerSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c SolutionContainerState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionContainerState)
	if !ok {
		return false, errors.New("parameter is not a SolutionContainerState type")
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
