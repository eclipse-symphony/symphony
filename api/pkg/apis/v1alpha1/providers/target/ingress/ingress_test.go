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

package ingress

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
)

// TestIngressTargetProviderConfigFromMapNil tests that passing nil to IngressTargetProviderConfigFromMap returns a valid config
func TestIngressTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := IngressTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

// TestIngressTargetProviderConfigFromMapEmpty tests that passing an empty map to IngressTargetProviderConfigFromMap returns a valid config
func TestIngressTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := IngressTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}

// TestInitWithBadConfigType tests that passing an invalid config type to Init returns an error
func TestInitWithBadConfigType(t *testing.T) {
	config := IngressTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyFile tests that passing an empty file to Init returns an error
func TestInitWithEmptyFile(t *testing.T) {
	getConfigMap := os.Getenv("TEST_INGRESS")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_INGRESS environment variable is not set")
	}
	config := IngressTargetProviderConfig{
		ConfigType: "path",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}

// TestInitWithBadFile tests that passing a bad file to Init returns an error
func TestInitWithBadFile(t *testing.T) {
	config := IngressTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyData tests that passing empty data to Init returns an error
func TestInitWithEmptyData(t *testing.T) {
	getConfigMap := os.Getenv("TEST_INGRESS")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_INGRESS environment variable is not set")
	}
	config := IngressTargetProviderConfig{
		ConfigType: "inline",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithBadData tests that passing bad data to Init returns an error
func TestInitWithBadData(t *testing.T) {
	config := IngressTargetProviderConfig{
		ConfigType: "inline",
		ConfigData: "bad data",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestIngressTargetProviderApply tests that applying a configmap works
func TestIngressTargetProviderApply(t *testing.T) {
	getConfigMap := os.Getenv("TEST_INGRESS")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_INGRESS environment variable is not set")
	}

	config := IngressTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-ingress",
		Type: "ingress",
		Metadata: map[string]string{
			"annotations.nginx.ingress.kubernetes.io/rewrite-target": "/",
		},
		Properties: map[string]interface{}{
			"rules.0": map[string]interface{}{
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"path":     "/testpath",
							"pathType": "Prefix",
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "test-service1",
									"port": map[string]interface{}{
										"number": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "test-ingress",
			Scope: "ingresses",
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

// TestIngressTargetProviderDelete tests that deleting a ingress works
func TestIngressTargetProviderDelete(t *testing.T) {
	getConfigMap := os.Getenv("TEST_INGRESS")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_INGRESS environment variable is not set")
	}

	config := IngressTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-ingress",
		Type: "ingress",
		Metadata: map[string]string{
			"annotations.nginx.ingress.kubernetes.io/rewrite-target": "/",
		},
		Properties: map[string]interface{}{
			"rules.0": map[string]interface{}{
				"http": map[string]interface{}{
					"paths": []interface{}{
						map[string]interface{}{
							"path":     "/testpath",
							"pathType": "Prefix",
							"backend": map[string]interface{}{
								"service": map[string]interface{}{
									"name": "test-service1",
									"port": map[string]interface{}{
										"number": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "test-ingress",
			Scope: "ingresses",
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

// TestIngressTargetProviderGet tests that getting a configmap works
func TestIngressTargetProviderGet(t *testing.T) {
	getConfigMap := os.Getenv("TEST_INGRESS")
	if getConfigMap == "" {
		t.Skip("Skipping because TEST_INGRESS environment variable is not set")
	}

	config := IngressTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "test-ingress",
		Type: "ingresses",
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name:  "ingress-test",
			Scope: "ingresses",
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
	assert.Equal(t, "/testpath", components[0].Properties["rules.0"].(networkingv1.IngressRule).HTTP.Paths[0].Path)
}
