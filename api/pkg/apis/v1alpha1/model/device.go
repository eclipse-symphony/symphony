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
	// DeviceState defines the current state of the device
	DeviceState struct {
		Id   string      `json:"id"`
		Spec *DeviceSpec `json:"spec,omitempty"`
	}
	// DeviceSpec defines the spec properties of the DeviceState
	// +kubebuilder:object:generate=true
	DeviceSpec struct {
		DisplayName string            `json:"displayName,omitempty"`
		Properties  map[string]string `json:"properties,omitempty"`
		Bindings    []BindingSpec     `json:"bindings,omitempty"`
	}
)

func (c DeviceSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(DeviceSpec)
	if !ok {
		return false, errors.New("parameter is not a DeviceSpec type")
	}

	if c.DisplayName != otherC.DisplayName {
		return false, nil
	}

	if !StringMapsEqual(c.Properties, otherC.Properties, nil) {
		return false, nil
	}

	if !SlicesEqual(c.Bindings, otherC.Bindings) {
		return false, nil
	}

	return true, nil
}
