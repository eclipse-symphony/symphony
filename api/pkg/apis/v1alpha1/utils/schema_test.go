/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIntType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"int": Rule{
				Type: "int",
			},
		},
	}
	properties := map[string]interface{}{
		"int": "1",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestCheckIntTypeNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"int": Rule{
				Type: "int",
			},
		},
	}
	properties := map[string]interface{}{
		"int": "1.1",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["int"].Valid == false)
	assert.True(t, result.Errors["int"].Error != "")
}
func TestCheckFloatType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"float": Rule{
				Type: "float",
			},
		},
	}
	properties := map[string]interface{}{
		"float": "1.1",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestCheckFloatTypeNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"float": Rule{
				Type: "float",
			},
		},
	}
	properties := map[string]interface{}{
		"float": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["float"].Valid == false)
	assert.True(t, result.Errors["float"].Error != "")
}
func TestCheckBoolType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"bool": Rule{
				Type: "bool",
			},
		},
	}
	properties := map[string]interface{}{
		"bool": "true",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestCheckBoolTypeNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"bool": Rule{
				Type: "bool",
			},
		},
	}
	properties := map[string]interface{}{
		"bool": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["bool"].Valid == false)
	assert.True(t, result.Errors["bool"].Error != "")
}
func TestCheckUintType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"uint": Rule{
				Type: "uint",
			},
		},
	}
	properties := map[string]interface{}{
		"uint": "1",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestCheckUintTypeNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"uint": Rule{
				Type: "uint",
			},
		},
	}
	properties := map[string]interface{}{
		"uint": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["uint"].Valid == false)
	assert.True(t, result.Errors["uint"].Error != "")
}
func TestCheckStringType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"string": Rule{
				Type: "string",
			},
		},
	}
	properties := map[string]interface{}{
		"string": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestInvalidType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"invalid": Rule{
				Type: "invalid",
			},
		},
	}
	properties := map[string]interface{}{
		"invalid": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["invalid"].Valid == false)
	assert.True(t, result.Errors["invalid"].Error != "")
}
func TestNestedType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.string`": Rule{
				Type: "string",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"string": "a",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestNestedInvalidType(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.string`": Rule{
				Type: "invalid",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"string": "a",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["`.person.string`"].Valid == false)
	assert.True(t, result.Errors["`.person.string`"].Error != "")
}
func TestRequired(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"required": Rule{
				Required: true,
			},
		},
	}
	properties := map[string]interface{}{
		"required": "a",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestNestedRequired(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.required`": Rule{
				Required: true,
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"required": "a",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestRequiredMissing(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"required": Rule{
				Required: true,
			},
		},
	}
	properties := map[string]interface{}{}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["required"].Valid == false)
	assert.True(t, result.Errors["required"].Error != "")
}
func TestNestedRequiredMissing(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.required`": Rule{
				Required: true,
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["`.person.required`"].Valid == false)
	assert.True(t, result.Errors["`.person.required`"].Error != "")
}
func TestPattern(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"pattern": Rule{
				Pattern: "^[a-z]+$",
			},
		},
	}
	properties := map[string]interface{}{
		"pattern": "abc",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestNestedPattern(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.pattern`": Rule{
				Pattern: "^[a-z]+$",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"pattern": "abc",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestPatternNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"pattern": Rule{
				Pattern: "^[a-z]+$",
			},
		},
	}
	properties := map[string]interface{}{
		"pattern": "123",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["pattern"].Valid == false)
	assert.True(t, result.Errors["pattern"].Error != "")
}
func TestEmailPattern(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"pattern": Rule{
				Pattern: "<email>",
			},
		},
	}
	properties := map[string]interface{}{
		"pattern": "test@abc.com",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestEmailPatternNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"pattern": Rule{
				Pattern: "<email>",
			},
		},
	}
	properties := map[string]interface{}{
		"pattern": "test",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["pattern"].Valid == false)
	assert.True(t, result.Errors["pattern"].Error != "")
}
func TestNestedEmailPatternNotMatch(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.pattern`": Rule{
				Pattern: "<email>",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"pattern": "test",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
	assert.True(t, result.Errors["`.person.pattern`"].Valid == false)
	assert.True(t, result.Errors["`.person.pattern`"].Error != "")
}
func TestExpression(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"expression": Rule{
				Expression: "${{$and($gt($val(),10),$lt($val(),20))}}",
			},
		},
	}
	properties := map[string]interface{}{
		"expression": "13",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestNestedExpression(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.expression`": Rule{
				Expression: "${{$and($gt($val(),10),$lt($val(),20))}}",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"expression": "13",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestInExpression(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"expression": Rule{
				Expression: "${{$in($val(), 'foo', 'bar', 'baz')}}",
			},
		},
	}
	properties := map[string]interface{}{
		"expression": "bar",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestNestedInExpression(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.expression`": Rule{
				Expression: "${{$in($val(), 'foo', 'bar', 'baz')}}",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"expression": "bar",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.True(t, result.Valid)
}
func TestInExpressionMiss(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"expression": Rule{
				Expression: "${{$in($val(), 'foo', 'bar', 'baz')}}",
			},
		},
	}
	properties := map[string]interface{}{
		"expression": "barbar",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
}
func TestNestedInExpressionMiss(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"`.person.expression`": Rule{
				Expression: "${{$in($val(), 'foo', 'bar', 'baz')}}",
			},
		},
	}
	properties := map[string]interface{}{
		"person": map[string]interface{}{
			"expression": "barbar",
		},
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
}
func TestDottedNameInExpressionMiss(t *testing.T) {
	schema := Schema{
		Rules: map[string]Rule{
			"person.expression": Rule{
				Expression: "${{$in($val(), 'foo', 'bar', 'baz')}}",
			},
		},
	}
	properties := map[string]interface{}{
		"person.expression": "barbar",
	}
	result, err := schema.CheckProperties(properties, nil)
	assert.Nil(t, err)
	assert.False(t, result.Valid)
}
