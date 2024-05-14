/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package list

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestListInitFromMap(t *testing.T) {
	UseServiceAccountTokenEnvName := os.Getenv(constants.UseServiceAccountTokenEnvName)
	if UseServiceAccountTokenEnvName != "false" {
		t.Skip("Skipping becasue UseServiceAccountTokenEnvName is not false")
	}
	provider := ListStageProvider{}
	input := map[string]string{
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	input = map[string]string{}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"user": "",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"user": "admin",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)
}

func TestListProcessInstances(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	utils.UpdateApiClientUrl(ts.URL + "/")
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
	})
	assert.Nil(t, err)
	instances, ok := outputs["items"].([]model.InstanceState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))
	assert.Equal(t, "instance1", instances[0].ObjectMeta.Name)
	assert.Equal(t, "instance2", instances[1].ObjectMeta.Name)

	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	instanceNames, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))
	assert.Equal(t, "instance1", instanceNames[0])
	assert.Equal(t, "instance2", instanceNames[1])
}

func TestListProcessSites(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	utils.UpdateApiClientUrl(ts.URL + "/")
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "sites",
	})
	assert.Nil(t, err)
	sites, ok := outputs["items"].([]model.SiteState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(sites))
	assert.Equal(t, "hq", sites[0].Id)
	assert.Equal(t, "child", sites[1].Id)

	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "sites",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	siteNames, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(siteNames))
	assert.Equal(t, "hq", siteNames[0])
	assert.Equal(t, "child", siteNames[1])
}

func TestListProcessCatalogs(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	utils.UpdateApiClientUrl(ts.URL + "/")
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "catalogs",
	})
	assert.Nil(t, err)
	catalogs, ok := outputs["items"].([]model.CatalogState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(catalogs))
	assert.Equal(t, "catalog1", catalogs[0].ObjectMeta.Name)
	assert.Equal(t, "catalog2", catalogs[1].ObjectMeta.Name)

	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "catalogs",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	catalogNames, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(catalogNames))
	assert.Equal(t, "catalog1", catalogNames[0])
	assert.Equal(t, "catalog2", catalogNames[1])
}

func TestListProcessUnsupported(t *testing.T) {
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "target",
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unsupported object type")
	assert.Nil(t, outputs)
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
		switch r.URL.Path {
		case "/instances":
			response = []model.InstanceState{
				{
					ObjectMeta: model.ObjectMeta{
						Name: "instance1",
					},
					Spec:   &model.InstanceSpec{},
					Status: model.InstanceStatus{},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "instance2",
					},
					Spec:   &model.InstanceSpec{},
					Status: model.InstanceStatus{},
				}}
		case "/federation/registry":
			response = []model.SiteState{
				{
					Id: "hq",
					Spec: &model.SiteSpec{
						Name: "hq",
					},
					Status: &model.SiteStatus{},
				},
				{
					Id: "child",
					Spec: &model.SiteSpec{
						Name: "child",
					},
					Status: &model.SiteStatus{},
				}}
		case "/catalogs/registry":
			response = []model.CatalogState{
				{
					ObjectMeta: model.ObjectMeta{
						Name: "catalog1",
					},
					Spec:   &model.CatalogSpec{},
					Status: &model.CatalogStatus{},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "catalog2",
					},
					Spec:   &model.CatalogSpec{},
					Status: &model.CatalogStatus{},
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
