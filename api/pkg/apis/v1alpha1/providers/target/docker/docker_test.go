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

package docker

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
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

func TestConformanceSuite(t *testing.T) {
	provider := &DockerTargetProvider{}
	err := provider.Init(DockerTargetProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
