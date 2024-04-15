/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package kubectl

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	dfake "k8s.io/client-go/dynamic/fake"
	kfake "k8s.io/client-go/kubernetes/fake"
)

// TestKubectlTargetProviderConfigFromMapNil tests that passing nil to KubectlTargetProviderConfigFromMap returns a valid config
func TestKubectlTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

// TestKubectlTargetProviderConfigFromMapEmpty tests that passing an empty map to KubectlTargetProviderConfigFromMap returns a valid config
func TestKubectlTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := KubectlTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}

// TestInitWithBadConfigType tests that passing an invalid config type to Init returns an error
func TestInitWithBadConfigType(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "Bad",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyFile tests that passing an empty file to Init returns an error
func TestInitWithEmptyFile(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		ConfigType: "path",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err) //This should succeed on machines where kubectl is configured
}

// TestInitWithBadFile tests that passing a bad file to Init returns an error
func TestInitWithBadFile(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "/doesnt/exist/config.yaml",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithEmptyData tests that passing empty data to Init returns an error
func TestInitWithEmptyData(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		ConfigType: "inline",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestInitWithBadData tests that passing bad data to Init returns an error
func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":       "name",
		"configType": "type",
		"configData": "",
		"context":    "context",
		"inCluster":  "false",
	}
	provider := KubectlTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)
}

// TestInitWithBadData tests that passing bad data to Init returns an error
func TestInitWithBadData(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "inline",
		ConfigData: "bad data",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestInitWithInlineEmptyData(t *testing.T) {
	config := KubectlTargetProviderConfig{
		ConfigType: "inline",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// TestReadYamlFromUrl tests that reading yaml from a url works
func TestReadYamlFromUrl(t *testing.T) {
	testReadYaml := os.Getenv("TEST_READ_YAML")
	if testReadYaml == "" {
		t.Skip("Skipping because TEST_READ_YAML environment variable is not set")
	}
	msgChan, errChan := readYaml("https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml")
	totalSize := 0
	for {
		select {
		case data, ok := <-msgChan:
			assert.True(t, ok)
			totalSize += len(data)

		case err, ok := <-errChan:
			assert.True(t, ok)
			if err == io.EOF {
				assert.True(t, totalSize > 10000)
				return
			}

			assert.Nil(t, err)
		}
	}
}

// TestKubectlTargetProviderApply tests that applying a deployment works
func TestKubectlTargetProviderPathApplyAndDelete(t *testing.T) {
	testKubectl := os.Getenv("TEST_KUBECTL")
	if testKubectl == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			// use nginx deployment as an example to replace gatekeeper tests since insufficient cleanup for gatekeeper deployment may break other cases.
			"yaml": "https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/deployment.yaml",
			// "yaml": "https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Scope: "nginx-deployment",
				Name:  "nginx",
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
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	time.Sleep(3 * time.Second)
	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	// sleep 30s and wait for sufficient cleanup for gatekeeper
	time.Sleep(30 * time.Second)
}

func TestKubectlTargetProviderInlineApply(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path"}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx",
				},
				"spec": map[string]interface{}{
					"replicas": 3,
					"selector": map[string]interface{}{
						"matchLabels": map[string]string{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"app": "nginx",
							},
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.16.1",
									"ports": []map[string]interface{}{
										{
											"containerPort": 80,
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
			Spec: &model.InstanceSpec{
				Name:  "nginx-deployment",
				Scope: "default",
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
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

// TestKubectlTargetProviderInlineUpdate tests that updating a component works
func TestKubectlTargetProviderInlineUpdate(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}
	// wait for 5 sec before updating the deployment
	time.Sleep(5 * time.Second)

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path"}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx",
				},
				"spec": map[string]interface{}{
					"replicas": 4,
					"selector": map[string]interface{}{
						"matchLabels": map[string]string{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"app": "nginx",
							},
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.17.0",
									"ports": []map[string]interface{}{
										{
											"containerPort": 80,
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
			Spec: &model.InstanceSpec{
				Name:  "test-instance-iu",
				Scope: "test-scope-iu",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	// update
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	time.Sleep(3 * time.Second)

	// cleanup
	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

// TestKubectlTargetProviderInlineStatusProbeApply tests that applying a deployment with a status probe works
func TestKubectlTargetProviderInlineStatusProbeApply(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path"}
	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginxtest",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginxtest",
				},
				"spec": map[string]interface{}{
					"replicas": 3,
					"selector": map[string]interface{}{
						"matchLabels": map[string]string{
							"app": "nginxtest",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"app": "nginxtest",
							},
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.16.1",
									"ports": []map[string]interface{}{
										{
											"containerPort": 80,
										},
									},
								},
							},
						},
					},
				},
			},
			"statusProbe": map[string]interface{}{
				"succeededValues":  []string{"True"},
				"failedValues":     []string{"Failed"},
				"statusPath":       "$.status.conditions[0].status",
				"errorMessagePath": "$.status.conditions[0].message",
				"timeout":          "1m",
				"interval":         "2s",
				"initialWait":      "10s",
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Name:  "test-instance-spa",
				Scope: "test-scope-spa",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	// update
	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)

	time.Sleep(3 * time.Second)

	// cleanup
	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestKubectlTargetProviderClusterLevelInlineApply(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinition",
				"metadata": map[string]interface{}{
					"name": "crontabs.stable.example.com",
				},
				"spec": map[string]interface{}{
					"group": "stable.example.com",
					"scope": "Namespaced",
					"names": map[string]interface{}{
						"plural":   "crontabs",
						"singular": "crontab",
						"kind":     "CronTab",
						"shortNames": []string{
							"ct",
						},
					},
					"versions": []map[string]interface{}{
						{
							"name":    "v1",
							"served":  true,
							"storage": true,
							"schema": map[string]interface{}{
								"openAPIV3Schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"spec": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"cronSpec": map[string]interface{}{
													"type": "string",
												},
												"image": map[string]interface{}{
													"type": "string",
												},
												"replicas": map[string]interface{}{
													"type": "integer",
												},
											},
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
			Spec: &model.InstanceSpec{
				Name: "gatekeeper",
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
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

// TestKubectlTargetProviderApplyPolicy tests that applying a policy works
func TestKubectlTargetProviderApplyPolicy(t *testing.T) {
	testPolicy := os.Getenv("TEST_KUBECTL")
	if testPolicy == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "policies",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"yaml": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Name: "policies",
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
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

func TestKubectlTargetProviderDeleteInline(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "nginx-deployment",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx-deployment",
				},
				"spec": map[string]interface{}{
					"replicas": 3,
					"selector": map[string]interface{}{
						"matchLabels": map[string]string{
							"app": "nginx",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]string{
								"app": "nginx",
							},
						},
						"spec": map[string]interface{}{
							"containers": []map[string]interface{}{
								{
									"name":  "nginx",
									"image": "nginx:1.16.1",
									"ports": []map[string]interface{}{
										{
											"containerPort": 80,
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
			Spec: &model.InstanceSpec{
				Name: "nginx-deployment",
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
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

// TestKubectlTargetProviderDeletePolicies tests that deleting a policy works
func TestKubectlTargetProviderDeletePolicies(t *testing.T) {
	testPolicy := os.Getenv("TEST_KUBECTL")
	if testPolicy == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "policies",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"yaml": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Name: "policies",
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
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}

	provider := &KubectlTargetProvider{}
	err := provider.Init(KubectlTargetProviderConfig{
		ConfigType: "path",
		ConfigData: "",
	})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}

func TestKubectlTargetProviderApplyFailed(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	client := kfake.NewSimpleClientset()
	provider.Client = client
	dynamicClient := dfake.NewSimpleDynamicClient(runtime.NewScheme())
	provider.DynamicClient = dynamicClient

	component := model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx-deployment",
				},
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Scope: "nginx-system",
				Name:  "nginx",
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
				Action:    model.ComponentUpdate,
				Component: component,
			},
		},
	}

	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.NotNil(t, err)

	step = model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action:    model.ComponentDelete,
				Component: component,
			},
		},
	}
	_, err = provider.Apply(context.Background(), deployment, step, false)
	assert.Nil(t, err)
}

func TestKubectlTargetProviderGet(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_KUBECTL")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_KUBECTL environment variable is not set")
	}
	config := KubectlTargetProviderConfig{
		InCluster:  false,
		ConfigType: "path",
		ConfigData: "",
	}

	provider := KubectlTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	client := kfake.NewSimpleClientset()
	provider.Client = client
	dynamicClient := dfake.NewSimpleDynamicClient(runtime.NewScheme())
	provider.DynamicClient = dynamicClient

	component := model.ComponentSpec{
		Name: "policies",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"yaml": "https://raw.githubusercontent.com/eclipse-symphony/symphony/main/docs/samples/k8s/gatekeeper/policy.yaml",
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Name: "policies",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	reference := []model.ComponentStep{
		{
			Action:    model.ComponentUpdate,
			Component: component,
		},
	}
	_, err = provider.Get(context.Background(), deployment, reference)
	assert.NotNil(t, err)

	component = model.ComponentSpec{
		Name: "nginx",
		Type: "yaml.k8s",
		Properties: map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "nginx-deployment",
				},
			},
		},
	}
	deployment = model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				Name:  "nginx",
				Scope: "nginx-system",
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{component},
			},
		},
	}
	reference = []model.ComponentStep{
		{
			Action:    model.ComponentUpdate,
			Component: component,
		},
	}
	_, err = provider.Get(context.Background(), deployment, reference)
	assert.Nil(t, err)
}
