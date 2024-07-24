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
func TestFormatObjectDictJson(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "")
	assert.Nil(t, err)
	assert.Equal(t, "\"bar\"", string(val))
}
func TestFormatObjectDictYaml(t *testing.T) {
	obj := map[string]interface{}{
		"foo": "bar",
	}
	val, err := FormatObject(obj, false, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar\n", string(val))
}
func TestFormatObjectArrayDictJson(t *testing.T) {
	obj := []map[string]interface{}{
		{
			"foo": "bar1",
		},
		{
			"foo": "bar2",
		},
		{
			"notfoo": "bar",
		},
	}
	val, err := FormatObject(obj, true, "$.foo", "")
	assert.Nil(t, err)
	assert.Equal(t, "[\"bar1\",\"bar2\",null]", string(val))
}

func TestFormatObjectArrayDictYaml(t *testing.T) {
	obj := []map[string]interface{}{
		{
			"foo": "bar1",
		},
		{
			"foo": "bar2",
		},
		{
			"notfoo": "bar",
		},
	}
	val, err := FormatObject(obj, true, "$.foo", "yaml")
	assert.Nil(t, err)
	assert.Equal(t, "bar1\n---\nbar2\n---\nnull\n", string(val))
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
