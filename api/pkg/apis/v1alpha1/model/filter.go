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

type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (c FilterSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	var otherC FilterSpec
	var ok bool
	if otherC, ok = other.(FilterSpec); !ok {
		return false, errors.New("parameter is not a FilterSpec type")
	}

	if c.Direction != otherC.Direction {
		return false, nil
	}
	if c.Type != otherC.Type {
		return false, nil
	}
	if !StringMapsEqual(c.Parameters, otherC.Parameters, nil) {
		return false, nil
	}
	return true, nil
}
