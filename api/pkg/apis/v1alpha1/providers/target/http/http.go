/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.target.http"
	providerName = "P (HTTP Target)"
	httpProvider = "http"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

type HttpTargetProviderConfig struct {
	Name string `json:"name"`
}

type HttpTargetProvider struct {
	Config  HttpTargetProviderConfig
	Context *contexts.ManagerContext
}

func HttpTargetProviderConfigFromMap(properties map[string]string) (HttpTargetProviderConfig, error) {
	ret := HttpTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	return ret, nil
}

func (i *HttpTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := HttpTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (HTTP Target): expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	return i.Init(config)
}

func (s *HttpTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *HttpTargetProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("Http Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (HTTP Target): Init()")

	updateConfig, err := toHttpTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (HTTP Target): expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (HTTP Target): failed to create metrics: %+v", err)
			}
		}
	})

	return err
}
func toHttpTargetProviderConfig(config providers.IProviderConfig) (HttpTargetProviderConfig, error) {
	ret := HttpTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *HttpTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (HTTP Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	// This provider doesn't remember what it does, so it always return nil when asked
	return nil, nil
}

func (i *HttpTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (HTTP Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.ObjectMeta.Name,
		SolutionId: deployment.Instance.Spec.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	functionName := utils.GetFunctionName()
	applyTime := time.Now().UTC()
	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (HTTP Target): failed to validate components: %+v", err)
		err = v1alpha2.NewCOAError(err, fmt.Sprintf("%s: the rule validation failed", providerName), v1alpha2.ValidateFailed)
		providerOperationMetrics.ProviderOperationErrors(
			httpProvider,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.UpdateOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		return nil, err
	}
	if isDryRun {
		sLog.DebugCtx(ctx, "  P (HTTP Target): dryRun is enabled, skipping apply")
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	for _, component := range step.Components {
		if component.Action == "update" {
			body := model.ReadPropertyCompat(component.Component.Properties, "http.body", injections)
			url := model.ReadPropertyCompat(component.Component.Properties, "http.url", injections)
			method := model.ReadPropertyCompat(component.Component.Properties, "http.method", injections)

			sLog.InfofCtx(ctx, "  P (HTTP Target):  start to send request to %s", url)
			utils.EmitUserAuditsLogs(ctx, fmt.Sprintf("  P (HTTP Target): Start to send request to %s", url))

			if url == "" {
				err = errors.New("component doesn't have a http.url property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.ErrorfCtx(ctx, "  P (HTTP Target): %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ApplyOperation,
					metrics.UpdateOperationType,
					v1alpha2.BadConfig.String(),
				)
				return ret, err
			}
			if method == "" {
				method = "POST"
			}
			jsonData := []byte(body)
			var request *http.Request
			request, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
			if err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.ErrorfCtx(ctx, "  P (HTTP Target): %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ApplyOperation,
					metrics.UpdateOperationType,
					v1alpha2.HttpNewRequestFailed.String(),
				)
				return ret, err
			}
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")

			client := &http.Client{}
			var resp *http.Response
			resp, err = client.Do(request)
			if err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.ErrorfCtx(ctx, "  P (HTTP Target): failed to process http request: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ApplyOperation,
					metrics.UpdateOperationType,
					v1alpha2.HttpSendRequestFailed.String(),
				)
				return ret, err
			}
			if resp.StatusCode != http.StatusOK {
				bodyBytes, err := io.ReadAll(resp.Body)
				var message string
				if err != nil {
					sLog.ErrorfCtx(ctx, "  P (HTTP Target): failed to read response body: %v", err)
					message = err.Error()
				} else {
					message = string(bodyBytes)
				}
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: message,
				}
				err = errors.New("HTTP request didn't respond 200 OK")
				sLog.ErrorfCtx(ctx, "  P (HTTP Target): %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					httpProvider,
					functionName,
					metrics.ApplyOperation,
					metrics.UpdateOperationType,
					v1alpha2.HttpErrorResponse.String(),
				)
				return ret, err
			}

			ret[component.Component.Name] = model.ComponentResultSpec{
				Status:  v1alpha2.Updated,
				Message: "HTTP request succeeded",
			}
		} else {
			sLog.InfofCtx(ctx, "  P (HTTP Target): component %s is not in update action, skipping", component.Component.Name)
		}
	}
	providerOperationMetrics.ProviderOperationLatency(
		applyTime,
		httpProvider,
		metrics.ApplyOperation,
		metrics.UpdateOperationType,
		functionName,
	)
	return ret, nil
}
func (*HttpTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{"http.url"},
			OptionalProperties:    []string{"http.method", "http.body"},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
		},
	}
}
