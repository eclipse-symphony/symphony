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

func TestCampaignContainersEndpoints(t *testing.T) {
	vendor := createCampaignContainersVendor()
	vendor.Route = "campaigncontainers"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestCampaignContainersInfo(t *testing.T) {
	vendor := createCampaignContainersVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func createCampaignContainersVendor() CampaignContainersVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := CampaignContainersVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "campaign-container-manager",
				Type: "managers.symphony.campaigncontainers",
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
		"campaign-container-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}

func TestOnCampaignContainers(t *testing.T) {
	vendor := createCampaignContainersVendor()
	vendor.Context = &contexts.VendorContext{}
	vendor.Context.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "fake",
	}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	campaign := model.CampaignContainerState{
		Spec: &model.CampaignContainerSpec{},
		ObjectMeta: model.ObjectMeta{
			Name:      "campaign1",
			Namespace: "scope1",
		},
	}
	data, _ := json.Marshal(campaign)
	resp := vendor.onCampaignContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onCampaignContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	var campaigns model.CampaignContainerState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &campaigns)
	assert.Nil(t, err)
	assert.Equal(t, "campaign1", campaigns.ObjectMeta.Name)
	assert.Equal(t, "scope1", campaigns.ObjectMeta.Namespace)

	resp = vendor.onCampaignContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaignsList []model.CampaignContainerState
	err = json.Unmarshal(resp.Body, &campaignsList)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(campaignsList))
	assert.Equal(t, "campaign1", campaignsList[0].ObjectMeta.Name)
	assert.Equal(t, "scope1", campaignsList[0].ObjectMeta.Namespace)

	resp = vendor.onCampaignContainers(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
