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

type TargetRefSpec struct {
	Name     string            `json:"name,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
}

func (c TargetRefSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC TargetRefSpec
	var ok bool
	if otherC, ok = other.(TargetRefSpec); !ok {
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

type TopologySpec struct {
	Device   string            `json:"device,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
	Bindings []BindingSpec     `json:"bindings,omitempty"`
}

func (c TopologySpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC TopologySpec
	var ok bool
	if otherC, ok = other.(TopologySpec); !ok {
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

type PipelineSpec struct {
	Name       string            `json:"name"`
	Skill      string            `json:"skill"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (c PipelineSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC PipelineSpec
	var ok bool
	if otherC, ok = other.(PipelineSpec); !ok {
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

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Solution    string            `json:"solution"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Target      TargetRefSpec     `json:"target,omitempty"`
	Topologies  []TopologySpec    `json:"topologies,omitempty"`
	Pipelines   []PipelineSpec    `json:"pipelines,omitempty"`
	Stage       string            `json:"stage,omitempty"`
	Schedule    string            `json:"schedule,omitempty"`
}

func (c InstanceSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC InstanceSpec
	var ok bool
	if otherC, ok = other.(InstanceSpec); !ok {
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
	if c.Solution != otherC.Solution {
		return false, nil
	}
	if c.Stage != otherC.Stage {
		return false, nil
	}
	if c.Schedule != otherC.Schedule {
		return false, nil
	}
	if !StringMapsEqual(c.Parameters, otherC.Parameters, nil) {
		return false, nil
	}
	if !StringMapsEqual(c.Metadata, otherC.Metadata, nil) {
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
	return true, nil
}
