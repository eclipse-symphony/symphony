/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogversion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestCatalogVersionProviderInitWithMap(t *testing.T) {
	config := map[string]string{
		"baseUrl":  "http://localhost:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	provider := CatalogVersionConfigProvider{}
	err := provider.InitWithMap(config)
	assert.Nil(t, err)

}

func TestRead(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogversions/registry/catalogversion1-v-version1":
			response = model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "catalogversion1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					ParentName: "parent:version1",
					Properties: map[string]interface{}{
						"components": []model.ComponentSpec{
							{
								Name: "name",
								Type: "type",
							},
						},
						"a": map[string]interface{}{
							"b": map[string]interface{}{
								"c": "nested",
							},
						},
						"a.b.d": "dot",
					},
				},
			}
		case "/catalogversions/registry/parent-v-version1":
			response = model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
					},
				},
			}
		default:
			response = api_utils.AuthResponse{
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
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CatalogVersionConfigProvider{}
	err := provider.Init(CatalogVersionConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.Read(ctx, "catalogversion1:version1", "components", nil)
	assert.Nil(t, err)
	data, err := json.Marshal(res)
	assert.Nil(t, err)
	var summary []model.ComponentSpec
	err = json.Unmarshal(data, &summary)
	assert.Nil(t, err)
	assert.Equal(t, "name", summary[0].Name)

	res, err = provider.Read(ctx, "catalogversion1:version1", "`.a.b.c`", nil)
	assert.Nil(t, err)
	data, err = json.Marshal(res)
	assert.Nil(t, err)
	var val string
	err = json.Unmarshal(data, &val)
	assert.Nil(t, err)
	assert.Equal(t, "nested", val)

	res, err = provider.Read(ctx, "catalogversion1:version1", "`.\"a.b.d\"`", nil)
	assert.Nil(t, err)
	data, err = json.Marshal(res)
	assert.Nil(t, err)
	err = json.Unmarshal(data, &val)
	assert.Nil(t, err)
	assert.Equal(t, "dot", val)

	res, err = provider.Read(ctx, "catalogversion1:version1", "`.parentAttribute`", nil)
	assert.Nil(t, err)
	v, ok := res.(string)
	assert.True(t, ok)
	assert.Equal(t, "This is father", v)

	res, err = provider.Read(ctx, "catalogversion1:version1", "notExist", nil)
	coaErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
	assert.Empty(t, res)
}

func TestReadObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogversions/registry/catalogversion1-v-version1":
			response = model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "catalogversion1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					ParentName: "parent:version1",
					Properties: map[string]interface{}{
						"components": map[string]interface{}{
							"Name": "name",
							"Type": "type",
						},
					},
				},
			}
		case "/catalogversions/registry/parent-v-version1":
			response = model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "parent-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					Properties: map[string]interface{}{
						"parentAttribute": "This is father",
						"components": map[string]interface{}{
							"Name": "notaname",
						},
					},
				},
			}
		default:
			response = api_utils.AuthResponse{
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
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CatalogVersionConfigProvider{}
	err := provider.Init(CatalogVersionConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	res, err := provider.ReadObject(ctx, "catalogversion1:version1", nil)
	assert.Nil(t, err)
	assert.Equal(t, "name", res["components"].(map[string]interface{})["Name"])
}

func TestSetandRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogversions/registry/catalogversion1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogVersionState{
					ObjectMeta: model.ObjectMeta{
						Name: "catalogversion1",
					},
					Spec: &model.CatalogVersionSpec{
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
			response = api_utils.AuthResponse{
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
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CatalogVersionConfigProvider{}
	err := provider.Init(CatalogVersionConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)

	err = provider.Set(ctx, "catalogversion1", "random", "random")
	assert.Nil(t, err)

	err = provider.Remove(ctx, "catalogversion1", "components")
	assert.Nil(t, err)

	err = provider.Remove(ctx, "catalogversion1", "notExist")
	coeErr := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.NotFound, coeErr.State)
}

func TestSetandRemoveObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogversions/registry/catalogversion1":
			if r.Method == http.MethodPost {
				response = nil
			} else {
				response = model.CatalogVersionState{
					ObjectMeta: model.ObjectMeta{
						Name: "catalogversion1",
					},
					Spec: &model.CatalogVersionSpec{
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
			response = api_utils.AuthResponse{
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
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CatalogVersionConfigProvider{}
	err := provider.Init(CatalogVersionConfigProviderConfig{})
	provider.Context = &contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	}
	assert.Nil(t, err)
	var data map[string]interface{} = make(map[string]interface{})
	data["random"] = "random"
	err = provider.SetObject(ctx, "catalogversion1", data)
	assert.Nil(t, err)

	err = provider.RemoveObject(ctx, "catalogversion1")
	assert.Nil(t, err)
}
