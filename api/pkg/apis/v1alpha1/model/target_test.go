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

func TestTargetDeepEquals(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestTargetDeepEqualsOneEmpty(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a TargetSpec type")
	assert.False(t, res)
}

func TestTargetDeepEqualsDisplayNameNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName1",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsScopeNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default1",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsMetadataKeyNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo1": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsMetadataValueNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar1"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsPropertiesKeyNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Components:  []ComponentSpec{{}},
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo1": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsPropertiesValueNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar1"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsComponentNameNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName1",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsTopologiestNameNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName1",
		}},
		ForceRedeploy: false,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsConstraintsNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Components:  []ComponentSpec{{}},
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Components:  []ComponentSpec{{}},
		Constraints: "Constraints",
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsForceRedeployNotMatch(t *testing.T) {
	Target := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: false,
	}
	other := TargetSpec{
		DisplayName: "TargetName",
		Scope:       "Default",
		Metadata:    map[string]string{"foo": "bar"},
		Properties:  map[string]string{"foo": "bar"},
		Components: []ComponentSpec{{
			Name: "ComponentName",
		}},
		Constraints: "",
		Topologies: []TopologySpec{{
			Device: "DeviceName",
		}},
		ForceRedeploy: true,
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}
