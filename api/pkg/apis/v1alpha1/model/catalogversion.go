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
type CatalogVersionState struct {
	ObjectMeta ObjectMeta     `json:"metadata,omitempty"`
	Spec       *CatalogVersionSpec   `json:"spec,omitempty"`
	Status     *CatalogVersionStatus `json:"status,omitempty"`
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
type CatalogVersionSpec struct {
	CatalogType  string                 `json:"catalogType"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	Properties   map[string]interface{} `json:"properties"`
	ParentName   string                 `json:"parentName,omitempty"`
	ObjectRef    ObjectRef              `json:"objectRef,omitempty"`
	Version      string                 `json:"version,omitempty"`
	RootResource string                 `json:"rootResource,omitempty"`
}

type CatalogVersionStatus struct {
	Properties map[string]string `json:"properties"`
}

func (c CatalogVersionSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogVersionSpec)
	if !ok {
		return false, errors.New("parameter is not a CatalogVersionSpec type")
	}

	if c.ParentName != otherC.ParentName {
		return false, nil
	}

	if !reflect.DeepEqual(c.Properties, otherC.Properties) {
		return false, nil
	}

	return true, nil
}

func (c CatalogVersionState) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(CatalogVersionState)
	if !ok {
		return false, errors.New("parameter is not a CatalogVersionState type")
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

// INode interface
func (s CatalogVersionState) GetId() string {
	return s.ObjectMeta.Name
}
func (s CatalogVersionState) GetParent() string {
	if s.Spec != nil {
		return s.Spec.ParentName
	}
	return ""
}
func (s CatalogVersionState) GetType() string {
	if s.Spec != nil {
		return s.Spec.CatalogType
	}
	return ""
}
func (s CatalogVersionState) GetProperties() map[string]interface{} {
	if s.Spec != nil {
		return s.Spec.Properties
	}
	return nil
}

// IEdge interface
func (s CatalogVersionState) GetFrom() string {
	if s.Spec != nil {
		if s.Spec.CatalogType == "edge" {
			if s.Spec.Metadata != nil {
				if from, ok := s.Spec.Metadata["from"]; ok {
					return from
				}
			}
		}
	}
	return ""
}

func (s CatalogVersionState) GetTo() string {
	if s.Spec != nil {
		if s.Spec.CatalogType == "edge" {
			if s.Spec.Metadata != nil {
				if to, ok := s.Spec.Metadata["to"]; ok {
					return to
				}
			}
		}
	}
	return ""
}
