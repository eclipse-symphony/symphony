/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStringWithOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "HIJ", val)
}

func TestReadStringWithNoOverrides(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "ABC", "")
	assert.Equal(t, "DEF", val)
}
func TestReadStringOverrideOnly(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"CDE": "HIJ",
	}, "CDE", "")
	assert.Equal(t, "HIJ", val)
}

func TestReadStringMissWithDefault(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "HIJ",
	}, "DEF", "HE")
	assert.Equal(t, "HE", val)
}
func TestReadStringEmptyOverride(t *testing.T) {
	val := ReadStringWithOverrides(map[string]string{
		"ABC": "DEF",
	}, map[string]string{
		"ABC": "",
	}, "DEF", "")
	assert.Equal(t, "", val)
}

func TestFormatObjectEmpty(t *testing.T) {
	obj := new(interface{})
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "null", string(val))
}
func TestFormatObjectEmptyDict(t *testing.T) {
	obj := map[string]interface{}{}
	val, err := FormatObject(obj, false, "", "")
	assert.Nil(t, err)
	assert.Equal(t, "{}", string(val))
}

func TestFormatObjectEmptyPath(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "", "json")
	assert.Nil(t, err)
	assert.Equal(t, "{\"foo\":\"bar\"}", string(val))
	val, err = FormatObject(obj, false, "", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "foo: bar\n", string(val))
}

func TestFormatObjectDictJson(t *testing.T) {
	obj := map[string]interface{}{
		"foo":    "bar",
		"man":    "what  ",
		"can":    "\"I\"",
		"say":    true,
		"number": 24,
		"mamba": map[string]interface{}{
			"out": "out",
		},
		"numberused": []int{8, 24},
	}
	val, err := FormatObject(obj, false, "$.foo", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"bar\"", string(val))
	val, err = FormatObject(obj, false, "$.man", "json")
	assert.Nil(t, err)
	assert.Equal(t, "\"what  \"", string(val))
	val, err = FormatObject(obj, false, "$.can", "json")
	assert.Nil(t, err)
	assert.Equal(t, "\"\\\"I\\\"\"", string(val))
	val, err = FormatObject(obj, false, "$.say", "json")
	assert.Nil(t, err)
	assert.Equal(t, "true", string(val))
	val, err = FormatObject(obj, false, "$.notexist", "json")
	assert.Nil(t, err)
	assert.Equal(t, "null", string(val))
	val, err = FormatObject(obj, false, "$.number", "json")
	assert.Nil(t, err)
	assert.Equal(t, "24", string(val))
	val, err = FormatObject(obj, false, "$.mamba", "json")
	assert.Nil(t, err)
	assert.Equal(t, "{\"out\":\"out\"}", string(val))
	val, err = FormatObject(obj, false, "$.mamba.out", "json")
	assert.Nil(t, err)
	assert.Equal(t, "\"out\"", string(val))
	val, err = FormatObject(obj, false, "$.numberused", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[8,24]", string(val))
}

func TestFormatObjectDictYaml(t *testing.T) {
	obj := map[string]interface{}{
		"foo":    "bar",
		"man":    "what  ",
		"can":    "\"I\"",
		"say":    true,
		"number": 24,
		"mamba": map[string]interface{}{
			"out": "out",
		},
		"numberused": []int{8, 24},
	}
	val, err := FormatObject(obj, false, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar\n", string(val))
	val, err = FormatObject(obj, false, "$.man", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "what\n", string(val))
	val, err = FormatObject(obj, false, "$.can", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "I\n", string(val))
	val, err = FormatObject(obj, false, "$.say", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "true\n", string(val))
	val, err = FormatObject(obj, false, "$.notexist", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n", string(val))
	val, err = FormatObject(obj, false, "$.number", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "24\n", string(val))
	val, err = FormatObject(obj, false, "$.mamba", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "out: out\n", string(val))
	val, err = FormatObject(obj, false, "$.mamba.out", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "out\n", string(val))
	val, err = FormatObject(obj, false, "$.numberused", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "- 8\n- 24\n", string(val))
}

func TestFormatObjectArrayDictJson(t *testing.T) {
	obj := []map[string]interface{}{
		{
			"foo": "bar1   ",
		},
		{
			"foo": "bar2",
		},
		{
			"notfoo": "bar",
		},
		{
			"number": 24,
		},
		{
			"question": []string{"man", "what", "can", "I", "say", "?"},
		},
		{
			"mamba": map[string]interface{}{
				"out": "out",
			},
		},
	}
	val, err := FormatObject(obj, true, "$.foo", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[\"bar1   \",\"bar2\",null,null,null,null]", string(val))
	val, err = FormatObject(obj, true, "$.bar", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,null,null,null,null]", string(val))
	val, err = FormatObject(obj, true, "$.notfoo", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,\"bar\",null,null,null]", string(val))
	val, err = FormatObject(obj, true, "$.number", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,null,24,null,null]", string(val))
	val, err = FormatObject(obj, true, "$.question", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,null,null,[\"man\",\"what\",\"can\",\"I\",\"say\",\"?\"],null]", string(val))
	val, err = FormatObject(obj, true, "$.mamba", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,null,null,null,{\"out\":\"out\"}]", string(val))
	val, err = FormatObject(obj, true, "$.mamba.out", "json")
	assert.Nil(t, err)
	assert.Equal(t, "[null,null,null,null,null,\"out\"]", string(val))
}

func TestFormatObjectArrayDictYaml(t *testing.T) {
	obj1 := []map[string]interface{}{
		{
			"foo": "bar1   ",
		},
		{
			"foo": "bar2",
		},
		{
			"notfoo": "bar",
		},
		{
			"number": 24,
		},
		{
			"question": []string{"man", "what", "can", "I", "say", "?"},
		},
		{
			"mamba": map[string]interface{}{
				"out": "out",
			},
		},
	}
	val, err := FormatObject(obj1, true, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar1\n---\nbar2\n---\nnull\n---\nnull\n---\nnull\n---\nnull\n", string(val))
	val, err = FormatObject(obj1, true, "$.bar", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nnull\n---\nnull\n---\nnull\n---\nnull\n", string(val))
	val, err = FormatObject(obj1, true, "$.notfoo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nbar\n---\nnull\n---\nnull\n---\nnull\n", string(val))
	val, err = FormatObject(obj1, true, "$.number", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nnull\n---\n24\n---\nnull\n---\nnull\n", string(val))
	val, err = FormatObject(obj1, true, "$.question", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nnull\n---\nnull\n---\n- man\n- what\n- can\n- I\n- say\n- '?'\n---\nnull\n", string(val))
	val, err = FormatObject(obj1, true, "$.mamba", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nnull\n---\nnull\n---\nnull\n---\nout: out\n", string(val))
	val, err = FormatObject(obj1, true, "$.mamba.out", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "null\n---\nnull\n---\nnull\n---\nnull\n---\nnull\n---\nout\n", string(val))
	obj2 := []map[string]interface{}{
		{
			"foo": "bar1",
		},
	}
	val, err = FormatObject(obj2, true, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar1\n", string(val))
}

func TestFormatObjectFirstEmbeddedPath(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"components": []map[string]interface{}{
				{
					"properties": map[string]interface{}{
						"embedded": "value",
					},
				},
			},
		},
	}
	val, err := FormatObject(obj, false, "first_embedded", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"value\"", string(val))
	val, err = FormatObject(obj, false, "first_embedded", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "value\n", string(val))
}
func TestJsonPathBasic(t *testing.T) {
	jData := `
	{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "$[?(@.status=='queued')].status")
	assert.Nil(t, err)
	assert.Equal(t, "queued", result)
}

func TestJsonPathBasicDirectQuery(t *testing.T) {
	jData := `
	{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "$.status")
	assert.Nil(t, err)
	assert.Equal(t, "queued", result)
}

func TestJsonPathObjectInArray(t *testing.T) {
	jData := `
	[{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}]`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "$[?(@.status=='queued')].status")
	assert.Nil(t, err)
	assert.Equal(t, "queued", result)
}
func TestJsonPathQueryInBracket(t *testing.T) {
	jData := `
	[{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}]`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "{$[?(@.status=='queued')].status}")
	assert.Nil(t, err)
	assert.Equal(t, "queued", result)
}
func TestJsonPathInvalidJsonPath(t *testing.T) {
	jData := `
	[{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}]`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "{$[?(@.status=='queued')].status")
	assert.NotNil(t, err)
	assert.Equal(t, nil, result)
}
func TestJsonPathBadJsonPath(t *testing.T) {
	jData := `
	[{
		"id": 30433642,
		"name": "Build",
		"head_branch": "main",
		"run_number": 562,
		"event": "push",
		"display_title": "Update README.md",
		"status": "queued",
		"conclusion": null,
		"workflow_id": 159038
	}]`
	var obj interface{}
	json.Unmarshal([]byte(jData), &obj)
	result, err := JsonPathQuery(obj, "sdgsgsdg")
	assert.NotNil(t, err)
	assert.Equal(t, nil, result)
}
func TestMatchString(t *testing.T) {
	assert.True(t, matchString("abc", "abc"))
	assert.True(t, matchString("a.*c", "abc"))
	assert.True(t, matchString("a%", "abc"))
	assert.False(t, matchString("a.*c", "ab"))
}
func TestReadInt32(t *testing.T) {
	mapData := map[string]string{
		"abc": "#123",
		"def": "N/A",
		"mno": "#N/A",
	}
	val := ReadInt32(mapData, "abc", 0)
	assert.Equal(t, int32(123), val)
	val = ReadInt32(mapData, "def", 0)
	assert.Equal(t, int32(0), val)
	val = ReadInt32(mapData, "mno", 0)
	assert.Equal(t, int32(0), val)
}
func TestGetString(t *testing.T) {
	mapData := map[string]string{
		"a": "def",
		"b": "{abc}",
		"c": `[{"key":"value"}]`,
	}
	val, err := GetString(mapData, "a")
	assert.Nil(t, err)
	assert.Equal(t, "def", val)

	val, err = GetString(mapData, "b")
	assert.NotNil(t, err)

	val, err = GetString(mapData, "c")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Bad Config: value of c is not a string")

	val, err = GetString(mapData, "d")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Bad Config: key d is not found")
}
func TestReadStringFromMapCompat(t *testing.T) {
	mapData := map[string]interface{}{
		"a": "def",
		"b": []map[string]string{
			{
				"key": "value",
			},
		},
	}
	val := ReadStringFromMapCompat(mapData, "a", "")
	assert.Equal(t, "def", val)

	val = ReadStringFromMapCompat(mapData, "b", "")
	assert.Equal(t, "", val)

	val = ReadStringFromMapCompat(mapData, "c", "")
	assert.Equal(t, "", val)
}

func TestReadString(t *testing.T) {
	mapData := map[string]string{
		"a": "def",
		"b": "{abc}",
		"c": `[{"key":"value"}]`,
	}
	val := ReadString(mapData, "a", "")
	assert.Equal(t, "def", val)

	val = ReadString(mapData, "b", "")
	assert.Equal(t, "", val)

	val = ReadString(mapData, "c", "")
	assert.Equal(t, "", val)

	val = ReadString(mapData, "d", "")
	assert.Equal(t, "", val)
}
func TestMergeCollection(t *testing.T) {
	mapData := map[string]string{
		"a": "def",
		"b": "abc",
	}
	mapData2 := map[string]string{
		"c": "def",
		"d": "abc",
	}
	merged := MergeCollection(mapData, mapData2)
	assert.Equal(t, 4, len(merged))
	assert.Equal(t, "def", merged["a"])
	assert.Equal(t, "abc", merged["b"])
	assert.Equal(t, "def", merged["c"])
	assert.Equal(t, "abc", merged["d"])
}

func TestDeepCopyCollection(t *testing.T) {
	// Test 1: DeepCopy then modify original map
	mapData := map[string]interface{}{
		"a": "def",
		"b": "abc",
	}
	copied := DeepCopyCollection(mapData)
	assert.Equal(t, 2, len(copied))
	assert.Equal(t, "def", copied["a"])
	assert.Equal(t, "abc", copied["b"])
	// Modify original map to ensure deep copy
	mapData["c"] = "xyz"
	assert.Equal(t, 2, len(copied)) // copied should not change
	assert.NotContains(t, copied, "c")

	// Test 2: DeepCopy with exclude key
	copied_exclude := DeepCopyCollection(mapData, "b")
	assert.Equal(t, 2, len(copied_exclude))
	assert.Equal(t, "def", copied_exclude["a"])
	assert.NotContains(t, copied_exclude, "b")

	// Test 3: DeepCopy with multiple exclude keys
	copied_exclude_multi := DeepCopyCollection(mapData, "b", "c")
	assert.Equal(t, 1, len(copied_exclude_multi))
	assert.Equal(t, "def", copied_exclude_multi["a"])
	assert.NotContains(t, copied_exclude_multi, "b")
	assert.NotContains(t, copied_exclude_multi, "c")

	// Test 4: DeepCopy with nil map
	copied_empty := DeepCopyCollection(nil)
	assert.NotNil(t, copied_empty)
	assert.Equal(t, 0, len(copied_empty)) // should be an empty map
}

func TestDeepCopyCollectionWithPrefixExclude(t *testing.T) {
	mapData := map[string]interface{}{
		"prefix_a": "def",
		"prefix_b": "abc",
		"c":        "xyz",
	}

	copied := DeepCopyCollectionWithPrefixExclude(mapData, "prefix")
	assert.Equal(t, 1, len(copied))
	assert.Equal(t, "xyz", copied["c"])
	assert.NotContains(t, copied, "prefix_a")
	assert.NotContains(t, copied, "prefix_b")

	copied2 := DeepCopyCollectionWithPrefixExclude(mapData)
	assert.Equal(t, 3, len(copied2))
	assert.Equal(t, "def", copied2["prefix_a"])
	assert.Equal(t, "abc", copied2["prefix_b"])
	assert.Equal(t, "xyz", copied2["c"])
	mapData["d"] = "new_value"
	assert.Contains(t, mapData, "d")
	assert.Equal(t, 3, len(copied2)) // copied2 should not change
	assert.NotContains(t, copied2, "d")

	copied3 := DeepCopyCollectionWithPrefixExclude(nil)
	assert.NotNil(t, copied3)
	assert.Equal(t, 0, len(copied3)) // should be an empty map
}

func TestMergeCollection_StringAny(t *testing.T) {
	mapData := map[string]interface{}{
		"a": "def",
		"b": "abc",
	}
	mapData2 := map[string]interface{}{
		"c": "def",
		"d": "abc",
	}
	merged := MergeCollection_StringAny(mapData, mapData2)
	assert.Equal(t, 4, len(merged))
	assert.Equal(t, "def", merged["a"])
	assert.Equal(t, "abc", merged["b"])
	assert.Equal(t, "def", merged["c"])
	assert.Equal(t, "abc", merged["d"])

	merged2 := MergeCollection_StringAny(mapData, nil)
	assert.Equal(t, 2, len(merged2))
	assert.Equal(t, "def", merged2["a"])
	assert.Equal(t, "abc", merged2["b"])

	merged3 := MergeCollection_StringAny(nil)
	assert.Equal(t, 0, len(merged3))

	mapData3_override := map[string]interface{}{
		"a": "xyz",
	}
	merged4 := MergeCollection_StringAny(mapData, mapData3_override)
	assert.Equal(t, 2, len(merged4))
	assert.Equal(t, "xyz", merged4["a"])
	assert.Equal(t, "abc", merged4["b"])
}

func TestCollectStringMap(t *testing.T) {
	mapData := map[string]string{
		"a1": "def",
		"a2": "abc",
		"b":  "xxx",
	}

	merged := CollectStringMap(mapData, "a")
	assert.Equal(t, 2, len(merged))
	assert.Equal(t, "def", merged["a1"])
	assert.Equal(t, "abc", merged["a2"])
}

func TestParseValue(t *testing.T) {
	// bool = $true
	v, err := ParseValue("$true")
	assert.Nil(t, err)
	assert.Equal(t, true, v)

	// bool = $false
	v, err = ParseValue("$false")
	assert.Nil(t, err)
	assert.Equal(t, false, v)

	// environment variable = $foo
	os.Setenv("foo", "bar")
	v, err = ParseValue("$foo")
	assert.Nil(t, err)
	assert.Equal(t, "bar", v)

	v, err = ParseValue("$foo1")
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

func TestFormatAsString(t *testing.T) {
	assert.Equal(t, "abc", FormatAsString("abc"))
	assert.Equal(t, "123", FormatAsString(123))
	assert.Equal(t, "123", FormatAsString(int32(123)))
	assert.Equal(t, "123", FormatAsString(int64(123)))
	assert.Equal(t, "123.456", FormatAsString(float32(123.456)))
	assert.Equal(t, "123.456", FormatAsString(float64(123.456)))
	assert.Equal(t, "true", FormatAsString(true))

	obj := map[string]interface{}{
		"foo": "bar",
		"abc": 123,
	}
	ret, _ := json.Marshal(obj)
	assert.Equal(t, string(ret), FormatAsString(obj))

	obj2 := []interface{}{
		"foo",
		123,
	}
	ret, _ = json.Marshal(obj2)
	assert.Equal(t, string(ret), FormatAsString(obj2))

	type customType struct {
		Foo string `json:"foo"`
	}
	obj3 := customType{
		Foo: "bar",
	}
	assert.Equal(t, fmt.Sprintf("%v", obj3), FormatAsString(obj3))
}

func TestToInterfaceMap(t *testing.T) {
	m := map[string]string{
		"foo": "bar",
		"abc": "123",
	}
	m2 := toInterfaceMap(m)
	assert.Equal(t, "bar", m2["foo"])
	assert.Equal(t, "123", m2["abc"])
}

func TestReadEmptyStringProperty(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "foo",
			},
		},
	}
	m2, ok := JsonParseProperty(m, "")
	assert.False(t, ok)
	assert.Equal(t, nil, m2)
}

func TestReadIilFormatStringProperty(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "foo",
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`")
	assert.False(t, ok)
	assert.Equal(t, nil, m2)
}

func TestReadRootProperty(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "foo",
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.`")
	assert.True(t, ok)
	assert.Equal(t, m, m2)
}

func TestReadNestedJsonStringProperty(t *testing.T) {
	value := "123"
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": value,
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.a.b.c`")
	assert.True(t, ok)
	assert.Equal(t, value, m2)
}

func TestReadNestedJsonNumberProperty(t *testing.T) {
	value := 123
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": value,
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.a.b.c`")
	assert.True(t, ok)
	assert.Equal(t, value, m2)
}

func TestReadNestedJsonPropertyNotExsits(t *testing.T) {
	value := 123
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": value,
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.a.b.d`")
	assert.False(t, ok)
	assert.Equal(t, m2, nil)
}

func TestReadNestedJsonPropertyThrowError(t *testing.T) {
	value := 123
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": value,
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.a..b.c`")
	assert.False(t, ok)
	assert.Equal(t, m2, nil)
}

func TestReadMiddleProperty(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "foo",
			},
		},
	}
	m2, ok := JsonParseProperty(m, "`.a.b`")
	assert.True(t, ok)
	assert.Equal(t, m["a"].(map[string]interface{})["b"], m2)
}

func TestReadPropertyNameWithDotIdentifier(t *testing.T) {
	value := "123"
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b.c": map[string]interface{}{
				"c": value,
			},
		},
		"a.b.c": value,
	}
	m2, ok := JsonParseProperty(m, "`.a.[\"b.c\"].c`")
	assert.True(t, ok)
	assert.Equal(t, value, m2)

	m3, ok := JsonParseProperty(m, "`.\"a.b.c\"`")
	assert.True(t, ok)
	assert.Equal(t, value, m3)
}

func TestReadPropertyNameWithDotIdentifierAndQuotationMark(t *testing.T) {
	value := "123"
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b.c": value,
		},
	}

	m2, ok := JsonParseProperty(m, "`.a.[\"b.c\"]`")
	assert.True(t, ok)
	assert.Equal(t, value, m2)

	m3, ok := JsonParseProperty(m, "`.a.\"b.c\"`")
	assert.True(t, ok)
	assert.Equal(t, value, m3)
}

func TestReadPropertyNameWithArraySlicing(t *testing.T) {
	jsonData := `{
		"a": {
			"b": [
				{"id": 1},
				{"id": 2},
				{"id": 3}
			]
		}
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		log.Fatal(err)
	}

	obj := map[string]interface{}(map[string]interface{}{"id": 1.})
	val := 3.

	m2, ok := JsonParseProperty(data, "`.a.b[0]`")
	assert.True(t, ok)
	assert.Equal(t, obj, m2)

	m3, ok := JsonParseProperty(data, "`.a.b[] | select(.id > 2) | .id`")
	assert.True(t, ok)
	assert.Equal(t, val, m3)
}
