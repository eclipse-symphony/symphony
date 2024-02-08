/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
)

type (
	SolutionState struct {
		Id       string            `json:"id"`
		Scope    string            `json:"scope"`
		Spec     *SolutionSpec     `json:"spec,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	SolutionSpec struct {
		DisplayName string          `json:"displayName,omitempty"`
		Components  []ComponentSpec `json:"components,omitempty"`
	}
)

func (c SolutionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionSpec)
	if !ok {
		return false, errors.New("parameter is not a SolutionSpec type")
	}

	if c.DisplayName != otherC.DisplayName {
		return false, nil
	}

	if !SlicesEqual(c.Components, otherC.Components) {
		return false, nil
	}

	return true, nil
}

func (c SolutionState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionState)
	if !ok {
		return false, errors.New("parameter is not a SolutionState type")
	}

	if c.Id != otherC.Id {
		return false, nil
	}

	if c.Scope != otherC.Scope {
		return false, nil
	}

	if !StringMapsEqual(c.Metadata, otherC.Metadata, nil) {
		return false, nil
	}

	equal, err := c.Spec.DeepEquals(otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}

	return true, nil
}
