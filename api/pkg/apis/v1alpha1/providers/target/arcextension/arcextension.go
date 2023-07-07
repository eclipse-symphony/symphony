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
	"strings"

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
		Name           string `json:"extensionName"`
		Type           string `json:"extensionType"`
		Cluster        string `json:"cluster"`
		ResourceGroup  string `json:"resourceGroup"`
		SubscriptionID string `json:"subscriptionID"`
	}

	// ArcExtensionTargetProvider ARC extension provider
	ArcExtensionTargetProvider struct {
		Config  ArcExtensionTargetProviderConfig
		Context *contexts.ManagerContext
	}
)

const (
	clusterKey        = "cluster"
	resourceGroupKey  = "resourceGroup"
	versionKey        = "version"
	subscriptionIDKey = "subscriptionID"
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

// Get gets the ARC extension details from connected k8s cluster
func (i *ArcExtensionTargetProvider) Get(ctx context.Context, dep model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
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

	for _, c := range references {
		if c.Component.Type == "arc-extension" {
			// deployment has ARC extension properties from component
			deployment, err := getDeploymentFromComponent(c.Component)
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

			clusterDetails := strings.Split(deployment.Cluster, "/")
			if len(clusterDetails) < 3 {
				err = errors.New("ArcExtensionTargetProvider cluster details are missing")
				return ret, err
			}

			_, err = clientFactory.NewExtensionsClient().Get(ctx, deployment.ResourceGroup, clusterDetails[0], clusterDetails[1], clusterDetails[2], deployment.Name, nil)
			if err != nil {
				sLog.Errorf(" P (Arc Extension Target Deployment):", err)
				return nil, err
			}

			ret = append(ret, model.ComponentSpec{
				Name: deployment.Name,
				Type: deployment.Type,
				Properties: map[string]interface{}{
					"cluster":       deployment.Cluster,
					"resourceGroup": deployment.ResourceGroup,
				},
			})
		}
	}

	return ret, nil
}

// Apply installs the ARC extension on the connected k8s cluster
func (i *ArcExtensionTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
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

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return nil, err
	}
	if isDryRun {
		return nil, nil
	}

	ret := step.PrepareResultMap()

	// opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}
	// cred had the managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Arc Extension Target Managed Identity Credential):", err)
		return nil, err
	}

	updated := step.GetUpdatedComponents()
	if len(updated) > 0 {
		for _, c := range components {
			if c.Type == "arc-extension" {
				// deployment has the ARC extension properties from component
				deployment, err := getDeploymentFromComponent(c)
				if err != nil {
					sLog.Errorf(" P (Arc Extension Target Deployment):", err)
					return ret, err
				}

				// clientFactory is a new Azure client
				clientFactory, err := armkubernetesconfiguration.NewClientFactory(deployment.SubscriptionID, cred, nil)
				if err != nil {
					sLog.Errorf(" P (Extension Target Subscription ID):", err)
					return ret, err
				}

				clusterDetails := strings.Split(deployment.Cluster, "/")
				if len(clusterDetails) < 3 {
					err = errors.New("ArcExtensionTargetProvider cluster details are missing")
					return ret, err
				}

				extensionDetails, err := toExtensionProperties(c)
				if err != nil {
					return ret, err
				}

				_, err = clientFactory.NewExtensionsClient().BeginCreate(ctx, deployment.ResourceGroup, clusterDetails[0], clusterDetails[1], clusterDetails[2], deployment.Name, extensionDetails, nil)
				if err != nil {
					sLog.Errorf(" P (Arc Extension Target Deployment):", err)
					return ret, err
				}
			}
		}
	}
	deleted := step.GetDeletedComponents()
	if len(deleted) > 0 {
		for _, c := range components {
			if c.Type == "arc-extension" {
				// deployment has ARC extension properties from the component
				deployment, err := getDeploymentFromComponent(c)
				if err != nil {
					sLog.Errorf(" P (Arc Extension Target Deletion):", err)
					return ret, err
				}

				// clientFactory is a new Aure client
				clientFactory, err := armkubernetesconfiguration.NewClientFactory(deployment.SubscriptionID, cred, nil)
				if err != nil {
					sLog.Errorf(" P (Arc Extension Target Subscription ID):", err)
					return ret, err
				}

				clusterDetails := strings.Split(deployment.Cluster, "/")
				if len(clusterDetails) < 3 {
					err = errors.New("ArcExtensionTargetProvider cluster details are missing")
					return ret, err
				}

				_, err = clientFactory.NewExtensionsClient().BeginDelete(ctx, deployment.ResourceGroup, clusterDetails[0], clusterDetails[1], clusterDetails[2], deployment.Name, &armkubernetesconfiguration.ExtensionsClientBeginDeleteOptions{ForceDelete: nil})
				if err != nil {
					sLog.Errorf(" P (Arc Extension Target Deployment):", err)
					return ret, err
				}
			}
		}
	}
	return ret, nil
}

// toExtensionProperties sets the arc extension properties
func toExtensionProperties(c model.ComponentSpec) (armkubernetesconfiguration.Extension, error) {
	ret := armkubernetesconfiguration.Extension{Properties: &armkubernetesconfiguration.ExtensionProperties{}}
	if c.Properties["arcExtension"] == nil {
		return ret, errors.New("Arc extension properties are not set")
	}

	extension := c.Properties["arcExtension"]
	arcExt, ok := extension.(map[string]interface{})
	if !ok {
		return ret, errors.New("The Arc extension properties are not set")
	}

	extType, ok := arcExt["extensionType"].(string)
	if !ok {
		return ret, errors.New("The Arc extension type property is not set")
	}

	ret.Properties.ExtensionType = &extType
	upgradeVersion, ok := arcExt["autoUpgradeMinorVersion"].(bool)
	if !ok {
		return ret, errors.New("The Arc extension autoUpgradeMinorVersion property is not set")
	}

	ret.Properties.AutoUpgradeMinorVersion = &upgradeVersion
	version, ok := arcExt["version"].(string)
	if !ok {
		return ret, errors.New("The Arc extension version property is not set")
	}

	ret.Properties.Version = &version
	releaseTrain, ok := arcExt["releaseTrain"].(string)
	if !ok {
		return ret, errors.New("The Arc extension releaseTrain property is not set")
	}

	ret.Properties.ReleaseTrain = &releaseTrain
	if arcExt["configurationSettings"] != nil {
		configurationSettings, ok := arcExt["configurationSettings"].(map[string]string)
		if !ok {
			return ret, errors.New("The Arc extension configuration settings are not set")
		}

		settings := map[string]*string{}
		for index, data := range configurationSettings {
			settings[index] = &data
		}

		ret.Properties.ConfigurationSettings = settings
	}

	if arcExt["configurationProtectedSettings"] != nil {
		configurationProtectedSettings, ok := arcExt["configurationProtectedSettings"].(map[string]string)
		if !ok {
			return ret, errors.New("The Arc extension configuration protected settings are not set")
		}

		protectedSettings := map[string]*string{}
		for index, data := range configurationProtectedSettings {
			protectedSettings[index] = &data
		}

		ret.Properties.ConfigurationSettings = protectedSettings
	}

	return ret, nil
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

func (*ArcExtensionTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{"arc-extension"},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
		ChangeDetectionProperties: []model.PropertyDesc{
			{Name: "arc-extension", IgnoreCase: false, SkipIfMissing: false},
			{Name: clusterKey, IgnoreCase: false, SkipIfMissing: true},
			{Name: resourceGroupKey, IgnoreCase: false, SkipIfMissing: true},
			{Name: versionKey, IgnoreCase: false, SkipIfMissing: true},
			{Name: subscriptionIDKey, IgnoreCase: false, SkipIfMissing: true}, // TODO: add capability to skip comparision for deletion
		},
	}
}
