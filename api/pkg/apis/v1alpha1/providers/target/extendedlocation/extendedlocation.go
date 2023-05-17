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
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/extendedlocation/armextendedlocation"
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
	// ExtendedTargetProviderConfig the extended location config
	ExtendedLocationTargetProviderConfig struct {
		ClientID string `json:"clientID"`
	}

	// ExtendedLocationProperty the extended location property
	ExtendedLocationProperty struct {
		Name                    string `json:"name"`
		Type                    string `json:"type"`
		ResourceGroupName       string `json:"resourceGroupName"`
		ResourceName            string `json:"ResourceName"`
		SubscriptionID          string `json:"subscriptionID"`
		Location                string `json:"location,omitempty"`
		Namespace               string `json:"namespace,omitempty"`
		HostResourceID          string `json:"hostResourceID,omitempty"`
		ClusterExtensionID      string `json:"clusterExtensionID,omitempty"`
		Priority                string `json:"priority,omitempty"`
		MatchExpressionKey      string `json:"matchExpressionKey,omitempty"`
		MatchExpressionValue    string `json:"matchExpressionValue,omitempty"`
		MatchExpressionOperator string `json:"matchExpressionOperator,omitempty"`
		MatchLabelKey           string `json:"matchLabelKey,omitempty"`
		TargetResourceGroup     string `json:"targetResourceGroup,omitempty"`
	}

	// ExtendedLocationTargetProvider the target location provider
	ExtendedLocationTargetProvider struct {
		Config  ExtendedLocationTargetProviderConfig
		Context *contexts.ManagerContext
	}
)

const (
	resourceNameKey      = "resourceName"
	resourceGroupNameKey = "resourceGroupName"
)

// ExtendedLocationTargetProviderConfigFromMap sets the config map for extended location provider
func ExtendedLocationTargetProviderConfigFromMap(properties map[string]string) (ExtendedLocationTargetProviderConfig, error) {
	ret := ExtendedLocationTargetProviderConfig{}
	v, ok := properties["clientID"]
	if !ok {
		return ret, v1alpha2.NewCOAError(nil, "Extended Location provider clientID is not set", v1alpha2.BadConfig)
	}

	ret.ClientID = v
	return ret, nil
}

// InitWithMap initializes the config map for extended location
func (i *ExtendedLocationTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ExtendedLocationTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}

	return i.Init(config)
}

// Init initializes the config for extended location provider
func (i *ExtendedLocationTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Extended Location Target Provider",
		context.Background(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	sLog.Info(" P (Extended Location Target) : Init()")

	updateConfig, err := toExtendedLocationTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf(" P (Extended Location Target) : expected ExtendedLocationTargetProviderConfig: %+v", err)
		return err
	}

	i.Config = updateConfig
	return nil
}

// toExtendedLocationTargetProviderConfig sets the provider config
func toExtendedLocationTargetProviderConfig(config providers.IProviderConfig) (ExtendedLocationTargetProviderConfig, error) {
	ret := ExtendedLocationTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("ExtendedLocationTargetProviderConfig is null")
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

// Get gets the extended location details from ARC enabled cluster
func (i *ExtendedLocationTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan(
		"Extended Location Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)
	ret := make([]model.ComponentSpec, 0)

	// opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}

	// cred sets the managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Extended Location Target Managed Identity Credential):", err)
		return ret, err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		// deployment gets the extended location properties from component
		deployment, err := getDeploymentFromComponent(c)
		if err != nil {
			sLog.Errorf(" P (Extended Location Custom Location Target):", err)
			return ret, err
		}

		// clientFactory creates a new client for Azure API
		clientFactory, err := armextendedlocation.NewClientFactory(deployment.SubscriptionID, cred, nil)
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Subscription ID):", err)
			return ret, err
		}

		switch c.Type {
		case "custom-location":
			_, err = clientFactory.NewCustomLocationsClient().Get(ctx, deployment.ResourceGroupName, deployment.ResourceName, nil)

		case "resource-sync-rule":
			_, err = clientFactory.NewResourceSyncRulesClient().Get(ctx, deployment.ResourceGroupName, deployment.ResourceName, deployment.Name, nil)
		}
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Get):", err)
			return nil, err
		}

		ret = append(ret, model.ComponentSpec{
			Name: deployment.Name,
			Type: deployment.Type,
			Properties: map[string]string{
				"resourceGroupName": deployment.ResourceGroupName,
				"resourceName":      deployment.ResourceName,
			},
		})
	}

	return ret, nil
}

// Remove deletes the extended location from ARC enabled cluster
func (i *ExtendedLocationTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan(
		"Extended Location Provider",
		ctx,
		&map[string]string{
			"method": "Remove",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)

	//opts sets the system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}

	//cred gets the managed identity credentials
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		sLog.Errorf(" P (Extended Location Target Managed Identity Credential):", err)
		return err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		deployment, err := getDeploymentFromComponent(c)
		if err != nil {
			sLog.Errorf(" P (Extended Location Custom Location Target):", err)
			return err
		}

		// clientFactory creates a new client for Azure API
		clientFactory, err := armextendedlocation.NewClientFactory(deployment.SubscriptionID, cred, nil)
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Subscription ID):", err)
			return err
		}

		switch c.Type {
		case "custom-location":
			_, err = clientFactory.NewCustomLocationsClient().BeginDelete(ctx, deployment.ResourceGroupName, deployment.ResourceName, nil)

		case "resource-sync-rule":
			_, err = clientFactory.NewResourceSyncRulesClient().Delete(ctx, deployment.ResourceGroupName, deployment.ResourceName, deployment.Name, nil)
		}
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Remove):", err)
			return err
		}
	}

	return nil
}

// NeedsUpdate checks for any updates for the extended location
func (i *ExtendedLocationTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	sLog.Infof("  P (Extended Location Target Provider): NeedsUpdate: %d - %d", len(desired), len(current))
	locationProperty := []string{resourceNameKey, resourceGroupNameKey}
	for _, dc := range desired {
		for _, cc := range current {
			for _, param := range locationProperty {
				if cc.Properties[param] == dc.Properties[param] {
					sLog.Info("  P (Extended Location Target Provider): NeedsUpdate: returning true")
					return true
				}
			}
		}
	}

	sLog.Info("  P (Extended Location Target Provider): NeedsUpdate: returning false")
	return false
}

// NeedsRemove checks if the solution component needs to be removed
func (i *ExtendedLocationTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	sLog.Infof("  P (Extended Location Target Provider): NeedsRemove: %d - %d", len(desired), len(current))
	locationProperty := []string{resourceNameKey, resourceGroupNameKey}
	for _, dc := range desired {
		for _, cc := range current {
			for _, param := range locationProperty {
				if cc.Properties[param] == dc.Properties[param] {
					sLog.Info("  P (Extended Location Target Provider): NeedsRemove: returning false")
					return false
				}
			}
		}
	}

	sLog.Info("  P (Extended Location Target Provider): NeedsRemove: returning true")
	return true
}

// Apply creates the extended location on the ARC enabled cluster
func (i *ExtendedLocationTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan(
		"Extended Location Provider",
		ctx, &map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, err)

	// opts sets system assigned managed identity
	opts := azidentity.ManagedIdentityCredentialOptions{
		ID: azidentity.ClientID(i.Config.ClientID),
	}

	// cred gets the managed identity credential
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		return err
	}

	components := deployment.GetComponentSlice()
	for _, c := range components {
		// deployment get extended location property from component
		deployment, err := getDeploymentFromComponent(c)
		if err != nil {
			sLog.Errorf(" P (Extended Location Custom Location Target):", err)
			return err
		}

		// clientFactory gets a new client for Azure API
		clientFactory, err := armextendedlocation.NewClientFactory(deployment.SubscriptionID, cred, nil)
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Subscription ID):", err)
			return err
		}

		switch c.Type {
		case "custom-location":
			// customLocation sets the custom location property object
			customLocation := armextendedlocation.CustomLocation{
				Location: &deployment.Location,
				Properties: &armextendedlocation.CustomLocationProperties{
					ClusterExtensionIDs: []*string{&deployment.ClusterExtensionID},
					DisplayName:         &deployment.Name,
					HostResourceID:      &deployment.HostResourceID,
					Namespace:           &deployment.Namespace,
				},
			}

			_, err = clientFactory.NewCustomLocationsClient().BeginCreateOrUpdate(ctx, deployment.ResourceGroupName, deployment.ResourceName, customLocation, nil)

		case "resource-sync-rule":
			//resourceSyncRule sets the resource sync rule property object
			resourceSyncRule := armextendedlocation.ResourceSyncRule{
				Location: &deployment.Location,
				Properties: &armextendedlocation.ResourceSyncRuleProperties{
					Selector: &armextendedlocation.ResourceSyncRulePropertiesSelector{
						MatchExpressions: []*armextendedlocation.MatchExpressionsProperties{
							{
								Key:      &deployment.MatchExpressionKey,
								Operator: &deployment.MatchExpressionOperator,
								Values: []*string{
									&deployment.MatchExpressionValue},
							}},
						MatchLabels: map[string]*string{
							"key1": &deployment.MatchLabelKey,
						},
					},
					TargetResourceGroup: &deployment.TargetResourceGroup,
				},
			}

			_, err = clientFactory.NewResourceSyncRulesClient().BeginCreateOrUpdate(ctx, deployment.ResourceGroupName, deployment.ResourceName, deployment.Name, resourceSyncRule, nil)
		}
		if err != nil {
			sLog.Errorf(" P (Extended Location Target Deployment):", err)
			return err
		}
	}

	return nil
}

// getDeploymentFromComponent gets the extended location properties from the component
func getDeploymentFromComponent(component model.ComponentSpec) (ExtendedLocationProperty, error) {
	ret := ExtendedLocationProperty{}
	if component.Name == "" {
		return ret, errors.New("Extended Location Name is not set")
	}

	if component.Type == "" {
		return ret, errors.New("Extended Location Type is not set")
	}

	if component.Properties == nil {
		return ret, errors.New("ExtendedLocationProperty is null")
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
