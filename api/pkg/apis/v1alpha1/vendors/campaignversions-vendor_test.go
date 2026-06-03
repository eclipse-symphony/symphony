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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createCampaignVersionsVendor() CampaignVersionsVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := CampaignVersionsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "campaignversions-manager",
				Type: "managers.symphony.campaignversions",
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
		"campaignversions-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	vendor.CampaignVersionsManager.CampaignVersionValidator = validation.NewCampaignVersionValidator(nil, nil)
	return vendor
}
func TestCampaignVersionsEndpoints(t *testing.T) {
	vendor := createCampaignVersionsVendor()
	vendor.Route = "campaignversions"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}
func TestCampaignVersionsInfo(t *testing.T) {
	vendor := createCampaignVersionsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestCampaignVersionsOnCampaignVersions(t *testing.T) {
	vendor := createCampaignVersionsVendor()
	campaignversionState := model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "campaignversion1-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "campaignversion1",
		},
	}
	data, _ := json.Marshal(campaignversionState)
	resp := vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaignversion model.CampaignVersionState
	err := json.Unmarshal(resp.Body, &campaignversion)
	assert.Nil(t, err)
	assert.Equal(t, "campaignversion1-v-version1", campaignversion.ObjectMeta.Name)

	resp = vendor.onCampaignVersions(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var campaignversions []model.CampaignVersionState
	err = json.Unmarshal(resp.Body, &campaignversions)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(campaignversions))
	assert.Equal(t, "campaignversion1-v-version1", campaignversions[0].ObjectMeta.Name)

	resp = vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
func TestCampaignVersionsOnCampaignVersionsFailure(t *testing.T) {
	vendor := createCampaignVersionsVendor()
	campaignversionState := model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "campaignversion1-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "campaignversion1",
		},
	}
	data, _ := json.Marshal(campaignversionState)
	resp := vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Body:   data,
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.NotFound, resp.State)
	assert.Equal(t, "Not Found: entry 'campaignversion1-v-version1' is not found in namespace default", string(resp.Body))

	resp = vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   []byte("bad data"),
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "invalid character 'b' looking for beginning of value", string(resp.Body))

	resp = vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Body:   data,
		Parameters: map[string]string{
			"__name": "campaignversion1-v-version1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.NotFound, resp.State)
	assert.Equal(t, "Not Found: entry 'campaignversion1-v-version1' is not found in namespace default", string(resp.Body))
}

func TestCampaignVersionsWrongMethod(t *testing.T) {
	vendor := createCampaignVersionsVendor()
	campaignversionSpec := model.CampaignVersionSpec{}
	data, _ := json.Marshal(campaignversionSpec)
	resp := vendor.onCampaignVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodPut,
		Body:   data,
		Parameters: map[string]string{
			"__name": "campaignversion1-v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
