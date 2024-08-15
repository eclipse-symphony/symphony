/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type ObjectMeta struct {
	Namespace   string            `json:"namespace,omitempty"`
	Name        string            `json:"name,omitempty"`
	Generation  string            `json:"generation,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// UnmarshalJSON custom unmarshaller to handle Generation field(accept both of number and string)
func (o *ObjectMeta) UnmarshalJSON(data []byte) error {
	type Alias ObjectMeta
	aux := &struct {
		Generation interface{} `json:"generation,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Generation == nil {
		o.Generation = ""
	} else {
		switch v := aux.Generation.(type) {
		case string:
			o.Generation = v
		case float64:
			o.Generation = strconv.FormatInt(int64(v), 10)
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("unexpected type for generation field: %T", v), v1alpha2.BadConfig)
		}
	}

	return nil
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

func (c *ObjectMeta) UpdateAnnotation(key string, value string) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}

	c.Annotations[key] = value
}
