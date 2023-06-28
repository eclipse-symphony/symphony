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
	"fmt"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type ValidationRule struct {
	RequiredComponentType string   `json:"requiredType"`
	RequiredProperties    []string `json:"requiredProperties"`
	OptionalProperties    []string `json:"optionalProperties"`
	RequiredMetadata      []string `json:"requiredMetadata"`
	OptionalMetadata      []string `json:"optionalMetadata"`
}

func (v ValidationRule) Validate(components []ComponentSpec) error {
	for _, c := range components {
		err := v.validateComponent(c)
		if err != nil {
			return err
		}
	}

	return nil
}
func (v ValidationRule) validateComponent(component ComponentSpec) error {
	//required component type must be set
	if v.RequiredComponentType != "" && v.RequiredComponentType != component.Type {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("provider requires component type '%s', but '%s' is found instead", v.RequiredComponentType, component.Type), v1alpha2.BadRequest)
	}

	// required properties must all present
	for _, p := range v.RequiredProperties {
		if ReadPropertyCompat(component.Properties, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required property '%s' is missing", p), v1alpha2.BadRequest)
		}
	}

	// required metadata must all present
	for _, p := range v.RequiredMetadata {
		if ReadProperty(component.Metadata, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required metadata '%s' is missing", p), v1alpha2.BadRequest)
		}
	}

	return nil
}
