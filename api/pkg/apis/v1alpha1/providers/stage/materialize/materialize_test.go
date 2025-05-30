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

const catalogNotFoundMsg = "catalog not found"

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
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
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
		"names":           []interface{}{"instance1:version1", "target1:version1", "solution1:version1", "catalog1:version1"},
		"__origin":        "hq",
		"objectNamespace": stageNs,
	})
	assert.Nil(t, err)
	assert.False(t, paused)
}

func TestMaterializeProcessWithoutStageNs(t *testing.T) {
	ts := InitializeMockSymphonyAPI(t, "objNS")
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
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
		"names":    []interface{}{"instance1:version1", "target1:version1", "solution1:version1", "catalog1:version1"},
		"__origin": "hq",
	})
	assert.Nil(t, err)
	assert.False(t, paused)
}

func TestMaterializeProcessFailedCase(t *testing.T) {
	ts := InitializeMockSymphonyAPI(t, "objNS")
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := MaterializeStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"names":    []interface{}{"instance1:version1", "target1:version1", "notexist"},
		"__origin": "hq",
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), catalogNotFoundMsg)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func InitializeMockSymphonyAPI(t *testing.T, expectNs string) *httptest.Server {
	requestCounts := make(map[string]int)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		body, _ := io.ReadAll(r.Body)
		requestCounts[r.URL.Path]++
		count := requestCounts[r.URL.Path]
		switch r.URL.Path {
		case "/instances/instance1-v-version1":
			instance := model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1-v-version1",
				},
			}
			if r.Method == "POST" {
				// Check create instance correctness
				err := json.Unmarshal(body, &instance)
				assert.Nil(t, err)
				assert.Equal(t, expectNs, instance.ObjectMeta.Namespace)
			} else if r.Method == "GET" && count == 1 {
				// Mock GUID for get
				instance.ObjectMeta.SetGuid("test-guid")
				instance.ObjectMeta.SetSummaryJobId("1")
			} else if r.Method == "GET" && count >= 2 {
				// Mock GUID for get
				instance.ObjectMeta.SetGuid("test-guid")
				instance.ObjectMeta.SetSummaryJobId("2")
			}
			response = instance
		case "/targets/registry/target1-v-version1":
			target := model.TargetState{
				ObjectMeta: model.ObjectMeta{
					Name: "target1-v-version1",
				},
			}
			if r.Method == "POST" {
				err := json.Unmarshal(body, &target)
				assert.Nil(t, err)
				assert.Equal(t, expectNs, target.ObjectMeta.Namespace)
			} else if r.Method == "GET" && count == 1 {
				// Mock GUID for get
				target.ObjectMeta.SetGuid("test-guid")
				target.ObjectMeta.SetSummaryJobId("1")
			} else if r.Method == "GET" && count >= 2 {
				// Mock GUID for get
				target.ObjectMeta.SetGuid("test-guid")
				target.ObjectMeta.SetSummaryJobId("2")
			}
			response = target
		case "/solutions/solution1-v-version1":
			var solution model.SolutionState
			err := json.Unmarshal(body, &solution)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, solution.ObjectMeta.Namespace)
			response = solution
		case "/catalogs/registry/hq-target1-v-version1":
			catalog := model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-target1-v-version1",
				},
				Spec: &model.CatalogSpec{
					CatalogType: "target",
					Properties: map[string]interface{}{
						"spec": &model.TargetSpec{
							DisplayName: "target1",
						},
						"metadata": &model.ObjectMeta{
							Name:      "target1:version1",
							Namespace: "objNS",
						},
					},
				},
			}
			response = catalog
		case "/catalogs/registry/hq-instance1-v-version1":
			catalog := model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-instance1-v-version1",
				},
				Spec: &model.CatalogSpec{
					CatalogType: "instance",
					Properties: map[string]interface{}{
						"spec": model.InstanceSpec{},
						"metadata": &model.ObjectMeta{
							Namespace: "objNS",
							Name:      "instance1:version1",
						},
					},
				},
			}
			response = catalog
		case "/catalogs/registry/hq-solution1-v-version1":
			catalog := model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-solution1-v-version1",
				},
				Spec: &model.CatalogSpec{
					CatalogType: "solution",
					Properties: map[string]interface{}{
						"spec": model.SolutionSpec{
							DisplayName: "solution1",
						},
						"metadata": &model.ObjectMeta{
							Namespace: "objNS",
							Name:      "instance1:version1",
						},
					},
				},
			}
			response = catalog
		case "/catalogs/registry/hq-catalog1-v-version1":
			catalog := model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-catalog1-v-version1",
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
							Name:      "catalog1:version1",
						},
					},
				},
			}
			response = catalog
		case "/catalogs/registry/catalog1-v-version1":
			var catalog model.CatalogState
			err := json.Unmarshal(body, &catalog)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, catalog.ObjectMeta.Namespace)
			response = catalog
		case "/catalogs/registry/hq-notexist":
			http.Error(w, catalogNotFoundMsg, http.StatusNotFound)
			return
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
