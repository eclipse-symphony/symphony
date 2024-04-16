/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

type SkillState struct {
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`
	Spec       *SkillSpec `json:"spec,omitempty"`
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

func (c SkillState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(SkillState)
	if !ok {
		return false, errors.New("parameter is not a SkillState type")
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
