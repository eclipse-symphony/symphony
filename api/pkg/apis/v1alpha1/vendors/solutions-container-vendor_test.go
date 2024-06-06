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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestSolutionContainersEndpoints(t *testing.T) {
	vendor := createSolutionContainersVendor()
	vendor.Route = "solutioncontainers"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestSolutionContainersInfo(t *testing.T) {
	vendor := createSolutionContainersVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func createSolutionContainersVendor() SolutionContainersVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := SolutionContainersVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "solution-container-manager",
				Type: "managers.symphony.solutioncontainers",
				Properties: map[string]string{
					"providers.state": "mem-state",
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
		"solution-container-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}

func TestOnSolutionContainers(t *testing.T) {
	vendor := createSolutionContainersVendor()
	vendor.Context = &contexts.VendorContext{}
	vendor.Context.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "fake",
	}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	solution := model.SolutionContainerState{
		Spec: &model.SolutionContainerSpec{},
		ObjectMeta: model.ObjectMeta{
			Name:      "solution1",
			Namespace: "scope1",
		},
	}
	data, _ := json.Marshal(solution)
	resp := vendor.onSolutionContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "solution1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onSolutionContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "solution1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	var solutions model.SolutionContainerState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &solutions)
	assert.Nil(t, err)
	assert.Equal(t, "solution1", solutions.ObjectMeta.Name)
	assert.Equal(t, "scope1", solutions.ObjectMeta.Namespace)

	resp = vendor.onSolutionContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var solutionsList []model.SolutionContainerState
	err = json.Unmarshal(resp.Body, &solutionsList)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(solutionsList))
	assert.Equal(t, "solution1", solutionsList[0].ObjectMeta.Name)
	assert.Equal(t, "scope1", solutionsList[0].ObjectMeta.Namespace)

	resp = vendor.onSolutionContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "solution1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
