/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package staging

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const loggerName = "providers.target.staging"

var sLog = logger.NewLogger(loggerName)

type StagingTargetProviderConfig struct {
	Name       string `json:"name"`
	TargetName string `json:"targetName"`
}

type StagingTargetProvider struct {
	Config    StagingTargetProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
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
	_, span := observability.StartSpan("Staging Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Info("  P (Staging Target): Init()")

	updateConfig, err := toStagingTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Staging Target): expected StagingTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig
	i.ApiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
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
	ctx, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Staging Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name, span.SpanContext().TraceID().String())

	var err error
	defer observ_utils.CloseSpanWithError(span, &err)

	scope := deployment.Instance.Spec.Scope
	if scope == "" {
		scope = "default"
	}
	catalog, err := i.ApiClient.GetCatalog(
		ctx,
		deployment.Instance.ObjectMeta.Name+"-"+i.Config.TargetName,
		scope,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password)

	if err != nil {
		if v1alpha2.IsNotFound(err) {
			sLog.Infof("  P (Staging Target): no staged artifact found: %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return nil, nil
		}
		sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	if spec, ok := catalog.Spec.Properties["reported"]; ok {
		var components []model.ComponentSpec
		jData, _ := json.Marshal(spec)
		err = json.Unmarshal(jData, &components)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v, traceId: %s", err, span.SpanContext().TraceID().String())
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
	return nil, nil
}
func (i *StagingTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Staging Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("  P (Staging Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name, span.SpanContext().TraceID().String())

	var err error
	defer observ_utils.CloseSpanWithError(span, &err)

	err = i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components	TODO: is this right?
	if err != nil {
		sLog.Errorf("  P (Staging Target): failed to validate components: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		sLog.Infof("  P (Staging Target): dry run, skipping apply")
		return nil, nil
	}
	ret := step.PrepareResultMap()

	scope := deployment.Instance.Spec.Scope
	if scope == "" {
		scope = "default"
	}

	var catalog model.CatalogState

	catalog, err = i.ApiClient.GetCatalog(
		ctx,
		deployment.Instance.ObjectMeta.Name+"-"+i.Config.TargetName,
		scope,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password)

	if err != nil && !v1alpha2.IsNotFound(err) {
		sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return ret, err
	}

	if catalog.Spec == nil {
		catalog.ObjectMeta.Name = deployment.Instance.ObjectMeta.Name + "-" + i.Config.TargetName
		catalog.Spec = &model.CatalogSpec{
			Type: "staged",
		}
	}
	if catalog.Spec.Properties == nil {
		catalog.Spec.Properties = make(map[string]interface{})
	}
	if catalog.Spec.Metadata == nil {
		catalog.Spec.Metadata = make(map[string]string)
	}
	if catalog.ObjectMeta.Annotations == nil {
		catalog.ObjectMeta.Annotations = make(map[string]string)
	}

	if catalog.ObjectMeta.Labels == nil {
		catalog.ObjectMeta.Labels = make(map[string]string)
	}
	catalog.ObjectMeta.Labels["staged_target"] = i.Config.TargetName

	var existing []model.ComponentSpec
	if v, ok := catalog.Spec.Properties["components"]; ok {
		jData, _ := json.Marshal(v)
		err = json.Unmarshal(jData, &existing)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to unmarshall catalog components: %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}

	var deleted []model.ComponentSpec
	if v, ok := catalog.Spec.Properties["removed-components"]; ok {
		jData, _ := json.Marshal(v)
		err = json.Unmarshal(jData, &deleted)
		if err != nil {
			sLog.Errorf("  P (Staging Target): failed to get staged artifact: %v, traceId: %s", err, span.SpanContext().TraceID().String())
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
			for j, c := range deleted {
				if c.Name == component.Name {
					deleted = append(deleted[:j], deleted[j+1:]...)
				}
			}
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

	catalog.Spec.Properties["deployment"] = deployment
	catalog.Spec.Properties["staged"] = map[string]interface{}{
		"components":         existing,
		"removed-components": deleted,
	}
	jData, _ := json.Marshal(catalog)
	err = i.ApiClient.UpsertCatalog(
		ctx,
		deployment.Instance.ObjectMeta.Name+"-"+i.Config.TargetName,
		jData,
		i.Context.SiteInfo.CurrentSite.Username,
		i.Context.SiteInfo.CurrentSite.Password)
	if err != nil {
		sLog.Errorf("  P (Staging Target): failed to upsert staged artifact: %v, traceId: %s", err, span.SpanContext().TraceID().String())
	}
	return ret, err
}

func (*StagingTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{
					Name: "*",
				},
			},
		},
	}
}
