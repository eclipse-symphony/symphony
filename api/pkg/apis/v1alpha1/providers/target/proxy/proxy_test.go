/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	testProxy := os.Getenv("TEST_PROXY")
	if testProxy != "yes" {
		t.Skip("Skipping becasue TEST_PROXY is missing or not set to 'yes'")
	}
	provider := ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{
		Name:      "proxy",
		ServerURL: "http://localhost:8090/v1alpha2/solution/",
	})
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "HomeHub_1.0.4.0_x64",
					Properties: map[string]interface{}{
						"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
					},
				},
			},
		},
		Assignments: map[string]string{
			"target1": "{HomeHub_1.0.4.0_x64}",
		},
		Targets: map[string]model.TargetSpec{
			"target1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "providers.target.win10.sideload",
								Config: map[string]string{
									"name":                "win10sideload",
									"ipAddress":           "192.168.50.55",
									"winAppDeployCmdPath": "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
								},
							},
						},
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "HomeHub_1.0.4.0_x64",
				Properties: map[string]interface{}{
					"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
				},
			},
		},
	})
	assert.Equal(t, 1, len(components))
	assert.Nil(t, err)
}

func TestRemove(t *testing.T) {
	testProxy := os.Getenv("TEST_PROXY")
	if testProxy != "yes" {
		t.Skip("Skipping becasue TEST_PROXY is missing or not set to 'yes'")
	}
	provider := ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{
		Name:      "proxy",
		ServerURL: "http://localhost:8090/v1alpha2/solution/",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
		},
	}
	deployment := model.DeploymentSpec{
		Assignments: map[string]string{
			"target1": "{HomeHub_1.0.4.0_x64}",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
		Targets: map[string]model.TargetSpec{
			"target1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "providers.target.win10.sideload",
								Config: map[string]string{
									"name":                "win10sideload",
									"ipAddress":           "192.168.50.55",
									"winAppDeployCmdPath": "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
								},
							},
						},
					},
				},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "delete",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}
func TestApply(t *testing.T) {
	testProxy := os.Getenv("TEST_PROXY")
	if testProxy != "yes" {
		t.Skip("Skipping becasue TEST_PROXY is missing or not set to 'yes'")
	}
	provider := ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{
		Name:      "proxy",
		ServerURL: "http://localhost:8090/v1alpha2/solution/",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "HomeHub_1.0.4.0_x64",
		Properties: map[string]interface{}{
			"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
		},
	}
	deployment := model.DeploymentSpec{
		Assignments: map[string]string{
			"target1": "{HomeHub_1.0.4.0_x64}",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
		Targets: map[string]model.TargetSpec{
			"target1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "providers.target.win10.sideload",
								Config: map[string]string{
									"name":                "win10sideload",
									"ipAddress":           "192.168.50.55",
									"winAppDeployCmdPath": "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22621.0\\x64\\WinAppDeployCmd.exe",
								},
							},
						},
					},
				},
			},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "delete",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":      "name",
		"serverUrl": "127.0.0.0",
	}
	provider := ProxyUpdateProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestCallRestAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	provider := &ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{ServerURL: ts.URL + "/"})
	assert.Nil(t, err)

	_, err = provider.callRestAPI("", "GET", []byte{})
	assert.Nil(t, err)
}

func TestProxyUpdateProviderApplyGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			componentSpecs := []model.ComponentSpec{
				{
					Name: "name",
				},
			}
			jsonData, _ := json.Marshal(componentSpecs)
			w.Write(jsonData)
		} else {
			w.Write([]byte("OK"))

		}
	}))
	defer ts.Close()

	provider := ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{
		Name:      "proxy",
		ServerURL: ts.URL + "/",
	})
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test",
	}
	deployment := model.DeploymentSpec{
		Assignments: map[string]string{
			"target1": "test",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	components := []model.ComponentStep{
		{
			Action:    "update",
			Component: component,
		},
	}
	step := model.DeploymentStep{
		Components: components,
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	_, err = provider.Get(context.Background(), deployment, components)
	assert.Nil(t, err)

	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "delete",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
