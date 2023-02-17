package staging

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestStagingTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(nil)
	assert.NotNil(t, err)
}
func TestStagingTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(map[string]string{})
	assert.NotNil(t, err)
}
func TestInitWithBadConfigType(t *testing.T) {
	config := StagingTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyFile(t *testing.T) {
	config := StagingTargetProviderConfig{
		ConfigType: "path",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}
func TestInitWithBadFile(t *testing.T) {
	config := StagingTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithEmptyData(t *testing.T) {
	config := StagingTargetProviderConfig{
		ConfigType: "bytes",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestInitWithBadData(t *testing.T) {
	config := StagingTargetProviderConfig{
		ConfigType: "bytes",
		ConfigData: "bad data",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestStagingTargetProviderGet(t *testing.T) {
	testStaging := os.Getenv("TEST_STAGING")
	if testStaging == "" {
		t.Skip("Skipping because TEST_STAGING enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		TargetName: "target-3f3a2c67-227f-4d2b-92cf-55c7abfa47de",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components)) // To make this test work, you need a target with a single component
}
func TestKubectlTargetProviderApply(t *testing.T) {
	testRedis := os.Getenv("TEST_STAGING")
	if testRedis == "" {
		t.Skip("Skipping because TEST_STAGING enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		TargetName: "target-3f3a2c67-227f-4d2b-92cf-55c7abfa47de",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			DisplayName: "policies",
			Scope:       "",
			Components: []model.ComponentSpec{
				{
					Name: "policies",
					Type: "yaml.k8s",
					Properties: map[string]string{
						"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}

func TestKubectlTargetProviderRemove(t *testing.T) {
	testRedis := os.Getenv("TEST_STAGING")
	if testRedis == "" {
		t.Skip("Skipping because TEST_STAGING enviornment variable is not set")
	}
	config := StagingTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		TargetName: "target-3f3a2c67-227f-4d2b-92cf-55c7abfa47de",
	}
	provider := StagingTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			DisplayName: "policies",
			Scope:       "",
			Components: []model.ComponentSpec{
				{
					Name: "policies",
					Type: "yaml.k8s",
					Properties: map[string]string{
						"yaml.url": "https://demopolicies.blob.core.windows.net/gatekeeper/policy.yaml",
					},
				},
			},
		},
	}, nil)
	assert.Nil(t, err)
}
