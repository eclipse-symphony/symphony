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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var sLog = logger.NewLogger("coa.runtime")

type MQTTTargetProviderConfig struct {
	Name               string `json:"name"`
	BrokerAddress      string `json:"brokerAddress"`
	ClientID           string `json:"clientID"`
	RequestTopic       string `json:"requestTopic"`
	ResponseTopic      string `json:"responseTopic"`
	TimeoutSeconds     int    `json:"timeoutSeconds,omitempty"`
	KeepAliveSeconds   int    `json:"keepAliveSeconds,omitempty"`
	PingTimeoutSeconds int    `json:"pingTimeoutSeconds,omitempty"`
}

var lock sync.Mutex

type ProxyResponse struct {
	IsOK    bool
	State   v1alpha2.State
	Payload interface{}
}
type MQTTTargetProvider struct {
	Config          MQTTTargetProviderConfig
	Context         *contexts.ManagerContext
	MQTTClient      gmqtt.Client
	GetChan         chan ProxyResponse
	RemoveChan      chan ProxyResponse
	NeedsUpdateChan chan ProxyResponse
	NeedsRemoveChan chan ProxyResponse
	ApplyChan       chan ProxyResponse
	Initialized     bool
}

func MQTTTargetProviderConfigFromMap(properties map[string]string) (MQTTTargetProviderConfig, error) {
	ret := MQTTTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
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
	return ret, nil
}

func (i *MQTTTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := MQTTTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *MQTTTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *MQTTTargetProvider) Init(config providers.IProviderConfig) error {
	lock.Lock()
	defer lock.Unlock()

	_, span := observability.StartSpan("MQTT Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (MQTT Target): Init()")

	if i.Initialized {
		return nil
	}
	updateConfig, err := toMQTTTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (MQTT Target): expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig
	id := uuid.New()
	opts := gmqtt.NewClientOptions().AddBroker(i.Config.BrokerAddress).SetClientID(id.String())
	opts.SetKeepAlive(time.Duration(i.Config.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(i.Config.PingTimeoutSeconds) * time.Second)
	opts.CleanSession = true
	i.MQTTClient = gmqtt.NewClient(opts)
	if token := i.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		sLog.Errorf("  P (MQTT Target): faild to connect to MQTT broker - %+v", err)
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	i.GetChan = make(chan ProxyResponse)
	i.RemoveChan = make(chan ProxyResponse)
	i.NeedsUpdateChan = make(chan ProxyResponse)
	i.NeedsRemoveChan = make(chan ProxyResponse)
	i.ApplyChan = make(chan ProxyResponse)

	if token := i.MQTTClient.Subscribe(i.Config.ResponseTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var response v1alpha2.COAResponse
		json.Unmarshal(msg.Payload(), &response)
		proxyResponse := ProxyResponse{
			IsOK:  response.State == v1alpha2.OK || response.State == v1alpha2.Accepted,
			State: response.State,
		}
		if !proxyResponse.IsOK {
			proxyResponse.Payload = string(response.Body)
		}
		switch response.Metadata["call-context"] {
		case "TargetProvider-Get":
			if proxyResponse.IsOK {
				var ret []model.ComponentSpec
				err = json.Unmarshal(response.Body, &ret)
				if err != nil {
					sLog.Errorf("  P (MQTT Target): faild to deserialize components from MQTT - %+v, %s", err, string(response.Body))
				}
				proxyResponse.Payload = ret
			}
			i.GetChan <- proxyResponse
		case "TargetProvider-Remove":
			i.RemoveChan <- proxyResponse
		case "TargetProvider-NeedsUpdate":
			i.NeedsUpdateChan <- proxyResponse
		case "TargetProvider-NeedsRemove":
			i.NeedsRemoveChan <- proxyResponse
		case "TargetProvider-Apply":
			i.ApplyChan <- proxyResponse
		}
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			sLog.Errorf("  P (MQTT Target): faild to connect to subscribe to the response topic - %+v", token.Error())
			err = v1alpha2.NewCOAError(token.Error(), "failed to subscribe to response topic", v1alpha2.InternalError)
			return err
		}
	}
	i.Initialized = true
	return nil
}
func toMQTTTargetProviderConfig(config providers.IProviderConfig) (MQTTTargetProviderConfig, error) {
	ret := MQTTTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *MQTTTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (MQTT Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	data, _ := json.Marshal(deployment)
	request := v1alpha2.COARequest{
		Route:  "instances",
		Method: "GET",
		Body:   data,
		Metadata: map[string]string{
			"call-context": "TargetProvider-Get",
		},
	}
	data, _ = json.Marshal(request)

	if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
		sLog.Errorf("  P (MQTT Target): failed to getting artifacts - %s, traceId: %s", token.Error(), span.SpanContext().TraceID().String())
		err = token.Error()
		return nil, err
	}

	timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
	select {
	case resp := <-i.GetChan:
		if resp.IsOK {
			var data []byte
			data, err = json.Marshal(resp.Payload)
			if err != nil {
				sLog.Errorf("  P (MQTT Target): failed to serialize payload - %s - %s, traceId: %s", err.Error(), fmt.Sprint(resp.Payload), span.SpanContext().TraceID().String())
				err = v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.InternalError)
				return nil, err
			}
			var ret []model.ComponentSpec
			err = json.Unmarshal(data, &ret)
			if err != nil {
				sLog.Errorf("  P (MQTT Target): failed to deserialize components - %s - %s, traceId: %s", err.Error(), fmt.Sprint(data), span.SpanContext().TraceID().String())
				err = v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.InternalError)
				return nil, err
			}
			return ret, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
			sLog.Errorf("  P (MQTT Target): failed to get response - %s - %s, traceId: %s", err.Error(), fmt.Sprint(data), span.SpanContext().TraceID().String())
			return nil, err
		}
	case <-timeout:
		err = v1alpha2.NewCOAError(nil, "didn't get response to Get() call over MQTT", v1alpha2.InternalError)
		sLog.Errorf("  P (MQTT Target): request timeout - %s - %s, traceId: %s", err.Error(), fmt.Sprint(data), span.SpanContext().TraceID().String())
		return nil, err
	}
}
func (i *MQTTTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (MQTT Target): deleting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	data, _ := json.Marshal(deployment)
	request := v1alpha2.COARequest{
		Route:  "instances",
		Method: "DELETE",
		Body:   data,
		Metadata: map[string]string{
			"call-context": "TargetProvider-Remove",
		},
	}
	data, _ = json.Marshal(request)

	if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
		err = token.Error()
		return err
	}

	timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
	select {
	case resp := <-i.RemoveChan:
		if resp.IsOK {
			err = nil
			return err
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
			sLog.Errorf("  P (MQTT Target): failed to get correct response - %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return err
		}
	case <-timeout:
		err = v1alpha2.NewCOAError(nil, "didn't get response to Remove() call over MQTT", v1alpha2.InternalError)
		sLog.Errorf("  P (MQTT Target): request timeout - %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return err
	}
}

func (i *MQTTTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (MQTT Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	data, _ := json.Marshal(deployment)

	components = step.GetUpdatedComponents()
	if len(components) > 0 {

		request := v1alpha2.COARequest{
			Route:  "instances",
			Method: "POST",
			Body:   data,
			Metadata: map[string]string{
				"call-context": "TargetProvider-Apply",
			},
		}
		data, _ = json.Marshal(request)

		if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			return ret, err
		}

		timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
		select {
		case resp := <-i.ApplyChan:
			if resp.IsOK {
				err = nil
				return ret, err
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
				sLog.Errorf("  P (MQTT Target): failed to get correct response from Apply() - %v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
		case <-timeout:
			err = v1alpha2.NewCOAError(nil, "didn't get response to Apply()-Update call over MQTT", v1alpha2.InternalError)
			sLog.Errorf("  P (MQTT Target): request timeout - %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		request := v1alpha2.COARequest{
			Route:  "instances",
			Method: "DELETE",
			Body:   data,
			Metadata: map[string]string{
				"call-context": "TargetProvider-Remove",
			},
		}
		data, _ = json.Marshal(request)

		if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			return ret, err
		}

		timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
		select {
		case resp := <-i.RemoveChan:
			if resp.IsOK {
				err = nil
				return ret, err
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
				sLog.Errorf("  P (MQTT Target): failed to get correct reponse from Apply() delete action - %v, traceId: %s", err, span.SpanContext().TraceID().String())
				return ret, err
			}
		case <-timeout:
			err = v1alpha2.NewCOAError(nil, "didn't get response to Apply()-Delete call over MQTT", v1alpha2.InternalError)
			sLog.Errorf("  P (MQTT Target): request timeout - %v, traceId: %s", err, span.SpanContext().TraceID().String())
			return ret, err
		}
	}
	//TODO: Should we remove empty namespaces?
	err = nil
	return ret, nil
}

func (*MQTTTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
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
