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

type (
	ConnectionSpec struct {
		Node  string `json:"node"`
		Route string `json:"route"`
	}
	EdgeSpec struct {
		Source ConnectionSpec `json:"source"`
		Target ConnectionSpec `json:"target"`
	}
)

func (c ConnectionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(*ConnectionSpec)
	if !ok {
		return false, nil
	}
	if c.Node != otherSpec.Node {
		return false, nil
	}
	if c.Route != otherSpec.Route {
		return false, nil
	}
	return true, nil
}
func (c EdgeSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherSpec, ok := other.(*EdgeSpec)
	if !ok {
		return false, nil
	}
	equal, err := c.Source.DeepEquals(&otherSpec.Source)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	equal, err = c.Target.DeepEquals(&otherSpec.Target)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	return true, nil
}
