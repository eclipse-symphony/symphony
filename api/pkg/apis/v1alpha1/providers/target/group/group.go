/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package group

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	log                      = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	loggerName   string = "providers.target.group"
	providerName        = "P (Group Target)"
	group               = "group"
)

type GroupTargetProviderConfig struct {
	User           string               `json:"user"`
	Password       string               `json:"password"`
	TargetSelector model.TargetSelector `json:"targetSelector"`
}

type GroupTargetProvider struct {
	Config    GroupTargetProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func GroupTargetProviderConfigFromMap(properties map[string]string) (GroupTargetProviderConfig, error) {
	ret := GroupTargetProviderConfig{}
	if api_utils.ShouldUseUserCreds() {
		user, err := api_utils.GetString(properties, "user")
		if err != nil {
			return ret, err
		}
		ret.User = user
		if ret.User == "" && !api_utils.ShouldUseSATokens() {
			return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := api_utils.GetString(properties, "password")
		ret.Password = password
		if err != nil {
			return ret, err
		}
	}
	var targetSelector model.TargetSelector
	if _, ok := properties["targetSelector"]; !ok {
		return ret, v1alpha2.NewCOAError(nil, "targetSelector is required", v1alpha2.BadConfig)
	}
	err := json.Unmarshal([]byte(properties["targetSelector"]), &targetSelector)
	if err != nil {
		return ret, v1alpha2.NewCOAError(err, fmt.Sprintf("expected targetSelector: %+v", err), v1alpha2.BadConfig)
	}
	if targetSelector.Name == "" && len(targetSelector.PropertySelector) == 0 && len(targetSelector.LabelSelector) == 0 {
		return ret, v1alpha2.NewCOAError(nil, "targetSelector must have either name or propertySelector", v1alpha2.BadConfig)
	}
	ret.TargetSelector = targetSelector

	return ret, nil
}
func (i *GroupTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := GroupTargetProviderConfigFromMap(properties)
	if err != nil {
		log.Errorf("  P (Group Target): expected GroupProviderConfig: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("expected GroupProviderConfig: %+v", err), v1alpha2.InitFailed)
	}
	return i.Init(config)
}
func (s *GroupTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func (i *GroupTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"Group Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfoCtx(ctx, "  P (Group Target): Init()")
	providerConfig, err := toGroupTargetProviderConfig(config)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): expected GroupProviderConfig - %+v", err)
		return err
	}
	i.Config = providerConfig
	i.ApiClient, err = api_utils.GetApiClient()
	if err != nil {
		return err
	}
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Group Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
}

func (i *GroupTargetProvider) Get(ctx context.Context, reference model.TargetProviderGetReference) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Group Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (Group Target): getting artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	targets, err := i.matchTargets(ctx, reference.Deployment.Instance.Spec.Scope, i.Config.TargetSelector, true)

	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get targets: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get targets", providerName), v1alpha2.GroupActionFailed)
		return nil, err
	}

	ret := make([]model.ComponentSpec, 0)

	for _, target := range targets {
		for k, prop := range target.Status.Properties {
			if strings.HasPrefix(k, "component:") {
				var component model.ComponentSpec
				err = json.Unmarshal([]byte(prop), &component)
				if err != nil {
					log.ErrorfCtx(ctx, "  P (Group Target): failed to unmarshal component %+v: %s", err, prop)
					continue
				}
				ret = append(ret, component)
			}
		}
	}

	return ret, nil
}
func (i *GroupTargetProvider) matchTargets(ctx context.Context, namespace string, targetSelector model.TargetSelector, checkState bool) ([]model.TargetState, error) {
	matches := []model.TargetState{}
	targets, err := i.ApiClient.GetTargets(ctx, namespace, i.Config.User, i.Config.Password)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get targets: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get targets", providerName), v1alpha2.GroupActionFailed)
		return nil, err
	}
	for _, target := range targets {
		if api_utils.IsTargetMatch(target, targetSelector, checkState) {
			matches = append(matches, target)
		}
	}
	return matches, nil
}

func (i *GroupTargetProvider) Apply(ctx context.Context, reference model.TargetProviderApplyReference) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"Group Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (Group Target): applying artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	functionName := observ_utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		group,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)

	components := reference.Step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to validate components: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			group,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)

		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}

	if reference.IsDryRun {
		log.DebugCtx(ctx, "  P (Group Target): dryRun is enabled, skipping apply")
		return nil, nil
	}

	ret := reference.Step.PrepareResultMap()

	targets, err := i.matchTargets(ctx, reference.Deployment.Instance.Spec.Scope, i.Config.TargetSelector, true)

	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get targets: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get targets", providerName), v1alpha2.GroupActionFailed)
		return nil, err
	}

	updatedComponents := reference.Step.GetUpdatedComponents()
	assignments, err := i.assignComponents(updatedComponents, targets)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to assign components: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			group,
			functionName,
			metrics.ApplyOperation,
			metrics.ApplyOperationType,
			v1alpha2.GroupActionFailed.String(),
		)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to assign components", providerName), v1alpha2.GroupActionFailed)
		return nil, err
	}

	for k, v := range assignments {
		for _, target := range targets {
			if target.ObjectMeta.Name == k {
				// Assign the components to the target
				for _, componentName := range v {
					for _, component := range updatedComponents {
						if component.Name == componentName {
							found := false
							for _, existingComponent := range target.Spec.Components {
								if existingComponent.Name == component.Name {
									found = true
									break
								}
							}
							if !found {
								target.Spec.Components = append(target.Spec.Components, component)
							}
						}
					}
				}
				log.InfofCtx(ctx, "  P (Group Target): assigned %d components to target %s", i, target.ObjectMeta.Name)
			}

			if len(target.Spec.Components) > 1 {
				for i, component := range target.Spec.Components {
					if strings.HasPrefix(component.Name, "probe-") {
						target.Spec.Components = append(target.Spec.Components[:i], target.Spec.Components[i+1:]...)
					}
				}
			}

			targetData, _ := json.Marshal(target)
			err = i.ApiClient.CreateTarget(ctx, target.ObjectMeta.Name, targetData, target.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Group Target): failed to update target %s: %+v", target.ObjectMeta.Name, err)
				ret[target.ObjectMeta.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: fmt.Sprintf("failed to update target %s: %v", target.ObjectMeta.Name, err),
				}
				providerOperationMetrics.ProviderOperationErrors(
					group,
					functionName,
					metrics.ApplyOperation,
					metrics.ApplyOperationType,
					v1alpha2.UpdateFailed.String(),
				)
				return ret, err
			}
			log.InfofCtx(ctx, "  P (Group Target): target %s has been updated", target.ObjectMeta.Name)
			ret[target.ObjectMeta.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: fmt.Sprintf("No error. %s has been updated", target.ObjectMeta.Name),
			}
		}
	}

	for _, target := range targets {
		if len(target.Spec.Components) == 0 {
			dummySpec := model.ComponentSpec{
				Name: fmt.Sprintf("probe-%s", target.ObjectMeta.Name),
				Type: "container",
				Metadata: map[string]string{
					"probe": "true",
				},
				Properties: map[string]interface{}{},
			}
			for _, component := range updatedComponents {
				dummySpec.Properties[component.Name] = ""
			}
			target.Spec.Components = append(target.Spec.Components, dummySpec)
			targetData, _ := json.Marshal(target)
			err = i.ApiClient.CreateTarget(ctx, target.ObjectMeta.Name, targetData, target.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Group Target): failed to update target %s: %+v", target.ObjectMeta.Name, err)
				ret[target.ObjectMeta.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: fmt.Sprintf("failed to update target %s: %v", target.ObjectMeta.Name, err),
				}
				providerOperationMetrics.ProviderOperationErrors(
					group,
					functionName,
					metrics.ApplyOperation,
					metrics.ApplyOperationType,
					v1alpha2.UpdateFailed.String(),
				)
				return ret, err
			}
		}
	}
	return ret, nil
}

func (i *GroupTargetProvider) assignComponents(components []model.ComponentSpec, targets []model.TargetState) (map[string][]string, error) {
	assignments := make(map[string][]string)
	for _, target := range targets {
		assignments[target.ObjectMeta.Name] = []string{}
	}
	for _, component := range components {
		sort.SliceStable(targets, func(i, j int) bool {
			return len(assignments[targets[i].ObjectMeta.Name]) < len(assignments[targets[j].ObjectMeta.Name])
		})
		assigned := false
		for _, target := range targets {
			if target.Status.Properties != nil && target.Status.Properties["component:"+component.Name] != "" {
				assignments[target.ObjectMeta.Name] = append(assignments[target.ObjectMeta.Name], component.Name)
				assigned = true
				break
			}
		}
		if !assigned {
			for _, target := range targets {
				for _, component := range target.Spec.Components {
					if component.Name == component.Name {
						assignments[target.ObjectMeta.Name] = append(assignments[target.ObjectMeta.Name], component.Name)
						assigned = true
						break
					}
				}
				if assigned {
					break
				}
			}
		}
		if !assigned {
			for _, target := range targets {
				if len(target.Spec.Components) != 0 {
					continue // TODO: specific rule: skip targets that already have a component assigned. genearlize/externalize this
				}
				hasComponent := false
				for k, _ := range target.Status.Properties {
					if strings.HasPrefix(k, "component:") {
						hasComponent = true //specific rule: skip targets that already have a component assigned. generalize/externalize this
						break
					}
				}
				if !hasComponent {
					//check anti-affinity rule
					assignments[target.ObjectMeta.Name] = append(assignments[target.ObjectMeta.Name], component.Name)
					assigned = true
					break
				}
			}
		}
		if !assigned {
			return nil, fmt.Errorf("no target available for component %s", component.Name)
		}
	}
	return assignments, nil
}

func (*GroupTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
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
					Name: "*", //react to all property changes
				},
			},
		},
	}
}

func toGroupTargetProviderConfig(config providers.IProviderConfig) (GroupTargetProviderConfig, error) {
	ret := GroupTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}
