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

func createSkillsVendor(route string) SkillsVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	vendor := SkillsVendor{}
	_ = vendor.Init(vendors.VendorConfig{
		Route: route,
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "skills-manager",
				Type: "managers.symphony.skills",
				Properties: map[string]string{
					"providers.state": "memory",
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
		"skills-manager": {
			"memory": stateProvider,
		},
	}, nil)
	return vendor
}

func TestSkillsVendorInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	vendor := SkillsVendor{}
	err = vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "skills-manager",
				Type: "managers.symphony.skills",
				Properties: map[string]string{
					"providers.state": "memory",
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
		"skills-manager": {
			"memory": stateProvider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestSkillsVendorInitFail(t *testing.T) {
	vendor := SkillsVendor{}
	// missing skills manager
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
	assert.Equal(t, "skills manager is not supplied", coaError.Message)
}

func TestSkillsVendorGetInfo(t *testing.T) {
	vendor := createSkillsVendor("")
	info := vendor.GetInfo()
	assert.Equal(t, "Skills", info.Name)
	assert.Equal(t, "Microsoft", info.Producer)
}

func TestSkillsVendorGetEndpoints(t *testing.T) {
	vendor := createSkillsVendor("")
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))

	vendor = createSkillsVendor("skills")
	endpoints = vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, "skills", endpoints[0].Route)
}

func TestSkillsVendorOnSkills(t *testing.T) {
	s1 := model.SkillSpec{
		DisplayName: "skill",
		Parameters: map[string]string{
			"foo": "bar",
		},
		Nodes: []model.NodeSpec{
			{
				Id: "node1",
			},
			{
				Id: "node2",
			},
		},
		Properties: map[string]string{
			"foo": "bar",
		},
		Bindings: []model.BindingSpec{
			{
				Role:     "role",
				Provider: "provider",
			},
		},
		Edges: []model.EdgeSpec{
			{
				Source: model.ConnectionSpec{
					Node:  "node1",
					Route: "route1",
				},
				Target: model.ConnectionSpec{
					Node:  "node2",
					Route: "route1",
				},
			},
		},
	}

	data, err := json.Marshal(s1)
	assert.Nil(t, err)

	vendor := createSkillsVendor("")

	// create
	createReq := v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": s1.DisplayName,
		},
		Context: context.Background(),
	}
	resp := vendor.onSkills(createReq)
	assert.Equal(t, v1alpha2.OK, resp.State)

	// get
	getReq := v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": s1.DisplayName,
		},
		Context: context.Background(),
	}
	resp = vendor.onSkills(getReq)
	assert.Equal(t, v1alpha2.OK, resp.State)
	var s1State model.SkillState
	err = json.Unmarshal(resp.Body, &s1State)
	assert.Nil(t, err)
	equal, err := s1.DeepEquals(*s1State.Spec)
	assert.Nil(t, err)
	assert.True(t, equal)

	// list
	listReq := v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	resp = vendor.onSkills(listReq)
	assert.Equal(t, v1alpha2.OK, resp.State)
	var skillsState []model.SkillState
	err = json.Unmarshal(resp.Body, &skillsState)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(skillsState))
	equal, err = s1.DeepEquals(*skillsState[0].Spec)
	assert.Nil(t, err)
	assert.True(t, equal)

	// delete
	deleteReq := v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": s1.DisplayName,
		},
		Context: context.Background(),
	}
	resp = vendor.onSkills(deleteReq)
	assert.Equal(t, v1alpha2.OK, resp.State)

}

func TestSkillsVendorOnSkills_PostErrorSkill(t *testing.T) {
	vendor := createSkillsVendor("")

	// create
	createReq := v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   []byte("invalid"),
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onSkills(createReq)
	assert.Equal(t, v1alpha2.InternalError, resp.State)
}

func TestSkillsVendorOnSkills_GetSkillNotExists(t *testing.T) {
	vendor := createSkillsVendor("")

	// get
	getReq := v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onSkills(getReq)
	assert.Equal(t, v1alpha2.InternalError, resp.State)
}

func TestSkillsVendorOnSkills_DeleteSkillNotExists(t *testing.T) {
	vendor := createSkillsVendor("")

	// delete
	deleteReq := v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onSkills(deleteReq)
	assert.Equal(t, v1alpha2.InternalError, resp.State)
}

func TestSkillsVendorOnSkills_InvalidMethod(t *testing.T) {
	vendor := createSkillsVendor("")

	// invalid method
	req := v1alpha2.COARequest{
		Method: fasthttp.MethodHead,
		Parameters: map[string]string{
			"__name": "invalid",
		},
		Context: context.Background(),
	}
	resp := vendor.onSkills(req)
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
