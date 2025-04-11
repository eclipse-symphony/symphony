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
	"math"
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
	User     string `json:"user"`
	Password string `json:"password"`
}

type GroupTargetProvider struct {
	Config    GroupTargetProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

type GroupPatchAction struct {
	SparePatch  map[string]string `json:"sparePatch"`
	TargetPatch map[string]string `json:"targetPatch"`
}
type TargetGroupProperty struct {
	TargetPropertySelector map[string]string     `json:"targetPropertySelector"`
	TargetStateSelector    map[string]string     `json:"targetStateSelector"`
	SparePropertySelector  map[string]string     `json:"sparePropertySelector"`
	SpareStateSelector     map[string]string     `json:"spareStateSelector"`
	MinMatchCount          int                   `json:"minMatchCount"`
	MaxMatchCount          int                   `json:"maxMatchCount"`
	LowMatchAction         GroupPatchAction      `json:"lowMatchAction"`
	HighMatchAction        GroupPatchAction      `json:"highMatchAction"`
	SpareComponents        []model.ComponentSpec `json:"spareComponents"`
	MemberComponents       []model.ComponentSpec `json:"memberComponents"`
}

func getTargetGroupPropertyFromComponent(component model.ComponentSpec) (*TargetGroupProperty, error) {
	ret := TargetGroupProperty{}
	data, err := json.Marshal(component.Properties)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
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

func (i *GroupTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"Group Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	log.InfofCtx(ctx, "  P (Group Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	namespace := deployment.Instance.Spec.Scope
	for _, component := range references {

		lowDeficite, highDeficite, _, _, err := i.matchTriggers(ctx, namespace, component.Component)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (Group Target): failed to match triggers: %+v", err)
			return nil, err
		}
		if lowDeficite >= 0 || highDeficite >= 0 {
			// Nothing to do, return the original compontn so no changes will be detected
			ret = append(ret, component.Component)
		}
	}

	return ret, nil
}
func (i *GroupTargetProvider) matchTargets(ctx context.Context, namespace string, propertySelector map[string]string, stateSelector map[string]string, matchState bool) ([]model.TargetState, error) {
	matches := []model.TargetState{}
	targets, err := i.ApiClient.GetTargets(ctx, namespace, i.Config.User, i.Config.Password)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get targets: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get targets", providerName), v1alpha2.GroupActionFailed)
		return nil, err
	}
	for _, target := range targets {
		if isTargetMatch(target, propertySelector) {
			if !matchState && isTargetStateMatch(target, stateSelector) {
				matches = append(matches, target)
			}
		}
	}
	return matches, nil
}
func (i *GroupTargetProvider) matchTriggers(ctx context.Context, namespace string, component model.ComponentSpec) (int, int, []model.TargetState, TargetGroupProperty, error) {
	failedStateMatch := []model.TargetState{}
	var groupProp *TargetGroupProperty
	groupProp, err := getTargetGroupPropertyFromComponent(component)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get group properties: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get group properties", providerName), v1alpha2.GroupActionFailed)
		return 0, 0, failedStateMatch, *groupProp, err
	}
	targets, err := i.matchTargets(ctx, namespace, groupProp.TargetPropertySelector, groupProp.TargetStateSelector, false)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get targets: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to get targets", providerName), v1alpha2.GroupActionFailed)
		return 0, 0, failedStateMatch, *groupProp, err
	}
	activeCount := 0

	for _, target := range targets {
		if isTargetStateMatch(target, groupProp.TargetStateSelector) {
			activeCount += 1
		} else {
			failedStateMatch = append(failedStateMatch, target)
		}
	}
	return activeCount - groupProp.MinMatchCount, groupProp.MaxMatchCount - activeCount, failedStateMatch, *groupProp, nil
}
func (i *GroupTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
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
	log.InfofCtx(ctx, "  P (Group Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := observ_utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		group,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)

	components := step.GetComponents()
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

	if isDryRun {
		log.DebugCtx(ctx, "  P (Group Target): dryRun is enabled, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()

	for _, component := range step.Components {
		if component.Action == model.ComponentUpdate {
			overLowMark, _, failedStateTargets, groupProp, err := i.matchTriggers(ctx, deployment.Instance.Spec.Scope, component.Component)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Group Target): failed to match triggers: %+v", err)
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				providerOperationMetrics.ProviderOperationErrors(
					group,
					functionName,
					metrics.ApplyOperation,
					metrics.ApplyOperationType,
					v1alpha2.HelmChartPullFailed.String(),
				)
				return ret, err
			}
			if overLowMark < 0 {
				log.InfofCtx(ctx, "  P (Group Target): low trigger fired")
				err = i.applyLowTrigger(ctx, deployment.Instance.Spec.Scope, groupProp, overLowMark, failedStateTargets)
				if err != nil {
					log.ErrorfCtx(ctx, "  P (Group Target): failed to apply low trigger: %+v", err)
					ret[component.Component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					providerOperationMetrics.ProviderOperationErrors(
						group,
						functionName,
						metrics.ApplyOperation,
						metrics.ApplyOperationType,
						v1alpha2.HelmChartPullFailed.String(),
					)
					return ret, err
				}
			}
			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: fmt.Sprintf("No error. %s has been updated", component.Component.Name),
			}
		}
	}
	return ret, nil
}
func (i *GroupTargetProvider) patchTargetProperty(target model.TargetState, patch map[string]string) (model.TargetState, error) {
	for k, v := range patch {
		if v == "~REMOVE" {
			delete(target.Spec.Properties, k)
			continue
		}
		target.Spec.Properties[k] = v
	}
	return target, nil
}
func (i *GroupTargetProvider) applyLowTrigger(ctx context.Context, namespace string, groupProperty TargetGroupProperty, deficate int, failedTargets []model.TargetState) error {
	spares, err := i.matchTargets(ctx, namespace, groupProperty.SparePropertySelector, groupProperty.SpareStateSelector, true)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Group Target): failed to get spares: %+v", err)
		return err
	}
	sparsNeeded := int(math.Abs(float64(deficate)))
	if len(spares) < sparsNeeded {
		log.ErrorfCtx(ctx, "  P (Group Target): not enough spares: %d", len(spares))
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("not enough spares: %d", len(spares)), v1alpha2.GroupActionFailed)
	}
	spareCount := 0
	for _, spare := range spares {
		spare, err = i.patchTargetProperty(spare, groupProperty.LowMatchAction.SparePatch)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (Group Target): failed to patch target: %+v", err)
			return err
		}
		objectData, _ := json.Marshal(spare)
		err = i.ApiClient.CreateTarget(ctx, spare.ObjectMeta.Name, objectData, spare.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (Group Target): failed to update target: %+v", err)
			return err
		}
		spareCount++
		if spareCount >= sparsNeeded {
			break
		}
	}
	for _, target := range failedTargets {
		target, err = i.patchTargetProperty(target, groupProperty.LowMatchAction.TargetPatch)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (Group Target): failed to patch target: %+v", err)
			return err
		}
		objectData, _ := json.Marshal(target)
		err = i.ApiClient.CreateTarget(ctx, target.ObjectMeta.Name, objectData, target.ObjectMeta.Namespace, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (Group Target): failed to update target: %+v", err)
			return err
		}
	}
	return nil
}

func (*GroupTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{"targetSelector", "targetState", "spareSelector", "spareState", "minMatchCount", "lowMatchAction"},
			OptionalProperties:    []string{"maxMatchCount"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "targetSelector", PropChanged: mapMatch},
				{Name: "targetState", PropChanged: mapMatch},
				{Name: "spareSelector", PropChanged: mapMatch},
				{Name: "spareState", PropChanged: mapMatch},
				{Name: "minMatchCount"},
				//TODO: compare lowMatchAction as well
			},
		},
	}
}

func mapMatch(a interface{}, b interface{}) bool {
	aMap, aOk := toMap(a)
	bMap, bOK := toMap(b)
	if !aOk || !bOK || aMap == nil || bMap == nil {
		return false
	}
	if len(aMap) != len(bMap) {
		return false
	}
	for k, v := range aMap {
		if vb, ok := bMap[k]; !ok || (vb != v) {
			return false
		}
	}
	return true
}
func toMap(a interface{}) (map[string]string, bool) {
	if a == nil {
		return nil, false
	}
	valueMap, ok := a.(map[string]string)
	return valueMap, ok
}

func isTargetMatch(target model.TargetState, propSelector map[string]string) bool {
	for k, v := range propSelector {
		if tv, ok := target.Spec.Properties[k]; !ok || (v != tv) {
			return false
		}
	}
	return true
}
func isTargetStateMatch(target model.TargetState, propSelector map[string]string) bool {
	for k, v := range propSelector {
		if tv, ok := target.Status.Properties[k]; !ok || (v != tv) {
			return false
		}
	}
	return true
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
