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
	Constraints  string                 `json:"constraints,omitempty"`
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

	// if c.Constraints != otherC.Constraints {	Can't compare constraints as components from actual envrionments don't have constraints
	// 	return false, nil
	// }
	return true, nil
}
