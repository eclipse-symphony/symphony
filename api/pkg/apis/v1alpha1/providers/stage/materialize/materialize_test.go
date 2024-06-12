/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package materialize

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestMaterializeInitForNonServiceAccount(t *testing.T) {
	UseServiceAccountTokenEnvName := os.Getenv(constants.UseServiceAccountTokenEnvName)
	if UseServiceAccountTokenEnvName != "false" {
		t.Skip("Skipping becasue UseServiceAccountTokenEnvName is not false")
	}
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	assert.Equal(t, "admin", provider.Config.User)
	assert.Equal(t, "", provider.Config.Password)

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
}
func TestMaterializeProcessWithStageNs(t *testing.T) {
	stageNs := "testns"
	ts := InitializeMockSymphonyAPI(t, stageNs)
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
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
	_, paused, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":           []interface{}{"instance1:v1", "target1:v1", "solution1:v1", "catalog1:v1"},
		"__origin":        "hq",
		"objectNamespace": stageNs,
	})
	assert.Nil(t, err)
	assert.False(t, paused)
}

func TestMaterializeProcessWithoutStageNs(t *testing.T) {
	ts := InitializeMockSymphonyAPI(t, "objNS")
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
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
	_, paused, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":    []interface{}{"instance1:v1", "target1:v1", "solution1:v1", "catalog1:v1"},
		"__origin": "hq",
	})
	assert.Nil(t, err)
	assert.False(t, paused)
}

func TestMaterializeProcessFailedCase(t *testing.T) {
	ts := InitializeMockSymphonyAPI(t, "objNS")
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":    []interface{}{"instance1:v1", "target1:v1", "solution1:v1, target2:v1"},
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

func InitializeMockSymphonyAPI(t *testing.T, expectNs string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		body, _ := io.ReadAll(r.Body)
		switch r.URL.Path {
		case "/instances/instance1-v1":
			instance := model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1-v1",
				},
			}
			err := json.Unmarshal(body, &instance)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, instance.ObjectMeta.Namespace)
			response = instance
		case "/targets/registry/target1-v1":
			target := model.TargetState{
				ObjectMeta: model.ObjectMeta{
					Name: "target1-v1",
				},
			}
			err := json.Unmarshal(body, &target)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, target.ObjectMeta.Namespace)
			response = target
		case "/solutions/solution1-v1":
			var solution model.SolutionState
			err := json.Unmarshal(body, &solution)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, solution.ObjectMeta.Namespace)
			response = solution
		case "/catalogs/registry":
			response = []model.CatalogState{
				{
					ObjectMeta: model.ObjectMeta{
						Name: "hq-target1-v1",
					},
					Spec: &model.CatalogSpec{
						CatalogType: "target",
						Properties: map[string]interface{}{
							"spec": &model.TargetSpec{
								DisplayName: "target1",
							},
							"metadata": &model.ObjectMeta{
								Name:      "target1:v1",
								Namespace: "objNS",
							},
						},
					},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "hq-instance1-v1",
					},
					Spec: &model.CatalogSpec{
						CatalogType: "instance",
						Properties: map[string]interface{}{
							"spec": model.InstanceSpec{},
							"metadata": &model.ObjectMeta{
								Namespace: "objNS",
								Name:      "instance1:v1",
							},
						},
					},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "hq-solution1-v1",
					},
					Spec: &model.CatalogSpec{
						CatalogType: "solution",
						Properties: map[string]interface{}{
							"spec": model.SolutionSpec{
								DisplayName: "solution1",
							},
							"metadata": &model.ObjectMeta{
								Namespace: "objNS",
								Name:      "instance1:v1",
							},
						},
					},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "hq-catalog1-v1",
					},
					Spec: &model.CatalogSpec{
						CatalogType: "catalog",
						Properties: map[string]interface{}{
							"spec": model.CatalogSpec{
								CatalogType: "config",
								Properties:  map[string]interface{}{},
							},
							"metadata": &model.ObjectMeta{
								Namespace: "objNS",
								Name:      "catalog1:v1",
							},
						},
					},
				},
			}
		case "catalogs/registry/catalog1-v1":
			var catalog model.CatalogState
			err := json.Unmarshal(body, &catalog)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, catalog.ObjectMeta.Namespace)
			response = catalog
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
