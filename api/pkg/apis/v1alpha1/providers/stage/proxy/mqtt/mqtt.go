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
	coalogcontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	loggerName   = "providers.stage.proxy.mqtt"
	providerName = "P (MQTT Proxy Stage)"
)

var (
	msLock                   sync.Mutex
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

type MQTTProxyStageProviderConfig struct {
}

type MQTTProxyProperties struct {
	BrokerAddress      string `json:"brokerAddress"`
	RequestTopic       string `json:"requestTopic"`
	ResponseTopic      string `json:"responseTopic"`
	TimeoutSeconds     int    `json:"timeoutSeconds,omitempty"`
	KeepAliveSeconds   int    `json:"keepAliveSeconds,omitempty"`
	PingTimeoutSeconds int    `json:"pingTimeoutSeconds,omitempty"`
}

type MQTTProxyStageProvider struct {
	Config        MQTTProxyStageProviderConfig
	Context       *contexts.ManagerContext
	ResponseChans sync.Map
}

type ProxyResponse struct {
	IsOK    bool
	State   v1alpha2.State
	Payload interface{}
}

func (s *MQTTProxyStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	ctx, span := observability.StartSpan("[Stage] MQTT Proxy Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (MQTT Proxy Stage): Init()")

	mockConfig, err := toProxyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): failed to create metrics - %v", err)
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
	if len(properties) > 0 {
		return v1alpha2.NewCOAError(nil, "properties are not supported", v1alpha2.BadRequest)
	}
	return i.Init(MQTTProxyStageProviderConfig{})
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

	ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)

	proxyProperties := MQTTProxyProperties{}

	jData, _ := json.Marshal(activationdata.Proxy.Config)
	err = json.Unmarshal(jData, &proxyProperties)
	if err != nil {
		coaError := v1alpha2.NewCOAError(err, "error unmarshalling proxy properties", v1alpha2.BadRequest)
		sLog.Errorf("  P (MQTT Proxy Stage): error unmarshalling proxy properties %s", coaError.Error())
		return nil, false, coaError
	}
	if proxyProperties.TimeoutSeconds == 0 {
		proxyProperties.TimeoutSeconds = 5
	}
	if proxyProperties.KeepAliveSeconds == 0 {
		proxyProperties.KeepAliveSeconds = 2
	}
	if proxyProperties.PingTimeoutSeconds == 0 {
		proxyProperties.PingTimeoutSeconds = 1
	}

	id := uuid.New()
	opts := gmqtt.NewClientOptions().AddBroker(proxyProperties.BrokerAddress).SetClientID(id.String())
	opts.SetKeepAlive(time.Duration(proxyProperties.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(proxyProperties.PingTimeoutSeconds) * time.Second)
	opts.CleanSession = true
	client := gmqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): faild to connect to MQTT broker - %+v", err)
		return nil, false, v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}
	defer client.Disconnect(250)
	if token := client.Subscribe(proxyProperties.ResponseTopic, 1, func(client gmqtt.Client, msg gmqtt.Message) {
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
			if ch, ok := i.ResponseChans.Load(reqId); ok {
				ch.(chan ProxyResponse) <- proxyResponse
				i.ResponseChans.Delete(reqId)
			}
		}
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): faild to connect to subscribe to the response topic - %+v", token.Error())
			err = v1alpha2.NewCOAError(token.Error(), "failed to subscribe to response topic", v1alpha2.InternalError)
			return nil, false, err
		}
	}

	reqId := uuid.New().String()
	responseChan := make(chan ProxyResponse)
	i.ResponseChans.Store(reqId, responseChan)
	// clear proxy before sending
	activationdata.Proxy = nil
	data, _ := json.Marshal(activationdata)
	request := v1alpha2.COARequest{
		Route:  "processor",
		Method: "POST",
		Body:   data,
		Metadata: map[string]string{
			"request-id": reqId,
		},
		Context: ctx,
	}
	data, _ = json.Marshal(request)

	if token := client.Publish(proxyProperties.RequestTopic, 1, false, data); token.Wait() && token.Error() != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): failed to publish process request - %s", token.Error())
		err = token.Error()
		return nil, false, err
	}

	timeout := time.After(time.Duration(proxyProperties.TimeoutSeconds) * time.Second)
	select {
	case resp := <-responseChan:
		if resp.IsOK {
			data := []byte(resp.Payload.(string))
			var ret model.StageStatus
			err = json.Unmarshal(data, &ret)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Proxy Stage): failed to deserialize components - %s - %s", err.Error(), fmt.Sprint(data))
				err = v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.InternalError)
				return nil, false, err
			}
			if ret.Status != v1alpha2.Done {
				return nil, false, v1alpha2.NewCOAError(nil, ret.StatusMessage, ret.Status)
			}
			return ret.Outputs, false, nil
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
