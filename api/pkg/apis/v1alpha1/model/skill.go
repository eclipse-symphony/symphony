/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
	otherSkillSpec, ok := other.(*SkillSpec)
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
