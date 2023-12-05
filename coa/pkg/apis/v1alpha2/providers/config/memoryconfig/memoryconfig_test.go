/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
}
func TestGetEmpty(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	_, err = provider.Read("obj", "field", nil)
	assert.NotNil(t, err)
}
func TestGet(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.Set("obj", "field", "obj::field")
	val, err := provider.Read("obj", "field", nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}
func TestGetObject(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.SetObject("obj", map[string]interface{}{"field": "obj::field"})
	val, err := provider.ReadObject("obj", nil)
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"field": "obj::field"}, val)
}
func TestDelete(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.Set("obj", "field", "obj::field")
	err = provider.Remove("obj", "field")
	assert.Nil(t, err)
	_, err = provider.Read("obj", "field", nil)
	assert.NotNil(t, err)
}
func TestDeleteObject(t *testing.T) {
	provider := MemoryConfigProvider{}
	err := provider.Init(MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider.SetObject("obj", map[string]interface{}{"field": "obj::field"})
	err = provider.RemoveObject("obj")
	assert.Nil(t, err)
	_, err = provider.ReadObject("obj", nil)
	assert.NotNil(t, err)
}
