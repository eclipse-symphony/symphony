/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package proxy

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
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

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &ProxyUpdateProvider{}
	err := provider.Init(ProxyUpdateProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
