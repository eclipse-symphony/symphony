/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
	"reflect"
)

type ComponentSpec struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Parameters   map[string]string      `json:"parameters,omitempty"`
	Routes       []RouteSpec            `json:"routes,omitempty"`
	Constraints  string                 `json:"constraints,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Skills       []string               `json:"skills,omitempty"`
	Sidecars     []SidecarSpec          `json:"sidecars,omitempty"`
}

func (c ComponentSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	otherC, ok := other.(ComponentSpec)
	if !ok {
		return false, errors.New("parameter is not a ComponentSpec type")
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if !reflect.DeepEqual(c.Properties, otherC.Properties) {
		return false, nil
	}
	if !StringMapsEqual(c.Metadata, otherC.Metadata, []string{"SYMPHONY_AGENT_ADDRESS"}) {
		return false, nil
	}

	if !SlicesEqual(c.Routes, otherC.Routes) {
		return false, nil
	}

	if !SlicesEqual(c.Sidecars, otherC.Sidecars) {
		return false, nil
	}
	// if c.Constraints != otherC.Constraints {	Can't compare constraints as components from actual envrionments don't have constraints
	// 	return false, nil
	// }
	return true, nil
}
