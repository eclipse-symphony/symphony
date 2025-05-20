/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"encoding/json"
	"fmt"
	"strings"

	go_slices "golang.org/x/exp/slices"
	"helm.sh/helm/v3/pkg/strvals"
)

type (
	// IDeepEquals interface defines an interface for memberwise equality comparision
	IDeepEquals interface {
		DeepEquals(other IDeepEquals) (bool, error)
	}

	ValueInjections struct {
		InstanceId   string
		SolutionId   string
		TargetId     string
		ActivationId string
		CatalogId    string
		CampaignId   string
		DeviceId     string
		SkillId      string
		ModelId      string
		SiteId       string
	}
)

// stringMapsEqual compares two string maps for equality
func StringMapsEqual(a map[string]string, b map[string]string, ignoredMissingKeys []string) bool {
	// if len(a) != len(b) {
	// 	return false
	// }

	//TODO: I don't think we need this anymore

	for k, v := range a {
		if bv, ok := b[k]; ok {
			if bv != v {
				if !strings.Contains(bv, "${{$instance()}}") &&
					!strings.Contains(v, "${{$instance()}}") &&
					!strings.Contains(bv, "${{$solution()}}") &&
					!strings.Contains(v, "${{$solution()}}") &&
					!strings.Contains(bv, "${{$target()}}") &&
					!strings.Contains(v, "${{$target()}}") { // Skip comparision because $instance is filled by different instances
					return false
				}
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				return false
			}
		}
	}

	for k, v := range b {
		if bv, ok := a[k]; ok {
			if bv != v {
				if !strings.Contains(bv, "${{$instance()}}") &&
					!strings.Contains(v, "${{$instance()}}") &&
					!strings.Contains(bv, "${{$solution()}}") &&
					!strings.Contains(v, "${{$solution()}}") &&
					!strings.Contains(bv, "${{$target()}}") &&
					!strings.Contains(v, "${{$target()}}") { // Skip comparision because $instance is filled by different instances
					return false
				}
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				return false
			}
		}
	}

	return true
}

func StringStringMapsEqual(a map[string]map[string]string, b map[string]map[string]string, ignoredMissingKeys []string) bool {
	for k, v := range a {
		if bv, ok := b[k]; ok {
			if !StringMapsEqual(v, bv, ignoredMissingKeys) {
				return false
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				return false
			}
		}
	}

	for k, v := range b {
		if bv, ok := a[k]; ok {
			if !StringMapsEqual(v, bv, ignoredMissingKeys) {
				return false
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				return false
			}
		}
	}

	return true
}

// Compatibility Helper
func ExtractRawEnvFromProperties(properties map[string]interface{}) map[string]string {
	env := make(map[string]string)
	for k, v := range properties {
		if strings.HasPrefix(k, "env.") {
			env[k] = fmt.Sprintf("%v", v)
		}
	}

	return env
}

func EnvMapsEqual(a map[string]string, b map[string]string) bool {
	// if len(a) != len(b) {
	// 	return false
	// }

	for k, v := range a {
		if strings.HasPrefix(k, "env.") {
			if bv, ok := b[k]; ok {
				if bv != v {
					if !strings.Contains(bv, "${{$instance()}}") &&
						!strings.Contains(v, "${{$instance()}}") &&
						!strings.Contains(bv, "${{$solution()}}") &&
						!strings.Contains(v, "${{$solution()}}") &&
						!strings.Contains(bv, "${{$target()}}") &&
						!strings.Contains(v, "${{$target()}}") { // Skip comparision because $instance is filled by different instances
						return false
					}
				}
			}
		}
	}

	for k, v := range b {
		if strings.HasPrefix(k, "env.") {
			if bv, ok := a[k]; ok {
				if bv != v {
					if !strings.Contains(bv, "${{$instance()}}") &&
						!strings.Contains(v, "${{$instance()}}") &&
						!strings.Contains(bv, "${{$solution()}}") &&
						!strings.Contains(v, "${{$solution()}}") &&
						!strings.Contains(bv, "${{$target()}}") &&
						!strings.Contains(v, "${{$target()}}") { // Skip comparision because $instance is filled by different instances
						return false
					}
				}
			}
		}
	}

	return true
}

func ExtractReferenceSlice[K IDeepEquals](a []K) []*K {
	ret := make([]*K, 0)
	for _, v := range a {
		ret = append(ret, &v)
	}
	return ret
}

// SliceEuql compares two slices of IDeepEqual items, ignoring the order of items
// It returns two if the two slices are exactly the same, otherwise it returns false
func SlicesEqual[K IDeepEquals](a []K, b []K) bool {
	if len(a) != len(b) {
		return false
	}

	used := make(map[int]bool)
	for _, ia := range a {
		found := false
		for j, jb := range b {
			if _, ok := used[j]; !ok {
				t, e := ia.DeepEquals(jb)
				if e != nil {
					return false
				}

				if t {
					used[j] = true
					found = true
					break
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func SlicesCover[K IDeepEquals](src []K, dest []K) bool {
	for _, ia := range src {
		found := false
		for _, jb := range dest {
			t, e := ia.DeepEquals(jb)
			if e == nil && t {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func SlicesAny[K IDeepEquals](src []K, dest []K) bool {
	for _, ia := range src {
		for _, jb := range dest {
			t, e := ia.DeepEquals(jb)
			if e == nil && t {
				return true
			}
		}
	}

	return false
}

func CheckProperty(a map[string]string, b map[string]string, key string, ignoreCase bool) bool {
	if va, oka := a[key]; oka {
		if vb, okb := b[key]; okb {
			if ignoreCase {
				return strings.EqualFold(va, vb)
			} else {
				return va == vb
			}
		}

		return false
	}

	return true
}

func CheckPropertyCompat(a map[string]interface{}, b map[string]interface{}, key string, ignoreCase bool) bool {
	if va, oka := a[key]; oka {
		if vb, okb := b[key]; okb {
			if ignoreCase {
				return strings.EqualFold(fmt.Sprintf("%v", va), fmt.Sprintf("%v", vb))
			} else {
				return va == vb
			}
		}

		return false
	}

	return true
}

func HasSameProperty(a map[string]string, b map[string]string, key string) bool {
	va, oka := a[key]
	vb, okb := b[key]
	if oka && okb {
		return va == vb
	} else if !oka && !okb {
		return true
	} else {
		return false
	}

}

func HasSamePropertyCompat(a map[string]interface{}, b map[string]interface{}, key string) bool {
	va, oka := a[key]
	vb, okb := b[key]
	if oka && okb {
		return fmt.Sprintf("%v", va) == fmt.Sprintf("%v", vb)
	} else if !oka && !okb {
		return true
	} else {
		return false
	}

}

func CollectPropertiesWithPrefix(col map[string]interface{}, prefix string, injections *ValueInjections, withHierarchy bool) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range col {
		if v, ok := v.(string); ok && strings.HasPrefix(k, prefix) {
			key := k[len(prefix):]
			if withHierarchy {
				strvals.ParseInto(fmt.Sprintf("%s=%s", key, ResolveString(v, injections)), ret)
			} else {
				ret[key] = ResolveString(v, injections)
			}
		}
	}

	return ret
}

func ReadPropertyCompat(col map[string]interface{}, key string, injections *ValueInjections) string {
	if v, ok := col[key]; ok {
		return ResolveString(fmt.Sprintf("%v", v), injections)
	}

	return ""
}

func ReadProperty(col map[string]string, key string, injections *ValueInjections) string {
	if v, ok := col[key]; ok {
		return ResolveString(v, injections)
	}

	return ""
}

func ResolveString(value string, injections *ValueInjections) string {
	//TODO: future enhancement - analyze the syntax instead of doing simply string replacement
	if injections != nil {
		value = strings.ReplaceAll(value, "${{$instance()}}", injections.InstanceId)
		value = strings.ReplaceAll(value, "${{$solution()}}", injections.SolutionId)
		value = strings.ReplaceAll(value, "${{$target()}}", injections.TargetId)
		value = strings.ReplaceAll(value, "${{$activation()}}", injections.ActivationId)
		value = strings.ReplaceAll(value, "${{$catalog()}}", injections.CatalogId)
		value = strings.ReplaceAll(value, "${{$campaign()}}", injections.CampaignId)
		value = strings.ReplaceAll(value, "${{$device()}}", injections.DeviceId)
		value = strings.ReplaceAll(value, "${{$model()}}", injections.ModelId)
		value = strings.ReplaceAll(value, "${{$skill()}}", injections.SkillId)
		value = strings.ReplaceAll(value, "${{$site()}}", injections.SiteId)
	}

	return value
}

func ToDeployment(data []byte) (*DeploymentSpec, error) {
	var deployment DeploymentSpec
	err := json.Unmarshal(data, &deployment)
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}
