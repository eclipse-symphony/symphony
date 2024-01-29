/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package docker

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
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
func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name": "name",
	}
	provider := DockerTargetProvider{}
	err := provider.InitWithMap(configMap)
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
	component := model.ComponentSpec{
		Name: "redis-test",
		Type: "container",
		Properties: map[string]interface{}{
			model.ContainerImage: "redis:latest",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
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

func TestDockerTargetProviderGet(t *testing.T) {
	// NOTE: To run this test case successfully, you need to have Docker and Redis container running:
	// docker run -d --name redis-test -p 6379:6379 redis:latest
	// Then, comment out the next 4 lines of code and run the test case.
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
					Properties: map[string]interface{}{
						model.ContainerImage: "redis:latest",
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "redis-test",
				Type: "container",
				Properties: map[string]interface{}{
					model.ContainerImage: "redis:latest",
					"env.REDIS_VERSION":  "7.0.12", // NOTE: Only environment variables passed in by the reference are returned.
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
	assert.Equal(t, "redis:latest", components[0].Properties[model.ContainerImage])
	assert.NotEqual(t, "", components[0].Properties["env.REDIS_VERSION"])
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
	component := model.ComponentSpec{
		Name: "redis-test",
		Type: "container",
		Properties: map[string]interface{}{
			model.ContainerImage: "redis:latest",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
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

func TestUpdateGetDelete(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_ENABLED")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	// Update
	component := model.ComponentSpec{
		Name: "alpine-test",
		Type: "container",
		Properties: map[string]interface{}{
			model.ContainerImage: "alpine:3.18",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
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

	// Get
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "alpine-test",
					Type: "container",
					Properties: map[string]interface{}{
						model.ContainerImage: "alpine:3.18",
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "alpine-test",
				Type: "container",
				Properties: map[string]interface{}{
					model.ContainerImage: "alpine:3.18",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))

	// Delete
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

func TestApplyFailed(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_ENABLED")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	// invalid container image name
	component := model.ComponentSpec{
		Name: "",
		Type: "container",
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
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
	assert.NotNil(t, err)

	// unknown container image
	component = model.ComponentSpec{
		Name: "abcd:latest",
		Type: "container",
		Properties: map[string]interface{}{
			model.ContainerImage: "abc:latest",
		},
	}
	deployment = model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
		},
	}
	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    "update",
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

func TestApplyAlreadyRunning(t *testing.T) {
	testDockerProvider := os.Getenv("TEST_DOCKER_ENABLED")
	if testDockerProvider == "" {
		t.Skip("Skipping because TEST_DOCKER_PROVIDER enviornment variable is not set")
	}
	config := DockerTargetProviderConfig{}
	provider := DockerTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	component := model.ComponentSpec{
		Name: "alpine-test",
		Type: "container",
		Properties: map[string]interface{}{
			model.ContainerImage: "alpine:3.18",
		},
	}
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{component},
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

	// already running
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestConformanceSuite(t *testing.T) {
	provider := &DockerTargetProvider{}
	err := provider.Init(DockerTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
