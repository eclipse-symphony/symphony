/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareEmptyMaps(t *testing.T) {
	assert.True(t, StringMapsEqual(nil, nil, nil))
}

func TestCompareEmptyVSOne(t *testing.T) {
	assert.False(t, StringMapsEqual(nil, map[string]string{
		"A": "B",
	}, nil))
}
func TestCompareDifferentSizes(t *testing.T) {
	assert.False(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"A": "B",
	}, nil))
}
func TestCompareSameSizeDifferentKeys(t *testing.T) {
	assert.False(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"A": "B",
		"D": "C",
	}, nil))
}
func TestCompareEqual(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, nil))
}
func TestCompareEqualWithInstanceFunc(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "$instance()",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, nil))
}
func TestCompareEqualWithIgnoredKeysLeft(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
		"E": "F",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, []string{"E"}))
}
func TestCompareEqualWithIgnoredKeysRight(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
		"E": "F",
	}, []string{"E"}))
}
func TestEmptyComponentSpecSliceDeepEquals(t *testing.T) {
	c1 := []ComponentSpec{}
	c2 := []ComponentSpec{}
	assert.True(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNameOnlySliceDeepEquals(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	assert.True(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNameOnlySliceDeepNotEquals(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "def",
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesSliceDeepEquals(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	assert.True(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesSliceDeepNotEqual(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bb",
				"ccc": "ddd",
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceDeepEquals(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.True(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV1(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jc",
							},
						},
					},
				},
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV2(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "out",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV3(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "TYPE",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV4(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "if",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV5(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url2",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.False(t, SlicesEqual(c1, c2))
}
func TestComponentSpecNamePropertiesRoutesSliceNotDeepEqualV6(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
			Routes: []RouteSpec{
				{
					Route: "url1",
					Properties: map[string]string{
						"ia": "ib",
						"ic": "id",
					},
					Filters: []FilterSpec{
						{
							Direction: "in",
							Type:      "type",
							Parameters: map[string]string{
								"ja": "jb",
							},
						},
					},
				},
			},
		},
	}
	assert.True(t, SlicesEqual(c1, c2))
}

// TestEmptyComponentSpecSliceCover tests if c1 covers c2 when both slices are empty
func TestEmptyComponentSpecSliceCover(t *testing.T) {
	c1 := []ComponentSpec{}
	c2 := []ComponentSpec{}
	assert.True(t, SlicesCover(c1, c2))
}

// TestComponentSpecNameOnlySliceCover tests if c1 covers c2 when they have same components
// True should be returned in this case
func TestComponentSpecNameOnlySliceCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	assert.True(t, SlicesCover(c1, c2))
}

// TestComponentSpecNameOnlySliceNotCover tests if c1 covers c2 when they have different components
// False should be returned in this case
func TestComponentSpecNameOnlySliceNotCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "def",
		},
	}
	assert.False(t, SlicesCover(c1, c2))
}

// TestComponentSpecNameOnlySliceNotCover tests if c1 covers c2 when they have same components with same properties
// True should be returned in this case
func TestComponentSpecNamePropertiesSliceCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	assert.True(t, SlicesCover(c1, c2))
}

// TestComponentSpecNamePropertiesSliceNotCover tests if c1 covers c2 when they have different properties
// False should be returned in this case
func TestComponentSpecNamePropertiesSliceNotCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]string{
				"aaa": "bb",
				"ccc": "ddd",
			},
		},
	}
	assert.False(t, SlicesCover(c1, c2))
}

// TestSlicesCoverMissingComponent tests if c1 covers c2 when c2 doesn't have c1 components
// False should be returned as c2 doesn't contain c1 components
func TestSlicesCoverMissingComponent(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{}
	assert.False(t, SlicesCover(c1, c2))
}

func TestSlicesCoverNil(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	var c2 []ComponentSpec
	assert.False(t, SlicesCover(c1, c2))
}

// TestSlicesCoverMissingComponentMultiple tests if c1 covers c2 when c2 contains 1 c1 component
func TestSlicesCoverMissingComponentMultiple(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
		{
			Name: "def",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	assert.False(t, SlicesCover(c1, c2))
}

// TestSlicesCoverMissingComponentExtra tests if c1 covers c2 when c2 doesn't have any c1 components
// but has an extra component by itself
// False should be returned as c2 doesn't contain any c1 components
func TestSlicesCoverMissingComponentExtra(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
		{
			Name: "def",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "hij",
		},
	}
	assert.False(t, SlicesCover(c1, c2))
}

func TestSlicesCoverWithExtra(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
		{
			Name: "def",
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
		},
		{
			Name: "def",
		},
		{
			Name: "hij",
		},
	}
	assert.True(t, SlicesCover(c1, c2))
}
func TestFullComponentSpecCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "symphony-agent",
			Properties: map[string]string{
				"container.createOptions": "",
				"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
				"container.restartPolicy": "always",
				"container.type":          "docker",
				"container.version":       "1.0",
				"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
				"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
				"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
				"env.STORAGE_ACCOUNT":     "voestore",
				"env.STORAGE_CONTAINER":   "snapshots",
				"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
				"env.TARGET_NAME":         "symphony-k8s-target",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "symphony-agent",
			Properties: map[string]string{
				"container.createOptions": "",
				"container.image":         "possprod.azurecr.io/symphony-agent:0.39.9",
				"container.restartPolicy": "always",
				"container.type":          "docker",
				"container.version":       "1.0",
				"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
				"env.AZURE_CLIENT_SECRET": "\\u003cSP Client Secret\\u003e",
				"env.AZURE_TENANT_ID":     "\\u003cSP Tenant ID\\u003e",
				"env.STORAGE_ACCOUNT":     "voestore",
				"env.STORAGE_CONTAINER":   "snapshots",
				"env.SYMPHONY_URL":        "http://20.118.178.8:8080/v1alpha2/agent/references",
				"env.TARGET_NAME":         "symphony-k8s-target",
			},
		},
	}
	assert.True(t, SlicesCover(c1, c2))
}

func TestCollectProperties(t *testing.T) {
	properties := map[string]string{
		"helm.values.abc": "ABC",
		"helm.values.def": "DEF",
	}
	ret := CollectPropertiesWithPrefix(properties, "helm.values.", nil)
	assert.Equal(t, "ABC", ret["abc"])
	assert.Equal(t, "DEF", ret["def"])
}
