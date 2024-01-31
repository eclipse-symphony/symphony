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

func TestSkillMatch(t *testing.T) {
	s1 := SkillSpec{
		DisplayName: "skill",
		Parameters: map[string]string{
			"foo": "bar",
		},
		Nodes: []NodeSpec{
			{
				Id: "node1",
			},
			{
				Id: "node2",
			},
		},
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{
			{
				Role:     "role",
				Provider: "provider",
			},
		},
		Edges: []EdgeSpec{
			{
				Source: ConnectionSpec{
					Node:  "node1",
					Route: "route1",
				},
				Target: ConnectionSpec{
					Node:  "node2",
					Route: "route1",
				},
			},
		},
	}

	s2 := SkillSpec{
		DisplayName: "skill",
		Parameters: map[string]string{
			"foo": "bar",
		},
		Nodes: []NodeSpec{
			{
				Id: "node1",
			},
			{
				Id: "node2",
			},
		},
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []BindingSpec{
			{
				Role:     "role",
				Provider: "provider",
			},
		},
		Edges: []EdgeSpec{
			{
				Source: ConnectionSpec{
					Node:  "node1",
					Route: "route1",
				},
				Target: ConnectionSpec{
					Node:  "node2",
					Route: "route1",
				},
			},
		},
	}

	equal, err := s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestSkillNotMatch(t *testing.T) {
	s1 := SkillSpec{
		DisplayName: "skill",
	}

	// not match empty
	equal, err := s1.DeepEquals(nil)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different display name
	s2 := SkillSpec{
		DisplayName: "skill2",
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different parameters
	s2.DisplayName = "skill"
	s2.Parameters = map[string]string{
		"foo": "bar2",
	}
	s1.Parameters = map[string]string{
		"foo": "bar",
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different nodes
	s2.Parameters = map[string]string{
		"foo": "bar",
	}
	s2.Nodes = []NodeSpec{
		{
			Id: "node1-test",
		},
		{
			Id: "node2-test",
		},
	}
	s1.Nodes = []NodeSpec{
		{
			Id: "node1",
		},
		{
			Id: "node2",
		},
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different properties
	s2.Nodes = []NodeSpec{
		{
			Id: "node1",
		},
		{
			Id: "node2",
		},
	}
	s2.Properties = map[string]string{
		"foo": "bar2",
	}
	s1.Properties = map[string]string{
		"foo": "bar",
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different bindings
	s2.Properties = map[string]string{
		"foo": "bar",
	}
	s1.Bindings = []BindingSpec{
		{
			Role:     "role",
			Provider: "provider",
		},
	}
	s2.Bindings = []BindingSpec{
		{
			Role:     "role-test",
			Provider: "provider-test",
		},
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// not match different edges
	s2.Bindings = []BindingSpec{
		{
			Role:     "role",
			Provider: "provider",
		},
	}
	s1.Edges = []EdgeSpec{
		{
			Source: ConnectionSpec{
				Node:  "node1",
				Route: "route1",
			},
			Target: ConnectionSpec{
				Node:  "node2",
				Route: "route1",
			},
		},
	}
	s2.Edges = []EdgeSpec{
		{
			Source: ConnectionSpec{
				Node:  "node1-test",
				Route: "route1-test",
			},
			Target: ConnectionSpec{
				Node:  "node2-test",
				Route: "route1-test",
			},
		},
	}
	equal, err = s1.DeepEquals(s2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
