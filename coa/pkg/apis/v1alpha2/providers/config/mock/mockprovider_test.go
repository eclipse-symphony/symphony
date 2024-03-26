/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockConfigProvider{}
	err := provider.Init(MockConfigProviderConfig{})
	assert.Nil(t, err)

	properties := map[string]string{
		"name": "test",
	}
	assert.Nil(t, err)
	err = provider.InitWithMap(properties)
	assert.Nil(t, err)
}
func TestGet(t *testing.T) {
	provider := MockConfigProvider{}
	err := provider.Init(MockConfigProviderConfig{})
	assert.Nil(t, err)
	val, err := provider.Get("obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)

	val, err = provider.Read("obj", "field", nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)

	val, err = provider.ReadObject("obj", nil)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"obj": "obj"}, val)
}

// TestMockConfigProviderConfigFromMap tests the MockConfigProviderConfigFromMap function
func TestMockConfigProviderConfigFromMap(t *testing.T) {
	_, err := MockConfigProviderConfigFromMap(map[string]string{
		"name": "test",
	})
	assert.Nil(t, err)
}

// TestInitWithMap tests the InitWithMap function
func TestInitWithMap(t *testing.T) {
	provider := MockConfigProvider{}
	err := provider.InitWithMap(map[string]string{
		"name": "test",
	})
	assert.Nil(t, err)
	name := provider.ID()
	assert.Equal(t, "test", name)
}

// TestMockConfigProviderConfig tests the MockConfigProviderConfig function
func TestMockConfigProviderConfig(t *testing.T) {
	_, err := toMockConfigProviderConfig(map[string]string{
		"name": "test",
	})
	assert.Nil(t, err)
}
