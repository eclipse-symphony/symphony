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

func TestNodeEqual(t *testing.T) {
	node1 := NodeSpec{
		Id:       "id",
		NodeType: "type",
		Name:     "name",
		Model:    "model",
		Configurations: map[string]string{
			"foo": "bar",
		},
		Inputs: []RouteSpec{
			{
				Route: "route",
				Type:  "type",
				Properties: map[string]string{
					"foo": "bar",
				},
				Filters: []FilterSpec{
					{
						Direction: "direction",
						Type:      "type",
						Parameters: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		},
	}

	node2 := NodeSpec{
		Id:       "id",
		NodeType: "type",
		Name:     "name",
		Model:    "model",
		Configurations: map[string]string{
			"foo": "bar",
		},
		Inputs: []RouteSpec{
			{
				Route: "route",
				Type:  "type",
				Properties: map[string]string{
					"foo": "bar",
				},
				Filters: []FilterSpec{
					{
						Direction: "direction",
						Type:      "type",
						Parameters: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
		},
	}

	equal, err := node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestNodeNotEqual(t *testing.T) {
	node1 := NodeSpec{
		Id: "id",
	}

	// Test empty
	equal, err := node1.DeepEquals(nil)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Id not equal
	node2 := NodeSpec{
		Id: "id2",
	}
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test NodeType not equal
	node2.Id = "id"
	node2.NodeType = "type2"
	node1.NodeType = "type"
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Name not equal
	node2.NodeType = "type"
	node2.Name = "name2"
	node1.Name = "name"
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Model not equal
	node2.Name = "name"
	node2.Model = "model2"
	node1.Model = "model"
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Configurations not equal
	node2.Model = "model"
	node2.Configurations = map[string]string{
		"foo": "bar2",
	}
	node1.Configurations = map[string]string{
		"foo": "bar",
	}
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Inputs not equal
	node2.Configurations = map[string]string{
		"foo": "bar",
	}
	node2.Inputs = []RouteSpec{
		{
			Route: "route2",
		},
	}
	node1.Inputs = []RouteSpec{
		{
			Route: "route",
		},
	}
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Outputs not equal
	node2.Inputs = []RouteSpec{
		{
			Route: "route",
		},
	}
	node2.Outputs = []RouteSpec{
		{
			Route: "route2",
		},
	}
	node1.Outputs = []RouteSpec{
		{
			Route: "route",
		},
	}
	equal, err = node1.DeepEquals(node2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
