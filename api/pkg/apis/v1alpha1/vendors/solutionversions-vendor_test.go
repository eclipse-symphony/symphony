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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestSolutionVersionsEndpoints(t *testing.T) {
	vendor := createSolutionVersionsVendor()
	vendor.Route = "solutionversions"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestSolutionVersionsInfo(t *testing.T) {
	vendor := createSolutionVersionsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func createSolutionVersionsVendor() SolutionVersionsVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := SolutionVersionsVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "solutionversions-manager",
				Type: "managers.symphony.solutionversions",
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
		"solutionversions-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	vendor.SolutionVersionsManager.SolutionVersionValidator = validation.NewSolutionVersionValidator(nil, nil, nil)
	return vendor
}
func TestSolutionVersionsOnSolutionVersions(t *testing.T) {
	vendor := createSolutionVersionsVendor()
	vendor.Context = &contexts.VendorContext{}
	vendor.Context.SiteInfo = v1alpha2.SiteInfo{
		SiteId: "fake",
	}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	solutionversion := model.SolutionVersionState{
		Spec: &model.SolutionVersionSpec{
			RootResource: "solutionversions1",
		},
		ObjectMeta: model.ObjectMeta{
			Name:      "solutionversions1-v-version1",
			Namespace: "scope1",
		},
	}
	data, _ := json.Marshal(solutionversion)
	resp := vendor.onSolutionVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "solutionversions1-v-version1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onSolutionVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "solutionversions1-v-version1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	var solutionversions model.SolutionVersionState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &solutionversions)
	assert.Nil(t, err)
	assert.Equal(t, "solutionversions1-v-version1", solutionversions.ObjectMeta.Name)
	assert.Equal(t, "scope1", solutionversions.ObjectMeta.Namespace)

	resp = vendor.onSolutionVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
	var solutionversionsList []model.SolutionVersionState
	err = json.Unmarshal(resp.Body, &solutionversionsList)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(solutionversionsList))
	assert.Equal(t, "solutionversions1-v-version1", solutionversionsList[0].ObjectMeta.Name)
	assert.Equal(t, "scope1", solutionversionsList[0].ObjectMeta.Namespace)

	resp = vendor.onSolutionVersions(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "solutionversions1-v-version1",
			"namespace": "scope1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
