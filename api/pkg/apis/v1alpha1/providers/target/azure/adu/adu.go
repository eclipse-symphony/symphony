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

package adu

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	azureutils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/cloudutils/azure"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

var sLog = logger.NewLogger("coa.runtime")

type ADUTargetProviderConfig struct {
	Name               string `json:"name"`
	TenantId           string `json:"tenantId"`
	ClientId           string `json:"clientId"`
	ClientSecret       string `json:"clientSecret"`
	ADUAccountEndpoint string `json:"aduAccountEndpoint"`
	ADUAccountInstance string `json:"aduAccountInstance"`
	ADUGroup           string `json:"aduGroup"`
}

type ADUTargetProvider struct {
	Config  ADUTargetProviderConfig
	Context *contexts.ManagerContext
}

func ADUTargetProviderConfigFromMap(properties map[string]string) (ADUTargetProviderConfig, error) {
	ret := ADUTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["tenantId"]; ok {
		ret.TenantId = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update provider tenant id is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["clientId"]; ok {
		ret.ClientId = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update provider client id is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["clientSecret"]; ok {
		ret.ClientSecret = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update provider client secret is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["aduAccountEndpoint"]; ok {
		ret.ADUAccountEndpoint = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update account endpoint is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["aduAccountInstance"]; ok {
		ret.ADUAccountInstance = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update account instance is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["aduGroup"]; ok {
		ret.ADUGroup = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "ADU update group is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (i *ADUTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := ADUTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *ADUTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *ADUTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("ADU Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("~~~ ADU Target Provider ~~~ : Init()")

	updateConfig, err := toADUTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ ADU Target Provider ~~~ : expected ADUTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func toADUTargetProviderConfig(config providers.IProviderConfig) (ADUTargetProviderConfig, error) {
	ret := ADUTargetProviderConfig{}
	if config == nil {
		return ret, errors.New("ADUTargetProviderConfig is null")
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

func (i *ADUTargetProvider) Get(ctx context.Context, dep model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("ADU Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Info("~~~ ADU Update Provider ~~~ : getting components")
	deployment, err := i.getDeployment()
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ ADU Target Provider ~~~ : %+v", err)
		return nil, err
	}

	ret := []model.ComponentSpec{}

	if deployment.DeploymentId != "" {
		ret = append(ret, model.ComponentSpec{
			Name: deployment.UpdateId.Name,
			Properties: map[string]interface{}{
				"update.name":     deployment.UpdateId.Name,
				"update.provider": deployment.UpdateId.Provider,
				"update.version":  deployment.UpdateId.Version,
			},
		})
	}

	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func getDeploymentFromComponent(c model.ComponentSpec) (azureutils.ADUDeployment, error) {
	provider := ""
	version := ""
	name := ""
	ok := false
	deployment := azureutils.ADUDeployment{}
	if provider, ok = c.Properties["update.provider"].(string); !ok {
		return deployment, errors.New("component doesn't contain a update.provider property")
	}
	if version, ok = c.Properties["update.version"].(string); !ok {
		return deployment, errors.New("component doesn't contain a update.version property")
	}
	if name, ok = c.Properties["update.name"].(string); !ok {
		return deployment, errors.New("component doesn't contain a update.name property")
	}
	deployment.DeploymentId = uuid.New().String()
	deployment.StartDateTime = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	deployment.UpdateId = azureutils.UpdateId{
		Name:     name,
		Provider: provider,
		Version:  version,
	}
	return deployment, nil
}

func (i *ADUTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	_, span := observability.StartSpan("ADU Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Info("  P (ADU Update): applying components")

	components := step.GetComponents()
	err := i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return nil, err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil, nil
	}

	ret := step.PrepareResultMap()

	for _, c := range step.Components {
		deployment, err := getDeploymentFromComponent(c.Component)
		if err != nil {
			ret[c.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.ValidateFailed,
				Message: err.Error(),
			}
			observ_utils.CloseSpanWithError(span, err)
			return ret, err
		}
		if c.Action == "update" {
			deployment.GroupId = i.Config.ADUGroup
			err = i.applyDeployment(deployment)
			if err != nil {
				ret[c.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				observ_utils.CloseSpanWithError(span, err)
				sLog.Errorf("  P (ADU Update): %+v", err)
				return ret, err
			}
		} else {
			err = i.deleteDeploymeent(deployment)
			if err != nil {
				ret[c.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.DeleteFailed,
					Message: err.Error(),
				}
				observ_utils.CloseSpanWithError(span, nil)
				return ret, nil //TODO: are we ignoring errors on purpose here?
			}
		}

	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (i *ADUTargetProvider) getDeployment() (azureutils.ADUDeployment, error) {
	ret := azureutils.ADUDeployment{}
	token, err := azureutils.GetAzureToken(i.Config.TenantId, i.Config.ClientId, i.Config.ClientSecret, "https://api.adu.microsoft.com/.default")
	if err != nil {
		return ret, err
	}
	group, err := azureutils.GetADUGroup(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup)
	if err != nil {
		return ret, err
	}
	if group.DeploymentId == "" {
		return ret, nil
	}
	deployment, err := azureutils.GetADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, group.DeploymentId)
	if err != nil {
		return ret, err
	}
	return deployment, nil
}
func (i *ADUTargetProvider) deleteDeploymeent(deployment azureutils.ADUDeployment) error {
	token, err := azureutils.GetAzureToken(i.Config.TenantId, i.Config.ClientId, i.Config.ClientSecret, "https://api.adu.microsoft.com/.default")
	if err != nil {
		return err
	}
	existing, err := i.getDeployment()
	if err != nil {
		return nil //Can't read existing deployment, ignore
	}
	if existing.UpdateId.Version == deployment.UpdateId.Version && existing.UpdateId.Name == deployment.UpdateId.Name && existing.UpdateId.Provider == deployment.UpdateId.Provider {
		return azureutils.DeleteADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, existing.DeploymentId)
	}
	return nil
}
func (i *ADUTargetProvider) applyDeployment(deployment azureutils.ADUDeployment) error {
	token, err := azureutils.GetAzureToken(i.Config.TenantId, i.Config.ClientId, i.Config.ClientSecret, "https://api.adu.microsoft.com/.default")
	if err != nil {
		return err
	}
	group, err := azureutils.GetADUGroup(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup)
	if err != nil {
		return err
	}
	if group.DeploymentId == "" {
		err = azureutils.CreateADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, deployment.DeploymentId, deployment)
		if err != nil {
			return err
		}
	} else {
		existing, err := azureutils.GetADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, group.DeploymentId)
		if err != nil {
			return err
		}
		if existing.UpdateId.Version != deployment.UpdateId.Version || existing.UpdateId.Name != deployment.UpdateId.Name || existing.UpdateId.Provider != deployment.UpdateId.Provider {
			err = azureutils.CreateADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, deployment.DeploymentId, deployment)
			if err != nil {
				return err
			}
		} else {
			if deployment.IsCanceled {
				deployment.DeploymentId = existing.DeploymentId
				err = azureutils.RetryADUDeployment(token, i.Config.ADUAccountEndpoint, i.Config.ADUAccountInstance, i.Config.ADUGroup, deployment.DeploymentId, deployment)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (*ADUTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{"update.provider", "update.name", "update.version"},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
