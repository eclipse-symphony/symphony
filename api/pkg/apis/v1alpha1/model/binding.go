/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

// +kubebuilder:object:generate=true
type BindingSpec struct {
	Role     string            `json:"role"`
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config,omitempty"`
}

func (c BindingSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(BindingSpec)
	if !ok {
		return false, errors.New("parameter is not a BindingSpec type")
	}

	if c.Role != otherC.Role {
		return false, nil
	}

	if c.Provider != otherC.Provider {
		return false, nil
	}

	if !StringMapsEqual(c.Config, otherC.Config, nil) {
		return false, nil
	}

	return true, nil
}
