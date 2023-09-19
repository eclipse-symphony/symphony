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

package staging

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
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
		TargetName: "tiny-edge",
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
		Instance: model.InstanceSpec{
			Name: "test",
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "policies",
				Type: "yaml.k8s",
				Properties: map[string]interface{}{
					"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components)) // To make this test work, you need a target with a single component
}
func TestKubectlTargetProviderApply(t *testing.T) {
	// os.Setenv("SYMPHONY_API_BASE_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("SYMPHONY_API_USER", "admin")
	// os.Setenv("SYMPHONY_API_PASSWORD", "")
	symphonyUrl := os.Getenv("SYMPHONY_API_BASE_URL")
	if symphonyUrl == "" {
		t.Skip("Skipping because SYMPHONY_API_BASE_URL enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		Name:       "tiny",
		TargetName: "tiny-edge",
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
			"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "test",
		},
		Solution: model.SolutionSpec{
			DisplayName: "policies",
			Scope:       "",
			Components:  []model.ComponentSpec{component},
		},
	}
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "update",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestKubectlTargetProviderRemove(t *testing.T) {
	// os.Setenv("SYMPHONY_API_BASE_URL", "http://localhost:8080/v1alpha2/")
	// os.Setenv("SYMPHONY_API_USER", "admin")
	// os.Setenv("SYMPHONY_API_PASSWORD", "")
	symphonyUrl := os.Getenv("SYMPHONY_API_BASE_URL")
	if symphonyUrl == "" {
		t.Skip("Skipping because SYMPHONY_API_BASE_URL enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		Name:       "tiny",
		TargetName: "tiny-edge",
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
			"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "test",
		},
		Solution: model.SolutionSpec{
			DisplayName: "policies",
			Scope:       "",
			Components:  []model.ComponentSpec{component},
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
	provider := &StagingTargetProvider{}
	_ = provider.Init(StagingTargetProviderConfig{})
	// assert.Nil(t, err) okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}
