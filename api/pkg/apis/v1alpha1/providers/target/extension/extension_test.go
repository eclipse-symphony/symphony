package extension

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestExtensionTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := ExtensionTargetProviderConfigFromMap(nil)
	assert.NotNil(t, err)
}
func TestExtensionTargetProviderConfigFromMapEmpty(t *testing.T) {
	subscriptionId := os.Getenv("TEST_SUBSCRIPTION_ID")
	if subscriptionId == "" {
		t.Skip("Skipping because TEST_SUBSCRIPTION_ID environment variable is not set")
	}
	clientId := os.Getenv("TEST_CLIENT_ID")
	if clientId == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}
	_, err := ExtensionTargetProviderConfigFromMap(map[string]string{
		"clientId":       clientId,
		"subscriptionId": subscriptionId,
	})
	assert.Nil(t, err)
}
func TestExtensionTargetProviderInitEmptyConfig(t *testing.T) {
	config := ExtensionTargetProviderConfig{}
	provider := ExtensionTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}
func TestExtensionTargetProviderGet(t *testing.T) {
	subscriptionId := os.Getenv("TEST_SUBSCRIPTION_ID")
	if subscriptionId == "" {
		t.Skip("Skipping because TEST_SUBSCRIPTION_ID environment variable is not set")
	}
	clientId := os.Getenv("TEST_CLIENT_ID")
	if clientId == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}
	config := ExtensionTargetProviderConfig{}
	config.ClientId = clientId
	config.SubscriptionId = subscriptionId

	provider := ExtensionTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extensions",
					Properties: map[string]string{
						"extensionName":       "ClusterMonitor",
						"extensionType":       "azuremonitor-containers",
						"clusterName":         "my-arc-cluster",
						"clusterRp":           "Microsoft.Kubernetes",
						"clusterResourceName": "connectedClusters",
						"resourceGroup":       "MyResourceGroup",
						"apiVersion":          "0.41.2",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))
}
func TestExtensionTargetProviderInstall(t *testing.T) {
	subscriptionId := os.Getenv("TEST_SUBSCRIPTION_ID")
	if subscriptionId == "" {
		t.Skip("Skipping because TEST_SUBSCRIPTION_ID environment variable is not set")
	}
	clientId := os.Getenv("TEST_CLIENT_ID")
	if clientId == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}
	config := ExtensionTargetProviderConfig{}
	config.ClientId = clientId
	config.SubscriptionId = subscriptionId
	provider := ExtensionTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extensions",
					Properties: map[string]string{
						"extensionName":       "ClusterMonitor",
						"extensionType":       "azuremonitor-containers",
						"clusterName":         "my-arc-cluster",
						"clusterRp":           "Microsoft.Kubernetes",
						"clusterResourceName": "connectedClusters",
						"resourceGroup":       "MyResourceGroup",
						"apiVersion":          "0.41.2",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestExtensionTargetProviderRemove(t *testing.T) {
	subscriptionId := os.Getenv("TEST_SUBSCRIPTION_ID")
	if subscriptionId == "" {
		t.Skip("Skipping because TEST_SUBSCRIPTION_ID environment variable is not set")
	}
	clientId := os.Getenv("TEST_CLIENT_ID")
	if clientId == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}
	config := ExtensionTargetProviderConfig{}
	config.ClientId = clientId
	config.SubscriptionId = subscriptionId
	provider := ExtensionTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extensions",
					Properties: map[string]string{
						"extensionName":       "ClusterMonitor",
						"extensionType":       "azuremonitor-containers",
						"clusterName":         "my-arc-cluster",
						"clusterRp":           "Microsoft.Kubernetes",
						"clusterResourceName": "connectedClusters",
						"resourceGroup":       "MyResourceGroup",
						"apiVersion":          "0.41.2",
					},
				},
			},
		},
	}, nil)
	assert.Nil(t, err)
}
