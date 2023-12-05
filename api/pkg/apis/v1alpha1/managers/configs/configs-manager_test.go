/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configs

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	memory "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config/memoryconfig"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
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
}
func TestGetObject(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	err := provider.Init(memory.MemoryConfigProviderConfig{})
	manager := ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	assert.Nil(t, err)
	manager.Set("obj", "field", "obj::field")
	val, err := manager.GetObject("obj", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val["field"])
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
