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

	go_slices "golang.org/x/exp/slices"
)

type DeploymentSpec struct {
	SolutionName        string                `json:"solutionName"`
	Solution            SolutionSpec          `json:"solution"`
	Instance            InstanceSpec          `json:"instance"`
	Targets             map[string]TargetSpec `json:"targets"`
	Devices             []DeviceSpec          `json:"devices,omitempty"`
	Assignments         map[string]string     `json:"assignments,omitempty"`
	ComponentStartIndex int                   `json:"componentStartIndex,omitempty"`
	ComponentEndIndex   int                   `json:"componentEndIndex,omitempty"`
	ActiveTarget        string                `json:"activeTarget,omitempty"`
	Generation          string                `json:"generation,omitempty"`
}

func (d DeploymentSpec) GetComponentSlice() []ComponentSpec {
	components := d.Solution.Components
	if d.ComponentStartIndex >= 0 && d.ComponentEndIndex >= 0 && d.ComponentEndIndex > d.ComponentStartIndex {
		components = components[d.ComponentStartIndex:d.ComponentEndIndex]
	}
	return components
}

func (c DeploymentSpec) DeepEquals(other IDeepEquals) (bool, error) {
	otherC, ok := other.(DeploymentSpec)
	if !ok {
		return false, errors.New("parameter is not a DeploymentSpec type")
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

	equal, err = c.Instance.DeepEquals(otherC.Instance)
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

func mapsEqual(a map[string]TargetSpec, b map[string]TargetSpec, ignoredMissingKeys []string) bool {
	for k, v := range a {
		if bv, ok := b[k]; ok {
			equal, err := bv.DeepEquals(v)
			if err != nil || !equal {
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
			equal, err := bv.DeepEquals(v)
			if err != nil || !equal {
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
