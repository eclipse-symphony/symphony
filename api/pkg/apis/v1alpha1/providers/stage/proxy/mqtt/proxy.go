/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	coalogcontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var msLock sync.Mutex

const (
	loggerName   = "providers.target.mqtt"
	providerName = "P (MQTT Stage)"
	mqtt         = "mqtt"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

type MQTTProxyStageProviderConfig struct {
	BrokerAddress      string `json:"brokerAddress"`
	ClientID           string `json:"clientID"`
	RequestTopic       string `json:"requestTopic"`
	ResponseTopic      string `json:"responseTopic"`
	TimeoutSeconds     int    `json:"timeoutSeconds,omitempty"`
	KeepAliveSeconds   int    `json:"keepAliveSeconds,omitempty"`
	PingTimeoutSeconds int    `json:"pingTimeoutSeconds,omitempty"`
}

type MQTTProxyStageProvider struct {
	Config        MQTTProxyStageProviderConfig
	Context       *contexts.ManagerContext
	MQTTClient    gmqtt.Client
	ResponseChans sync.Map
	Initialized   bool
}

type ProxyResponse struct {
	IsOK    bool
	State   v1alpha2.State
	Payload interface{}
}

func (s *MQTTProxyStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	ctx, span := observability.StartSpan("MQTT Stage Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (MQTT Stage): Init()")

	mockConfig, err := toProxyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	id := uuid.New()
	opts := gmqtt.NewClientOptions().AddBroker(s.Config.BrokerAddress).SetClientID(id.String())
	opts.SetKeepAlive(time.Duration(s.Config.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(s.Config.PingTimeoutSeconds) * time.Second)
	opts.CleanSession = true
	s.MQTTClient = gmqtt.NewClient(opts)
	if token := s.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Stage): faild to connect to MQTT broker - %+v", err)
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	if token := s.MQTTClient.Subscribe(s.Config.ResponseTopic, 1, func(client gmqtt.Client, msg gmqtt.Message) {
		var response v1alpha2.COAResponse
		json.Unmarshal(msg.Payload(), &response)
		proxyResponse := ProxyResponse{
			IsOK:    response.State == v1alpha2.OK || response.State == v1alpha2.Accepted,
			State:   response.State,
			Payload: response.String(),
		}

		if !proxyResponse.IsOK {
			proxyResponse.Payload = string(response.Body)
		}

		if reqId, ok := response.Metadata["request-id"]; ok {
			if ch, ok := s.ResponseChans.Load(reqId); ok {
				ch.(chan ProxyResponse) <- proxyResponse
				s.ResponseChans.Delete(reqId)
			}
		}
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			sLog.ErrorfCtx(ctx, "  P (MQTT Stage): faild to connect to subscribe to the response topic - %+v", token.Error())
			err = v1alpha2.NewCOAError(token.Error(), "failed to subscribe to response topic", v1alpha2.InternalError)
			return err
		}
	}
	s.Initialized = true

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Stage): failed to create metrics - %v", err)
			}
		}
	})

	return err
}
func (s *MQTTProxyStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toProxyStageProviderConfig(config providers.IProviderConfig) (MQTTProxyStageProviderConfig, error) {
	ret := MQTTProxyStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MQTTProxyStageProvider) InitWithMap(properties map[string]string) error {
	config, err := SymphonyStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func SymphonyStageProviderConfigFromMap(properties map[string]string) (MQTTProxyStageProviderConfig, error) {
	ret := MQTTProxyStageProviderConfig{}
	if v, ok := properties["brokerAddress"]; ok {
		ret.BrokerAddress = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'brokerAdress' is missing in MQTT provider config", v1alpha2.BadConfig)
	}
	if v, ok := properties["clientID"]; ok {
		ret.ClientID = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'clientID' is missing in MQTT provider config", v1alpha2.BadConfig)
	}
	if v, ok := properties["requestTopic"]; ok {
		ret.RequestTopic = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'requestTopic' is missing in MQTT provider config", v1alpha2.BadConfig)
	}
	if v, ok := properties["responseTopic"]; ok {
		ret.ResponseTopic = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "'responseTopic' is missing in MQTT provider config", v1alpha2.BadConfig)
	}
	if v, ok := properties["timeoutSeconds"]; ok {
		if num, err := strconv.Atoi(v); err == nil {
			ret.TimeoutSeconds = num
		} else {
			return ret, v1alpha2.NewCOAError(nil, "'timeoutSeconds' is not an integer in MQTT provider config", v1alpha2.BadConfig)
		}
	} else {
		ret.TimeoutSeconds = 8
	}
	if v, ok := properties["keepAliveSeconds"]; ok {
		if num, err := strconv.Atoi(v); err == nil {
			ret.KeepAliveSeconds = num
		} else {
			return ret, v1alpha2.NewCOAError(nil, "'keepAliveSeconds' is not an integer in MQTT provider config", v1alpha2.BadConfig)
		}
	} else {
		ret.KeepAliveSeconds = 2
	}
	if v, ok := properties["pingTimeoutSeconds"]; ok {
		if num, err := strconv.Atoi(v); err == nil {
			ret.PingTimeoutSeconds = num
		} else {
			return ret, v1alpha2.NewCOAError(nil, "'pingTimeoutSeconds' is not an integer in MQTT provider config", v1alpha2.BadConfig)
		}
	} else {
		ret.PingTimeoutSeconds = 1
	}
	if ret.TimeoutSeconds <= 0 {
		ret.TimeoutSeconds = 8
	}
	return ret, nil
}
func (m *MQTTProxyStageProvider) traceValue(v interface{}, ctx interface{}) (interface{}, error) {
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

func (i *MQTTProxyStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, activationdata v1alpha2.ActivationData) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] MQTT Proxy provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (MQTT Proxy Stage) process started")

	data, _ := json.Marshal(activationdata)
	ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)

	reqId := uuid.New().String()
	responseChan := make(chan ProxyResponse)
	i.ResponseChans.Store(reqId, responseChan)
	request := v1alpha2.COARequest{
		Route:  "process",
		Method: "POST",
		Body:   data,
		Metadata: map[string]string{
			"request-id": reqId,
		},
		Context: ctx,
	}
	data, _ = json.Marshal(request)

	if token := i.MQTTClient.Publish(i.Config.RequestTopic, 1, false, activationdata); token.Wait() && token.Error() != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): failed to getting artifacts - %s", token.Error())
		err = token.Error()
		return nil, false, err
	}

	timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
	select {
	case resp := <-responseChan:
		if resp.IsOK {
			data := []byte(resp.Payload.(string))
			var ret map[string]interface{}
			err = json.Unmarshal(data, &ret)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): failed to deserialize components - %s - %s", err.Error(), fmt.Sprint(data))
				err = v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.InternalError)
				return nil, false, err
			}
			return ret, false, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to get response - %s - %s", err.Error(), fmt.Sprint(string(data)))
			return nil, false, err
		}
	case <-timeout:
		err = v1alpha2.NewCOAError(nil, "didn't get response to Get() call over MQTT", v1alpha2.InternalError)
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): request timeout - %s - %s", err.Error(), fmt.Sprint(string(data)))
		return nil, false, err
	}
}
