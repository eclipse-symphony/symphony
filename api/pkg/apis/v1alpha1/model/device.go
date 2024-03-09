/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

type (
	// DeviceState defines the current state of the device
	DeviceState struct {
		ObjectMeta ObjectMeta        `json:"metadata,omitempty"`
		Spec       *DeviceSpec       `json:"spec,omitempty"`
		Status     map[string]string `json:"status,omitempty"`
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

func (c DeviceState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(DeviceState)
	if !ok {
		return false, errors.New("parameter is not a DeviceState type")
	}

	equal, err := c.ObjectMeta.DeepEquals(otherC.ObjectMeta)
	if err != nil || !equal {
		return equal, err
	}

	equal, err = c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}

	return true, nil
}
