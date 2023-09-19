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
					Name: "Bluefin",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID": "77969078-2897-47b0-9143-917252379303",
						"resourceGroup":  "MyResourceGroup",
						"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
					},
				},
			},
		},
	}, []model.ComponentStep{
		{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "Bluefin",
				Type: "arc-extension",
				Properties: map[string]interface{}{
					"subscriptionID": "77969078-2897-47b0-9143-917252379303",
					"resourceGroup":  "MyResourceGroup",
					"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
				},
			},
		},
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(components))
}

// TestExtensionProviderConfigSettingProperties tests the properties functions when config settings and protected settings are missing
func TestExtensionProviderConfigSettingProperties(t *testing.T) {
	_, err := toExtensionProperties(model.ComponentSpec{
		Name: "Bluefin",
		Type: "arc-extension",
		Properties: map[string]interface{}{
			"subscriptionID": "77969078-2897-47b0-9143-917252379303",
			"resourceGroup":  "MyResourceGroup",
			"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
			"arcExtension": map[string]interface{}{
				"extensionType":           "azuremonitor-containers",
				"autoUpgradeMinorVersion": false,
				"version":                 "0.1.1",
				"releaseTrain":            "dev",
			},
		},
	},
	)
	require.Nil(t, err)
}

// TestExtensionProviderProtectedConfigSettingProperties tests the properties functions when only the protected settings are present
func TestExtensionProviderProtectedConfigSettingProperties(t *testing.T) {
	_, err := toExtensionProperties(model.ComponentSpec{
		Name: "Bluefin",
		Type: "arc-extension",
		Properties: map[string]interface{}{
			"subscriptionID": "77969078-2897-47b0-9143-917252379303",
			"resourceGroup":  "MyResourceGroup",
			"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
			"arcExtension": map[string]interface{}{
				"extensionType":           "azuremonitor-containers",
				"autoUpgradeMinorVersion": false,
				"version":                 "0.1.1",
				"releaseTrain":            "dev",
				"configurationProtectedSettings": map[string]string{
					"secret.key": "secretKeyValue01",
				},
			},
		},
	},
	)
	require.Nil(t, err)
}

// TestExtensionTargetProviderProperties tests the properties function of ARC extension provider
func TestExtensionTargetProviderProperties(t *testing.T) {
	_, err := toExtensionProperties(model.ComponentSpec{
		Name: "Bluefin",
		Type: "arc-extension",
		Properties: map[string]interface{}{
			"subscriptionID": "77969078-2897-47b0-9143-917252379303",
			"resourceGroup":  "MyResourceGroup",
			"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
			"arcExtension": map[string]interface{}{
				"extensionType":           "azuremonitor-containers",
				"autoUpgradeMinorVersion": false,
				"version":                 "0.1.1",
				"releaseTrain":            "dev",
				"configurationSettings": map[string]string{
					"cluster": "my-arc-cluster",
				},
				"configurationProtectedSettings": map[string]string{
					"secret.key": "secretKeyValue01",
				},
			},
		},
	},
	)
	require.Nil(t, err)
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

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "Bluefin",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID": "77969078-2897-47b0-9143-917252379303",
						"resourceGroup":  "MyResourceGroup",
						"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
						"arcExtension": map[string]interface{}{
							"extensionType":                  "azuremonitor-containers",
							"autoUpgradeMinorVersion":        false,
							"version":                        "0.1.1",
							"releaseTrain":                   "dev",
							"configurationSettings":          map[string]string{},
							"configurationProtectedSettings": map[string]string{},
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
					Name: "Bluefin",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID": "77969078-2897-47b0-9143-917252379303",
						"resourceGroup":  "MyResourceGroup",
						"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
						"arcExtension": map[string]interface{}{
							"extensionType":                  "azuremonitor-containers",
							"autoUpgradeMinorVersion":        false,
							"version":                        "0.1.1",
							"releaseTrain":                   "dev",
							"configurationSettings":          map[string]string{},
							"configurationProtectedSettings": map[string]string{},
						},
					},
				},
			},
		},
	}, false)

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

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: []model.ComponentSpec{
				{
					Name: "Bluefin",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID": "77969078-2897-47b0-9143-917252379303",
						"resourceGroup":  "MyResourceGroup",
						"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
					},
				},
			},
		},
	}, model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Action: "delete",
				Component: model.ComponentSpec{
					Name: "Bluefin",
					Type: "arc-extension",
					Properties: map[string]interface{}{
						"subscriptionID": "77969078-2897-47b0-9143-917252379303",
						"resourceGroup":  "MyResourceGroup",
						"cluster":        "Microsoft.Kubernetes/connectedClusters/my-arc-cluster",
					},
				},
			},
		},
	}, false)

	require.Nil(t, err)
}
