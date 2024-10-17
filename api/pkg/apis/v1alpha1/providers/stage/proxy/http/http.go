/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.proxy.http"
	providerName = "P (HTTP Proxy Stage)"
)

var (
	msLock                   sync.Mutex
	sLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type HTTPProxyStageProviderConfig struct {
}

type HTTPProxyStageProvider struct {
	Config  HTTPProxyStageProviderConfig
	Context *contexts.ManagerContext
}

type HTTPPRoxyProperties struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (s *HTTPProxyStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	ctx, span := observability.StartSpan("[Stage] HTTP Proxy Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	mockConfig, err := toProxyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (HTTP Proxy Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return nil
}
func (s *HTTPProxyStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toProxyStageProviderConfig(config providers.IProviderConfig) (HTTPProxyStageProviderConfig, error) {
	ret := HTTPProxyStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *HTTPProxyStageProvider) InitWithMap(properties map[string]string) error {
	if len(properties) > 0 {
		return v1alpha2.NewCOAError(nil, "properties are not supported", v1alpha2.BadRequest)
	}
	return i.Init(HTTPProxyStageProviderConfig{})
}

func (m *HTTPProxyStageProvider) traceValue(v interface{}, ctx interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := m.Context.VencorContext.EvaluationContext.Clone()
		context.Value = ctx
		v, err := parser.Eval(*context)
		if err != nil {
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return m.traceValue(v, ctx)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := m.traceValue(v, ctx)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := m.traceValue(v, ctx)
			if err != nil {
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
}

func (i *HTTPProxyStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, activationdata v1alpha2.ActivationData) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] HTTP Proxy Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	var ret model.StageStatus
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (HTTP Proxy Stage): start process request")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()

	proxyProperties := HTTPPRoxyProperties{}

	jData, _ := json.Marshal(activationdata.Proxy.Config)
	err = json.Unmarshal(jData, &proxyProperties)
	if err != nil {
		coaError := v1alpha2.NewCOAError(err, "error unmarshalling proxy properties", v1alpha2.BadRequest)
		sLog.Errorf("  P (HTTP Proxy Stage): error unmarshalling proxy properties %s", coaError.Error())
		return nil, false, coaError
	}

	ret, err = utils.CallRemoteProcessor(ctx,
		proxyProperties.BaseUrl,
		proxyProperties.User,
		proxyProperties.Password,
		activationdata)
	if err != nil {
		sLog.Errorf("  P (HTTP Proxy Stage): error calling remote stage processor %s", err.Error())
		return nil, false, err
	}
	if ret.Status != v1alpha2.Done {
		sLog.Errorf("  P (HTTP Proxy Stage): remote stage processor returned an error %s", ret.StatusMessage)
		return nil, false, v1alpha2.NewCOAError(nil, ret.StatusMessage, ret.Status)
	}

	outputs := ret.Outputs
	sLog.InfoCtx(ctx, "  P (HTTP Proxy Stage): end process request")
	providerOperationMetrics.ProviderOperationLatency(
		processTime,
		"http-proxy",
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)
	return outputs, false, nil
}
