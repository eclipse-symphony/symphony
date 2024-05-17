/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package staging

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestStagingTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := StagingProviderConfigFromMap(nil)
	assert.NotNil(t, err)
}
func TestStagingTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := StagingProviderConfigFromMap(map[string]string{})
	assert.NotNil(t, err)
}
func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":       "tiny",
		"targetName": "tiny-edge",
	}
	provider := StagingTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}
func TestStagingTargetProviderGet(t *testing.T) {
	// os.Setenv("SYMPHONY_API_BASE_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("SYMPHONY_API_USER", "admin")
	// os.Setenv("SYMPHONY_API_PASSWORD", "")
	symphonyUrl := os.Getenv("SYMPHONY_API_BASE_URL")
	if symphonyUrl == "" {
		t.Skip("Skipping because SYMPHONY_API_BASE_URL enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		Name:       "tiny",
		TargetName: "tiny-edge:v1",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  os.Getenv("SYMPHONY_API_BASE_URL"),
				Username: os.Getenv("SYMPHONY_API_USER"),
				Password: os.Getenv("SYMPHONY_API_PASSWORD"),
			},
		},
	}
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{},
		},
	}, []model.ComponentStep{
		{
			Action: model.ComponentUpdate,
			Component: model.ComponentSpec{
				Name: "policies",
				Type: "yaml.k8s",
				Properties: map[string]interface{}{
					"yaml.url": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components)) // To make this test work, you need a target with a single component
}
func TestStagingTargetProviderApply(t *testing.T) {
	// os.Setenv("SYMPHONY_API_BASE_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("SYMPHONY_API_USER", "admin")
	// os.Setenv("SYMPHONY_API_PASSWORD", "")
	symphonyUrl := os.Getenv("SYMPHONY_API_BASE_URL")
	if symphonyUrl == "" {
		t.Skip("Skipping because SYMPHONY_API_BASE_URL enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		Name:       "tiny",
		TargetName: "tiny-edge:v1",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  os.Getenv("SYMPHONY_API_BASE_URL"),
				Username: os.Getenv("SYMPHONY_API_USER"),
				Password: os.Getenv("SYMPHONY_API_PASSWORD"),
			},
		},
	}
	component := model.ComponentSpec{
		Name: "policies",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"yaml.url": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "policies-v1",
				Namespace: "",
			},
			Spec: &model.SolutionSpec{
				DisplayName: "policies",
				Components:  []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestStagingTargetProviderRemove(t *testing.T) {
	// os.Setenv("SYMPHONY_API_BASE_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("SYMPHONY_API_USER", "admin")
	// os.Setenv("SYMPHONY_API_PASSWORD", "")
	symphonyUrl := os.Getenv("SYMPHONY_API_BASE_URL")
	if symphonyUrl == "" {
		t.Skip("Skipping because SYMPHONY_API_BASE_URL enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		Name:       "tiny",
		TargetName: "tiny-edge:v1",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  os.Getenv("SYMPHONY_API_BASE_URL"),
				Username: os.Getenv("SYMPHONY_API_USER"),
				Password: os.Getenv("SYMPHONY_API_PASSWORD"),
			},
		},
	}
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "policies",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"yaml.url": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "policies-v1",
				Namespace: "",
			},
			Spec: &model.SolutionSpec{
				DisplayName: "policies",
				Components:  []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func TestApply(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/test-v1-target/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "abc-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"components": []model.ComponentSpec{
							{
								Name: "name",
								Type: "type",
							},
						},
					},
					Version:      "v1",
					RootResource: "abc",
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

	config := StagingTargetProviderConfig{
		Name:       "default",
		TargetName: "target:v1",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)

	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "name",
		Type: "type",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{
				Version:      "v1",
				RootResource: "test",
			},
		},
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Namespace: "name-v1",
			},
			Spec: &model.SolutionSpec{
				DisplayName:  "name-v1",
				Components:   []model.ComponentSpec{component},
				Version:      "v1",
				RootResource: "name",
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/test-v1-target/v1":
			response = model.CatalogState{
				ObjectMeta: model.ObjectMeta{
					Name: "test-v1-target-v1",
				},
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"staged": map[string]interface{}{
							"components": []model.ComponentSpec{
								{
									Name: "name",
									Type: "type",
								},
							},
						},
					},
					Version:      "v1",
					RootResource: "test-target",
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

	config := StagingTargetProviderConfig{
		Name:       "default",
		TargetName: "target:v1",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "name",
		Type: "type",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{
				Version:      "v1",
				RootResource: "test",
			},
		},
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Namespace: "name-v1",
			},
			Spec: &model.SolutionSpec{
				DisplayName:  "name-v1",
				Components:   []model.ComponentSpec{component},
				Version:      "v1",
				RootResource: "name",
			},
		},
	}
	step := []model.ComponentStep{
		{
			Action:    model.ComponentUpdate,
			Component: component,
		},
	}
	_, err = provider.Get(context.Background(), deployment, step)
	assert.Nil(t, err)
}

func TestGetCatalogsFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/catalogs/registry/test-v1-target/v1":
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
	defer ts.Close()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")

	config := StagingTargetProviderConfig{
		Name:       "default",
		TargetName: "target",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	provider.Context = &contexts.ManagerContext{
		SiteInfo: v1alpha2.SiteInfo{
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "name",
		Type: "type",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-v1",
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "name-v1",
				Namespace: "",
			},
			Spec: &model.SolutionSpec{
				DisplayName: "name",
				Components:  []model.ComponentSpec{component},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &StagingTargetProvider{}
	_ = provider.Init(StagingTargetProviderConfig{})
	// assert.Nil(t, err) okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}
