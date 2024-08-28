/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/devices"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createDevicesVendor() DevicesVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := devices.DevicesManager{
		StateProvider: stateProvider,
	}
	vendor := DevicesVendor{
		DevicesManager: &manager,
	}
	return vendor
}

func TestDevicesVendorInit(t *testing.T) {
	provider := memorystate.MemoryStateProvider{}
	provider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := DevicesVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "devices-manager",
				Type: "managers.symphony.devices",
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
		"devices-manager": {
			"mem-state": &provider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestGetEndpoints(t *testing.T) {
	vendor := createDevicesVendor()
	vendor.Route = "route"
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "route", endpoints[len(endpoints)-1].Route)
}

func TestGetInfo(t *testing.T) {
	vendor := createDevicesVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestPostAndGet(t *testing.T) {
	vendor := createDevicesVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
		},
	}
	res := vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)
	deviceState := model.DeviceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.DeviceSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"type": "sensor",
			},
		},
	}
	data, err := json.Marshal(deviceState)
	request.Body = data
	res = vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
	assert.Nil(t, err)

	request = &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name":   "test",
			"doc-type": "yaml",
		},
	}
	res = vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
	var state model.DeviceState

	err = json.Unmarshal(res.Body, &state)
	assert.Nil(t, err)
	equal, err := deviceState.DeepEquals(state)
	assert.Nil(t, err)
	assert.True(t, equal)

	request = &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	res = vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
	var states []model.DeviceState
	err = json.Unmarshal(res.Body, &states)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(states))
}

func TestPostAndDelete(t *testing.T) {
	vendor := createDevicesVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
		},
	}
	deviceSpec := model.DeviceSpec{
		DisplayName: "device",
		Properties: map[string]string{
			"type": "sensor",
		},
	}
	data, err := json.Marshal(deviceSpec)
	request.Body = data
	res := vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
	assert.Nil(t, err)

	request = &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "unknown",
		},
	}
	res = vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)

	requestGet := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
		},
	}
	res = vendor.onDevices(*requestGet)
	assert.Equal(t, v1alpha2.OK, res.State)
	request = &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
		Parameters: map[string]string{
			"__name": "test",
		},
	}
	res = vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
	res = vendor.onDevices(*requestGet)
	assert.Equal(t, v1alpha2.InternalError, res.State)
}

func TestNotAllowed(t *testing.T) {
	vendor := createDevicesVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPatch,
		Context: context.Background(),
	}
	res := vendor.onDevices(*request)
	assert.Equal(t, v1alpha2.MethodNotAllowed, res.State)
}
