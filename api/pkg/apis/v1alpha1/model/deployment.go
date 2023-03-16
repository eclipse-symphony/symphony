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
	"fmt"

	go_slices "golang.org/x/exp/slices"
)

type DeploymentStage struct {
	SolutionName        string                `json:"solutionName"`
	Solution            SolutionSpec          `json:"solution"`
	Targets             map[string]TargetSpec `json:"targets"`
	Devices             []DeviceSpec          `json:"devices,omitempty"`
	Assignments         map[string]string     `json:"assignments,omitempty"`
	ComponentStartIndex int                   `json:"componentStartIndex,omitempty"`
	ComponentEndIndex   int                   `json:"componentEndIndex,omitempty"`
	ActiveTarget        string                `json:"activeTarget,omitempty"`
}

func (c DeploymentStage) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC DeploymentStage
	var ok bool
	if otherC, ok = other.(DeploymentStage); !ok {
		return false, errors.New("parameter is not a DeploymentStage type")
	}
	if c.SolutionName != otherC.SolutionName {
		return false, nil
	}
	equal, err := c.Solution.DeepEquals(otherC.Solution)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	if !mapsEqual(c.Targets, otherC.Targets, nil) {
		return false, nil
	}
	if !SlicesEqual(c.Devices, otherC.Devices) {
		return false, nil
	}
	if !StringMapsEqual(c.Assignments, otherC.Assignments, nil) {
		return false, nil
	}
	if c.ComponentStartIndex != otherC.ComponentStartIndex {
		return false, nil
	}
	if c.ComponentEndIndex != otherC.ComponentEndIndex {
		return false, nil
	}
	if c.ActiveTarget != otherC.ActiveTarget {
		return false, nil
	}
	return true, nil
}

type DeploymentSpec struct {
	Instance InstanceSpec      `json:"instance"`
	Stages   []DeploymentStage `json:"stages"`
}

func (d DeploymentSpec) GetComponentSlice() []ComponentSpec {
	components := d.Stages[0].Solution.Components
	if d.Stages[0].ComponentStartIndex >= 0 && d.Stages[0].ComponentEndIndex >= 0 && d.Stages[0].ComponentEndIndex > d.Stages[0].ComponentStartIndex {
		components = components[d.Stages[0].ComponentStartIndex:d.Stages[0].ComponentEndIndex]
	}
	return components
}

func (c DeploymentSpec) DeepEquals(other IDeepEquals) (bool, error) {
	var otherC DeploymentSpec
	var ok bool
	if otherC, ok = other.(DeploymentSpec); !ok {
		return false, errors.New("parameter is not a DeploymentSpec type")
	}
	equal, err := c.Instance.DeepEquals(otherC.Instance)
	if err != nil {
		return false, err
	}
	if !equal {
		return false, nil
	}
	if !SlicesEqual(c.Stages, otherC.Stages) {
		return false, nil
	}
	return true, nil
}
func mapsEqual(a map[string]TargetSpec, b map[string]TargetSpec, ignoredMissingKeys []string) bool {
	for k, v := range a {
		if bv, ok := b[k]; ok {
			equal, err := bv.DeepEquals(v)
			if err != nil || !equal {
				fmt.Println("10")
				return false
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				fmt.Println("11")
				return false
			}
		}
	}
	for k, v := range b {
		if bv, ok := a[k]; ok {
			equal, err := bv.DeepEquals(v)
			if err != nil || !equal {
				fmt.Println("12")
				return false
			}
		} else {
			if !go_slices.Contains(ignoredMissingKeys, k) {
				fmt.Println("14")
				return false
			}
		}
	}
	return true
}
