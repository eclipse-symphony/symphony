/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE

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
	val, err := manager.Get("obj", "field")
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
	val, err := manager.Get("obj", "field")
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
	val, err := manager.Get("obj", "field")
	assert.Nil(t, err)
	assert.Equal(t, "overlay::field", val)
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
	val, err := manager.Get("obj", "field")
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
	val, err := manager.GetObject("obj")
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val["field"])
}
