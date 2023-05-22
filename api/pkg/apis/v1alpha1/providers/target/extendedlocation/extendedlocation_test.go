package extendedlocation

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/require"
)

// TestAzureResourceTargetProviderConfigFromMapNil tests the null config map for provider
func TestAzureResourceTargetProviderConfigFromMapNil(t *testing.T) {
	_, err := ExtendedLocationTargetProviderConfigFromMap(nil)
	require.NotNil(t, err)
}

// TestAzureResourceTargetProviderConfigFromMapEmpty tests the empty config map for provider
func TestAzureResourceTargetProviderConfigFromMapEmpty(t *testing.T) {
	_, err := ExtendedLocationTargetProviderConfigFromMap(map[string]string{})
	require.NotNil(t, err)
}

// TestAzureResourceTargetProviderInitEmptyConfig tests the config initialization for provider
func TestAzureResourceTargetProviderInitEmptyConfig(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ExtendedLocationTargetProviderConfig{
		ClientID: clientID,
	}
	provider := ExtendedLocationTargetProvider{}
	err := provider.Init(config)
	require.NotNil(t, err)
}

// TestAzureResourceTargetProviderGet tests the get extended location functionality
func TestAzureResourceTargetProviderGet(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ExtendedLocationTargetProviderConfig{
		ClientID: clientID,
	}
	provider := ExtendedLocationTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	components, err := provider.Get(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "customLocation01",
					Type: "custom-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
					},
				},
				{
					Name: "resourceSyncRule01",
					Type: "resource-sync-rule",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
					},
				},
			},
		},
	},
	)

	require.Nil(t, err)
	require.Equal(t, 1, len(components))
}

// TestAzureResourceTargetProviderInstall tests the extended location installation for provider
func TestAzureResourceTargetProviderInstall(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ExtendedLocationTargetProviderConfig{
		ClientID: clientID,
	}

	provider := ExtendedLocationTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "customLocation01",
					Type: "custom-location",
					Properties: map[string]interface{}{
						"subscriptionID":     "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName":  "MyResourceGroup",
						"resourceName":       "customLocation01",
						"location":           "West US",
						"namespace":          "namespace01",
						"hostResourceID":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01",
						"clusterExtensionID": "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01/Microsoft.KubernetesConfiguration/clusterExtensions/ClusterMonitor",
					},
				},
				{
					Name: "resourceSyncRule01",
					Type: "resource-sync-rule",
					Properties: map[string]interface{}{
						"subscriptionID":          "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName":       "MyResourceGroup",
						"resourceName":            "customLocation01",
						"resourceSyncRuleName":    "resourceSyncRule01",
						"priority":                "999",
						"location":                "West Us",
						"matchExpressionKey":      "key4",
						"matchExpressionOperator": "In",
						"matchExpressionValue":    "value4",
						"matchLabelKey":           "value1",
						"targetResourceGroup":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup",
					},
				},
			},
		},
	},
	)

	require.Nil(t, err)
}

// TestAzureResourceTargetProviderRemove tests the delete functionality for extended location
func TestAzureResourceTargetProviderRemove(t *testing.T) {
	clientID := os.Getenv("TEST_CLIENT_ID")
	if clientID == "" {
		t.Skip("Skipping because TEST_CLIENT_ID environment variable is not set")
	}

	config := ExtendedLocationTargetProviderConfig{
		ClientID: clientID,
	}
	provider := ExtendedLocationTargetProvider{}
	err := provider.Init(config)
	require.Nil(t, err)

	err = provider.Remove(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "customLocation01",
					Type: "custom-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
					},
				},
				{
					Name: "resourceSyncRule01",
					Type: "resource-sync-rule",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
					},
				},
			},
		},
	},
		nil)

	require.Nil(t, err)
}
