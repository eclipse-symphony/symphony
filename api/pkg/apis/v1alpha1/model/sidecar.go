/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "reflect"

type SidecarSpec struct {
	Name       string                 `json:"name,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

func (s SidecarSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherS, ok := other.(SidecarSpec)
	if !ok {
		return false, nil
	}

	if s.Name != otherS.Name {
		return false, nil
	}

	if s.Type != otherS.Type {
		return false, nil
	}

	if !reflect.DeepEqual(s.Properties, otherS.Properties) {
		return false, nil
	}

	return true, nil
}
