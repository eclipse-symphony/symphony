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

func TestCampaignsEndpoints(t *testing.T) {
	vendor := createCampaignsVendor()
	vendor.Route = "campaigns"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestCampaignsInfo(t *testing.T) {
	vendor := createCampaignsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func createCampaignsVendor() CampaignsVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := CampaignsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "campaignversion-container-manager",
				Type: "managers.symphony.campaigns",
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
		"campaignversion-container-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}

func TestOnCampaigns(t *testing.T) {
	vendor := createCampaignsVendor()
	vendor.Context = &contexts.VendorContext{}
	vendor.Context.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "fake",
	}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	campaignversion := model.CampaignState{
		Spec: &model.CampaignSpec{},
		ObjectMeta: model.ObjectMeta{
			Name:      "campaignversion1",
			Namespace: "scope1",
		},
	}
	data, _ := json.Marshal(campaignversion)
	resp := vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaignversion1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "campaignversion1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	var campaignversions model.CampaignState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &campaignversions)
	assert.Nil(t, err)
	assert.Equal(t, "campaignversion1", campaignversions.ObjectMeta.Name)
	assert.Equal(t, "scope1", campaignversions.ObjectMeta.Namespace)

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaignversionsList []model.CampaignState
	err = json.Unmarshal(resp.Body, &campaignversionsList)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(campaignversionsList))
	assert.Equal(t, "campaignversion1", campaignversionsList[0].ObjectMeta.Name)
	assert.Equal(t, "scope1", campaignversionsList[0].ObjectMeta.Namespace)

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "campaignversion1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
