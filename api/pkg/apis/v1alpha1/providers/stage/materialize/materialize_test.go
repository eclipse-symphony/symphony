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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

const catalogversionNotFoundMsg = "catalogversion not found"

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
		"names":           []interface{}{"instance1:version1", "target1:version1", "solutionversion1:version1", "catalogversion1:version1"},
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
		"names":    []interface{}{"instance1:version1", "target1:version1", "solutionversion1:version1", "catalogversion1:version1"},
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
	assert.Contains(t, err.Error(), catalogversionNotFoundMsg)
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
		case "/solutionversions/solutionversion1-v-version1":
			var solutionversion model.SolutionVersionState
			err := json.Unmarshal(body, &solutionversion)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, solutionversion.ObjectMeta.Namespace)
			response = solutionversion
		case "/catalogversions/registry/hq-target1-v-version1":
			catalogversion := model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-target1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
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
			response = catalogversion
		case "/catalogversions/registry/hq-instance1-v-version1":
			catalogversion := model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-instance1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
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
			response = catalogversion
		case "/catalogversions/registry/hq-solutionversion1-v-version1":
			catalogversion := model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-solutionversion1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					CatalogType: "solutionversion",
					Properties: map[string]interface{}{
						"spec": model.SolutionVersionSpec{
							DisplayName: "solutionversion1",
						},
						"metadata": &model.ObjectMeta{
							Namespace: "objNS",
							Name:      "instance1:version1",
						},
					},
				},
			}
			response = catalogversion
		case "/catalogversions/registry/hq-catalogversion1-v-version1":
			catalogversion := model.CatalogVersionState{
				ObjectMeta: model.ObjectMeta{
					Name: "hq-catalogversion1-v-version1",
				},
				Spec: &model.CatalogVersionSpec{
					CatalogType: "catalogversion",
					Properties: map[string]interface{}{
						"spec": model.CatalogVersionSpec{
							CatalogType: "config",
							Properties:  map[string]interface{}{},
						},
						"metadata": &model.ObjectMeta{
							Namespace: "objNS",
							Name:      "catalogversion1:version1",
						},
					},
				},
			}
			response = catalogversion
		case "/catalogversions/registry/catalogversion1-v-version1":
			var catalogversion model.CatalogVersionState
			err := json.Unmarshal(body, &catalogversion)
			assert.Nil(t, err)
			assert.Equal(t, expectNs, catalogversion.ObjectMeta.Namespace)
			response = catalogversion
		case "/catalogversions/registry/hq-notexist":
			http.Error(w, catalogversionNotFoundMsg, http.StatusNotFound)
			return
		default:
			response = utils.AuthResponse{
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
