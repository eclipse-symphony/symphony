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

package extension

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
	client_factory "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
	extensions_client "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/kubernetesconfiguration/armkubernetesconfiguration"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type ExtensionTargetProviderConfig struct {
	ClientId       string `json:"clientId"`
	SubscriptionId string `json:"subscriptionId"`
}
type ExtensionTargetProviderProperty struct {
	Name                string `json:"extensionName"`
	Type                string `json:"extensionType"`
	ClusterName         string `json:"clusterName"`
	ClusterRp           string `json:"clusterRp"`
	ClusterResourceName string `json:"clusterResourceName"`
	ResourceGroup       string `json:"resourceGroup"`
	Version             string `json:"apiVersion"`
}

type ExtensionTargetProvider struct {
	Config  ExtensionTargetProviderConfig
	Context *contexts.ManagerContext
}

func ExtensionTargetProviderConfigFromMap(properties map[string]string) (ExtensionTargetProviderConfig, error) {
	ret := ExtensionTargetProviderConfig{}
	if v, ok := properties["clientId"]; ok {
		ret.ClientId = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Extension client id for provider is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["subscriptionId"]; ok {
		ret.SubscriptionId = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Extension subscription id for provider is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}
func (i *ExtensionTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ExtensionTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func (i *ExtensionTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Extension Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("  P (Extension Target): Init()")
	extensionConfig, err := toExtensionTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Extension Target): expected ExtensionTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = extensionConfig
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toExtensionTargetProviderConfig(config providers.IProviderConfig) (ExtensionTargetProviderConfig, error) {
	ret := ExtensionTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("ExtensionTargetProviderConfig is null")
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
func (i *ExtensionTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	ctx, span := observability.StartSpan("Extension Target Provider", ctx, &map[string]string{
		"method": "NeedsUpdate",
	})
	sLog.Infof(" P (Extension Target): NeedsUpdate: %d - %d", len(desired), len(current))
	for _, dc := range desired {
		found := false
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Type == "arc-extensions" {
				if cc.Properties["clusterName"] != "" && cc.Properties["clusterName"] != dc.Properties["clusterName"] {
					found = true
					break
				}
				if cc.Properties["clusterRp"] != "" && cc.Properties["clusterRp"] != dc.Properties["clusterRp"] {
					found = true
					break
				}
				if cc.Properties["clusterResourceName"] != "" && cc.Properties["clusterResourceName"] != dc.Properties["clusterResourceName"] {
					found = true
					break
				}
				if cc.Properties["resourceGroup"] != "" && cc.Properties["resourceGroup"] != dc.Properties["resourceGroup"] {
					found = true
					break
				}
				if cc.Properties["apiVersion"] != "" && cc.Properties["apiVersion"] != dc.Properties["apiVersion"] {
					found = true
					break
				}
			}
		}
		if found {
			// extension needs an update
			sLog.Info(" P (Extension Target): NeedsUpdate: returning true")
			observ_utils.CloseSpanWithError(span, nil)
			return true
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	sLog.Info(" P (Extension Target): NeedsUpdate: returning false")
	return false
}

func (i *ExtensionTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Extension Target Provider", ctx, &map[string]string{
		"method": "NeedsRemove",
	})
	sLog.Infof("  P (Extension Target Provider): NeedsRemove: %d - %d", len(desired), len(current))
	for _, dc := range desired {
		for _, cc := range current {
			if cc.Name == dc.Name && cc.Type == "arc-extensions" {
				if cc.Properties["apiVersion"] == dc.Properties["apiVersion"] {
					observ_utils.CloseSpanWithError(span, nil)
					sLog.Info("  P (Extension Target Provider): NeedsRemove: returning true")
					return true
				}
				if cc.Properties["clusterName"] == dc.Properties["clusterName"] {
					observ_utils.CloseSpanWithError(span, nil)
					sLog.Info("  P (Extension Target Provider): NeedsRemove: returning true")
					return true
				}
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	sLog.Info("  P (Extension Target Provider): NeedsRemove: returning false")
	return false
}

func (i *ExtensionTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Extension Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	ret := []model.ComponentSpec{}
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientId),
	}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf(" P (Extension Target Managed Identity Credential):", err)
		return ret, err
	}
	clientFactory, err := client_factory.NewClientFactory(i.Config.SubscriptionId, cred, nil)
	components := deployment.GetComponentSlice()
	for _, c := range components {
		if c.Type == "arc-extensions" {
			deployment, err := getDeploymentFromComponent(ctx, c)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target):", err)
				return ret, err
			}
			ret, err := clientFactory.NewExtensionsClient().Get(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, nil)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target Deployment):", err)
				return nil, err
			}
			_ = ret
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (i *ExtensionTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Extension Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("  P (Extension Target): deleting components")
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientId),
	}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf(" P (Extension Target Managed Identity Credential):", err)
		return err
	}
	clientFactory, err := client_factory.NewClientFactory(i.Config.SubscriptionId, cred, nil)
	components := deployment.GetComponentSlice()
	if len(components) == 0 {
		observ_utils.CloseSpanWithError(span, nil)
		return nil
	}
	for _, c := range components {
		if c.Type == "arc-extensions" {
			deployment, err := getDeploymentFromComponent(ctx, c)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target Deletion):", err)
				return err
			}
			ret, err := clientFactory.NewExtensionsClient().BeginDelete(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, &extensions_client.ExtensionsClientBeginDeleteOptions{ForceDelete: nil})
			_ = ret
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target Deployment):", err)
				return err
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func (i *ExtensionTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan("Extension Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Extension Target): applying artifacts: applying components")
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientId),
	}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf(" P (Extension Target Managed Identity Credential):", err)
		return err
	}
	clientFactory, err := client_factory.NewClientFactory(i.Config.SubscriptionId, cred, nil)
	components := deployment.GetComponentSlice()
	for _, c := range components {
		if c.Type == "arc-extensions" {
			deployment, err := getDeploymentFromComponent(ctx, c)
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target Deployment):", err)
				return err
			}
			ret, err := clientFactory.NewExtensionsClient().BeginCreate(ctx, deployment.ResourceGroup, deployment.ClusterRp, deployment.ClusterResourceName, deployment.ClusterName, deployment.Name, armkubernetesconfiguration.Extension{}, nil)
			_ = ret
			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf(" P (Extension Target Deployment):", err)
				return err
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func getDeploymentFromComponent(ctx context.Context, c model.ComponentSpec) (ExtensionTargetProviderProperty, error) {
	ok := false
	deployment := ExtensionTargetProviderProperty{}
	if deployment.ResourceGroup, ok = c.Properties["resourceGroup"]; !ok {
		return deployment, errors.New("component doesn't contain a resource group property")
	}
	if deployment.ClusterName, ok = c.Properties["clusterName"]; !ok {
		return deployment, errors.New("component doesn't contain a cluster name property")
	}
	if deployment.ClusterRp, ok = c.Properties["clusterRp"]; !ok {
		return deployment, errors.New("component doesn't contain a cluster resource provider property")
	}
	if deployment.ClusterResourceName, ok = c.Properties["clusterResourceName"]; !ok {
		return deployment, errors.New("component doesn't contain a cluster resource name property")
	}
	if deployment.Name, ok = c.Properties["extensionName"]; !ok {
		return deployment, errors.New("component doesn't contain an extension name property")
	}
	if deployment.Type, ok = c.Properties["extensionType"]; !ok {
		return deployment, errors.New("component doesn't contain an extension type property")
	}
	if deployment.Version, ok = c.Properties["apiVersion"]; !ok {
		return deployment, errors.New("component doesn't contain an api version property")
	}
	return deployment, nil
}
