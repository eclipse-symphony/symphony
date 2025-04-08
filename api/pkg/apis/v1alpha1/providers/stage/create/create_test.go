/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package create

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestDeployInstance(t *testing.T) {
	testDeploy := os.Getenv("TEST_DEPLOY_INSTANCE")
	if testDeploy != "yes" {
		t.Skip("Skipping becasue TEST_DEPLOY_INSTANCE is missing or not set to 'yes'")
	}
	provider := CreateStageProvider{}
	err := provider.Init(CreateStageProviderConfig{
		WaitCount:    3,
		WaitInterval: 5,
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "redis-server",
		"object": map[string]interface{}{
			"displayName": "redis-server",
			"solution":    "sample-redis",
			"target": map[string]interface{}{
				"name": "sample-docker-target",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "OK", outputs["status"])
}

func TestCreateInitFromVendorMap(t *testing.T) {
	provider := CreateStageProvider{}
	input := map[string]string{
		"wait.interval": "1",
		"wait.count":    "3",
	}
	config, err := SymphonyStageProviderConfigFromMap(input)
	assert.Nil(t, err)
	assert.Equal(t, 1, config.WaitInterval)
	assert.Equal(t, 3, config.WaitCount)
	err = provider.InitWithMap(input)
	assert.Nil(t, err)

	input = map[string]string{}
	config, err = SymphonyStageProviderConfigFromMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"wait.count": "abc",
	}
	config, err = SymphonyStageProviderConfigFromMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"wait.count":    "15",
		"wait.interval": "abc",
	}
	config, err = SymphonyStageProviderConfigFromMap(input)
	assert.NotNil(t, err)
}

func TestCreateInitFromVendorMapForNonServiceAccount(t *testing.T) {
	UseServiceAccountTokenEnvName := os.Getenv(constants.UseServiceAccountTokenEnvName)
	if UseServiceAccountTokenEnvName != "false" {
		t.Skip("Skipping becasue UseServiceAccountTokenEnvName is not false")
	}
	provider := CreateStageProvider{}
	input := map[string]string{
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	config, err := SymphonyStageProviderConfigFromMap(input)
	assert.Nil(t, err)
	assert.Equal(t, "admin", config.User)
	assert.Equal(t, 1, config.WaitInterval)
	assert.Equal(t, 3, config.WaitCount)
	err = provider.InitWithMap(input)
	assert.Nil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "",
	}
	config, err = SymphonyStageProviderConfigFromMap(input)
	assert.NotNil(t, err)
	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "admin",
	}
	config, err = SymphonyStageProviderConfigFromMap(input)
	assert.NotNil(t, err)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func TestCreateProcessCreate(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	instance := model.InstanceState{
		Spec: &model.InstanceSpec{
			DisplayName: "instance1",
			Solution:    "sample:version1",
		},
	}
	_, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "instance1",
		"action":     "create",
		"object":     instance,
	})
	assert.Nil(t, err)
}

func TestCreateProcessCreateFailedCase(t *testing.T) {
	ts := InitializeMockSymphonyAPIFailedCase()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	instance := model.InstanceState{
		Spec: &model.InstanceSpec{
			DisplayName: "instance1",
			Solution:    "sample:version1",
		},
	}
	_, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "instance1",
		"action":     "create",
		"object":     instance,
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Instance creation reconcile failed:")
}

func TestCreateProcessRemove(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	instance := model.InstanceState{
		Spec: &model.InstanceSpec{
			DisplayName: "instance1",
		},
	}
	_, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "instance1",
		"action":     "remove",
		"object":     instance,
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Instance deletion reconcile timeout")
}

func TestCreateSolutionRemove(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	solution := model.SolutionState{
		Spec: &model.SolutionSpec{
			DisplayName: "solution1",
		},
	}
	_, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "solution",
		"objectName": "solution1",
		"action":     "remove",
		"object":     solution,
	})
	assert.Nil(t, err)
}

func TestCreateSolutionCreate(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       ts.URL + "/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	solution := model.SolutionState{
		Spec: &model.SolutionSpec{
			DisplayName: "solution1",
		},
	}
	outpts, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "solution",
		"objectName": "sample:version1",
		"action":     "create",
		"object":     solution,
	})
	assert.Nil(t, err)
	assert.Equal(t, "sample-v-version1", outpts["objectName"])
	assert.Equal(t, "solution", outpts["objectType"])
}

func TestCreateProcessUnsupported(t *testing.T) {
	provider := CreateStageProvider{}
	input := map[string]string{
		"baseUrl":       "http://symphony-service:8080/v1alpha2/",
		"user":          "admin",
		"password":      "",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	provider.InitWithMap(input)
	instance := model.InstanceState{
		Spec: &model.InstanceSpec{
			DisplayName: "instance1",
			Solution:    "solution1",
		},
	}
	_, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "instance",
		"objectName": "instance1",
		"action":     "upsert",
		"object":     instance,
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unsupported action:")

	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType": "catalog",
		"objectName": "catalog1",
		"action":     "delete",
		"object":     model.SolutionState{},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unsupported object type:")

}

func InitializeMockSymphonyAPI() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		statuCode := 200
		switch r.URL.Path {
		case "/instances/instance1":
			response = model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1",
					Annotations: map[string]string{
						"Guid": "test-guid",
					},
				},
				Spec:   &model.InstanceSpec{},
				Status: model.InstanceStatus{},
			}
		case "/solution/queue":
			response = model.SummaryResult{
				Summary: model.SummarySpec{
					TargetCount:         1,
					SuccessCount:        1,
					AllAssignedDeployed: true,
				},
				State: model.SummaryStateDone,
			}
		case "/solutions/sample-v-version1":
			response = nil
		case "/solutioncontainers/sample":
			if r.Method == http.MethodGet {
				statuCode = 404
			} else {
				statuCode = 200
			}
		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}
		w.WriteHeader(statuCode)
		json.NewEncoder(w).Encode(response)
	}))
	return ts
}

func InitializeMockSymphonyAPIFailedCase() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/instances/instance1":
			response = model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1",
					Annotations: map[string]string{
						"Guid": "test-guid",
					},
				},
				Spec:   &model.InstanceSpec{},
				Status: model.InstanceStatus{},
			}
		case "/solution/queue":
			response = model.SummaryResult{
				Summary: model.SummarySpec{
					TargetCount:  2,
					SuccessCount: 1,
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
