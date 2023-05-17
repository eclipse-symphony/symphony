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

type VersionSpec struct {
	Solution   string `json:"solution"`
	Percentage int    `json:"percentage"`
}

func (c VersionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC VersionSpec
	var ok bool
	if otherC, ok = other.(VersionSpec); !ok {
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

type InstanceState struct {
	Id     string            `json:"id"`
	Spec   *InstanceSpec     `json:"spec,omitempty"`
	Status map[string]string `json:"status,omitempty"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"` //TODO: Do we still need this?
	Metadata    map[string]string `json:"metadata,omitempty"`
	Solution    string            `json:"solution"`
	// Instead of a single solution, users can specify a mixture of multiple solutions with percentage weight. This is for
	// scenarios like canary deployment and blue-green deployment (used in conjunction with the Compagin object)
	Versions   []VersionSpec                `json:"versions,omitempty"`
	Target     TargetRefSpec                `json:"target,omitempty"`
	Topologies []TopologySpec               `json:"topologies,omitempty"`
	Pipelines  []PipelineSpec               `json:"pipelines,omitempty"`
	Arguments  map[string]map[string]string `json:"arguments,omitempty"`
	// We allow optining out reconciliation to support deployment campaign, which operates on a collection of instances. As
	// the compaign progresses, it needs to turn off reconciliation of all but one instances so that reconciliation is driven
	// by only one instance.
	OptOutReconciliation bool `json:"optOutReconciliation,omitempty"`
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
