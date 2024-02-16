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
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestTargetDeepEqualsOneEmpty(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a TargetState type")
	assert.False(t, res)
}

func TestTargetDeepEqualsDisplayNameNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName1",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsNamespaceNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default1",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsMetadataKeyNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo1": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsMetadataValueNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar1"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsPropertiesKeyNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo1": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsPropertiesValueNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar1"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsComponentNameNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName1",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsTopologiestNameNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName1",
			}},
			ForceRedeploy: false,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsConstraintsNotMatch(t *testing.T) {
	Target := TargetState{
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
		},
	}
	other := TargetState{
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Components:  []ComponentSpec{{}},
			Constraints: "Constraints",
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}

func TestTargetDeepEqualsForceRedeployNotMatch(t *testing.T) {
	Target := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: false,
		},
	}
	other := TargetState{
		Metadata: map[string]string{"foo": "bar"},
		Spec: &TargetSpec{
			DisplayName: "TargetName",
			Scope:       "Default",
			Properties:  map[string]string{"foo": "bar"},
			Components: []ComponentSpec{{
				Name: "ComponentName",
			}},
			Constraints: "",
			Topologies: []TopologySpec{{
				Device: "DeviceName",
			}},
			ForceRedeploy: true,
		},
	}
	res, err := Target.DeepEquals(other)
	assert.Nil(t, err)
	assert.False(t, res)
}
