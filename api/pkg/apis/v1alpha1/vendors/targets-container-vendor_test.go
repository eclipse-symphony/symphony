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

func TestTargetContainersEndpoints(t *testing.T) {
	vendor := createTargetContainersVendor()
	vendor.Route = "targetcontainers"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestTargetContainersInfo(t *testing.T) {
	vendor := createTargetContainersVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func createTargetContainersVendor() TargetContainersVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := TargetContainersVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "target-container-manager",
				Type: "managers.symphony.targetcontainers",
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
		"target-container-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}

func TestOnTargetContainers(t *testing.T) {
	vendor := createTargetContainersVendor()
	vendor.Context = &contexts.VendorContext{}
	vendor.Context.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "fake",
	}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	target := model.TargetContainerState{
		Spec: &model.TargetContainerSpec{},
		ObjectMeta: model.ObjectMeta{
			Name:      "target1",
			Namespace: "scope1",
		},
	}
	data, _ := json.Marshal(target)
	resp := vendor.onTargetContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "target1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onTargetContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "target1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	var targets model.TargetContainerState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &targets)
	assert.Nil(t, err)
	assert.Equal(t, "target1", targets.ObjectMeta.Name)
	assert.Equal(t, "scope1", targets.ObjectMeta.Namespace)

	resp = vendor.onTargetContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var targetsList []model.TargetContainerState
	err = json.Unmarshal(resp.Body, &targetsList)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(targetsList))
	assert.Equal(t, "target1", targetsList[0].ObjectMeta.Name)
	assert.Equal(t, "scope1", targetsList[0].ObjectMeta.Namespace)

	resp = vendor.onTargetContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "target1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
