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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	coalogcontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	loggerName   = "providers.target.mqtt"
	providerName = "P (MQTT Target)"
	mqtt         = "mqtt"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

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
	Config        MQTTTargetProviderConfig
	Context       *contexts.ManagerContext
	MQTTClient    gmqtt.Client
	ResponseChans sync.Map
	Initialized   bool
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
	if ret.TimeoutSeconds <= 0 {
		ret.TimeoutSeconds = 8
	}
	return ret, nil
}

func (i *MQTTTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := MQTTTargetProviderConfigFromMap(properties)
	if err != nil {
		sLog.Errorf("  P (MQTT Target): expected MQTTTargetProviderConfig: %+v", err)
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

	ctx, span := observability.StartSpan("MQTT Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (MQTT Target): Init()")

	if i.Initialized {
		return nil
	}
	updateConfig, err := toMQTTTargetProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): expected MQTTTargetProviderConfig: %+v", err)
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
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): faild to connect to MQTT broker - %+v", err)
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	if token := i.MQTTClient.Subscribe(i.Config.ResponseTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
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
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): faild to connect to subscribe to the response topic - %+v", token.Error())
			err = v1alpha2.NewCOAError(token.Error(), "failed to subscribe to response topic", v1alpha2.InternalError)
			return err
		}
	}
	i.Initialized = true

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to create metrics - %v", err)
			}
		}
	})

	return err
}
func toMQTTTargetProviderConfig(config providers.IProviderConfig) (MQTTTargetProviderConfig, error) {
	ret := MQTTTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	if ret.TimeoutSeconds <= 0 {
		ret.TimeoutSeconds = 8
	}
	return ret, err
}

func (i *MQTTTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (MQTT Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	data, _ := json.Marshal(deployment)
	ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)

	reqId := uuid.New().String()
	responseChan := make(chan ProxyResponse)
	i.ResponseChans.Store(reqId, responseChan)
	request := v1alpha2.COARequest{
		Route:  "instances",
		Method: "GET",
		Body:   data,
		Metadata: map[string]string{
			"active-target": deployment.ActiveTarget,
			"request-id":    reqId,
		},
		Context: ctx,
	}
	data, _ = json.Marshal(request)

	sLog.InfofCtx(ctx, "  P (MQTT Target): start to publish on topic %s", i.Config.RequestTopic)
	if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to getting artifacts - %s", token.Error())
		err = token.Error()
		return nil, err
	}
	timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
	select {
	case resp := <-responseChan:
		if resp.IsOK {
			data := []byte(resp.Payload.(string))
			var ret []model.ComponentSpec
			err = json.Unmarshal(data, &ret)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to deserialize components - %s - %s", err.Error(), fmt.Sprint(data))
				err = v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.InternalError)
				return nil, err
			}
			return ret, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to get response - %s - %s", err.Error(), fmt.Sprint(string(data)))
			return nil, err
		}
	case <-timeout:
		err = v1alpha2.NewCOAError(nil, "didn't get response to Get() call over MQTT", v1alpha2.InternalError)
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): request timeout - %s - %s", err.Error(), fmt.Sprint(string(data)))
		return nil, err
	}
}
func (i *MQTTTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	ctx, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (MQTT Target): deleting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	data, _ := json.Marshal(deployment)
	ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)

	reqId := uuid.New().String()
	responseChan := make(chan ProxyResponse)
	i.ResponseChans.Store(reqId, responseChan)
	request := v1alpha2.COARequest{
		Route:  "instances",
		Method: "DELETE",
		Body:   data,
		Metadata: map[string]string{
			"active-target": deployment.ActiveTarget,
			"request-id":    reqId,
		},
		Context: ctx,
	}
	data, _ = json.Marshal(request)

	sLog.InfofCtx(ctx, "  P (MQTT Target): start to publish on topic %s", i.Config.RequestTopic)
	if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
		err = token.Error()
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to publish - %v", err)
		return err
	}

	timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
	select {
	case resp := <-responseChan:
		if resp.IsOK {
			err = nil
			return err
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to get correct response - %v", err)
			return err
		}
	case <-timeout:
		err = v1alpha2.NewCOAError(nil, "didn't get response to Remove() call over MQTT", v1alpha2.InternalError)
		sLog.ErrorfCtx(ctx, "  P (MQTT Target): request timeout - %v", err)
		return err
	}
}

func (i *MQTTTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("MQTT Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (MQTT Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := observ_utils.GetFunctionName()
	startTime := time.Now().UTC()
	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		mqtt,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			mqtt,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		return nil, err
	}
	if isDryRun {
		sLog.DebugCtx(ctx, "  P (MQTT Target): dryRun is enabled, skipping apply")
		return nil, nil
	}

	ret := step.PrepareResultMap()
	data, _ := json.Marshal(deployment)

	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (MQTT Target): get updated components: count - %d", len(components))
		ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)
		requestId := uuid.New().String()
		request := v1alpha2.COARequest{
			Route:  "instances",
			Method: "POST",
			Body:   data,
			Metadata: map[string]string{
				"request-id":    requestId,
				"active-target": deployment.ActiveTarget,
			},
			Context: ctx,
		}
		data, _ = json.Marshal(request)

		utils.EmitUserAuditsLogs(ctx, "  P (MQTT Target): Start to send Apply()-Update request over MQTT on topic %s", i.Config.RequestTopic)

		responseChan := make(chan ProxyResponse)
		i.ResponseChans.Store(requestId, responseChan)

		sLog.InfofCtx(ctx, "  P (MQTT Target): start to publish on topic %s", i.Config.RequestTopic)
		if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			providerOperationMetrics.ProviderOperationErrors(
				mqtt,
				functionName,
				metrics.ApplyOperation,
				metrics.ApplyOperationType,
				v1alpha2.MqttPublishFailed.String(),
			)
			return ret, err
		}

		timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
		select {
		case resp := <-responseChan:
			if resp.IsOK {
				data := []byte(resp.Payload.(string))
				var summary model.SummarySpec
				err = json.Unmarshal(data, &summary)
				if err == nil {
					// Update ret
					for target, targetResult := range summary.TargetResults {
						for _, componentResults := range targetResult.ComponentResults {
							ret[target] = componentResults
						}
					}
				}
				return ret, err
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
				sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to get correct response from Apply() - %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					mqtt,
					functionName,
					metrics.ApplyOperation,
					metrics.ApplyOperationType,
					v1alpha2.MqttApplyFailed.String(),
				)
				return ret, err
			}
		case <-timeout:
			err = v1alpha2.NewCOAError(nil, "didn't get response to Apply()-Update call over MQTT", v1alpha2.InternalError)
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): request timeout - %v", err)
			providerOperationMetrics.ProviderOperationErrors(
				mqtt,
				functionName,
				metrics.ApplyOperation,
				metrics.ApplyOperationType,
				v1alpha2.MqttApplyTimeout.String(),
			)
			return ret, err
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (MQTT Target): get deleted components: count - %d", len(components))
		ctx = coalogcontexts.GenerateCorrelationIdToParentContextIfMissing(ctx)
		requestId := uuid.New().String()
		request := v1alpha2.COARequest{
			Route:  "instances",
			Method: "DELETE",
			Body:   data,
			Metadata: map[string]string{
				"request-id": requestId,
			},
			Context: ctx,
		}
		data, _ = json.Marshal(request)

		utils.EmitUserAuditsLogs(ctx, "  P (MQTT Target): Start to send Apply()-Delete action over MQTT on topic %s", i.Config.RequestTopic)

		responseChan := make(chan ProxyResponse)
		i.ResponseChans.Store(requestId, responseChan)

		if token := i.MQTTClient.Publish(i.Config.RequestTopic, 0, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			providerOperationMetrics.ProviderOperationErrors(
				mqtt,
				functionName,
				metrics.ApplyOperation,
				metrics.ApplyOperationType,
				v1alpha2.MqttPublishFailed.String(),
			)
			return ret, err
		}

		timeout := time.After(time.Duration(i.Config.TimeoutSeconds) * time.Second)
		select {
		case resp := <-responseChan:
			if resp.IsOK {
				err = nil
				return ret, err
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprint(resp.Payload), resp.State)
				sLog.ErrorfCtx(ctx, "  P (MQTT Target): failed to get correct reponse from Apply() delete action - %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					mqtt,
					functionName,
					metrics.ApplyOperation,
					metrics.ApplyOperationType,
					v1alpha2.MqttApplyFailed.String(),
				)
				return ret, err
			}
		case <-timeout:
			err = v1alpha2.NewCOAError(nil, "didn't get response to Apply()-Delete call over MQTT", v1alpha2.InternalError)
			sLog.ErrorfCtx(ctx, "  P (MQTT Target): request timeout - %v", err)
			providerOperationMetrics.ProviderOperationErrors(
				mqtt,
				functionName,
				metrics.ApplyOperation,
				metrics.ApplyOperationType,
				v1alpha2.MqttApplyTimeout.String(),
			)
			return ret, err
		}
	}
	//TODO: Should we remove empty namespaces?
	err = nil
	return ret, nil
}

func (*MQTTTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
		},
	}
}

type TwoComponentSlices struct {
	Current []model.ComponentSpec `json:"current"`
	Desired []model.ComponentSpec `json:"desired"`
}
