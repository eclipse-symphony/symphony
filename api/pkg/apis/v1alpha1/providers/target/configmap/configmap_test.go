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

package configmap

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

// TestConfiMapTargetProviderConfigFromMapNil tests that passing nil to ConfigMapTargetProviderConfigFromMap returns a valid config
func TestConfiMapTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := ConfigMapTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

// TestConfigMapTargetProviderConfigFromMapEmpty tests that passing an empty map to ConfigMapTargetProviderConfigFromMap returns a valid config
func TestConfigMapTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := ConfigMapTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}

// TestInitWithBadConfigType tests that passing an invalid config type to Init returns an error
func TestInitWithBadConfigType(t *testing.T) {
	config := ConfigMapTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyFile tests that passing an empty file to Init returns an error
func TestInitWithEmptyFile(t *testing.T) {
	getConfigMap := os.Getenv("TEST_CONFIGMAP")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_CONFIGMAP environment variable is not set")
	}
	config := ConfigMapTargetProviderConfig{
		ConfigType: "path",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}

// TestInitWithBadFile tests that passing a bad file to Init returns an error
func TestInitWithBadFile(t *testing.T) {
	config := ConfigMapTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyData tests that passing empty data to Init returns an error
func TestInitWithEmptyData(t *testing.T) {
	getConfigMap := os.Getenv("TEST_CONFIGMAP")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_CONFIGMAP environment variable is not set")
	}

	config := ConfigMapTargetProviderConfig{
		ConfigType: "inline",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithBadData tests that passing bad data to Init returns an error
func TestInitWithBadData(t *testing.T) {
	config := ConfigMapTargetProviderConfig{
		ConfigType: "inline",
		ConfigData: "bad data",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestConfigMapTargetProviderApply tests that applying a configmap works
func TestConfigMapTargetProviderApply(t *testing.T) {
	getConfigMap := os.Getenv("TEST_CONFIGMAP")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_CONFIGMAP environment variable is not set")
	}

	config := ConfigMapTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-config",
		Type: "config",
		Properties: map[string]interface{}{
			"foo": "bar",
			"complex": map[string]interface{}{
				"easy": "as",
				"123":  456,
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "config-test",
			Scope: "configs",
		},
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

// TestConfigMapTargetProviderDekete tests that deleting a configmap works
func TestConfigMapTargetProviderDekete(t *testing.T) {
	getConfigMap := os.Getenv("TEST_CONFIGMAP")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_CONFIGMAP environment variable is not set")
	}

	config := ConfigMapTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-config",
		Type: "config",
		Properties: map[string]interface{}{
			"foo": "bar",
			"complex": map[string]interface{}{
				"easy": "as",
				"123":  456,
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "config-test",
			Scope: "configs",
		},
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

// TestConfigMapTargetProviderGet tests that getting a configmap works
func TestConfigMapTargetProviderGet(t *testing.T) {
	getConfigMap := os.Getenv("TEST_CONFIGMAP")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_CONFIGMAP environment variable is not set")
	}

	config := ConfigMapTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := ConfigMapTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-config",
		Type: "config",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "config-test",
			Scope: "configs",
		},
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
	components, err := provider.Get(context.Background(), deployment, step.Components)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
	assert.Equal(t, "bar", components[0].Properties["foo"])
	assert.Equal(t, "as", components[0].Properties["complex"].(map[string]interface{})["easy"])
	// TODO: This could be problematic as integers are probably preferred
	assert.Equal(t, 456.0, components[0].Properties["complex"].(map[string]interface{})["123"])
}
