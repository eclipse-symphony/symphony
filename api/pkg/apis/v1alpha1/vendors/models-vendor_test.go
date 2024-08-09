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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createModelsVendor(route string) ModelsVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	vendor := ModelsVendor{}
	_ = vendor.Init(vendors.VendorConfig{
		Route: route,
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "models-manager",
				Type: "managers.symphony.models",
				Properties: map[string]string{
					"providers.persistentstate": "memory",
				},
				Providers: map[string]managers.ProviderConfig{
					"memory": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"models-manager": {
			"memory": stateProvider,
		},
	}, nil)
	return vendor
}

func TestModelsVendorInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	vendor := ModelsVendor{}
	err = vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "models-manager",
				Type: "managers.symphony.models",
				Properties: map[string]string{
					"providers.persistentstate": "memory",
				},
				Providers: map[string]managers.ProviderConfig{
					"memory": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"models-manager": {
			"memory": stateProvider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestModelsVendorInitFail(t *testing.T) {
	vendor := ModelsVendor{}
	// missing models manager
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{}, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)
	assert.Equal(t, "models manager is not supplied", coaError.Message)
}

func TestModelsVendorGetInfo(t *testing.T) {
	vendor := createModelsVendor("")
	info := vendor.GetInfo()
	assert.Equal(t, "Models", info.Name)
	assert.Equal(t, "Microsoft", info.Producer)
}

func TestModelsVendorGetEndpoints(t *testing.T) {
	vendor := createModelsVendor("")
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))

	vendor = createModelsVendor("models")
	endpoints = vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, "models", endpoints[0].Route)
}

func TestModelsVendorOnModels(t *testing.T) {
	m1 := model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "model",
		},
		Spec: &model.ModelSpec{
			DisplayName: "model",
			Properties: map[string]string{
				"foo": "bar",
			},
			Constraints: "constraints",
			Bindings: []model.BindingSpec{
				{
					Role:     "role",
					Provider: "provider",
					Config: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
	}

	data, err := json.Marshal(m1)
	assert.Nil(t, err)

	vendor := createModelsVendor("")

	// create
	createReq := v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": m1.Spec.DisplayName,
		},
		Context: context.Background(),
	}
	resp := vendor.onModels(createReq)
	assert.Equal(t, v1alpha2.OK, resp.State)

	// get
	getReq := v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": m1.Spec.DisplayName,
		},
		Context: context.Background(),
	}
	resp = vendor.onModels(getReq)
	assert.Equal(t, v1alpha2.OK, resp.State)
	var m1State model.ModelState
	err = json.Unmarshal(resp.Body, &m1State)
	assert.Nil(t, err)
	equal, err := m1.Spec.DeepEquals(*m1State.Spec)
	assert.Nil(t, err)
	assert.True(t, equal)

	// list
	listReq := v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	resp = vendor.onModels(listReq)
	assert.Equal(t, v1alpha2.OK, resp.State)
	var modelsState []model.ModelState
	err = json.Unmarshal(resp.Body, &modelsState)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(modelsState))
	equal, err = m1.Spec.DeepEquals(*modelsState[0].Spec)
	assert.Nil(t, err)
	assert.True(t, equal)

	// delete
	deleteReq := v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": m1.Spec.DisplayName,
		},
		Context: context.Background(),
	}
	resp = vendor.onModels(deleteReq)
	assert.Equal(t, v1alpha2.OK, resp.State)

}

func TestModelsVendorOnModels_PostErrorModel(t *testing.T) {
	vendor := createModelsVendor("")

	// create
	createReq := v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   []byte("invalid"),
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onModels(createReq)
	assert.Equal(t, v1alpha2.InternalError, resp.State)
}

func TestModelsVendorOnModels_GetModelNotExists(t *testing.T) {
	vendor := createModelsVendor("")

	// get
	getReq := v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onModels(getReq)
	assert.Equal(t, v1alpha2.InternalError, resp.State)
}

func TestModelsVendorOnModels_DeleteModelNotExists(t *testing.T) {
	vendor := createModelsVendor("")

	// delete
	deleteReq := v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onModels(deleteReq)
	assert.Equal(t, v1alpha2.OK, resp.State)
}

func TestModelsVendorOnModels_InvalidMethod(t *testing.T) {
	vendor := createModelsVendor("")

	// invalid method
	req := v1alpha2.COARequest{
		Method: fasthttp.MethodHead,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onModels(req)
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
