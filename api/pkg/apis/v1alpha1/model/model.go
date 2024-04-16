/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

type ModelState struct {
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`
	Spec       *ModelSpec `json:"spec,omitempty"`
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
	otherModelSpec, ok := other.(ModelSpec)
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

func (c ModelState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(ModelState)
	if !ok {
		return false, errors.New("parameter is not a ModelState type")
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
