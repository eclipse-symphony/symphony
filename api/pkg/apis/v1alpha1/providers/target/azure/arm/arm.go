/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package arm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/contexts"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

const (
	radius       = "arm"
	providerName = "P (ARM Target)"
	loggerName   = "providers.target.azure.arm"
)

type (
	ArmTargetProviderConfig struct {
	}
	ArmTargetProvider struct {
		Config  ArmTargetProviderConfig
		Context *contexts.ManagerContext
	}
)

func RadiusTargetProviderConfigFromMap(properties map[string]string) (ArmTargetProviderConfig, error) {
	return ArmTargetProviderConfig{}, nil
}

func (r *ArmTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := RadiusTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (ARM Target): expected ArmTargetProviderConfig: %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init", providerName), v1alpha2.InitFailed)
	}

	return r.Init(config)
}

func (s *ArmTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (r *ArmTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		context.TODO(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfoCtx(ctx, "  P (ARM Target): Init()")

	updateConfig, err := toARMTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (ARM Target): expected ArmTargetProviderConfig - %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to convert to ArmTargetProviderConfig", providerName), v1alpha2.InitFailed)
		return err
	}
	r.Config = updateConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorCtx(ctx, err)
				err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: failed to init metrics", providerName), v1alpha2.InitFailed)
			}
		}
	})
	return err
}

func (r *ArmTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		ctx, &map[string]string{
			"method": "Get",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (ARM Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)

	return ret, nil
}

func (r *ArmTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan(
		"ARM Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error
	defer utils.CloseSpanWithError(span, &err)
	defer utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (ARM Target):  applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		radius,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)
	components := step.GetComponents()
	err = r.GetValidationRule(ctx).Validate(components)
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			radius,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)

		sLog.ErrorfCtx(ctx, "  P (ARM Target): failed to validate components, error: %v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		return nil, err
	}
	if isDryRun {
		sLog.DebugfCtx(ctx, "  P (ARM Target): dryRun is enabled,, skipping apply")
		return nil, nil
	}

	return nil, nil
}

func toARMTargetProviderConfig(config providers.IProviderConfig) (ArmTargetProviderConfig, error) {
	ret := ArmTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (*ArmTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{"yaml", "resource"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
			ChangeDetectionProperties: []model.PropertyDesc{
				{Name: "yaml", IgnoreCase: false, SkipIfMissing: true},
				{Name: "resource", IgnoreCase: false, SkipIfMissing: true},
			},
		},
	}
}
