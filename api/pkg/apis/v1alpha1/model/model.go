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

type ModelState struct {
	Id   string     `json:"id"`
	Spec *ModelSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:generate=true
type ModelSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Constraints string            `json:"constraints,omitempty"`
	Bindings    []BindingSpec     `json:"bindings,omitempty"`
}

const (
	AppPackage     = "app.package"
	AppImage       = "app.image"
	ContainerImage = "container.image"
)

func (c ModelSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherModelSpec, ok := other.(*ModelSpec)
	if !ok {
		return false, nil
	}
	if c.DisplayName != otherModelSpec.DisplayName {
		return false, nil
	}
	if c.Constraints != otherModelSpec.Constraints {
		return false, nil
	}
	if !StringMapsEqual(c.Properties, otherModelSpec.Properties, nil) {
		return false, nil
	}
	if !SlicesEqual(c.Bindings, otherModelSpec.Bindings) {
		return false, nil
	}
	return true, nil
}
