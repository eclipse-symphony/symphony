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

	// InstanceState defines the current state of the instance
	InstanceState struct {
		Id     string            `json:"id"`
		Scope  string            `json:"scope"`
		Spec   *InstanceSpec     `json:"spec,omitempty"`
		Status map[string]string `json:"status,omitempty"`
	}

	// InstanceSpec defines the spec property of the InstanceState
	// +kubebuilder:object:generate=true
	InstanceSpec struct {
		Name        string                       `json:"name"`
		DisplayName string                       `json:"displayName,omitempty"`
		Scope       string                       `json:"scope,omitempty"`
		Parameters  map[string]string            `json:"parameters,omitempty"` //TODO: Do we still need this?
		Metadata    map[string]string            `json:"metadata,omitempty"`
		Solution    string                       `json:"solution"`
		Target      TargetSelector               `json:"target,omitempty"`
		Topologies  []TopologySpec               `json:"topologies,omitempty"`
		Pipelines   []PipelineSpec               `json:"pipelines,omitempty"`
		Arguments   map[string]map[string]string `json:"arguments,omitempty"`
		Generation  string                       `json:"generation,omitempty"`
		// Defines the version of a particular resource
		Version string `json:"version,omitempty"`
	}

	// TargertRefSpec defines the target the instance will deploy to
	// +kubebuilder:object:generate=true
	TargetSelector struct {
		Name     string            `json:"name,omitempty"`
		Selector map[string]string `json:"selector,omitempty"`
	}

	// PipelineSpec defines the desired pipeline of the instance
	// +kubebuilder:object:generate=true
	PipelineSpec struct {
		Name       string            `json:"name"`
		Skill      string            `json:"skill"`
		Parameters map[string]string `json:"parameters,omitempty"`
	}
)

func (c TargetSelector) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(TargetSelector)
	if !ok {
		return false, errors.New("parameter is not a TargetSelector type")
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if !StringMapsEqual(c.Selector, otherC.Selector, nil) {
		return false, nil
	}

	return true, nil
}

func (c TopologySpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(TopologySpec)
	if !ok {
		return false, errors.New("parameter is not a TopologySpec type")
	}

	if c.Device != otherC.Device {
		return false, nil
	}

	if !StringMapsEqual(c.Selector, otherC.Selector, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Bindings, otherC.Bindings) {
		return false, nil
	}

	return true, nil
}

func (c PipelineSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(PipelineSpec)
	if !ok {
		return false, errors.New("parameter is not a PipelineSpec type")
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if c.Skill != otherC.Skill {
		return false, nil
	}

	if !StringMapsEqual(c.Parameters, otherC.Parameters, nil) {
		return false, nil
	}

	return true, nil
}

func (c InstanceSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(InstanceSpec)
	if !ok {
		return false, errors.New("parameter is not a InstanceSpec type")
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if c.DisplayName != otherC.DisplayName {
		return false, nil
	}

	if c.Scope != otherC.Scope {
		return false, nil
	}

	// TODO: These are not compared in current version. Metadata is usually not considred part of the state so
	// it's reasonable not to compare. The parameters (same arguments apply to arguments below) are dynamic so
	// comparision is unpredictable. Should we not compare the arguments as well? Or, should we get rid of the
	// dynamic things altoghter so everyting is explicitly declared? I feel we are mixing some templating features
	// into the object model.

	// if !StringMapsEqual(c.Parameters, otherC.Parameters, nil) {
	// 	return false, nil
	// }

	// if !StringMapsEqual(c.Metadata, otherC.Metadata, nil) {
	// 	return false, nil
	// }

	equal, err := c.Target.DeepEquals(otherC.Target)
	if err != nil {
		return false, err
	}

	if !equal {
		return false, nil
	}

	if !SlicesEqual(c.Topologies, otherC.Topologies) {
		return false, nil
	}

	if !SlicesEqual(c.Pipelines, otherC.Pipelines) {
		return false, nil
	}

	if !StringStringMapsEqual(c.Arguments, otherC.Arguments, nil) {
		return false, nil
	}

	return true, nil
}
