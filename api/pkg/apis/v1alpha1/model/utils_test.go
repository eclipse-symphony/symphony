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

func TestStringMapsEqualCompareEmptyMaps(t *testing.T) {
	assert.True(t, StringMapsEqual(nil, nil, nil))
}

func TestStringMapsEqualCompareEmptyVSOne(t *testing.T) {
	assert.False(t, StringMapsEqual(nil, map[string]string{
		"A": "B",
	}, nil))
}

func TestStringMapsEqualCompareDifferentSizes(t *testing.T) {
	assert.False(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"A": "B",
	}, nil))
}

func TestStringMapsEqualCompareSameSizeDifferentKeys(t *testing.T) {
	assert.False(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"A": "B",
		"D": "C",
	}, nil))
}

func TestStringMapsEqualCompareEqual(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, nil))
}

func TestStringMapsEqualCompareEqualWithInstanceFunc(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "$instance()",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, nil))
}

func TestStringMapsEqualCompareEqualWithIgnoredKeysLeft(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
		"E": "F",
	}, map[string]string{
		"C": "D",
		"A": "B",
	}, []string{"E"}))
}

func TestStringMapsEqualCompareEqualWithIgnoredKeysRight(t *testing.T) {
	assert.True(t, StringMapsEqual(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
		"E": "F",
	}, []string{"E"}))
}

func TestStringStringMapsEqualCompareEmptyMaps(t *testing.T) {
	assert.True(t, StringStringMapsEqual(nil, nil, nil))
}

func TestStringStringMapsEqualCompareEmptyVSOne(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}
	assert.False(t, StringStringMapsEqual(outerMapA, nil, nil))
}

func TestStringStringMapsEqualCompareDifferentSizes(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}
	outerMapB := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
	}

	assert.False(t, StringStringMapsEqual(outerMapA, outerMapB, nil))
}

func TestStringStringMapsEqualCompareSameSizeDifferentKeys(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
	}
	outerMapB := map[string]map[string]string{
		"B": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
	}
	assert.False(t, StringStringMapsEqual(outerMapA, outerMapB, nil))
}

func TestStringStringMapsEqualCompareEqual(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}
	outerMapB := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}

	assert.True(t, StringStringMapsEqual(outerMapA, outerMapB, nil))
}

func TestStringStringMapsEqualCompareEqualWithInstanceFunc(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
		},
		"B": {
			"foo1": "bar1",
		},
		"C": {
			"foo1": "$instance()",
		},
	}
	outerMapB := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
		},
		"B": {
			"foo1": "bar1",
		},
		"C": {
			"foo1": "bar1",
		},
	}
	assert.True(t, StringStringMapsEqual(outerMapA, outerMapB, nil))
}

func TestStringStringMapsEqualCompareEqualWithIgnoredKeysLeft(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}
	outerMapB := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}

	assert.True(t, StringStringMapsEqual(outerMapA, outerMapB, []string{"C"}))
}

func TestStringStringMapsEqualCompareEqualWithIgnoredKeysRight(t *testing.T) {
	outerMapA := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
		"C": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}
	outerMapB := map[string]map[string]string{
		"A": {
			"foo1": "bar1",
			"foo2": "bar1",
		},
		"B": {
			"foo1": "bar1",
			"foo2": "bar2",
		},
	}

	assert.True(t, StringStringMapsEqual(outerMapA, outerMapB, []string{"C"}))
}

func TestEnvMapsEmptyMaps(t *testing.T) {
	assert.True(t, EnvMapsEqual(nil, nil))
}

func TestEnvMapsCompareDifferentSizes(t *testing.T) {
	assert.True(t, EnvMapsEqual(map[string]string{
		"env.AZURE_CLIENT_ID":   "\\u003cSP App ID\\u003e",
		"env.AZURE_TENANT_ID":   "\\u003cSP Tenant ID\\u003e",
		"env.STORAGE_ACCOUNT":   "voestore",
		"env.STORAGE_CONTAINER": "snapshots",
		"env.SYMPHONY_URL":      "http://20.118.178.8:8080/v1alpha2/agent/references",
		"env.TARGET_NAME":       "symphony-k8s-target",
	}, map[string]string{
		"env.AZURE_CLIENT_ID":   "\\u003cSP App ID\\u003e",
		"env.AZURE_TENANT_ID":   "\\u003cSP Tenant ID\\u003e",
		"env.STORAGE_ACCOUNT":   "voestore",
		"env.STORAGE_CONTAINER": "snapshots",
		"env.SYMPHONY_URL":      "http://20.118.178.8:8080/v1alpha2/agent/references",
	}))
}

func TestEnvMapsCompareSameSizeDifferentKeys(t *testing.T) {
	assert.True(t, EnvMapsEqual(map[string]string{
		"env.CLIENT_ID": "\\u003cSP App ID\\u003e",
		"env.TENANT_ID": "\\u003cSP Tenant ID\\u003e",
		"env.ACCOUNT":   "voestore",
		"env.CONTAINER": "snapshots",
		"env.URL":       "http://20.118.178.8:8080/v1alpha2/agent/references",
		"env.NAME":      "symphony-k8s-target",
	}, map[string]string{
		"env.AZURE_CLIENT_ID":   "\\u003cSP App ID\\u003e",
		"env.AZURE_TENANT_ID":   "\\u003cSP Tenant ID\\u003e",
		"env.STORAGE_ACCOUNT":   "voestore",
		"env.STORAGE_CONTAINER": "snapshots",
		"env.SYMPHONY_URL":      "http://20.118.178.8:8080/v1alpha2/agent/references",
		"env.TARGET_NAME":       "symphony-k8s-target",
	}))
}

func TestEnvMapsCompareDifferentSizesWithTarget(t *testing.T) {
	assert.True(t, EnvMapsEqual(map[string]string{
		"env.AZURE_CLIENT_ID":   "\\u003cSP App ID\\u003e",
		"env.STORAGE_CONTAINER": "snapshots",
		"env.TARGET_NAME":       "$target()",
	}, map[string]string{
		"env.AZURE_CLIENT_ID": "\\u003cSP App ID\\u003e",
		"env.TARGET_NAME":     "someRandomName",
	}))
}

func TestEnvMapsCompareDifferentSizesWithoutTarget(t *testing.T) {
	assert.False(t, EnvMapsEqual(map[string]string{
		"env.AZURE_CLIENT_ID":   "\\u003cSP App ID\\u003e",
		"env.STORAGE_CONTAINER": "snapshots",
		"env.TARGET_NAME":       "someRandomName",
	}, map[string]string{
		"env.AZURE_CLIENT_ID": "\\u003cSP App ID\\u003e",
		"env.TARGET_NAME":     "someOtherRandomName",
	}))
}

func TestEnvMapsCompareEqual(t *testing.T) {
	assert.True(t, EnvMapsEqual(map[string]string{
		"A":               "B",
		"env.TARGET_NAME": "someRandomName",
	}, map[string]string{
		"A":               "B",
		"env.TARGET_NAME": "someRandomName",
	}))
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
			Properties: map[string]interface{}{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]interface{}{
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
			Properties: map[string]interface{}{
				"aaa": "bbb",
				"ccc": "ddd",
			},
		},
	}
	c2 := []ComponentSpec{
		{
			Name: "abc",
			Properties: map[string]interface{}{
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

func TestSlicesAnySameSize(t *testing.T) {
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
	assert.True(t, SlicesAny(c1, c2))
}

func TestSlicesAnyMissingComponent(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	c2 := []ComponentSpec{}
	assert.False(t, SlicesAny(c1, c2))
}

func TestSlicesAnyNil(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "abc",
		},
	}
	var c2 []ComponentSpec
	assert.False(t, SlicesAny(c1, c2))
}

func TestCheckPropertyNil(t *testing.T) {
	assert.True(t, CheckProperty(nil, nil, "", false))
}

func TestCheckPropertyOneEmptyWithKeyExisting(t *testing.T) {
	assert.False(t, CheckProperty(map[string]string{
		"A": "B",
		"C": "D",
	}, nil, "A", false))
}

func TestCheckPropertyOneEmptyWithoutKeyExisting(t *testing.T) {
	assert.True(t, CheckProperty(map[string]string{
		"A": "B",
		"C": "D",
	}, nil, "E", false))
}

func TestCheckPropertyWithKeyExisting(t *testing.T) {
	assert.True(t, CheckProperty(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
		"E": "F",
	}, "A", false))
}

func TestCheckPropertyWithoutKeyExisting(t *testing.T) {
	assert.True(t, CheckProperty(map[string]string{
		"A": "B",
		"C": "D",
	}, map[string]string{
		"C": "D",
		"A": "B",
		"E": "F",
	}, "W", false))
}

func TestCheckPropertyWithoutIgnoreCase1(t *testing.T) {
	assert.True(t, CheckProperty(map[string]string{
		"A": "B",
	}, map[string]string{
		"A": "B",
	}, "A", true))
}

func TestCheckPropertyWithoutIgnoreCase2(t *testing.T) {
	assert.False(t, CheckProperty(map[string]string{
		"A": "B",
	}, map[string]string{
		"A": "C",
	}, "A", true))
}

func TestFullComponentSpecCover(t *testing.T) {
	c1 := []ComponentSpec{
		{
			Name: "symphony-agent",
			Properties: map[string]interface{}{
				"container.createOptions": "",
				ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
				"container.restartPolicy": "always",
				"container.type":          "docker",
				"container.version":       "1.0",
				"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
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
			Properties: map[string]interface{}{
				"container.createOptions": "",
				ContainerImage:            "possprod.azurecr.io/symphony-agent:0.39.9",
				"container.restartPolicy": "always",
				"container.type":          "docker",
				"container.version":       "1.0",
				"env.AZURE_CLIENT_ID":     "\\u003cSP App ID\\u003e",
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

func TestCollectPropertiesWithoutHierarchy(t *testing.T) {
	properties := map[string]interface{}{
		"helm.values.abc":     "ABC",
		"helm.values.def.geh": "DEF",
	}
	ret := CollectPropertiesWithPrefix(properties, "helm.values.", nil, false)
	assert.Equal(t, "ABC", ret["abc"])
	assert.Equal(t, "DEF", ret["def.geh"])
}

func TestCollectPropertiesWithHierarchy(t *testing.T) {
	properties := map[string]interface{}{
		"helm.values.abc":             "ABC",
		"helm.values.def.geh":         "DEF",
		"helm.values.def.somebool":    "true",
		"helm.values.def.some[0].int": "123",
	}
	ret := CollectPropertiesWithPrefix(properties, "helm.values.", nil, true)
	assert.Equal(t, "ABC", ret["abc"])
	assert.IsType(t, map[string]interface{}{}, ret["def"])
	assert.Equal(t, "DEF", ret["def"].(map[string]interface{})["geh"])
	assert.Equal(t, true, ret["def"].(map[string]interface{})["somebool"])
	assert.IsType(t, []interface{}{}, ret["def"].(map[string]interface{})["some"])
	assert.Equal(t, int64(123), ret["def"].(map[string]interface{})["some"].([]interface{})[0].(map[string]interface{})["int"])
}
