/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package wait

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestWaitInitFromVendorMap(t *testing.T) {
	input := map[string]string{
		"wait.wait.interval": "15",
		"wait.wait.count":    "10",
	}
	config, err := WaitStageProviderConfigFromVendorMap(input)
	assert.Nil(t, err)
	assert.Equal(t, 15, config.WaitInterval)
	assert.Equal(t, 10, config.WaitCount)

	input = map[string]string{
		"wait.wait.interval": "abc",
	}
	config, err = WaitStageProviderConfigFromVendorMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"wait.wait.interval": "15",
		"wait.wait.count":    "abc",
	}
	config, err = WaitStageProviderConfigFromVendorMap(input)
	assert.NotNil(t, err)
}

func TestWaitProcess(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	utils.UpdateApiClientUrl(ts.URL + "/")
	config := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider := WaitStageProvider{}
	err := provider.InitWithMap(config)
	assert.Nil(t, err)

	// instances exist
	input := map[string]interface{}{
		"objectType": "instance",
		"names":      []interface{}{"instance1"},
		"__origin":   "hq",
	}
	output, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.Nil(t, err)
	assert.Equal(t, "OK", output["status"])
	assert.Equal(t, "instance", output["objectType"])

	// instances not exist
	input = map[string]interface{}{
		"objectType": "instance",
		"names":      []interface{}{"instance2"},
		"__origin":   "hq",
	}
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to wait for")

	// sites exist
	input = map[string]interface{}{
		"objectType": "sites",
		"names":      []interface{}{"site1"},
		"__origin":   "hq",
	}
	output, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.Nil(t, err)
	assert.Equal(t, "OK", output["status"])
	assert.Equal(t, "sites", output["objectType"])

	// sites not exist
	input = map[string]interface{}{
		"objectType": "sites",
		"names":      []interface{}{"site2"},
		"__origin":   "hq",
	}
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to wait for")

	// catalogs exist
	input = map[string]interface{}{
		"objectType": "catalogs",
		"names":      []interface{}{"catalog1"},
		"__origin":   "hq",
	}
	output, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.Nil(t, err)
	assert.Equal(t, "OK", output["status"])
	assert.Equal(t, "catalogs", output["objectType"])

	// catalogs not exist
	input = map[string]interface{}{
		"objectType": "catalogs",
		"names":      []interface{}{"catalog2"},
		"__origin":   "hq",
	}
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, input)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to wait for")
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func InitializeMockSymphonyAPI() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		log.Info("Mock Symphony API called", "path", r.URL.Path)
		switch r.URL.Path {
		case "/instances":
			response = []model.InstanceState{{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-instance1",
				},
				Spec:   &model.InstanceSpec{},
				Status: model.InstanceStatus{},
			}}
		case "/federation/registry":
			response = []model.SiteState{{
				Id: "site1",
				Spec: &model.SiteSpec{
					Name: "hq-site1",
				},
				Status: &model.SiteStatus{},
			}}
		case "/catalogs/registry":
			response = []model.CatalogState{{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-catalog1",
				},
				Spec: &model.CatalogSpec{},
			}}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	return ts
}
