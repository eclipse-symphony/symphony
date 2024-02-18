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
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestListInitFromMap(t *testing.T) {
	provider := ListStageProvider{}
	input := map[string]string{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)

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
func TestListProcessInstances(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
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
	instances, ok := outputs["items"].([]model.SiteState)
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))
	assert.Equal(t, "hq", instances[0].Id)
	assert.Equal(t, "child", instances[1].Id)

	outputs, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "sites",
		"namesOnly":  true,
	})
	assert.Nil(t, err)
	instanceNames, ok := outputs["items"].([]string)
	assert.True(t, ok)
	assert.Equal(t, 2, len(instances))
	assert.Equal(t, "hq", instanceNames[0])
	assert.Equal(t, "child", instanceNames[1])
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
					Spec: &model.InstanceSpec{
						Name: "instance1",
					},
					Status: model.InstanceStatus{},
				},
				{
					ObjectMeta: model.ObjectMeta{
						Name: "instance2",
					},
					Spec: &model.InstanceSpec{
						Name: "instance2",
					},
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
