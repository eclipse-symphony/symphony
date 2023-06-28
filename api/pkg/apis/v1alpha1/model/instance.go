/*
Copyright 2022 The COA Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import "errors"

type (

	// InstanceState defines the current state of the instance
	InstanceState struct {
		Id     string            `json:"id"`
		Spec   *InstanceSpec     `json:"spec,omitempty"`
		Status map[string]string `json:"status,omitempty"`
	}

	// InstanceSpec defines the spec property of the InstanceState
	InstanceSpec struct {
		Name        string            `json:"name"`
		DisplayName string            `json:"displayName,omitempty"`
		Scope       string            `json:"scope,omitempty"`
		Parameters  map[string]string `json:"parameters,omitempty"` //TODO: Do we still need this?
		Metadata    map[string]string `json:"metadata,omitempty"`

		// Instead of a single solution, users can specify a mixture of multiple solutions with percentage weight. This is for
		// scenarios like canary deployment and blue-green deployment (used in conjunction with the Campagin object)

		Solution   string                       `json:"solution"`
		Versions   []VersionSpec                `json:"versions,omitempty"`
		Target     TargetRefSpec                `json:"target,omitempty"`
		Topologies []TopologySpec               `json:"topologies,omitempty"`
		Pipelines  []PipelineSpec               `json:"pipelines,omitempty"`
		Arguments  map[string]map[string]string `json:"arguments,omitempty"`
	}

	// TargertRefSpec defines the target the instance will deploy to
	TargetRefSpec struct {
		Name     string            `json:"name,omitempty"`
		Selector map[string]string `json:"selector,omitempty"`
	}

	// TopologySpec defines the desired device topology the instance
	TopologySpec struct {
		Device   string            `json:"device,omitempty"`
		Selector map[string]string `json:"selector,omitempty"`
		Bindings []BindingSpec     `json:"bindings,omitempty"`
	}

	// PipelineSpec defines the desired pipeline of the instance
	PipelineSpec struct {
		Name       string            `json:"name"`
		Skill      string            `json:"skill"`
		Parameters map[string]string `json:"parameters,omitempty"`
	}

	// VersionSpec defines the desired version of the instance
	VersionSpec struct {
		Solution   string `json:"solution"`
		Percentage int    `json:"percentage"`
	}
)

func (c TargetRefSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(TargetRefSpec)
	if !ok {
		return false, errors.New("parameter is not a TargetRefSpec type")
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

func (c VersionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(VersionSpec)
	if !ok {
		return false, errors.New("parameter is not a VersionSpec type")
	}

	if c.Solution != otherC.Solution {
		return false, nil
	}

	if c.Percentage != otherC.Percentage {
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

	if !SlicesEqual(c.Versions, otherC.Versions) {
		return false, nil
	}

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
