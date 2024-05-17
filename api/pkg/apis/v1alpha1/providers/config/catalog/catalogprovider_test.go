/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func TestCatalogProviderInitWithMap(t *testing.T) {
	config := map[string]string{
		"baseUrl":  "http://localhost:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	provider := CatalogConfigProvider{}
	err := provider.InitWithMap(config)
	assert.Nil(t, err)

}

func TestRead(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "catalog1-v1",
				},
				Spec: &model.CatalogSpec{
					ParentName: "parent:v1",
					Properties: map[string]interface{}{
						"components": []model.ComponentSpec{
							{
								Name: "name",
								Type: "type",
							},
						},
					},
					Version:      "v1",
					RootResource: "catalog1",
				},
			}
		case "/catalogs/registry/parent/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
					},
					Version:      "v1",
					RootResource: "parent",
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
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.Read("catalog1:v1", "components", nil)
	assert.Nil(t, err)
	data, err := json.Marshal(res)
	assert.Nil(t, err)
	var summary []model.ComponentSpec
	err = json.Unmarshal(data, &summary)
	assert.Nil(t, err)
	assert.Equal(t, "name", summary[0].Name)

	res, err = provider.Read("catalog1:v1", "parentAttribute", nil)
	assert.Nil(t, err)
	v, ok := res.(string)
	assert.True(t, ok)
	assert.Equal(t, "This is father", v)

	res, err = provider.Read("catalog1:v1", "notExist", nil)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
	assert.Empty(t, res)
}

func TestReadObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "catalog1-v1",
				},
				Spec: &model.CatalogSpec{
					ParentName: "parent:v1",
					Properties: map[string]interface{}{
						"components": map[string]interface{}{
							"Name": "name",
							"Type": "type",
						},
					},
					Version:      "v1",
					RootResource: "catalog1",
				},
			}
		case "/catalogs/registry/parent/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
					},
					Version:      "v1",
					RootResource: "parent",
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
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.ReadObject("catalog1:v1", nil)
	assert.Nil(t, err)
	assert.Equal(t, "name", res["Name"])
}

func TestSetandRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1/v1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogState{
					ObjectMeta: model.ObjectMeta{
						Name: "catalog1-v",
					},
					Spec: &model.CatalogSpec{
						ParentName: "parent",
						Properties: map[string]interface{}{
							"components": []model.ComponentSpec{
								{
									Name: "name",
									Type: "type",
								},
							},
						},
						Version:      "v1",
						RootResource: "catalog1",
					},
				}
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
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	err = provider.Set("catalog1:v1", "random", "random")
	assert.Nil(t, err)

	err = provider.Remove("catalog1:v1", "components")
	assert.Nil(t, err)

	err = provider.Remove("catalog1:v1", "notExist")
	coeErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coeErr.State)
}

func TestSetandRemoveObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1/v1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogState{
					ObjectMeta: model.ObjectMeta{
						Name: "catalog1-v1",
					},
					Spec: &model.CatalogSpec{
						ParentName: "parent",
						Properties: map[string]interface{}{
							"components": []model.ComponentSpec{
								{
									Name: "name",
									Type: "type",
								},
							},
						},
						Version:      "v1",
						RootResource: "catalog1",
					},
				}
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
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)
	var data map[string]interface{} = make(map[string]interface{})
	data["random"] = "random"
	err = provider.SetObject("catalog1:v1", data)
	assert.Nil(t, err)

	err = provider.RemoveObject("catalog1:v1")
	assert.Nil(t, err)
}
