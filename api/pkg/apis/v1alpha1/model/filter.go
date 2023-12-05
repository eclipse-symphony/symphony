/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

// +kubebuilder:object:generate=true
type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (c FilterSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	otherC, ok := other.(FilterSpec)
	if !ok {
		return false, errors.New("parameter is not a FilterSpec type")
	}

	if c.Direction != otherC.Direction {
		return false, nil
	}

	if c.Type != otherC.Type {
		return false, nil
	}

	if !StringMapsEqual(c.Parameters, otherC.Parameters, nil) {
		return false, nil
	}

	return true, nil
}
