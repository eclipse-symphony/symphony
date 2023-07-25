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
	"regexp"
	"strings"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type PropertyDesc struct {
	Name            string `json:"name"`
	IgnoreCase      bool   `json:"ignoreCase,omitempty"`
	SkipIfMissing   bool   `json:"skipIfMissing,omitempty"`
	PrefixMatch     bool   `json:"prefixMatch,omitempty"`
	IsComponentName bool   `json:"isComponentName,omitempty"`
}
type ValidationRule struct {
	RequiredComponentType     string         `json:"requiredType"`
	ChangeDetectionProperties []PropertyDesc `json:"changeDetection,omitempty"`
	RequiredProperties        []string       `json:"requiredProperties"`
	OptionalProperties        []string       `json:"optionalProperties"`
	RequiredMetadata          []string       `json:"requiredMetadata"`
	OptionalMetadata          []string       `json:"optionalMetadata"`
	// a provider that supports scope isolation can deploy to specified scopes other than the "default" scope.
	// instances from different scopes are isolated from each other.
	ScopeIsolation bool `json:"supportScopes,omitempty"`
	// a provider that supports instance isolation can deploy multiple instances on the same target without conflicts.
	InstanceIsolation bool `json:"instanceIsolation,omitempty"`
}

func (v ValidationRule) ValidateInputs(inputs map[string]interface{}) error {
	// required properties must all present
	for _, p := range v.RequiredProperties {
		if ReadPropertyCompat(inputs, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required property '%s' is missing", p), v1alpha2.BadRequest)
		}
	}
	return nil
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

func (v ValidationRule) IsComponentChanged(old ComponentSpec, new ComponentSpec) bool {
	for _, c := range v.ChangeDetectionProperties {
		if strings.Contains(c.Name, "*") {
			escapedPattern := regexp.QuoteMeta(c.Name)
			// Replace the wildcard (*) with a regular expression pattern
			regexpPattern := strings.ReplaceAll(escapedPattern, `\*`, ".*")
			// Compile the regular expression
			regexpObject := regexp.MustCompile("^" + regexpPattern + "$")
			for k, _ := range old.Properties {
				if regexpObject.MatchString(k) {
					if compareProperties(c, old, new, k) {
						return true
					}
				}
			}
		} else {
			if c.IsComponentName {
				if !compareStrings(old.Name, new.Name, c.IgnoreCase, c.SkipIfMissing) {
					return true
				}
			} else {
				if compareProperties(c, old, new, c.Name) {
					return true
				}
			}
		}
	}
	return false
}
func compareStrings(a, b string, ignoreCase bool, prefixMatch bool) bool {
	ta := a
	tb := b
	if ignoreCase {
		ta = strings.ToLower(a)
		tb = strings.ToLower(b)
	}
	if !prefixMatch {
		return ta == tb
	} else {
		return strings.HasPrefix(tb, ta)
	}
}
func compareProperties(c PropertyDesc, old ComponentSpec, new ComponentSpec, key string) bool {
	if v, ok := old.Properties[key]; ok {
		if nv, nok := new.Properties[key]; nok {
			if !compareStrings(fmt.Sprintf("%v", v), fmt.Sprintf("%v", nv), c.IgnoreCase, c.PrefixMatch) {
				return true
			}
		}
	} else {
		if !c.SkipIfMissing {
			return true
		}
	}
	return false
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
