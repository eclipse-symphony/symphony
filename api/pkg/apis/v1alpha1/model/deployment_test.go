/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentDeepEquals(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestDeploymentDeepEqualsOneEmpty(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a DeploymentSpec type")
	assert.False(t, res)
}

func TestDeploymentDeepEqualsSolutionNameNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName1",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsSolutionNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName1",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsInstanceNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName1",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsTargetsNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo1": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsDevicesNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName1",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsComponentStartIndexNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 1,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsComponentEndIndexNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   1,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestDeploymentDeepEqualsActiveTargetNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget1",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestGetComponentSlice(t *testing.T) {
	deployment := DeploymentSpec{
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
	}
	res := deployment.GetComponentSlice()
	assert.Equal(t, 0, len(res))
}

func TestGetComponentSliceWithValues(t *testing.T) {
	deployment := DeploymentSpec{
		ComponentStartIndex: 1,
		ComponentEndIndex:   2,
		Solution: SolutionSpec{
			Components: []ComponentSpec{
				{Name: "Component1"},
				{Name: "Component2"},
				{Name: "Component3"},
				{Name: "Component4"},
				{Name: "Component5"},
			},
		},
	}
	res := deployment.GetComponentSlice()
	assert.Equal(t, 1, len(res))
}

func TestDeploymentDeepEqualsAssignmentsNotMatch(t *testing.T) {
	deployment1 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	deployment2 := DeploymentSpec{
		SolutionName: "SolutionName",
		Solution: SolutionSpec{
			DisplayName: "SolutionDisplayName",
		},
		Instance: InstanceSpec{
			Name: "InstanceName",
		},
		Targets: map[string]TargetSpec{
			"foo": {
				DisplayName: "TargetName",
			},
		},
		Devices: []DeviceSpec{{
			DisplayName: "DeviceName",
		}},
		Assignments: map[string]string{
			"foo": "bar1",
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   0,
		ActiveTarget:        "ActiveTarget",
	}
	res, err := deployment1.DeepEquals(deployment2)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestMapsEqualMap1Extra(t *testing.T) {
	map1 := map[string]TargetSpec{
		"foo": {
			DisplayName: "TargetName",
		},
	}
	map2 := map[string]TargetSpec{}
	res := mapsEqual(map1, map2, nil)
	assert.False(t, res)
}

func TestMapsNotEqualMap2Extra(t *testing.T) {
	map1 := map[string]TargetSpec{}
	map2 := map[string]TargetSpec{
		"foo": {
			DisplayName: "TargetName",
		},
	}
	res := mapsEqual(map1, map2, nil)
	assert.False(t, res)
}

func TestMapsNotEqual(t *testing.T) {
	map2 := map[string]TargetSpec{
		"foo": {
			DisplayName: "TargetName",
		},
	}
	map1 := map[string]TargetSpec{
		"foo": {
			DisplayName: "TargetName1",
		},
	}
	res := mapsEqual(map1, map2, nil)
	assert.False(t, res)
}
