package docker

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestDockerTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := DockerTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestDockerTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := DockerTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestDockerTargetProviderInitEmptyConfig(t *testing.T) {
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}
func TestDockerTargetProviderInstall(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_PROVIDER")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "redis-test",
					Type: "container",
					Properties: map[string]string{
						"container.image": "redis:latest",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}

func TestDockerTargetProviderGet(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_PROVIDER")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "redis-test",
					Type: "container",
					Properties: map[string]string{
						"container.image": "redis:latest",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
	assert.Equal(t, "redis:latest", components[0].Properties["container.image"])
	assert.Equal(t, "", components[0].Properties["container.ports"])
}

func TestDockerTargetProviderRemove(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_PROVIDER")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "redis-test",
					Type: "container",
					Properties: map[string]string{
						"container.image": "redis:latest",
					},
				},
			},
		},
	})
	assert.Nil(t, component)
}
