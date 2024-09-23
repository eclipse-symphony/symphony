/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package configs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/config/catalog"
	coa_contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	memory "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/memoryconfig"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

var ctx = context.Background()

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
	provider1.Set(ctx, "obj", "field", "obj::field")
	provider2.Set(ctx, "obj-overlay", "field", "overlay::field")
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
	provider1.Set(ctx, "obj", "field", "obj::field")
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
	provider1.Set(ctx, "obj", "field", "obj::field")
	provider2.Set(ctx, "obj-overlay", "field", "overlay::field")
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
	provider1.Set(ctx, "obj", "field", "obj::field1")
	provider2.Set(ctx, "obj", "field", "obj::field2")
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

func TestCircularCatalogReferences(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/config1-v-v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "config1-v-v1",
				},
				Spec: &model.CatalogSpec{
					ParentName: "parent:v1",
					Properties: map[string]interface{}{
						"image":     "${{$config('config2:v1','image')}}",
						"attribute": "value",
					},
				},
			}
		case "/catalogs/registry/config2-v-v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "config2-v-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"attribute": "${{$config('config1:v1','attribute')}}",
						"foo":       "bar",
					},
				},
			}
		case "/catalogs/registry/parent-v-v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentConfig": "${{$config('config1:v1','parentAttribute')}}",
					},
				},
			}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")

	evalContext := utils.EvaluationContext{}
	vendorContext := coa_contexts.VendorContext{
		EvaluationContext: &evalContext,
	}
	provider := catalog.CatalogConfigProvider{}

	provider.Context = &coa_contexts.ManagerContext{
		VencorContext: &vendorContext,
	}
	err := provider.Init(catalog.CatalogConfigProviderConfig{})
	assert.Nil(t, err)

	manager := ConfigsManager{}
	config := managers.ManagerConfig{
		Name: "config-name",
		Type: "config-type",
		Properties: map[string]string{
			"providers.config": "ConfigProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["ConfigProvider"] = &provider
	err = manager.Init(&vendorContext, config, providers)
	assert.Nil(t, err)

	evalContext.ConfigProvider = &manager

	_, err = manager.Get("config1:v1", "image", nil, evalContext)
	assert.Error(t, err, "Detect circular dependency, object: config1-v-v1, field: image")

	_, err = manager.Get("config1:v1", "attribute", nil, evalContext)
	assert.Nil(t, err, "Detect correct attribute, expect no error")

	_, err = manager.Get("config2:v1", "attribute", nil, evalContext)
	assert.Nil(t, err, "Detect correct attribute, expect no error")
}

func TestParentConfigEvaluation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/config-v-v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "config-v-v1",
				},
				Spec: &model.CatalogSpec{
					ParentName: "parent:v1",
					Properties: map[string]interface{}{
						"attribute": "value",
					},
				},
			}
		case "/catalogs/registry/parent-v-v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "${{$config('config:v1','attribute')}}",
						"parentCircular":  "${{$config('config:v1','parentCircular')}}",
					},
				},
			}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")

	evalContext := utils.EvaluationContext{}
	vendorContext := coa_contexts.VendorContext{
		EvaluationContext: &evalContext,
	}
	provider := catalog.CatalogConfigProvider{}

	provider.Context = &coa_contexts.ManagerContext{
		VencorContext: &vendorContext,
	}
	err := provider.Init(catalog.CatalogConfigProviderConfig{})
	assert.Nil(t, err)

	manager := ConfigsManager{}
	config := managers.ManagerConfig{
		Name: "config-name",
		Type: "config-type",
		Properties: map[string]string{
			"providers.config": "ConfigProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["ConfigProvider"] = &provider
	err = manager.Init(&vendorContext, config, providers)
	assert.Nil(t, err)

	evalContext.ConfigProvider = &manager

	val, err := manager.Get("config:v1", "parentAttribute", nil, evalContext)
	assert.Equal(t, "value", val)
	assert.Nil(t, err)

	_, err = manager.Get("config:v1", "parentCircular", nil, evalContext)
	assert.Error(t, err, "Circular dependency in config should throw error")
}
