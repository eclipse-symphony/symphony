/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type SolutionState struct {
	ObjectMeta ObjectMeta               `json:"metadata,omitempty"`
	Spec       *SolutionSpec   `json:"spec,omitempty"`
	Status     *SolutionStatus `json:"status,omitempty"`
}

type SolutionSpec struct {
}

type SolutionStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c SolutionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	return true, nil
}

func (c SolutionState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionState)
	if !ok {
		return false, errors.New("parameter is not a SolutionState type")
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
