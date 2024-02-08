/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"
	"fmt"

	go_slices "golang.org/x/exp/slices"
)

type DeploymentSpec struct {
	SolutionName        string                 `json:"solutionName"`
	Solution            SolutionState          `json:"solution"`
	Instance            InstanceState          `json:"instance"`
	Targets             map[string]TargetState `json:"targets"`
	Devices             []DeviceSpec           `json:"devices,omitempty"`
	Assignments         map[string]string      `json:"assignments,omitempty"`
	ComponentStartIndex int                    `json:"componentStartIndex,omitempty"`
	ComponentEndIndex   int                    `json:"componentEndIndex,omitempty"`
	ActiveTarget        string                 `json:"activeTarget,omitempty"`
	Generation          string                 `json:"generation,omitempty"`
}

func (d DeploymentSpec) GetComponentSlice() []ComponentSpec {
	components := d.Solution.Spec.Components
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
		fmt.Println(">>>>>>>>1")
		return false, nil
	}

	equal, err := c.Solution.DeepEquals(otherC.Solution)
	if err != nil {
		return false, err
	}

	if !equal {
		fmt.Println(">>>>>>>>2")
		return false, nil
	}

	equal, err = c.Instance.DeepEquals(otherC.Instance)
	if err != nil {
		return false, err
	}

	if !equal {
		fmt.Println(">>>>>>>>3")
		return false, nil
	}

	if !mapsEqual(c.Targets, otherC.Targets, nil) {
		fmt.Println(">>>>>>>>4")
		return false, nil
	}

	if !SlicesEqual(c.Devices, otherC.Devices) {
		fmt.Println(">>>>>>>>5")
		return false, nil
	}

	if !StringMapsEqual(c.Assignments, otherC.Assignments, nil) {
		fmt.Println(">>>>>>>>6")
		return false, nil
	}

	if c.ComponentStartIndex != otherC.ComponentStartIndex {
		fmt.Println(">>>>>>>>7")
		return false, nil
	}

	if c.ComponentEndIndex != otherC.ComponentEndIndex {
		fmt.Println(">>>>>>>>8")
		return false, nil
	}

	if c.ActiveTarget != otherC.ActiveTarget {
		fmt.Println(">>>>>>>>9")
		return false, nil
	}

	return true, nil
}

func mapsEqual(a map[string]TargetState, b map[string]TargetState, ignoredMissingKeys []string) bool {
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
