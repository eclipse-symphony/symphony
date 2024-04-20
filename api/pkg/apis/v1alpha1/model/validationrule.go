/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
)

type PropertyDesc struct {
	Name            string `json:"name"`
	IgnoreCase      bool   `json:"ignoreCase,omitempty"`
	SkipIfMissing   bool   `json:"skipIfMissing,omitempty"`
	PrefixMatch     bool   `json:"prefixMatch,omitempty"`
	IsComponentName bool   `json:"isComponentName,omitempty"`
	// This is a stop-gap solution to support change detection for advanced comparison scenarios.
	PropChanged func(oldProp, newProp any) bool `json:"-"`
}
type ComponentValidationRule struct {
	RequiredComponentType     string         `json:"requiredType"`
	ChangeDetectionProperties []PropertyDesc `json:"changeDetection,omitempty"`
	ChangeDetectionMetadata   []PropertyDesc `json:"changeDetectionMetadata,omitempty"`
	RequiredProperties        []string       `json:"requiredProperties"`
	OptionalProperties        []string       `json:"optionalProperties"`
	RequiredMetadata          []string       `json:"requiredMetadata"`
	OptionalMetadata          []string       `json:"optionalMetadata"`
}
type ValidationRule struct {
	RequiredComponentType   string                  `json:"requiredType"`
	ComponentValidationRule ComponentValidationRule `json:"componentValidationRule,omitempty"`
	SidecarValidationRule   ComponentValidationRule `json:"sidecarValidationRule,omitempty"`
	// a provider that supports sidecar can deploy sidecar containers with the main container.
	AllowSidecar bool `json:"allowSidecar,omitempty"`
	// a provider that supports scope isolation can deploy to specified scopes other than the "default" scope.
	// instances from different scopes are isolated from each other.
	ScopeIsolation bool `json:"supportScopes,omitempty"`
	// a provider that supports instance isolation can deploy multiple instances on the same target without conflicts.
	InstanceIsolation bool `json:"instanceIsolation,omitempty"`
}

func (v ValidationRule) ValidateInputs(inputs map[string]interface{}) error {
	// required properties must all present
	for _, p := range v.ComponentValidationRule.RequiredProperties {
		if ReadPropertyCompat(inputs, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required property '%s' is missing", p), v1alpha2.BadRequest)
		}
	}
	if v.AllowSidecar {
		for _, p := range v.SidecarValidationRule.RequiredProperties {
			if ReadPropertyCompat(inputs, p, nil) == "" {
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("required sidecar property '%s' is missing", p), v1alpha2.BadRequest)
			}
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

func detectChanges(properties []PropertyDesc, oldName string, newName string, oldValues map[string]interface{}, newValues map[string]interface{}) bool {
	for _, p := range properties {
		if strings.Contains(p.Name, "*") {
			escapedPattern := regexp.QuoteMeta(p.Name)
			// Replace the wildcard (*) with a regular expression pattern
			regexpPattern := strings.ReplaceAll(escapedPattern, `\*`, ".*")
			// Compile the regular expression
			regexpObject := regexp.MustCompile("^" + regexpPattern + "$")
			for k := range oldValues {
				if regexpObject.MatchString(k) {
					if compareProperties(p, oldValues, newValues, k) {
						return true
					}
				}
			}
		} else {
			if p.IsComponentName {
				if !compareStrings(oldName, newName, p.IgnoreCase, p.PrefixMatch) {
					return true
				}
			} else {
				if compareProperties(p, oldValues, newValues, p.Name) {
					return true
				}
			}
		}
	}

	return false
}
func convertMapStringToStringInterface(m map[string]string) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func (v ValidationRule) IsComponentChanged(old ComponentSpec, new ComponentSpec) bool {
	if detectChanges(v.ComponentValidationRule.ChangeDetectionProperties, old.Name, new.Name, old.Properties, new.Properties) {
		return true
	}
	if detectChanges(v.ComponentValidationRule.ChangeDetectionMetadata, old.Name, new.Name,
		convertMapStringToStringInterface(old.Metadata),
		convertMapStringToStringInterface(new.Metadata)) {
		return true
	}
	if v.AllowSidecar {
		touchCount := 0
		for _, sidecar := range new.Sidecars {
			foundOld := false
			for _, oldSidecar := range old.Sidecars {
				if sidecar.Name == oldSidecar.Name {
					if detectChanges(v.SidecarValidationRule.ChangeDetectionProperties, oldSidecar.Name, sidecar.Name, oldSidecar.Properties, sidecar.Properties) {
						return true
					}
					foundOld = true
					touchCount++
					break
				}
			}
			if !foundOld {
				return true
			}
		}
		if touchCount != len(old.Sidecars) {
			return true
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
		return strings.HasPrefix(tb, ta) || strings.HasPrefix(ta, tb)
	}
}
func compareProperties(c PropertyDesc, old map[string]interface{}, new map[string]interface{}, key string) bool {
	v, ook := old[key]
	nv, nok := new[key]
	if c.PropChanged != nil {
		return c.PropChanged(v, nv)
	}
	if ook {
		if nok {
			if !compareStrings(fmt.Sprintf("%v", v), fmt.Sprintf("%v", nv), c.IgnoreCase, c.PrefixMatch) {
				return true
			}
		} else if !c.SkipIfMissing {
			return true
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
	for _, p := range v.ComponentValidationRule.RequiredProperties {
		if ReadPropertyCompat(component.Properties, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required property '%s' is missing", p), v1alpha2.BadRequest)
		}
	}

	// required metadata must all present
	for _, p := range v.ComponentValidationRule.RequiredMetadata {
		if ReadProperty(component.Metadata, p, nil) == "" {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("required metadata '%s' is missing", p), v1alpha2.BadRequest)
		}
	}

	if v.AllowSidecar {
		for _, sidecar := range component.Sidecars {
			for _, p := range v.SidecarValidationRule.RequiredProperties {
				if ReadPropertyCompat(sidecar.Properties, p, nil) == "" {
					return v1alpha2.NewCOAError(nil, fmt.Sprintf("required sidecar property '%s' is missing in sidecar %s", p, sidecar.Name), v1alpha2.BadRequest)
				}
			}
		}
	}

	return nil
}
