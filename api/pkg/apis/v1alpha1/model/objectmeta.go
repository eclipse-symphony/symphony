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

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type ObjectMeta struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	// ETag is a string representing the version of the object, it bump whenever the object is updated.
	// All the state store should support auto-incrementing the version number.
	// For example, resourceVersion in kubernetes
	ETag string `json:"eTag,omitempty"`
	// ObjGeneration changes when Spec changes
	// object manager need to detect spec changes and update the generation
	// For example, generation in kubernetes
	ObjGeneration int64 `json:"objGeneraion,omitempty"`

	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// UnmarshalJSON custom unmarshaller to handle Generation field(accept both of number and string)
func (o *ObjectMeta) UnmarshalJSON(data []byte) error {
	type Alias ObjectMeta
	aux := &struct {
		ETag interface{} `json:"etag,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ETag == nil {
		o.ETag = ""
	} else {
		switch v := aux.ETag.(type) {
		case string:
			o.ETag = v
		case float64:
			o.ETag = strconv.FormatInt(int64(v), 10)
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

func (c *ObjectMeta) PreserveSystemMetadata(metadata ObjectMeta) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}

	// system annotations should only be updated ineternally
	for _, key := range constants.SystemReservedAnnotations() {
		if value, ok := metadata.Annotations[key]; ok {
			c.Annotations[key] = value
		}
	}
	// set the resoure version
	c.ETag = metadata.ETag
}
