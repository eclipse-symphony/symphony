/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/
package model

import (
	"errors"
	"reflect"
)

// TODO: all state objects should converge to this paradigm: id, spec and status
type CatalogState struct {
	Id     string         `json:"id"`
	Spec   *CatalogSpec   `json:"spec,omitempty"`
	Status *CatalogStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=true
type ObjectRef struct {
	SiteId     string            `json:"siteId"`
	Name       string            `json:"name"`
	Group      string            `json:"group"`
	Version    string            `json:"version"`
	Kind       string            `json:"kind"`
	Scope      string            `json:"scope"`
	Address    string            `json:"address,omitempty"`
	Status     map[string]string `json:"status,omitempty"`
	Generation string            `json:"generation,omitempty"`
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
