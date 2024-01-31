/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type ProxyUpdateProviderConfig struct {
	Name      string `json:"name"`
	ServerURL string `json:"serverUrl"`
}

type ProxyUpdateProvider struct {
	Config  ProxyUpdateProviderConfig
	Context *contexts.ManagerContext
}

func ProxyUpdateProviderConfigFromMap(properties map[string]string) (ProxyUpdateProviderConfig, error) {
	ret := ProxyUpdateProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["serverUrl"]; ok {
		ret.ServerURL = utils.ParseProperty(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "proxy update provider server url is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (i *ProxyUpdateProvider) InitWithMap(properties map[string]string) error {
	config, err := ProxyUpdateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *ProxyUpdateProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *ProxyUpdateProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Proxy Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Info("  P (Proxy Target): Init()")

	updateConfig, err := toProxyUpdateProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Proxy Target): expected ProxyUpdateProviderConfig - %+v", err)
		err = errors.New("expected ProxyUpdateProviderConfig")
		return err
	}
	i.Config = updateConfig

	return nil
}
func toProxyUpdateProviderConfig(config providers.IProviderConfig) (ProxyUpdateProviderConfig, error) {
	ret := ProxyUpdateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.ServerURL = utils.ParseProperty(ret.ServerURL)
	return ret, err
}

func (a *ProxyUpdateProvider) callRestAPI(route string, method string, payload []byte) ([]byte, error) {
	client := &http.Client{}
	url := a.Config.ServerURL + route
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	if resp.StatusCode >= 300 {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", string(bodyBytes)), v1alpha2.InternalError)
	}
	return bodyBytes, nil
}

func (i *ProxyUpdateProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Proxy Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Proxy Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	data, _ := json.Marshal(deployment)
	payload, err := i.callRestAPI("instances", "GET", data)
	if err != nil {
		sLog.Errorf("  P (Proxy Target): failed to get instances: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	ret := make([]model.ComponentSpec, 0)
	err = json.Unmarshal(payload, &ret)
	if err != nil {
		sLog.Errorf("  P (Proxy Target): failed to unmarshall get response: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	return ret, nil
}

func (i *ProxyUpdateProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Proxy Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Proxy Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.Errorf("  P (Proxy Target): failed to validate components: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		data, _ := json.Marshal(deployment)

		_, err = i.callRestAPI("instances", "POST", data)
		if err != nil {
			sLog.Errorf("  P (Proxy Target): failed to post instances: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		data, _ := json.Marshal(deployment)
		_, err = i.callRestAPI("instances", "DELETE", data)
		if err != nil {
			sLog.Errorf("  P (Proxy Target): failed to delete instances: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}
	//TODO: Should we remove empty namespaces?
	err = nil
	return ret, nil
}

func (*ProxyUpdateProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}

type TwoComponentSlices struct {
	Current []model.ComponentSpec `json:"current"`
	Desired []model.ComponentSpec `json:"desired"`
}
