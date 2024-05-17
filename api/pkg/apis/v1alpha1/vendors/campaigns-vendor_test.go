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
				Name: "campaigns-manager",
				Type: "managers.symphony.campaigns",
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
		"campaigns-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}
func TestCampaignsEndpoints(t *testing.T) {
	vendor := createCampaignsVendor()
	vendor.Route = "campaigns"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 2, len(endpoints))
}
func TestCampaignsInfo(t *testing.T) {
	vendor := createCampaignsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestCampaignsOnCampaigns(t *testing.T) {
	vendor := createCampaignsVendor()
	campaignState := model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name: "campaign1-v1",
		},
		Spec: &model.CampaignSpec{Version: "v1", RootResource: "campaign1"},
	}
	data, _ := json.Marshal(campaignState)
	resp := vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaign model.CampaignState
	err := json.Unmarshal(resp.Body, &campaign)
	assert.Nil(t, err)
	assert.Equal(t, "campaign1-v1", campaign.ObjectMeta.Name)

	resp = vendor.onCampaignsList(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaigns []model.CampaignState
	err = json.Unmarshal(resp.Body, &campaigns)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(campaigns))
	assert.Equal(t, "campaign1-v1", campaigns[0].ObjectMeta.Name)

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
func TestCampaignsOnCampaignsFailure(t *testing.T) {
	vendor := createCampaignsVendor()
	campaignState := model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name: "campaign1-v1",
		},
		Spec: &model.CampaignSpec{Version: "v1", RootResource: "campaign1"},
	}
	data, _ := json.Marshal(campaignState)
	resp := vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "Not Found: entry 'campaign1-v1' is not found in namespace default", string(resp.Body))

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   []byte("bad data"),
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "invalid character 'b' looking for beginning of value", string(resp.Body))

	resp = vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "Not Found: entry 'campaign1-v1' is not found in namespace default", string(resp.Body))
}

func TestCampaignsWrongMethod(t *testing.T) {
	vendor := createCampaignsVendor()
	campaignSpec := model.CampaignSpec{}
	data, _ := json.Marshal(campaignSpec)
	resp := vendor.onCampaigns(v1alpha2.COARequest{
		Method: fasthttp.MethodPut,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "campaign1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
