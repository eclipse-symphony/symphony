/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/configs"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	memory "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/memoryconfig"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

var ctx = context.Background()

func createSettingsVendor() SettingsVendor {
	provider := memory.MemoryConfigProvider{}
	provider.Init(memory.MemoryConfigProviderConfig{})
	manager := configs.ConfigsManager{
		ConfigProviders: map[string]config.IConfigProvider{
			"memory": &provider,
		},
	}
	vendor := SettingsVendor{
		EvaluationContext: &coa_utils.EvaluationContext{
			ConfigProvider: &manager,
		},
	}
	return vendor
}

func TestSettingsVendorInit(t *testing.T) {
	provider := memory.MemoryConfigProvider{}
	provider.Init(memory.MemoryConfigProviderConfig{})
	vendor := SettingsVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "configs-manager",
				Type: "managers.symphony.configs",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"configs-manager": {
			"mem-state": &provider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestSettingsEndpoints(t *testing.T) {
	vendor := createSettingsVendor()
	vendor.Route = "settings"
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "settings/config", endpoints[len(endpoints)-1].Route)
}

func TestSettingsInfo(t *testing.T) {
	vendor := createSettingsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestSettingsEvaluation(t *testing.T) {
	vendor := createSettingsVendor()
	context := vendor.GetEvaluationContext()
	manager := context.ConfigProvider.(*configs.ConfigsManager)
	assert.NotNil(t, manager.ConfigProviders["memory"])
}

func TestConfigNotAllowed(t *testing.T) {
	vendor := createSettingsVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
	}
	res := vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.MethodNotAllowed, res.State)
}

func TestConfigGet(t *testing.T) {
	vendor := createSettingsVendor()
	manager := vendor.EvaluationContext.ConfigProvider.(*configs.ConfigsManager)
	provider := manager.ConfigProviders["memory"]
	provider.Set(ctx, "test", "field", "obj::field")

	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
		},
	}
	res := vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.OK, res.State)

	request.Parameters["__name"] = "unknown"
	res = vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)
}

func TestConfigGetField(t *testing.T) {
	vendor := createSettingsVendor()
	manager := vendor.EvaluationContext.ConfigProvider.(*configs.ConfigsManager)
	provider := manager.ConfigProviders["memory"]
	provider.Set(ctx, "test", "field", "obj::field")

	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
			"field":  "field",
		},
	}
	res := vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.OK, res.State)

	request.Parameters["__name"] = "unknown"
	res = vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)
}
