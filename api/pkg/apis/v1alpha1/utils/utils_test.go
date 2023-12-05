/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
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
