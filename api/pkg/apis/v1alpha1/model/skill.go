/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

type SkillState struct {
	Id   string     `json:"id"`
	Spec *SkillSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true
type SkillSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Nodes       []NodeSpec        `json:"nodes"`
	Properties  map[string]string `json:"properties,omitempty"`
	Bindings    []BindingSpec     `json:"bindings,omitempty"`
	Edges       []EdgeSpec        `json:"edges"`
}

// +kubebuilder:object:generate=true
type SkillPackageSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Skill       string            `json:"skill"`
	Properties  map[string]string `json:"properties,omitempty"`
	Constraints string            `json:"constraints,omitempty"`
	Routes      []RouteSpec       `json:"routes,omitempty"`
}

func (c SkillSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSkillSpec, ok := other.(SkillSpec)
	if !ok {
		return false, nil
	}
	if c.DisplayName != otherSkillSpec.DisplayName {
		return false, nil
	}
	if !StringMapsEqual(c.Parameters, otherSkillSpec.Parameters, nil) {
		return false, nil
	}
	if !SlicesEqual(c.Nodes, otherSkillSpec.Nodes) {
		return false, nil
	}
	if !StringMapsEqual(c.Properties, otherSkillSpec.Properties, nil) {
		return false, nil
	}
	if !SlicesEqual(c.Bindings, otherSkillSpec.Bindings) {
		return false, nil
	}

	if !SlicesEqual(c.Edges, otherSkillSpec.Edges) {
		return false, nil
	}
	return true, nil
}
