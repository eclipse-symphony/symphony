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
	SolutionVersionState struct {
		ObjectMeta ObjectMeta    `json:"metadata,omitempty"`
		Spec       *SolutionVersionSpec `json:"spec,omitempty"`
	}

	SolutionVersionSpec struct {
		DisplayName  string            `json:"displayName,omitempty"`
		Metadata     map[string]string `json:"metadata,omitempty"`
		Components   []ComponentSpec   `json:"components,omitempty"`
		Version      string            `json:"version,omitempty"`
		RootResource string            `json:"rootResource,omitempty"`
	}
)

func (c SolutionVersionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionVersionSpec)
	if !ok {
		return false, errors.New("parameter is not a SolutionVersionSpec type")
	}

	if c.DisplayName != otherC.DisplayName {
		return false, nil
	}

	if !StringMapsEqual(c.Metadata, otherC.Metadata, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Components, otherC.Components) {
		return false, nil
	}

	return true, nil
}

func (c SolutionVersionState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SolutionVersionState)
	if !ok {
		return false, errors.New("parameter is not a SolutionVersionState type")
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
