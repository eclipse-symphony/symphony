/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE

*/

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/vendors"
	sym_mgr "github.com/azure/symphony/api/pkg/apis/v1alpha1/managers"
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
		sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"users-manager": map[string]providers.IProvider{
			"mem-state": &p,
		},
	})
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
