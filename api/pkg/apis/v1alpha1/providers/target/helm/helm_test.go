/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package helm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

// TestHelmTargetProviderConfigFromMapNil tests the HelmTargetProviderConfigFromMap function with nil input
func TestHelmTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := HelmTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}

// TestHelmTargetProviderConfigFromMapEmpty tests the HelmTargetProviderConfigFromMap function with empty input
func TestHelmTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := HelmTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}

// TestHelmTargetProviderInitEmptyConfig tests the Init function of HelmTargetProvider with empty config
func TestHelmTargetProviderInitEmptyConfig(t *testing.T) {
	config := HelmTargetProviderConfig{}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

func TestInitWithMap(t *testing.T) {
	configMap := map[string]string{
		"name":       "test",
		"configType": "bytes",
		"inCluster":  "false",
		"configData": "data",
		"context":    "context",
	}
	provider := HelmTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":       "test",
		"configType": "bytes",
		"inCluster":  "false",
		"context":    "context",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":       "test",
		"configType": "wrongtype",
		"inCluster":  "false",
		"context":    "context",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)
}

// TestHelmTargetProviderGetHelmProperty tests the getHelmValuesFromComponent function with valid input
func TestHelmTargetProviderGetHelmPropertyMissingRepo(t *testing.T) {
	_, err := getHelmPropertyFromComponent(model.ComponentSpec{
		Name: "bluefin-arc-extensions",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "", // blank repo
				"name":    "bluefin-arc-extension",
				"version": "0.1.1",
			},
			"values": map[string]interface{}{
				"CUSTOM_VISION_KEY": "BBB",
				"CLUSTER_SECRET":    "test",
				"CERTIFICATES":      []string{"a", "b"},
			},
		},
	})
	assert.NotNil(t, err)
}

func TestHelmTargetProviderGetHelmProperty(t *testing.T) {
	_, err := getHelmPropertyFromComponent(model.ComponentSpec{
		Name: "bluefin-arc-extensions",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
				"name":    "bluefin-arc-extension",
				"version": "0.1.1",
			},
			"values": map[string]interface{}{
				"CUSTOM_VISION_KEY": "BBB",
				"CLUSTER_SECRET":    "test",
				"CERTIFICATES":      []string{"a", "b"},
			},
		},
	})
	assert.Nil(t, err)
}

// TestHelmTargetProviderInstall tests the Apply function of HelmTargetProvider
func TestHelmTargetProviderInstall(t *testing.T) {
	// To run this test case successfully, you shouldn't have a symphony Helm chart already deployed to your current Kubernetes context
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSION")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "bluefin-arc-extensions",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
				"name":    "bluefin-arc-extension",
				"version": "0.1.1",
			},
			"values": map[string]interface{}{
				"CUSTOM_VISION_KEY": "BBB",
				"CLUSTER_SECRET":    "test",
				"CERTIFICATES":      []string{"a", "b"},
			},
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

// TestHelmTargetProviderGet tests the Get function of HelmTargetProvider
func TestHelmTargetProviderGet(t *testing.T) {
	// To run this test case successfully, you need to have a symphony Helm chart deployed to your current Kubernetes context
	testHelmChart := os.Getenv("TEST_HELM_CHART")
	if testHelmChart == "" {
		t.Skip("Skipping because TEST_HELM_CHART environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "bluefin-arc-extensions",
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "bluefin-arc-extensions",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}

// TestHelmTargetProviderInstallNoOci tests the Apply function of HelmTargetProvider with no OCI registry
func TestHelmTargetProviderInstallNoOci(t *testing.T) {
	// To run this test case successfully, you shouldn't have a symphony Helm chart already deployed to your current Kubernetes context
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSIONS")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "akri",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "https://project-akri.github.io/akri/akri",
				"name":    "akri",
				"version": "",
			},
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

func TestHelmTargetProviderInstallNginxIngress(t *testing.T) {
	// To run this test case successfully, you shouldn't have a symphony Helm chart already deployed to your current Kubernetes context
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSIONS")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "ingress-nginx",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "https://github.com/kubernetes/ingress-nginx/releases/download/helm-chart-4.7.1/ingress-nginx-4.7.1.tgz",
				"name":    "ingress-nginx",
				"version": "4.7.1",
			},
			"values": map[string]interface{}{
				"controller": map[string]interface{}{
					"replicaCount": 1,
					"nodeSelector": map[string]interface{}{
						"kubernetes.io/os": "linux",
					},
					"admissionWebhooks": map[string]interface{}{
						"patch": map[string]interface{}{
							"nodeSelector": map[string]interface{}{
								"kubernetes.io/os": "linux",
							},
						},
					},
					"service": map[string]interface{}{
						"annotations": map[string]interface{}{
							"service.beta.kubernetes.io/azure-load-balancer-health-probe-request-path": "/healthz",
						},
					},
				},
				"defaultBackend": map[string]interface{}{
					"nodeSelector": map[string]interface{}{
						"kubernetes.io/os": "linux",
					},
				},
			},
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

// TestHelmTargetProviderInstallDirectDownload tests the Apply function of HelmTargetProvider with direct download
func TestHelmTargetProviderInstallDirectDownload(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_HELM_GATEKEEPER")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_HELM_GATEKEEPER environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "gatekeeper",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo": "https://open-policy-agent.github.io/gatekeeper/charts/gatekeeper-3.10.0-beta.1.tgz",
				"name": "gatekeeper",
			},
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

// TestHelmTargetProviderRemove tests the Remove function of HelmTargetProvider
func TestHelmTargetProviderRemove(t *testing.T) {
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSION")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION environment variable is not set")
	}

	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "bluefin-arc-extensions",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
				"name":    "bluefin-arc-extension",
				"version": "0.1.1",
			},
		},
	}
	deployment := model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "symphony",
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

// TestHelmTargetProviderGetAnotherCluster tests the Get function of HelmTargetProvider with another cluster
func TestHelmTargetProviderGetAnotherCluster(t *testing.T) {
	//to run this test successfully, you need to fix the cluster config below, and the target cluster shouldn't have symphony Helm chart installed
	//THIS CASE IS BROKERN
	testotherK8s := os.Getenv("TEST_HELM_OTHER_K8S")
	if testotherK8s == "" {
		t.Skip("Skipping because TEST_HELM_OTHER_K8S environment variable is not set")
	}
	config := HelmTargetProviderConfig{
		InCluster:  false,
		ConfigType: "bytes",
		ConfigData: `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ...
    server: https://k12s-dns-6b5afdc5.hcp.westus3.azmk8s.io:443
  name: k12s
contexts:
- context:
    cluster: k12s
    user: clusterUser_symphony_k12s
  name: k12s
current-context: k12s
kind: Config
preferences: {}
users:
- name: clusterUser_symphony_k12s
  user:
    client-certificate-data: ...
    client-key-data: ...
    token: ...`,
	}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}

func TestHelmTargetProviderUpdateDelete(t *testing.T) {
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	component := model.ComponentSpec{
		Name: "bluefin-arc-extensions",
		Type: "helm.v3",
		Properties: map[string]interface{}{
			"chart": map[string]string{
				"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
				"name":    "bluefin-arc-extension",
				"version": "0.1.1",
			},
			"values": map[string]interface{}{
				"CUSTOM_VISION_KEY": "BBB",
				"CLUSTER_SECRET":    "test",
				"CERTIFICATES":      []string{"a", "b"},
			},
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

func TestHelmTargetProviderGetFail(t *testing.T) {
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "bluefin-arc-extensions",
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "bluefin-arc-extensions",
			},
		},
	})
	assert.NotNil(t, err)
}

func TestDownloadFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	err := downloadFile(ts.URL, "test")
	assert.Nil(t, err)
}

func TestGetActionConfig(t *testing.T) {
	config := &rest.Config{
		Host:        "host",
		BearerToken: "token",
	}
	_, err := getActionConfig(context.Background(), "default", config)
	assert.Nil(t, err)
}

// TestConformanceSuite tests the HelmTargetProvider for conformance
func TestConformanceSuite(t *testing.T) {
	provider := &HelmTargetProvider{}
	err := provider.Init(HelmTargetProviderConfig{InCluster: true})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
