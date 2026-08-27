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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
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
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
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
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
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

func TestListProcessCatalogVersions(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "catalogversions",
	})
	assert.Nil(t, err)
	catalogversions, ok := outputs["items"].([]model.CatalogVersionState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(catalogversions))
	assert.Equal(t, "catalogversion1", catalogversions[0].ObjectMeta.Name)
	assert.Equal(t, "catalogversion2", catalogversions[1].ObjectMeta.Name)

	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "catalogversions",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	catalogversionNames, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(catalogversionNames))
	assert.Equal(t, "catalogversion1", catalogversionNames[0])
	assert.Equal(t, "catalogversion2", catalogversionNames[1])
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

func TestListProcessActivations(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "activations",
	})
	assert.Nil(t, err)
	activations, ok := outputs["items"].([]model.ActivationState)
	assert.True(t, ok)
	assert.Equal(t, 3, len(activations))
	assert.Equal(t, 3, outputs["itemCount"])

	// names only
	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "activations",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	names, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"activation1", "activation2", "activation3"}, names)
}

func TestListProcessActivationsWithSingleFilter(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := ListStageProvider{}
	err := provider.InitWithMap(map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	})
	assert.Nil(t, err)

	// keep only completed (Done = 9996) activations using single-filter inputs
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":     "activations",
		"filterField":    "status.status",
		"filterValue":    "9996",
		"filterOperator": "eq",
	})
	assert.Nil(t, err)
	activations, ok := outputs["items"].([]model.ActivationState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(activations))
	assert.Equal(t, 2, outputs["itemCount"])
	assert.Equal(t, "activation1", activations[0].ObjectMeta.Name)
	assert.Equal(t, "activation3", activations[1].ObjectMeta.Name)
}

func TestListProcessActivationsWithMultipleFilters(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := ListStageProvider{}
	err := provider.InitWithMap(map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	})
	assert.Nil(t, err)

	// Combine a label filter (belongs to parent) with a status filter (completed).
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "activations",
		"filter": []interface{}{
			map[string]interface{}{"field": "metadata.labels.parentActivation", "value": "parent1", "operator": "eq"},
			map[string]interface{}{"field": "status.status", "value": "9996", "operator": "eq"},
		},
	})
	assert.Nil(t, err)
	activations, ok := outputs["items"].([]model.ActivationState)
	assert.True(t, ok)
	assert.Equal(t, 1, len(activations))
	assert.Equal(t, "activation1", activations[0].ObjectMeta.Name)
}

func TestListProcessActivationsWithExistsFilter(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := ListStageProvider{}
	err := provider.InitWithMap(map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	})
	assert.Nil(t, err)

	// "ne" against Running (9994) keeps everything that is not running.
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":     "activations",
		"filterField":    "status.status",
		"filterValue":    "9994",
		"filterOperator": "ne",
	})
	assert.Nil(t, err)
	activations, ok := outputs["items"].([]model.ActivationState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(activations))
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
		case "/catalogversions/registry":
			response = []model.CatalogVersionState{
				{
					ObjectMeta: model.ObjectMeta{
						Name: "catalogversion1",
					},
					Spec:   &model.CatalogVersionSpec{},
					Status: &model.CatalogVersionStatus{},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "catalogversion2",
					},
					Spec:   &model.CatalogVersionSpec{},
					Status: &model.CatalogVersionStatus{},
				}}
		case "/activations/registry":
			response = []model.ActivationState{
				{
					ObjectMeta: model.ObjectMeta{
						Name:   "activation1",
						Labels: map[string]string{"parentActivation": "parent1"},
					},
					Spec:   &model.ActivationSpec{CampaignVersion: "child:v1"},
					Status: &model.ActivationStatus{Status: v1alpha2.Done},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name:   "activation2",
						Labels: map[string]string{"parentActivation": "parent1"},
					},
					Spec:   &model.ActivationSpec{CampaignVersion: "child:v1"},
					Status: &model.ActivationStatus{Status: v1alpha2.Running},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name:   "activation3",
						Labels: map[string]string{"parentActivation": "parent2"},
					},
					Spec:   &model.ActivationSpec{CampaignVersion: "child:v1"},
					Status: &model.ActivationStatus{Status: v1alpha2.Done},
				}}
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
