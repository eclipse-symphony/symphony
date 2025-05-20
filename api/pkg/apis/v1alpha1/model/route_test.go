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

func TestRouteMatch(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestRouteRouteNotMatch(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
		Route: "route2",
		Type:  "type",
		Properties: map[string]string{
			"foo": "bar",
		},
		Filters: []FilterSpec{
			{
				Direction: "direction",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRouteTypeNotMatch(t *testing.T) {
	route1 := RouteSpec{
		Route: "route",
		Type:  "type2",
		Properties: map[string]string{
			"foo": "bar",
		},
		Filters: []FilterSpec{
			{
				Direction: "direction",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRoutePropertiesNotMatch(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
		Route: "route",
		Type:  "type",
		Properties: map[string]string{
			"foo": "bar2",
		},
		Filters: []FilterSpec{
			{
				Direction: "direction",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRouteExtraProperties(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
		Route: "route",
		Type:  "type",
		Properties: map[string]string{
			"foo":  "bar",
			"foo2": "bar2",
		},
		Filters: []FilterSpec{
			{
				Direction: "direction",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRouteExtraFilter(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
			{
				Direction: "direction2",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRouteMissingFilter(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
			{
				Direction: "direction2",
				Type:      "type",
				Parameters: map[string]string{
					"foo2": "bar2",
				},
			},
		},
	}
	route2 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	equal, err := route1.DeepEquals(route2)
	assert.Nil(t, err)
	assert.False(t, equal)
}

func TestRouteMatchOneEmpty(t *testing.T) {
	route1 := RouteSpec{
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
					"foo2": "bar2",
				},
			},
		},
	}
	res, err := route1.DeepEquals(nil)
	assert.EqualError(t, err, "parameter is not a RouteSpec type")
	assert.False(t, res)
}
