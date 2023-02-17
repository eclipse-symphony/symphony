/*
Copyright 2022 The COA Authors
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
	"errors"
	"sort"
	"strings"
)

type ConstraintSpec struct {
	Key       string   `json:"key"`
	Qualifier string   `json:"qualifier,omitempty"`
	Operator  string   `json:"operator"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"` //TODO: It seems kubebuilder has difficulties handling recursive defs. This is supposed to be an ConstraintSpec array
}

func (c ConstraintSpec) DeepEquals(other IDeepEquals) (bool, error) { // avoid using reflect, which has performance problems
	var otherC ConstraintSpec
	var ok bool
	if otherC, ok = other.(ConstraintSpec); !ok {
		return false, errors.New("parameter is not a ConstraintSpec type")
	}

	if c.Key != otherC.Key {
		return false, nil
	}
	if c.Qualifier != otherC.Qualifier {
		return false, nil
	}
	if c.Operator != otherC.Operator {
		return false, nil
	}
	if c.Value != otherC.Value {
		return false, nil
	}
	if len(c.Values) != len(otherC.Values) {
		return false, nil
	}
	//TODO: values should be order sensitive, so the following comparision needs to be modified
	// However as we plan to refactor the constraint alltogether, we can postpone this
	ca := make([]string, len(c.Values))
	cb := make([]string, len(otherC.Values))
	copy(ca, c.Values)
	copy(cb, otherC.Values)
	sort.Strings(ca)
	sort.Strings(cb)
	sort.Strings(c.Values)
	return strings.Join(ca, ",") == strings.Join(cb, ","), nil
}
func (c ConstraintSpec) Match(properties map[string]string) bool {
	//TODO: expand on this, the following logic only handles "must" conditions
	if v, ok := properties[c.Key]; ok {
		return v == c.Value
	}
	return false
}
