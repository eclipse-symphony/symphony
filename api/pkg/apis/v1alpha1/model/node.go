/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

// +kubebuilder:object:generate=true
type NodeSpec struct {
	Id             string            `json:"id"`
	NodeType       string            `json:"type"`
	Name           string            `json:"name"`
	Configurations map[string]string `json:"configurations,omitempty"`
	Inputs         []RouteSpec       `json:"inputs,omitempty"`
	Outputs        []RouteSpec       `json:"outputs,omitempty"`
	Model          string            `json:"model,omitempty"`
}

func (n NodeSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(NodeSpec)
	if !ok {
		return false, nil
	}
	if n.Id != otherSpec.Id {
		return false, nil
	}
	if n.NodeType != otherSpec.NodeType {
		return false, nil
	}
	if n.Name != otherSpec.Name {
		return false, nil
	}
	if n.Model != otherSpec.Model {
		return false, nil
	}
	if !StringMapsEqual(n.Configurations, otherSpec.Configurations, nil) {
		return false, nil
	}
	if !SlicesEqual(n.Inputs, otherSpec.Inputs) {
		return false, nil
	}
	if !SlicesEqual(n.Outputs, otherSpec.Outputs) {
		return false, nil
	}
	return true, nil
}
