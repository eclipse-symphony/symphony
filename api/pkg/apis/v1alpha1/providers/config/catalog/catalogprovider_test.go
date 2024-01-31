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
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

// func TestRead(t *testing.T) {
// 	// To make this test work, you'll need these configurations:
// 	// ai-config: flavor=cloud, model=gpt, version=4.5
// 	// ai-config-site: model=LLaMA, version=3.3
// 	// ai-config-line: flavor=mobile
// 	// combined: ai=<ai-config>, ai-mode=<ai-config>.model, com:<combined-1>.foo, less=<123, e4k=<e4k-config>, influxdb=<influx-db-config>, loop=<combined-1>.loop
// 	// combined-1: foo=<combined-2>.foo, loop=<combined-2>.loop
// 	// combined-2: foo=bar2, loop=<combined>.loop
// 	// os.Setenv("CATALOG_API_URL", "http://localhost:8080/v1alpha2/")
// 	// os.Setenv("CATALOG_API_USER", "admin")
// 	// os.Setenv("CATALOG_API_PASSWORD", "")
// 	catalogAPIUrl := os.Getenv("CATALOG_API_URL")
// 	if catalogAPIUrl == "" {
// 		t.Skip("Skipping becasue CATALOG_API_URL is missing or not set to 'yes'")
// 	}
// 	catalogAPIUser := os.Getenv("CATALOG_API_USER")
// 	catalogAPIPassword := os.Getenv("CATALOG_API_PASSWORD")

// 	provider := CatalogConfigProvider{}
// 	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: catalogAPIUrl, User: catalogAPIUser, Password: catalogAPIPassword})
// 	assert.Nil(t, err)

// 	value, err := provider.Read("ai-config", "model", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "gpt", value)
// 	value, err = provider.Read("ai-config-site", "model", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "LLaMA", value)
// 	value, err = provider.Read("ai-config-line", "model", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "LLaMA", value)
// 	value, err = provider.Read("ai-config", "flavor", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "cloud", value)
// 	value, err = provider.Read("ai-config-site", "flavor", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "cloud", value)
// 	value, err = provider.Read("ai-config-line", "flavor", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "mobile", value)
// 	value, err = provider.Read("ai-config", "version", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "4.5", value)
// 	value, err = provider.Read("ai-config-site", "version", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "3.3", value)
// 	value, err = provider.Read("ai-config-line", "version", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "3.3", value)
// 	value, err = provider.Read("combined", "ai-model", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "gpt", value)
// 	value, err = provider.Read("combined", "ai", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "{\"flavor\":\"cloud\",\"model\":\"gpt\",\"version\":\"4.5\"}", value)
// 	value, err = provider.Read("combined", "com", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "bar2", value)
// 	value, err = provider.Read("combined", "less", nil)
// 	assert.Nil(t, err)
// 	assert.Equal(t, "<123", value)
// 	// TODO: needs a good way to detect reference loops
// 	// value, err = provider.Read("combined", "loop")
// 	// assert.NotNil(t, err)
// }

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
		case "/catalogs/registry/catalog1":
			response = model.CatalogState{
				Id: "catalog1",
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
				},
			}
		case "/catalogs/registry/parent":
			response = model.CatalogState{
				Id: "parent",
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
					},
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

	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: ts.URL + "/", User: "admin", Password: ""})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.Read("catalog1", "components", nil)
	assert.Nil(t, err)
	data, err := json.Marshal(res)
	assert.Nil(t, err)
	var summary []model.ComponentSpec
	err = json.Unmarshal(data, &summary)
	assert.Nil(t, err)
	assert.Equal(t, "name", summary[0].Name)

	res, err = provider.Read("catalog1", "parentAttribute", nil)
	assert.Nil(t, err)
	v, ok := res.(string)
	assert.True(t, ok)
	assert.Equal(t, "This is father", v)

	res, err = provider.Read("catalog1", "notExist", nil)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
	assert.Empty(t, res)
}

func TestReadObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1":
			response = model.CatalogState{
				Id: "catalog1",
				Spec: &model.CatalogSpec{
					ParentName: "parent",
					Properties: map[string]interface{}{
						"components": map[string]interface{}{
							"Name": "name",
							"Type": "type",
						},
					},
				},
			}
		case "/catalogs/registry/parent":
			response = model.CatalogState{
				Id: "parent",
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
					},
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

	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: ts.URL + "/", User: "admin", Password: ""})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.ReadObject("catalog1", nil)
	assert.Nil(t, err)
	assert.Equal(t, "name", res["Name"])
}

func TestSetandRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogState{
					Id: "catalog1",
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

	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: ts.URL + "/", User: "admin", Password: ""})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	err = provider.Set("catalog1", "random", "random")
	assert.Nil(t, err)

	err = provider.Remove("catalog1", "components")
	assert.Nil(t, err)

	err = provider.Remove("catalog1", "notExist")
	coeErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coeErr.State)
}

func TestSetandRemoveObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/catalog1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogState{
					Id: "catalog1",
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

	provider := CatalogConfigProvider{}
	err := provider.Init(CatalogConfigProviderConfig{BaseUrl: ts.URL + "/", User: "admin", Password: ""})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)
	var data map[string]interface{} = make(map[string]interface{})
	data["random"] = "random"
	err = provider.SetObject("catalog1", data)
	assert.Nil(t, err)

	err = provider.RemoveObject("catalog1")
	assert.Nil(t, err)
}
