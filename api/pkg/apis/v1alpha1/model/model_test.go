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

func TestModelEqual(t *testing.T) {
	model1 := &ModelSpec{
		DisplayName: "model",
		Properties: map[string]string{
			"foo": "bar",
		},
		Constraints: "constraints",
		Bindings: []BindingSpec{
			{
				Role:     "role",
				Provider: "provider",
				Config: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	model2 := &ModelSpec{
		DisplayName: "model",
		Properties: map[string]string{
			"foo": "bar",
		},
		Constraints: "constraints",
		Bindings: []BindingSpec{
			{
				Role:     "role",
				Provider: "provider",
				Config: map[string]string{
					"foo": "bar",
				},
			},
		},
	}

	equal, err := model1.DeepEquals(model2)
	assert.Nil(t, err)
	assert.True(t, equal)
}

func TestModelNotEqual(t *testing.T) {
	// Test empty
	model1 := &ModelSpec{
		DisplayName: "model",
	}
	equal, err := model1.DeepEquals(nil)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test DisplayName not equal
	model2 := &ModelSpec{
		DisplayName: "model2",
	}
	equal, err = model1.DeepEquals(model2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Properties not equal
	model2.DisplayName = "model"
	model1.Properties = map[string]string{
		"foo": "bar",
	}
	model2.Properties = map[string]string{
		"foo": "bar2",
	}
	equal, err = model1.DeepEquals(model2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Constraints not equal
	model2.Properties = map[string]string{
		"foo": "bar",
	}
	model1.Constraints = "constraints"
	model2.Constraints = "constraints2"
	equal, err = model1.DeepEquals(model2)
	assert.Nil(t, err)
	assert.False(t, equal)

	// Test Bindings not equal
	model2.Constraints = "constraints"
	model1.Bindings = []BindingSpec{
		{
			Role:     "role",
			Provider: "provider",
			Config: map[string]string{
				"foo": "bar",
			},
		},
	}
	model2.Bindings = []BindingSpec{
		{
			Role:     "role2",
			Provider: "provider2",
			Config: map[string]string{
				"foo": "bar",
			},
		},
	}
	equal, err = model1.DeepEquals(model2)
	assert.Nil(t, err)
	assert.False(t, equal)
}
