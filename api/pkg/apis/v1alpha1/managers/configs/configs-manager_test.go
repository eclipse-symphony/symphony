/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configs

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	memory "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/memoryconfig"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	configProvider := memory.MemoryConfigProvider{}
	configProvider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.config": "ConfigProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["ConfigProvider"] = &configProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
}
func TestObjectFieldGetWithProviderSpecified(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)
	manager.Set("memory:obj", "field", "obj::field")
	val, err := manager.Get("memory:obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}

func TestObjectGetWithProviderSpecified(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)
	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	manager.SetObject("memory:obj", object)

	// GetObject
	val, err := manager.GetObject("memory:obj", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, object, val)

	// Get
	val2, err2 := manager.Get("memory:obj", "", nil, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object, val2)
}

func TestObjectFieldGetWithOneProvider(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)
	manager.Set("obj", "field", "obj::field")
	val, err := manager.Get("obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}

func TestObjectGetWithOneProvider(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)
	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	manager.SetObject("obj", object)

	// GetObject
	val, err := manager.GetObject("obj", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, object, val)

	// Get
	val2, err2 := manager.Get("obj", "", nil, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object, val2)
}

func TestObjectFieldGetWithMoreProviders(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory1", "memory2"},
	}
	assert.Nil(t, err)
	manager.Set("obj", "field", "obj::field")
	val, err := manager.Get("obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}

func TestObjectGetWithMoreProviders(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory1", "memory2"},
	}
	assert.Nil(t, err)
	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	manager.SetObject("obj", object)

	// GetObject
	val, err := manager.GetObject("obj", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, object, val)

	// Get
	val2, err2 := manager.Get("obj", "", nil, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object, val2)
}

func TestWithOverlay(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)

	manager.Set("obj", "field", "obj::field")
	manager.Set("obj-overlay", "field", "overlay::field")
	val, err := manager.Get("obj", "field", []string{"obj-overlay"}, nil)
	assert.Nil(t, err)
	assert.Equal(t, "overlay::field", val)

	object := map[string]interface{}{
		"key1": "value1",
	}
	manager.SetObject("obj2", object)
	object2 := map[string]interface{}{
		"key1": "overlay",
	}
	manager.SetObject("obj2-overlay", object2)
	val2, err2 := manager.GetObject("obj2", []string{"obj2-overlay"}, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object2, val2)

}
func TestOverlayWithMultipleProviders(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory2", "memory1"},
	}
	assert.Nil(t, err)
	provider1.Set("obj", "field", "obj::field")
	provider2.Set("obj-overlay", "field", "overlay::field")
	val, err := manager.Get("obj", "field", []string{"obj-overlay"}, nil)
	assert.Nil(t, err)
	assert.Equal(t, "overlay::field", val)

	object := map[string]interface{}{
		"key1": "value1",
	}
	manager.SetObject("obj2", object)
	object2 := map[string]interface{}{
		"key1": "overlay",
	}
	manager.SetObject("obj2-overlay", object2)
	val2, err2 := manager.GetObject("obj2", []string{"obj2-overlay"}, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object2, val2)
}
func TestOverlayMissWithMultipleProviders(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory2", "memory1"},
	}
	assert.Nil(t, err)
	provider1.Set("obj", "field", "obj::field")
	val, err := manager.Get("obj", "field", []string{"obj-overlay"}, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)

	object := map[string]interface{}{
		"key1": "value1",
	}
	manager.SetObject("obj2", object)
	object2 := map[string]interface{}{
		"key1": "overlay",
	}
	manager.SetObject("obj2-overlay", object2)
	val2, err2 := manager.GetObject("obj2", []string{"obj2-overlay"}, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object2, val2)
}
func TestOverlayWithMultipleProvidersReversedPrecedence(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory1", "memory2"},
	}
	assert.Nil(t, err)
	provider1.Set("obj", "field", "obj::field")
	provider2.Set("obj-overlay", "field", "overlay::field")
	val, err := manager.Get("obj", "field", []string{"obj-overlay"}, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)

	object := map[string]interface{}{
		"key1": "value1",
	}
	manager.SetObject("obj2", object)
	object2 := map[string]interface{}{
		"key1": "overlay",
	}
	manager.SetObject("obj2-overlay", object2)
	val2, err2 := manager.GetObject("obj2", []string{"obj2-overlay"}, nil)
	assert.Nil(t, err2)
	assert.Equal(t, object2, val2)
}

func TestMultipleProvidersSameKey(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory2", "memory1"},
	}
	assert.Nil(t, err)
	provider1.Set("obj", "field", "obj::field1")
	provider2.Set("obj", "field", "obj::field2")
	val, err := manager.Get("obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field2", val)
}

func TestObjectDeleteWithProviderSpecified(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)

	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	manager.SetObject("memory::obj", object)

	// Delete field
	err = manager.Delete("memory::obj", "key1")
	assert.Nil(t, err)
	val, err := manager.Get("memory::obj", "key1", nil, nil)
	assert.NotNil(t, err)
	assert.Empty(t, val)

	// Delete object
	err2 := manager.DeleteObject("memory::obj")
	assert.Nil(t, err2)
	val2, err2 := manager.GetObject("memory::obj", nil, nil)
	assert.NotNil(t, err2)
	assert.Empty(t, val2)
}

func TestObjectDeleteWithOneProvider(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)

	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	manager.SetObject("obj", object)

	// Delete field
	err = manager.Delete("obj", "key1")
	assert.Nil(t, err)
	val, err := manager.Get("obj", "key1", nil, nil)
	assert.NotNil(t, err)
	assert.Empty(t, val)

	// Delete object
	err2 := manager.DeleteObject("obj")
	assert.Nil(t, err2)
	val2, err2 := manager.GetObject("obj", nil, nil)
	assert.NotNil(t, err2)
	assert.Empty(t, val2)
}

func TestObjectDeleteWithMoreProviders(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory1", "memory2"},
	}
	assert.Nil(t, err)

	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	manager.SetObject("obj", object)

	// Delete field
	err = manager.Delete("obj", "key1")
	assert.Nil(t, err)
	val, err := manager.Get("obj", "key1", nil, nil)
	assert.NotNil(t, err)
	assert.Empty(t, val)

	// Delete object
	err2 := manager.DeleteObject("obj")
	assert.Nil(t, err2)
	val2, err2 := manager.GetObject("obj", nil, nil)
	assert.NotNil(t, err2)
	assert.Empty(t, val2)
}

func TestObjectReference(t *testing.T) {
	provider1 := memory.MemoryConfigProvider{}
	err := provider1.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	provider2 := memory.MemoryConfigProvider{}
	err = provider2.Init(memory.MemoryConfigProviderConfig{})
	assert.Nil(t, err)
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory1": &provider1,
			"memory2": &provider2,
		},
		Precedence: []string{"memory1", "memory2"},
	}
	assert.Nil(t, err)

	object := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	// Get field
	manager.SetObject("memory1::obj:v1", object)
	val, err := manager.Get("obj:v1", "key1", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "value1", val)

	// Delete field
	err = manager.Delete("memory1::obj:v1", "key1")
	assert.Nil(t, err)
	val, err = manager.Get("memory1::obj:v1", "key1", nil, nil)
	assert.NotNil(t, err)
	assert.Empty(t, val)

	// Delete object
	err2 := manager.DeleteObject("memory1::obj:v1")
	assert.Nil(t, err2)
	val2, err2 := manager.GetObject("memory1::obj:v1", nil, nil)
	assert.NotNil(t, err2)
	assert.Empty(t, val2)
}
