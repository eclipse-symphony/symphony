package proxy

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
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
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "HomeHub_1.0.4.0_x64",
							Properties: map[string]string{
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
			},
		},
	})
	assert.Equal(t, 1, len(components))
	assert.Nil(t, err)
}
func TestNeedsUpdate(t *testing.T) {
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

	deployment := model.DeploymentSpec{
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "HomeHub_1.0.4.0_x64",
							Properties: map[string]string{
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
			},
		},
	}
	current, err := provider.Get(context.Background(), deployment)
	assert.Nil(t, err)

	desired := []model.ComponentSpec{
		{
			Name: "HomeHub_1.0.4.0_x64",
			Properties: map[string]string{
				"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
			},
		},
	}
	assert.False(t, provider.NeedsUpdate(context.Background(), desired, current))
}
func TestNeedsRemove(t *testing.T) {
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

	deployment := model.DeploymentSpec{
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "HomeHub_1.0.4.0_x64",
							Properties: map[string]string{
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
			},
		},
	}
	current, err := provider.Get(context.Background(), deployment)
	assert.Nil(t, err)
	desired := []model.ComponentSpec{
		{
			Name: "HomeHub_1.0.4.0_x64",
			Properties: map[string]string{
				"app.package.path": "E:\\projects\\go\\github.com\\azure\\symphony-docs\\samples\\scenarios\\homehub\\HomeHub\\HomeHub.Package\\AppPackages\\HomeHub.Package_1.0.4.0_Debug_Test\\HomeHub.Package_1.0.4.0_x64_Debug.appxbundle",
			},
		},
	}
	assert.True(t, provider.NeedsRemove(context.Background(), desired, current))
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
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "HomeHub_1.0.4.0_x64",
							Properties: map[string]string{
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
			},
		},
	}, []model.ComponentSpec{
		{
			Name: "HomeHub_1.0.4.0_x64__gjd5ncee18d88",
			Type: "win.uwp",
		},
	})
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
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Stages: []model.DeploymentStage{
			{
				Solution: model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "HomeHub_1.0.4.0_x64",
							Properties: map[string]string{
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
			},
		},
	})
	assert.Nil(t, err)
}
