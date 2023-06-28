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
	Constraints  []ConstraintSpec       `json:"constraints,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Skills       []string               `json:"skills,omitempty"`
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

	// if !SlicesEqual(c.Constraints, otherC.Constraints) {	Can't compare constraints as components from actual envrionments don't have constraints
	// 	return false, nil
	// }
	return true, nil
}
