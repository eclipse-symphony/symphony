/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package proxy

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var sLog = logger.NewLogger("coa.runtime")

type ProxyStageProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ProxyStageProvider struct {
	Config  ProxyStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *ProxyStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toProxyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func (s *ProxyStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toProxyStageProviderConfig(config providers.IProviderConfig) (ProxyStageProviderConfig, error) {
	ret := ProxyStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *ProxyStageProvider) InitWithMap(properties map[string]string) error {
	config, err := SymphonyStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func SymphonyStageProviderConfigFromMap(properties map[string]string) (ProxyStageProviderConfig, error) {
	ret := ProxyStageProviderConfig{}
	baseUrl, err := utils.GetString(properties, "baseUrl")
	if err != nil {
		return ret, err
	}
	ret.BaseUrl = baseUrl
	if ret.BaseUrl == "" {
		return ret, v1alpha2.NewCOAError(nil, "baseUrl is required", v1alpha2.BadConfig)
	}
	user, err := utils.GetString(properties, "user")
	if err != nil {
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
	}
	password, err := utils.GetString(properties, "password")
	if err != nil {
		return ret, err
	}
	ret.Password = password
	return ret, nil
}
func (m *ProxyStageProvider) traceValue(v interface{}, ctx interface{}) (interface{}, error) {
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

func (i *ProxyStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, activationdata v1alpha2.ActivationData) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Proxy Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	var ret model.ActivationStatus
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (Proxy Stage): start process request")

	ret, err = utils.CallRemoteProcessor(ctx,
		activationdata.Proxy.Config.BaseUrl,
		activationdata.Proxy.Config.User,
		activationdata.Proxy.Config.Password,
		activationdata)
	if err != nil {
		sLog.Errorf("  P (Proxy Stage): error calling remote stage processor %s", err.Error())
		return nil, false, err
	}
	if ret.ErrorMessage != "" {
		sLog.Errorf("  P (Proxy Stage): remote stage processor returned an error %s", ret.ErrorMessage)
		return nil, false, v1alpha2.NewCOAError(nil, ret.ErrorMessage, v1alpha2.InternalError)
	}
	outputs := ret.Outputs

	sLog.Info("  P (Proxy Stage): end process request")
	return outputs, false, nil
}
