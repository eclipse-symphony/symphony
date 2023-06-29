package http

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHttpTargetProviderConfigFromMapNil tests that HttpTargetProviderConfigFromMap returns nil when passed nil
func TestHttpTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := HttpTargetProviderConfigFromMap(nil)
	require.Nil(t, err)
}

// TestHttpTargetProviderConfigFromMapEmpty tests that HttpTargetProviderConfigFromMap returns nil when passed an empty map
func TestHttpTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := HttpTargetProviderConfigFromMap(map[string]string{})
	require.Nil(t, err)
}

// TestHttpTargetProviderConfigFromMap tests that HttpTargetProviderConfigFromMap returns nil when passed a valid map
func TestHttpTargetProviderConfigFromMap(t *testing.T) {
	_, err := HttpTargetProviderConfigFromMap(map[string]string{
		"name": "test",
	})
	require.Nil(t, err)
}

// TestHttpTargetProviderInitWithMap tests that HttpTargetProvider.InitWithMap returns nil when passed a non empty map
func TestHttpTargetProviderInitWithMap(t *testing.T) {
	provider := HttpTargetProvider{}
	err := provider.InitWithMap(map[string]string{
		"name": "test",
	})
	require.Nil(t, err)
}

// TestHttpTargetproviderApply tests that HttpTargetProvider.Apply returns nil when passed a valid deployment spec
func TestHttpTargetProviderApply(t *testing.T) {
	config := HttpTargetProviderConfig{
		Name: "qa-target",
	}
	provider := HttpTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "http-component",
					Properties: map[string]interface{}{
						"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
						"http.method": "GET",
					},
				},
			},
		},
		Assignments: map[string]string{
			"target-1": "{http-component}",
		},
		Targets: map[string]model.TargetSpec{
			"target-1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "doesn't-matter",
								Config:   map[string]string{},
							},
						},
					},
				},
			},
		},
	}, false)
	assert.Nil(t, err)
}

// TestHttpTargetProviderIncorrectApply tests that HttpTargetProvider.Apply returns an error when passed an invalid deployment spec
func TestHttpTargetProviderIncorrectApply(t *testing.T) {
	config := HttpTargetProviderConfig{
		Name: "qa-target",
	}
	provider := HttpTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "http-component",
					Properties: map[string]interface{}{
						"http.url":    "",
						"http.method": "GET",
					},
				},
			},
		},
	}, false)
	assert.NotNil(t, err)
}

// TestHttpTargetProviderGet tests that HttpTargetProvider.Get returns nil when passed a valid deployment spec
func TestHttpTargetProviderGet(t *testing.T) {
	config := HttpTargetProviderConfig{
		Name: "qa-target",
	}
	provider := HttpTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "http-component",
					Properties: map[string]interface{}{
						"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
						"http.method": "GET",
					},
				},
			},
		},
		Assignments: map[string]string{
			"target-1": "{http-component}",
		},
		Targets: map[string]model.TargetSpec{
			"target-1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "doesn't-matter",
								Config:   map[string]string{},
							},
						},
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}

// TestHttpTargetProviderRemove tests that HttpTargetProvider.Remove returns nil when passed a valid deployment spec
func TestHttpTargetProviderRemove(t *testing.T) {
	config := HttpTargetProviderConfig{
		Name: "qa-target",
	}
	provider := HttpTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "http-component",
					Properties: map[string]interface{}{
						"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
						"http.method": "GET",
					},
				},
			},
		},
		Assignments: map[string]string{
			"target-1": "{http-component}",
		},
		Targets: map[string]model.TargetSpec{
			"target-1": {
				Topologies: []model.TopologySpec{
					{
						Bindings: []model.BindingSpec{
							{
								Role:     "instance",
								Provider: "doesn't-matter",
								Config:   map[string]string{},
							},
						},
					},
				},
			},
		},
	}, []model.ComponentSpec{})
	assert.Nil(t, err)
}

// TestHttpTargetProviderNeedsUpdate tests that HttpTargetProvider.NeedsUpdate returns false when passed a valid deployment spec
func TestHttpTargetProviderNeedsUpdate(t *testing.T) {
	current := []model.ComponentSpec{
		{
			Name: "http-component",
			Properties: map[string]interface{}{
				"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
				"http.method": "GET",
			},
		},
	}

	desired := []model.ComponentSpec{
		{
			Name: "http-component",
			Properties: map[string]interface{}{
				"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
				"http.method": "GET",
			},
		},
	}
	provider := HttpTargetProvider{}
	update := provider.NeedsUpdate(context.Background(), current, desired)
	require.False(t, update)
}

// TestHttpTargetProviderNeedsRemove tests that HttpTargetProvider.NeedsRemove returns false when passed a valid deployment spec
func TestHttpTargetProviderNeedsRemove(t *testing.T) {
	current := []model.ComponentSpec{
		{
			Name: "http-component",
			Properties: map[string]interface{}{
				"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
				"http.method": "GET",
			},
		},
	}

	desired := []model.ComponentSpec{
		{
			Name: "http-component",
			Properties: map[string]interface{}{
				"http.url":    "https://learn.microsoft.com/en-us/content-nav/azure.json?",
				"http.method": "GET",
			},
		},
	}
	provider := HttpTargetProvider{}
	remove := provider.NeedsRemove(context.Background(), current, desired)
	require.True(t, remove)
}

// TestReadProperty tests that ReadProperty returns the correct value
func TestReadProperty(t *testing.T) {
	url := "https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<redacted>"
	val := model.ReadProperty(map[string]string{
		"http.url": url,
	}, "http.url", &model.ValueInjections{
		InstanceId: "A",
		SolutionId: "B",
		TargetId:   "C",
	})
	assert.Equal(t, url, val)
}

// TestConformanceSuite tests that the HttpTargetProvider conforms to the TargetProvider interface
func TestConformanceSuite(t *testing.T) {
	provider := &HttpTargetProvider{}
	err := provider.Init(HttpTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
