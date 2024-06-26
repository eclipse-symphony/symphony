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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
)

func initVendor(t *testing.T) UsersVendor {
	p := memorystate.MemoryStateProvider{}
	p.Init(memorystate.MemoryStateProviderConfig{})
	vendor := UsersVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test-users": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "users-manager",
				Type: "managers.symphony.users",
				Properties: map[string]string{
					"providers.volatilestate": "mem-state",
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
		"users-manager": map[string]providers.IProvider{
			"mem-state": &p,
		},
	}, nil)
	assert.Nil(t, err)
	return vendor
}

func TestInit(t *testing.T) {
	initVendor(t)
}

func TestAuth(t *testing.T) {
	authRequest := AuthRequest{
		UserName: "admin",
		Password: "",
	}
	data, _ := json.Marshal(authRequest)
	vendor := initVendor(t)
	response := vendor.onAuth(v1alpha2.COARequest{
		Context: context.Background(),
		Method:  "POST",
		Body:    data,
	})
	assert.Equal(t, response.State, v1alpha2.OK)
}

func TestUnauthorized(t *testing.T) {
	authRequest := AuthRequest{
		UserName: "abc",
		Password: "",
	}
	data, _ := json.Marshal(authRequest)
	vendor := initVendor(t)
	response := vendor.onAuth(v1alpha2.COARequest{
		Context: context.Background(),
		Method:  "POST",
		Body:    data,
	})
	assert.Equal(t, response.State, v1alpha2.Unauthorized)
}

func TestUsersEndpoints(t *testing.T) {
	vendor := initVendor(t)
	vendor.Route = "user"
	endpoints := vendor.GetEndpoints()
	assert.NotNil(t, endpoints)
	assert.Equal(t, "user/auth", endpoints[len(endpoints)-1].Route)
}
