/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package materialize

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestMaterializeInit(t *testing.T) {
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	assert.Equal(t, "http://symphony-service:8080/v1alpha2/", provider.Config.BaseUrl)
	assert.Equal(t, "admin", provider.Config.User)
	assert.Equal(t, "", provider.Config.Password)

	input = map[string]string{}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "admin",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)
}

func TestMaterializeInitFromVendorMap(t *testing.T) {
	input := map[string]string{
		"wait.baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"wait.user":     "admin",
		"wait.password": "",
	}
	config, err := MaterializeStageProviderConfigFromVendorMap(input)
	assert.Nil(t, err)
	provider := MaterializeStageProvider{}
	provider.Init(config)
	assert.Equal(t, "http://symphony-service:8080/v1alpha2/", provider.Config.BaseUrl)
	assert.Equal(t, "admin", provider.Config.User)
	assert.Equal(t, "", provider.Config.Password)
}
func TestMaterializeProcess(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	provider.SetContext(&contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	})
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":    []interface{}{"instance1", "target1", "solution1", "catalog1"},
		"__origin": "hq",
	})
	assert.Nil(t, err)
}

func TestMaterializeProcessFailedCase(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":    []interface{}{"instance1", "target1", "solution1, target2"},
		"__origin": "hq",
	})
	assert.NotNil(t, err)
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
		case "/instances/instance1":
			response = model.InstanceState{
				Id: "hq-instance1",
				Spec: &model.InstanceSpec{
					Name: "hq-instance1",
				},
				Status: map[string]string{},
			}
		case "/targets/registry/target1":
			response = model.TargetState{
				Id: "hq-target1",
				Spec: &model.TargetSpec{
					DisplayName: "hq-target1",
				},
			}
		case "/solutions/solution1":
			response = model.SolutionState{
				Id: "hq-solution1",
				Spec: &model.SolutionSpec{
					DisplayName: "hq-solution1",
				},
			}
		case "/catalogs/registry":
			response = []model.CatalogState{
				{
					Id: "targetcatalog",
					Spec: &model.CatalogSpec{
						Type: "target",
						Name: "hq-target1",
						Properties: map[string]interface{}{
							"spec": &model.TargetSpec{
								DisplayName: "target1",
							},
						},
					},
				},
				{
					Id: "instancecatalog",
					Spec: &model.CatalogSpec{
						Type: "instance",
						Name: "hq-instance1",
						Properties: map[string]interface{}{
							"spec": model.InstanceSpec{
								Name: "instance1",
							},
						},
					},
				},
				{
					Id: "solutioncatalog",
					Spec: &model.CatalogSpec{
						Type: "solution",
						Name: "hq-solution1",
						Properties: map[string]interface{}{
							"spec": model.SolutionSpec{
								DisplayName: "solution1",
							},
						},
					},
				},
				{
					Id: "catalog1",
					Spec: &model.CatalogSpec{
						Type: "catalog",
						Name: "hq-catalog1",
						Properties: map[string]interface{}{
							"spec": model.SolutionSpec{
								DisplayName: "solution1",
							},
						},
					},
				},
			}
		case "catalogs/registry/catalog1":
			response = model.CatalogState{
				Id: "catalog1",
				Spec: &model.CatalogSpec{
					Name: "catalog1",
				},
			}
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
