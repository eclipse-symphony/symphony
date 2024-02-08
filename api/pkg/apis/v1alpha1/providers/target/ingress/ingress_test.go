/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package ingress

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes/fake"
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

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":       "name",
		"configType": "type",
		"configData": "",
		"context":    "context",
		"inCluster":  "false",
	}
	provider := IngressTargetProvider{}
	err := provider.InitWithMap(configMap)
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
			"ingressClassName": "nginx",
			"rules": []map[string]interface{}{
				{
					"http": map[string]interface{}{
						"paths": []interface{}{
							map[string]interface{}{
								"path":     "/testpath",
								"pathType": "Prefix",
								"backend": map[string]interface{}{
									"service": map[string]interface{}{
										"name": "test-service1",
										"port": map[string]interface{}{
											"number": 88,
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
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Scope: "ingresses",
			Spec: &model.InstanceSpec{
				Name: "test-ingress",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
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
			"ruless": []map[string]interface{}{
				{
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
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Scope: "ingresses",
			Spec: &model.InstanceSpec{
				Name: "test-ingress",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
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
		Instance: model.InstanceState{
			Scope: "ingresses",
			Spec: &model.InstanceSpec{
				Name: "ingress-test",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
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

func TestIngressTargetProviderApplyGet(t *testing.T) {
	testEnabled := os.Getenv("TEST_MINIKUBE_ENABLED")
	if testEnabled == "" {
		t.Skip("Skipping because TEST_MINIKUBE_ENABLED enviornment variable is not set")
	}
	config := IngressTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := IngressTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
	client := fake.NewSimpleClientset()
	provider.Client = client

	component := model.ComponentSpec{
		Name: "test-ingress",
		Type: "ingress",
		Metadata: map[string]string{
			"annotations.nginx.ingress.kubernetes.io/rewrite-target": "/",
		},
		Properties: map[string]interface{}{
			"ingressClassName": "nginx",
			"rules": []map[string]interface{}{
				{
					"http": map[string]interface{}{
						"paths": []interface{}{
							map[string]interface{}{
								"path":     "/testpath",
								"pathType": "Prefix",
								"backend": map[string]interface{}{
									"service": map[string]interface{}{
										"name": "test-service1",
										"port": map[string]interface{}{
											"number": 88,
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
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Scope: "ingresses",
			Spec: &model.InstanceSpec{
				Name: "test-ingress",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
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
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	reference := []model.ComponentStep{
		{
			Action:    "update",
			Component: component,
		},
	}
	componentSpec, err := provider.Get(context.Background(), deployment, reference)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(componentSpec))

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
