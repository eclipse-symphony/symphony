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

type BindingSpec struct {
	Role     string            `json:"role"`
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config,omitempty"`
}

func (c BindingSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC BindingSpec
	var ok bool
	if otherC, ok = other.(BindingSpec); !ok {
		return false, errors.New("parameter is not a BindingSpec type")
	}
	if c.Role != otherC.Role {
		return false, nil
	}
	if c.Provider != otherC.Provider {
		return false, nil
	}
	if !StringMapsEqual(c.Config, otherC.Config, nil) {
		return false, nil
	}
	return true, nil
}
