/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/scriptutils"
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

// validateProxyBaseUrl checks that the proxy baseUrl is safe to connect to.
// It enforces an http/https scheme and applies the server-wide SecurityPolicy
// allow-list/exclusive-mode rules to the resolved host addresses.
func validateProxyBaseUrl(rawURL string, policy *contexts.SecurityPolicy) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return v1alpha2.NewCOAError(err, "invalid proxy baseUrl", v1alpha2.BadConfig)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid proxy baseUrl scheme %q: only http and https are permitted", u.Scheme),
			v1alpha2.BadConfig)
	}

	var allowedNets []*net.IPNet
	exclusiveMode := false
	if policy != nil {
		allowedNets, err = scriptutils.ParseIPRanges(policy.AllowedIPRanges)
		if err != nil {
			return v1alpha2.NewCOAError(err, "invalid allowedIPRanges in security policy", v1alpha2.BadConfig)
		}
		exclusiveMode = policy.AllowListExclusive
	}

	return scriptutils.ValidateURLHost(u.Hostname(), allowedNets, exclusiveMode)
}

func (i *HTTPProxyStageProvider) InitWithMap(properties map[string]string) error {
	if len(properties) > 0 {
		return v1alpha2.NewCOAError(nil, "properties are not supported", v1alpha2.BadRequest)
	}
	return i.Init(HTTPProxyStageProviderConfig{})
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

	err = validateProxyBaseUrl(proxyProperties.BaseUrl, mgrContext.GetSecurityPolicy())
	if err != nil {
		sLog.Errorf("  P (HTTP Proxy Stage): invalid proxy baseUrl %s", err.Error())
		return nil, false, err
	}

	// the remote site address is only known per activation, so the API client
	// targeting the remote site is created here instead of in Init
	apiClient, err := utils.GetParentApiClient(proxyProperties.BaseUrl)
	if err != nil {
		sLog.Errorf("  P (HTTP Proxy Stage): error creating API client for remote site %s", err.Error())
		return nil, false, err
	}
	ret, err = apiClient.CallRemoteProcessor(ctx,
		activationdata,
		proxyProperties.User,
		proxyProperties.Password)
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
