/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "errors"

type ObjectMeta struct {
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (c *ObjectMeta) FixNames(name string) {
	c.Name = name
	if c.Namespace == "" {
		c.Namespace = "default"
	}
}
func (c *ObjectMeta) MergeFrom(other ObjectMeta) {
	if c.Name == "" {
		c.Name = other.Name
	}
	if c.Namespace == "" {
		c.Namespace = other.Namespace
	}
	if c.Labels == nil {
		c.Labels = other.Labels
	}
	if c.Annotations == nil {
		c.Annotations = other.Annotations
	}
}

func (c ObjectMeta) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(ObjectMeta)
	if !ok {
		return false, errors.New("parameter is not a ObjectMeta type")
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if c.Namespace != otherC.Namespace {
		if (c.Namespace == "" && otherC.Namespace == "default") || (c.Namespace == "default" && otherC.Namespace == "") {
			// default namespace is the same as empty namespace
		} else {
			return false, nil
		}
	}

	if !StringMapsEqual(c.Labels, otherC.Labels, nil) {
		return false, nil
	}

	// annotations are not compared
	// if !StringMapsEqual(c.Annotations, otherC.Annotations, nil) {
	// 	return false, nil
	// }

	return true, nil
}
