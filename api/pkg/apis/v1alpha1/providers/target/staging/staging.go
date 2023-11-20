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

package staging

import (
	"context"
	"encoding/json"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type StagingTargetProviderConfig struct {
	Name       string `json:"name"`
	TargetName string `json:"targetName"`
}

type StagingTargetProvider struct {
	Config  StagingTargetProviderConfig
	Context *contexts.ManagerContext
}

func StagingProviderConfigFromMap(properties map[string]string) (StagingTargetProviderConfig, error) {
	ret := StagingTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["targetName"]; ok {
		ret.TargetName = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid staging provider config, exptected 'targetName'", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (i *StagingTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := StagingProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *StagingTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *StagingTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Staging Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("  P (Staging Target): Init()")

	updateConfig, err := toStagingTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Staging Target): expected StagingTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toStagingTargetProviderConfig(config providers.IProviderConfig) (StagingTargetProviderConfig, error) {
	ret := StagingTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *StagingTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Staging Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	var err error
	defer observ_utils.CloseSpanWithError(span, err)

	scope := deployment.Instance.Scope
	if scope == "" {
		scope = "default"
	}
	catalog, err := utils.GetCatalog(
		ctx,
		i.Context.SiteInfo.CurrentSite.BaseUrl,
		deployment.Instance.Name+"-"+i.Config.TargetName,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password)

	if err != nil {
		if v1alpha2.IsNotFound(err) {
			sLog.Infof("  P (Staging Target): no staged artifact found")
			return nil, nil
		}
		sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
		return nil, err
	}

	if spec, ok := catalog.Spec.Properties["components"]; ok {
		var components []model.ComponentSpec
		jData, _ := json.Marshal(spec)
		err := json.Unmarshal(jData, &components)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
			return nil, err
		}
		ret := make([]model.ComponentSpec, len(references))
		for i, reference := range references {
			for _, component := range components {
				if component.Name == reference.Component.Name {
					ret[i] = component
					break
				}
			}
		}
		return ret, nil
	}
	err = v1alpha2.NewCOAError(nil, "staged artifact is not found as a 'spec' property", v1alpha2.NotFound)
	sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
	return nil, err
}
func (i *StagingTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Staging Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	var err error
	defer observ_utils.CloseSpanWithError(span, err)

	err = i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components	TODO: is this right?
	if err != nil {
		sLog.Errorf("  P (Staging Target): failed to validate components: %v", err)
		return nil, err
	}
	if isDryRun {
		sLog.Infof("  P (Staging Target): dry run, skipping apply")
		return nil, nil
	}
	ret := step.PrepareResultMap()

	scope := deployment.Instance.Scope
	if scope == "" {
		scope = "default"
	}

	var catalog model.CatalogState

	catalog, err = utils.GetCatalog(
		ctx,
		i.Context.SiteInfo.CurrentSite.BaseUrl,
		deployment.Instance.Name+"-"+i.Config.TargetName,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password)
	if err != nil && !v1alpha2.IsNotFound(err) {
		sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
		return ret, err
	}

	if catalog.Spec == nil {
		catalog.Id = deployment.Instance.Name + "-" + i.Config.TargetName
		catalog.Spec = &model.CatalogSpec{
			SiteId: i.Context.SiteInfo.SiteId,
			Type:   "staged",
			Name:   catalog.Id,
		}
	}
	if catalog.Spec.Properties == nil {
		catalog.Spec.Properties = make(map[string]interface{})
	}

	var existing []model.ComponentSpec
	if v, ok := catalog.Spec.Properties["components"]; ok {
		jData, _ := json.Marshal(v)
		err := json.Unmarshal(jData, &existing)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
			return ret, err
		}
	}

	components := step.GetUpdatedComponents()
	if len(components) > 0 {
		for i, component := range components {
			found := false
			for j, c := range existing {
				if c.Name == component.Name {
					found = true
					existing[j] = components[i]
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.Updated,
						Message: "",
					}
					break
				}
			}
			if !found {
				existing = append(existing, component)
			}
		}
	}

	var deleted []model.ComponentSpec
	if v, ok := catalog.Spec.Properties["removed-components"]; ok {
		jData, _ := json.Marshal(v)
		err := json.Unmarshal(jData, &deleted)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v", err)
			return ret, err
		}
	}

	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for i, component := range components {
			found := false
			for j, c := range deleted {
				if c.Name == component.Name {
					found = true
					deleted[j] = components[i]
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.Updated,
						Message: "",
					}
					break
				}
			}
			if !found {
				deleted = append(deleted, component)
			}
		}
	}

	catalog.Spec.Properties["components"] = existing
	catalog.Spec.Properties["removed-components"] = deleted
	jData, _ := json.Marshal(catalog.Spec)
	err = utils.UpsertCatalog(
		ctx,
		i.Context.SiteInfo.CurrentSite.BaseUrl,
		deployment.Instance.Name+"-"+i.Config.TargetName,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password, jData)
	if err != nil {
		sLog.Errorf("  P (Staging Target): failed to upsert staged artifact: %v", err)
	}
	return ret, err
}

func (*StagingTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
