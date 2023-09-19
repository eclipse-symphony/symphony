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

import "errors"

type (
	// TargetState defines the current state of the target
	TargetState struct {
		Id       string            `json:"id"`
		Metadata map[string]string `json:"metadata,omitempty"`
		Status   map[string]string `json:"status,omitempty"`
		Spec     *TargetSpec       `json:"spec,omitempty"`
	}

	// TargetSpec defines the spec property of the TargetState
	TargetSpec struct {
		DisplayName   string            `json:"displayName,omitempty"`
		Scope         string            `json:"scope,omitempty"`
		Metadata      map[string]string `json:"metadata,omitempty"`
		Properties    map[string]string `json:"properties,omitempty"`
		Components    []ComponentSpec   `json:"components,omitempty"`
		Constraints   string            `json:"constraints,omitempty"`
		Topologies    []TopologySpec    `json:"topologies,omitempty"`
		ForceRedeploy bool              `json:"forceRedeploy,omitempty"`
		Generation    string            `json:"generation,omitempty"`
		// Defines the version of a particular resource
		Version string `json:"version,omitempty"`
	}
)

func (c TargetSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(TargetSpec)
	if !ok {
		return false, errors.New("parameter is not a TargetSpec type")
	}

	if c.DisplayName != otherC.DisplayName {
		return false, nil
	}

	if c.Scope != otherC.Scope {
		return false, nil
	}

	if !StringMapsEqual(c.Metadata, otherC.Metadata, nil) {
		return false, nil
	}

	if !StringMapsEqual(c.Properties, otherC.Properties, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Components, otherC.Components) {
		return false, nil
	}

	if c.Constraints != otherC.Constraints {
		return false, nil
	}

	if !SlicesEqual(c.Topologies, otherC.Topologies) {
		return false, nil
	}

	if c.ForceRedeploy != otherC.ForceRedeploy {
		return false, nil
	}

	return true, nil
}
