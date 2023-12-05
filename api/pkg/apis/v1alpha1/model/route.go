/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

// +kubebuilder:object:generate=true
type RouteSpec struct {
	Route      string            `json:"route"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

func (c RouteSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	otherC, ok := other.(RouteSpec)
	if !ok {
		return false, errors.New("parameter is not a RouteSpec type")
	}

	if c.Route != otherC.Route {
		return false, nil
	}

	if c.Type != otherC.Type {
		return false, nil
	}

	if !StringMapsEqual(c.Properties, otherC.Properties, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Filters, otherC.Filters) {
		return false, nil
	}

	return true, nil
}
