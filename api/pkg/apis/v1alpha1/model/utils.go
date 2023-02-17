/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package model

import (
	"strings"

	go_slices "golang.org/x/exp/slices"
)

// IDeepEquals interface defines an interface for memberwise equality comparision
type IDeepEquals interface {
	DeepEquals(other IDeepEquals) (bool, error)
}

// stringMapsEqual compares two string maps for equality
func StringMapsEqual(a map[string]string, b map[string]string, ignoredMissingKeys []string) bool {
	// if len(a) != len(b) {
	// 	return false
	// }
	for k, v := range a {
		if bv, ok := b[k]; ok {
			if bv != v {
				if !strings.Contains(bv, "$instance()") && !strings.Contains(v, "$instance()") && !strings.Contains(bv, "$solution()") && !strings.Contains(v, "$solution()") && !strings.Contains(bv, "$target()") && !strings.Contains(v, "$target()") { // Skip comparision because $instance is filled by different instances
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
				if !strings.Contains(bv, "$instance()") && !strings.Contains(v, "$instance()") && !strings.Contains(bv, "$solution()") && !strings.Contains(v, "$solution()") && !strings.Contains(bv, "$target()") && !strings.Contains(v, "$target()") { // Skip comparision because $instance is filled by different instances
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

func EnvMapsEqual(a map[string]string, b map[string]string) bool {
	// if len(a) != len(b) {
	// 	return false
	// }
	for k, v := range a {
		if strings.HasPrefix(k, "env.") {
			if bv, ok := b[k]; ok {
				if bv != v {
					if !strings.Contains(bv, "$instance()") && !strings.Contains(v, "$instance()") && !strings.Contains(bv, "$solution()") && !strings.Contains(v, "$solution()") && !strings.Contains(bv, "$target()") && !strings.Contains(v, "$target()") { // Skip comparision because $instance is filled by different instances
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
					if !strings.Contains(bv, "$instance()") && !strings.Contains(v, "$instance()") && !strings.Contains(bv, "$solution()") && !strings.Contains(v, "$solution()") && !strings.Contains(bv, "$target()") && !strings.Contains(v, "$target()") { // Skip comparision because $instance is filled by different instances
						return false
					}
				}
			}
		}
	}
	return true
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

type ValueInjections struct {
	InstanceId string
	SolutionId string
	TargetId   string
}

func CollectPropertiesWithPrefix(col map[string]string, prefix string, injections *ValueInjections) map[string]string {
	ret := make(map[string]string)
	for k, v := range col {
		if strings.HasPrefix(k, prefix) {
			key := k[len(prefix):]
			ret[key] = ResolveString(v, injections)
		}
	}
	return ret
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
		value = strings.ReplaceAll(value, "$instance()", injections.InstanceId)
		value = strings.ReplaceAll(value, "$solution()", injections.SolutionId)
		value = strings.ReplaceAll(value, "$target()", injections.TargetId)
	}
	return value
}
