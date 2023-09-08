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
