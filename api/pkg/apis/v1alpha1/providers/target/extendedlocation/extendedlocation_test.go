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
					Type: "extended-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
						"resourceSyncRule":  "reosurceSyncRule01",
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "customLocation01",
				Type: "extended-location",
				Properties: map[string]interface{}{
					"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
					"resourceGroupName": "MyResourceGroup",
					"resourceName":      "customLocation01",
					"resourceSyncRule":  "reosurceSyncRule01",
				},
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(components))
}

// TestCustomLocationNilProperties tests the custom location properties are not set
func TestCustomLocationNilProperties(t *testing.T) {
	customLocation, err := toCustomLocationProperties(
		model.ComponentSpec{
			Name: "ExtendedLocation01",
			Type: "extended-location",
			Properties: map[string]interface{}{
				"location":       "West US",
				"customLocation": map[string]interface{}{},
			},
		},
	)
	_ = customLocation
	require.NotNil(t, err)
}

// TestCustomLocationProperties tests the custom location properties
func TestCustomLocationProperties(t *testing.T) {
	_, err := toCustomLocationProperties(
		model.ComponentSpec{
			Name: "ExtendedLocation01",
			Type: "extended-location",
			Properties: map[string]interface{}{
				"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
				"resourceGroupName": "MyResourceGroup",
				"location":          "West US",
				"customLocation": map[string]interface{}{
					"properties": map[string]interface{}{
						"namespace":          "namespace01",
						"displayName":        "customLocation01",
						"hostResourceID":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01",
						"clusterExtensionID": []string{"/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01/Microsoft.KubernetesConfiguration/clusterExtensions/ClusterMonitor"},
					},
				},
			},
		},
	)
	require.Nil(t, err)
}

// TestResourceSyncRuleProperties tests the resource sync rule properties
func TestResourceSyncRuleProperties(t *testing.T) {
	_, err := toResourceSyncRuleProperties(
		model.ComponentSpec{
			Name: "ExtendedLocation01",
			Type: "extended-location",
			Properties: map[string]interface{}{
				"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
				"resourceGroupName": "MyResourceGroup",
				"location":          "West US",
				"customLocation": map[string]interface{}{
					"properties": map[string]interface{}{
						"namespace":          "namespace01",
						"displayName":        "customLocation01",
						"hostResourceID":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01",
						"clusterExtensionID": []string{"/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01/Microsoft.KubernetesConfiguration/clusterExtensions/ClusterMonitor"},
					},
					"resourceSyncRule": map[string]interface{}{
						"name":     "resourceSyncRule01",
						"location": "West Us",
						"properties": map[string]interface{}{
							"priority": 999,
							"selector": map[string]interface{}{
								"matchLabels": map[string]string{
									"key1": "value1",
								},
								"matchExpressions": []map[string]interface{}{
									{
										"key":      "key4",
										"operator": "In",
										"values": []string{
											"value4",
										},
									},
								},
							},
							"targetResourceGroup": "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup",
						},
					},
				},
			},
		},
	)
	require.Nil(t, err)
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

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "ExtendedLocation01",
					Type: "extended-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"location":          "West US",
						"customLocation": map[string]interface{}{
							"properties": map[string]interface{}{
								"namespace":          "namespace01",
								"displayName":        "customLocation01",
								"hostResourceID":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01",
								"clusterExtensionID": []string{"/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01/Microsoft.KubernetesConfiguration/clusterExtensions/ClusterMonitor"},
							},
							"resourceSyncRule": map[string]interface{}{
								"name":     "resourceSyncRule01",
								"location": "West Us",
								"properties": map[string]interface{}{
									"priority": 999,
									"selector": map[string]interface{}{
										"matchLabels": map[string]string{
											"key1": "value1",
										},
										"matchExpressions": map[string]interface{}{
											"key":      "key4",
											"operator": "In",
											"values": []string{
												"value4",
											},
										},
									},
									"targetResourceGroup": "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup",
								},
							},
						},
					},
				},
			},
		},
	}, model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: "update",
				Component: model.ComponentSpec{
					Name: "ExtendedLocation01",
					Type: "extended-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"location":          "West US",
						"customLocation": map[string]interface{}{
							"properties": map[string]interface{}{
								"namespace":          "namespace01",
								"displayName":        "customLocation01",
								"hostResourceID":     "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01",
								"clusterExtensionID": []string{"/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup/providers/Microsoft.Kubernetes/connectedCluster/cluster01/Microsoft.KubernetesConfiguration/clusterExtensions/ClusterMonitor"},
							},
							"resourceSyncRule": map[string]interface{}{
								"name":     "resourceSyncRule01",
								"location": "West Us",
								"properties": map[string]interface{}{
									"priority": 999,
									"selector": map[string]interface{}{
										"matchLabels": map[string]string{
											"key1": "value1",
										},
										"matchExpressions": map[string]interface{}{
											"key":      "key4",
											"operator": "In",
											"values": []string{
												"value4",
											},
										},
									},
									"targetResourceGroup": "/subscriptions/77969078-2897-47b0-9143-917252379303/resourceGroups/MyResourceGroup",
								},
							},
						},
					},
				},
			},
		},
	}, false)

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

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "customLocation01",
					Type: "extended-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
						"resourceSyncRule":  "reosurceSyncRule01",
					},
				},
			},
		},
	}, model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: "delete",
				Component: model.ComponentSpec{
					Name: "customLocation01",
					Type: "extended-location",
					Properties: map[string]interface{}{
						"subscriptionID":    "77969078-2897-47b0-9143-917252379303",
						"resourceGroupName": "MyResourceGroup",
						"resourceName":      "customLocation01",
						"resourceSyncRule":  "reosurceSyncRule01",
					},
				},
			},
		},
	}, false)

	require.Nil(t, err)
}
