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
	"io"
	"net/http"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

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
		return err
	}
	return i.Init(config)
}

func (s *HttpTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *HttpTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Http Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (HTTP Target): Init()")

	updateConfig, err := toHttpTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (HTTP Target): expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

	return nil
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
	_, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (HTTP Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	// This provider doesn't remember what it does, so it always return nil when asked
	return nil, nil
}

func (i *HttpTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (HTTP Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.Errorf("  P (HTTP Target): failed to validate components: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	for _, component := range step.Components {
		if component.Action == "update" {
			body := model.ReadPropertyCompat(component.Component.Properties, "http.body", injections)
			url := model.ReadPropertyCompat(component.Component.Properties, "http.url", injections)
			method := model.ReadPropertyCompat(component.Component.Properties, "http.method", injections)

			if url == "" {
				err = errors.New("component doesn't have a http.url property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P (HTTP Target): %v, traceId: %s", err, span.SpanContext().TraceID().String())
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
				sLog.Errorf("  P (HTTP Target): %v, traceId: %s", err, span.SpanContext().TraceID().String())
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
				sLog.Errorf("  P (HTTP Target): %v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
			if resp.StatusCode != http.StatusOK {
				bodyBytes, err := io.ReadAll(resp.Body)
				var message string
				if err != nil {
					message = err.Error()
				} else {
					message = string(bodyBytes)
				}
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: message,
				}
				err = errors.New("HTTP request didn't respond 200 OK")
				sLog.Errorf("  P (HTTP Target): %v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
		}
	}
	return ret, nil
}
func (*HttpTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{"http.url"},
		OptionalProperties:    []string{"http.method", "http.body"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
