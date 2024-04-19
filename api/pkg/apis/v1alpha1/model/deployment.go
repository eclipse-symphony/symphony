/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"errors"

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
	ObjectNamespace     string                 `json:"objectNamespace,omitempty"`
	Hash                string                 `json:"hash,omitempty"`
}

func (d DeploymentSpec) GetComponentSlice() []ComponentSpec {
	if d.Solution.Spec == nil {
		return nil
	}
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
