/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
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
