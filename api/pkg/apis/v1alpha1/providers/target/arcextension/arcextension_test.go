package arcextension

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/require"
)

// TestExtensionTargetProviderConfigFromMapNil tests the null provider config
func TestExtensionTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := ArcExtensionTargetProviderConfigFromMap(nil)
	require.NotNil(t, err)
}

// TestExtensionTargetProviderConfigFromMapEmpty tests the empty provider config
func TestExtensionTargetProviderConfigFromMapEmpty(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	_, err := ArcExtensionTargetProviderConfigFromMap(map[string]string{
		"clientID": clientID,
	})

	require.Nil(t, err)
}

// TestExtensionTargetProviderInitEmptyConfig test the initialization of provider config
func TestExtensionTargetProviderInitEmptyConfig(t *testing.T) {
	config := ArcExtensionTargetProviderConfig{}
	provider := ArcExtensionTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)
}

// TestExtensionTargetProviderGet tests the get function of ARC extension provider
func TestExtensionTargetProviderGet(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ArcExtensionTargetProviderConfig{
		ClientID: clientID,
	}
	provider := ArcExtensionTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID":      "77969078-2897-47b0-9143-917252379303",
						"extensionName":       "ClusterMonitor",
						"extensionType":       "azuremonitor-containers",
						"clusterName":         "my-arc-cluster",
						"clusterRp":           "Microsoft.Kubernetes",
						"clusterResourceName": "connectedClusters",
						"resourceGroup":       "MyResourceGroup",
						"apiVersion":          "2023-05-02",
					},
				},
			},
		},
	},
	)
	require.Nil(t, err)
	require.Equal(t, 1, len(components))
}

// TestExtensionTargetProviderInstall tests the install function of the ARC extension provider
func TestExtensionTargetProviderInstall(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ArcExtensionTargetProviderConfig{
		ClientID: clientID,
	}
	provider := ArcExtensionTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID":          "77969078-2897-47b0-9143-917252379303",
						"extensionName":           "ClusterMonitor",
						"extensionType":           "azuremonitor-containers",
						"clusterName":             "my-arc-cluster",
						"clusterRp":               "Microsoft.Kubernetes",
						"clusterResourceName":     "connectedClusters",
						"resourceGroup":           "MyResourceGroup",
						"autoUpgradeMinorVersion": "false",
						"apiVersion":              "2023-05-02",
						"releaseTrain":            "preview",
					},
				},
			},
		},
	},
	)

	require.Nil(t, err)
}

// TestExtensionTargetProviderRemove tests the uninstall function of ARC extension provider
func TestExtensionTargetProviderRemove(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}
	config := ArcExtensionTargetProviderConfig{
		ClientID: clientID,
	}

	provider := ArcExtensionTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ClusterMonitor",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID":      "77969078-2897-47b0-9143-917252379303",
						"extensionName":       "ClusterMonitor",
						"extensionType":       "azuremonitor-containers",
						"clusterName":         "my-arc-cluster",
						"clusterRp":           "Microsoft.Kubernetes",
						"clusterResourceName": "connectedClusters",
						"resourceGroup":       "MyResourceGroup",
						"apiVersion":          "2023-05-02",
					},
				},
			},
		},
	}, nil)

	require.Nil(t, err)
}
