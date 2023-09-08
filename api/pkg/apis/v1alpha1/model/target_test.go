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
	assert.Errorf(t, err, "parameter is not a TargetSpec type")
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
