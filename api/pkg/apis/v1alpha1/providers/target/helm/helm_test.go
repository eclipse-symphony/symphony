package helm

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestHelmTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := HelmTargetProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestHelmTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := HelmTargetProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestHelmTargetProviderInitEmptyConfig(t *testing.T) {
	config := HelmTargetProviderConfig{}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}
func TestHelmTargetProviderGet(t *testing.T) {
	// To run this test case successfully, you need to have a symphony Helm chart deployed to your current Kubernetes context
	testHelmChart := os.Getenv("TEST_HELM_CHART")
	if testHelmChart == "" {
		t.Skip("Skipping because TEST_HELM_CHART enviornment variable is not set")
	}
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: testHelmChart,
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}
func TestHelmTargetProviderInstall(t *testing.T) {
	// To run this test case successfully, you shouldn't have a symphony Helm chart already deployed to your current Kubernetes context
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSION")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION enviornment variable is not set")
	}
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "symphony-com",
					Type: "helm.v3",
					Properties: map[string]string{
						"helm.chart.repo":               "possprod.azurecr.io/helm/symphony",
						"helm.chart.name":               "symphony",
						"helm.chart.version":            testSymphonyHelmVersion,
						"helm.values.CUSTOM_VISION_KEY": "BBB",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestHelmTargetProviderInstallNoOci(t *testing.T) {
	// To run this test case successfully, you shouldn't have a symphony Helm chart already deployed to your current Kubernetes context
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSION")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION enviornment variable is not set")
	}
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "akri",
					Type: "helm.v3",
					Properties: map[string]string{
						"helm.chart.repo":    "https://project-akri.github.io/akri/akri",
						"helm.chart.name":    "akri",
						"helm.chart.version": "",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestHelmTargetProviderInstallDirectDownload(t *testing.T) {
	testGatekeeper := os.Getenv("TEST_HELM_GATEKEEPER")
	if testGatekeeper == "" {
		t.Skip("Skipping because TEST_HELM_GATEKEEPER enviornment variable is not set")
	}
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "gatekeeper",
					Type: "helm.v3",
					Properties: map[string]string{
						"helm.chart.repo": "https://open-policy-agent.github.io/gatekeeper/charts/gatekeeper-3.10.0-beta.1.tgz",
						"helm.chart.name": "gatekeeper",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestHelmTargetProviderRemove(t *testing.T) {
	testSymphonyHelmVersion := os.Getenv("TEST_SYMPHONY_HELM_VERSION")
	if testSymphonyHelmVersion == "" {
		t.Skip("Skipping because TEST_SYMPHONY_HELM_VERSION enviornment variable is not set")
	}
	config := HelmTargetProviderConfig{InCluster: true}
	provider := HelmTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceSpec{
			Name: "symphony",
		},
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "symphony-com",
					Type: "helm.v3",
					Properties: map[string]string{
						"helm.chart.repo":    "possprod.azurecr.io/helm/symphony",
						"helm.chart.name":    "symphony",
						"helm.chart.version": testSymphonyHelmVersion,
					},
				},
			},
		},
	}, nil)
	assert.Nil(t, err)
}
func TestHelmTargetProviderGetAnotherCluster(t *testing.T) {
	//to run this test successfully, you need to fix the cluster config below, and the target cluster shouldn't have symphony Helm chart installed
	//THIS CASE IS BROKERN
	testotherK8s := os.Getenv("TEST_HELM_OTHER_K8S")
	if testotherK8s == "" {
		t.Skip("Skipping because TEST_HELM_OTHER_K8S enviornment variable is not set")
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
	components, err := provider.Get(context.Background(), model.DeploymentSpec{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}
