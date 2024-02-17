/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
	"reflect"
)

// TODO: all state objects should converge to this paradigm: id, spec and status
type CatalogState struct {
	Id        string                 `json:"id"`
	Namespace string                 `json:"namespace"`
	Spec      *CatalogSpec           `json:"spec,omitempty"`
	Status    *CatalogStatus         `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// +kubebuilder:object:generate=true
type ObjectRef struct {
	SiteId     string            `json:"siteId"`
	Name       string            `json:"name"`
	Group      string            `json:"group"`
	Version    string            `json:"version"`
	Kind       string            `json:"kind"`
	Namespace  string            `json:"namespace"`
	Address    string            `json:"address,omitempty"`
	Generation string            `json:"generation,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
type CatalogSpec struct {
	SiteId     string                 `json:"siteId"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	ParentName string                 `json:"parentName,omitempty"`
	ObjectRef  ObjectRef              `json:"objectRef,omitempty"`
	Generation string                 `json:"generation,omitempty"`
}

type CatalogStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CatalogSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogSpec)
	if !ok {
		return false, errors.New("parameter is not a CatalogSpec type")
	}

	if c.SiteId != otherC.SiteId {
		return false, nil
	}

	if c.Name != otherC.Name {
		return false, nil
	}

	if c.ParentName != otherC.ParentName {
		return false, nil
	}

	if c.Generation != otherC.Generation {
		return false, nil
	}

	if !reflect.DeepEqual(c.Properties, otherC.Properties) {
		return false, nil
	}

	return true, nil
}

func (c CatalogState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogState)
	if !ok {
		return false, errors.New("parameter is not a CatalogState type")
	}

	if c.Id != otherC.Id {
		return false, nil
	}

	if c.Namespace != otherC.Namespace {
		return false, nil
	}

	if !SimpleMapsEqual(c.Metadata, otherC.Metadata) {
		return false, nil
	}

	equal, err := c.Spec.DeepEquals(*otherC.Spec)
	if err != nil || !equal {
		return equal, err
	}
	return true, nil
}

// INode interface
func (s CatalogState) GetId() string {
	return s.Id
}
func (s CatalogState) GetParent() string {
	if s.Spec != nil {
		return s.Spec.ParentName
	}
	return ""
}
func (s CatalogState) GetType() string {
	if s.Spec != nil {
		return s.Spec.Type
	}
	return ""
}
func (s CatalogState) GetProperties() map[string]interface{} {
	if s.Spec != nil {
		return s.Spec.Properties
	}
	return nil
}

// IEdge interface
func (s CatalogState) GetFrom() string {
	if s.Spec != nil {
		if s.Spec.Type == "edge" {
			if s.Metadata != nil {
				if from, ok := s.Metadata["from"]; ok {
					return from.(string)
				}
			}
		}
	}
	return ""
}

func (s CatalogState) GetTo() string {
	if s.Spec != nil {
		if s.Spec.Type == "edge" {
			if s.Metadata != nil {
				if to, ok := s.Metadata["to"]; ok {
					return to.(string)
				}
			}
		}
	}
	return ""
}
