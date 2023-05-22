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
	"errors"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/goccy/go-json"
)

var sLog = logger.NewLogger("coa.runtime")

type (
	// ArcExtensionTargetProviderConfig ARC extension config
	ArcExtensionTargetProviderConfig struct {
		ClientID string `json:"clientID"`
	}

	// ArcExtensionTargetProviderProperty ARC extension property
	ArcExtensionTargetProviderProperty struct {
		Name                    string `json:"extensionName"`
		Type                    string `json:"extensionType"`
		ClusterName             string `json:"clusterName"`
		ClusterRp               string `json:"clusterRp"`
		ClusterResourceName     string `json:"clusterResourceName"`
		ResourceGroup           string `json:"resourceGroup"`
		Version                 string `json:"apiVersion"`
		SubscriptionID          string `json:"subscriptionID"`
		ReleaseTrain            string `json:"releaseTrain,omitempty"`
		AutoUpgradeMinorVersion string `json:"autoUpgradeMinorVersion,omitempty"`
	}

	// ArcExtensionTargetProvider ARC extension provider
	ArcExtensionTargetProvider struct {
		Config  ArcExtensionTargetProviderConfig
		Context *contexts.ManagerContext
	}
)

const (
	clusterNameKey         = "clusterName"
	clusterRpKey           = "clusterRp"
	clusterResourceNameKey = "clusterResourceName"
	resourceGroupKey       = "resourceGroup"
	apiVersionKey          = "apiVersion"
)

// ArcExtensionTargetProviderConfigFromMap creates the config map for ARC extension provider
func ArcExtensionTargetProviderConfigFromMap(properties map[string]string) (ArcExtensionTargetProviderConfig, error) {
	ret := ArcExtensionTargetProviderConfig{}
	v, ok := properties["clientID"]
	if !ok {
		return ret, v1alpha2.NewCOAError(nil, "Arc Extension Client ID for provider is not set", v1alpha2.BadConfig)
	}

	ret.ClientID = v

	return ret, nil
}

// InitWithMap initializes the config map for ARC extension provider
func (i *ArcExtensionTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ArcExtensionTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}

	return i.Init(config)
}

// Init initializes the config for the ARC extension provider
func (i *ArcExtensionTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Arc Extension Target Provider",
		context.Background(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Info("  P (Extension Target): Init()")

	// get the arc extension config
	extensionConfig, err := toArcExtensionTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Arc Extension Target): expected ArcExtensionTargetProviderConfig: %+v", err)
		return err
	}

	i.Config = extensionConfig
	return nil
}

// toArcExtensionTargetProviderConfig sets the ARC extension config
func toArcExtensionTargetProviderConfig(config providers.IProviderConfig) (ArcExtensionTargetProviderConfig, error) {
	ret := ArcExtensionTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("ArcExtensionTargetProviderConfig is null")
	}

	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

// NeedsUpdate checks if the ARC extension needs an update
func (i *ArcExtensionTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	sLog.Infof(" P (Arc Extension Target): NeedsUpdate: %d - %d", len(desired), len(current))
	extensionProperty := []string{clusterNameKey, clusterRpKey, clusterResourceNameKey, resourceGroupKey, apiVersionKey}
	for _, dc := range desired {
		found := false
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Type == "arc-extension" {
				for _, param := range extensionProperty {
					if cc.Properties[param] != "" && cc.Properties[param] != dc.Properties[param] {
						found = true
						break
					}
				}
			}
		}
		if found {
			sLog.Info(" P (Arc Extension Target): NeedsUpdate: returning true")
			return true
		}
	}

	sLog.Info(" P (Arc Extension Target): NeedsUpdate: returning false")
	return false
}

// NeedsRemove checks if the Arc extension component needs to be removed
func (i *ArcExtensionTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	sLog.Infof("  P (Arc Extension Target Provider): NeedsRemove: %d - %d", len(desired), len(current))
	extensionProperty := []string{clusterNameKey, apiVersionKey}
	for _, dc := range desired {
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Type == "arc-extension" {
				for _, param := range extensionProperty {
					if cc.Properties[param] == dc.Properties[param] {
						sLog.Info("  P (Arc Extension Target Provider): NeedsRemove: returning true")
						return true
					}
				}
			}
		}
	}

	sLog.Info("  P (Arc Extension Target Provider): NeedsRemove: returning false")
	return false
}

// Get gets the ARC extension details from connected k8s cluster
func (i *ArcExtensionTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan(
		"Arc Extension Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	ret := []model.ComponentSpec{}

	// opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}

	// cred gets the system assigned managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Arc Extension Target Managed Identity Credential):", err)
		return ret, err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		if c.Type == "arc-extension" {
			// deployment has ARC extension properties from component
			deployment, err := getDeploymentFromComponent(c)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target):", err)
				return ret, err
			}

			// clientFactory is a new Azure client using System Assigned Managed Identity credentials
			clientFactory, err := armkubernetesconfiguration.NewClientFactory(deployment.SubscriptionID, cred, nil)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Subscription ID):", err)
				return nil, err
			}

			_, err = clientFactory.NewExtensionsClient().Get(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, nil)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deployment):", err)
				return nil, err
			}

			ret = append(ret, model.ComponentSpec{
				Name: deployment.Name,
				Type: deployment.Type,
				Properties: map[string]interface{}{
					"clusterName":         deployment.ClusterName,
					"clusterRp":           deployment.ClusterRp,
					"clusterResourceName": deployment.ClusterResourceName,
					"resourceGroup":       deployment.ResourceGroup,
					"apiVersion":          deployment.Version,
				},
			})
		}
	}

	return ret, nil
}

// Remove deletes the ARC extension from the connected k8s cluster
func (i *ArcExtensionTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan(
		"Arc Extension Target Provider",
		ctx,
		&map[string]string{
			"method": "Remove",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("  P (Arc Extension Target): deleting components")
	// opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}

	// cred has the managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Arc Extension Target Managed Identity Credential):", err)
		return err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		if c.Type == "arc-extension" {
			// deployment has ARC extension properties from the component
			deployment, err := getDeploymentFromComponent(c)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deletion):", err)
				return err
			}

			// clientFactory is a new Aure client
			clientFactory, err := armkubernetesconfiguration.NewClientFactory(deployment.SubscriptionID, cred, nil)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Subscription ID):", err)
				return err
			}

			_, err = clientFactory.NewExtensionsClient().BeginDelete(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, &armkubernetesconfiguration.ExtensionsClientBeginDeleteOptions{ForceDelete: nil})
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deployment):", err)
				return err
			}
		}
	}

	return nil
}

// Apply installs the ARC extension on the connected k8s cluster
func (i *ArcExtensionTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan(
		"Arc Extension Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Infof("  P (Arc Extension Target): applying artifacts: applying components")

	// opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}
	// cred had the managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Arc Extension Target Managed Identity Credential):", err)
		return err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		if c.Type == "arc-extension" {
			// deployment has the ARC extension properties from component
			deployment, err := getDeploymentFromComponent(c)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deployment):", err)
				return err
			}

			// clientFactory is a new Azure client
			clientFactory, err := armkubernetesconfiguration.NewClientFactory(deployment.SubscriptionID, cred, nil)
			if err != nil {
				sLog.Errorf(" P (Extension Target Subscription ID):", err)
				return err
			}

			upgradeVersion, err := strconv.ParseBool(deployment.AutoUpgradeMinorVersion)
			// extension sets the ARC extension property object
			extension := armkubernetesconfiguration.Extension{
				Properties: &armkubernetesconfiguration.ExtensionProperties{
					AutoUpgradeMinorVersion: &upgradeVersion,
					ExtensionType:           &deployment.Type,
					ReleaseTrain:            &deployment.ReleaseTrain,
				},
			}

			_, err = clientFactory.NewExtensionsClient().BeginCreate(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, extension, nil)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deployment):", err)
				return err
			}
		}
	}

	return nil
}

// getDeploymentFromComponent gets the ARC extension component properties
func getDeploymentFromComponent(component model.ComponentSpec) (ArcExtensionTargetProviderProperty, error) {
	ret := ArcExtensionTargetProviderProperty{}
	if component.Name == "" {
		return ret, errors.New("Arc Extension Name is not set")
	}

	if component.Type == "" {
		return ret, errors.New("Arc Extension Type is not set")
	}

	if component.Properties == nil {
		return ret, errors.New("Arc Extension is null")
	}

	data, err := json.Marshal(component.Properties)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}
