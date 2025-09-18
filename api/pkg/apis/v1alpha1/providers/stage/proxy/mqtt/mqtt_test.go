/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestInitWithMap(t *testing.T) {
	config := MQTTProxyStageProviderConfig{}
	provider := MQTTProxyStageProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
}
func TestSuccessfulProcess(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT enviornment variable is not set")
	}

	config := MQTTProxyStageProviderConfig{}
	provider := MQTTProxyStageProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	proxyProperties := MQTTProxyProperties{
		BrokerAddress:      "tcp://localhost:1883",
		RequestTopic:       "test-request",
		ResponseTopic:      "test-response",
		TimeoutSeconds:     50,
		KeepAliveSeconds:   2,
		PingTimeoutSeconds: 1,
	}

	opts := gmqtt.NewClientOptions().AddBroker(proxyProperties.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(time.Duration(proxyProperties.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(proxyProperties.PingTimeoutSeconds) * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe(proxyProperties.RequestTopic, 1, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		json.Unmarshal(msg.Payload(), &request)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		state := model.StageStatus{
			Status: v1alpha2.Done,
			Outputs: map[string]interface{}{
				"foo": "bar",
			},
		}
		stateData, _ := json.Marshal(state)
		response.Body = stateData
		data, _ := json.Marshal(response)
		if token := client.Publish(proxyProperties.ResponseTopic, 1, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			panic(err)
		}

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	result, paused, err := provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"brokerAddress": proxyProperties.BrokerAddress,
				"requestTopic":  proxyProperties.RequestTopic,
				"responseTopic": proxyProperties.ResponseTopic,
				"timeout":       5,
			},
		},
	})
	assert.Nil(t, err)
	assert.False(t, paused)
	assert.NotNil(t, result)
	assert.Equal(t, "bar", result["foo"])
}
func TestFailedProcess(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT enviornment variable is not set")
	}

	config := MQTTProxyStageProviderConfig{}
	provider := MQTTProxyStageProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	proxyProperties := MQTTProxyProperties{
		BrokerAddress:      "tcp://localhost:1883",
		RequestTopic:       "test-request",
		ResponseTopic:      "test-response",
		TimeoutSeconds:     50,
		KeepAliveSeconds:   2,
		PingTimeoutSeconds: 1,
	}

	opts := gmqtt.NewClientOptions().AddBroker(proxyProperties.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(time.Duration(proxyProperties.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(proxyProperties.PingTimeoutSeconds) * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe(proxyProperties.RequestTopic, 1, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		json.Unmarshal(msg.Payload(), &request)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		state := model.StageStatus{
			Status: v1alpha2.Done,
			Outputs: map[string]interface{}{
				"foo": "bar",
			},
		}
		stateData, _ := json.Marshal(state)
		response.Body = stateData
		data, _ := json.Marshal(response)
		if token := client.Publish(proxyProperties.ResponseTopic, 1, false, data); token.Wait() && token.Error() != nil {
			err = token.Error()
			panic(err)
		}

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, _, err = provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"brokerAddress": proxyProperties.BrokerAddress,
				"requestTopic":  proxyProperties.RequestTopic,
				"responseTopic": proxyProperties.ResponseTopic,
				"timeout":       5,
			},
		},
	})
	assert.Equal(t, err.(v1alpha2.COAError).State, v1alpha2.InternalError)
}
func TestNoProcessor(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT enviornment variable is not set")
	}

	config := MQTTProxyStageProviderConfig{}
	provider := MQTTProxyStageProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	proxyProperties := MQTTProxyProperties{
		BrokerAddress:      "tcp://localhost:1883",
		RequestTopic:       "test-request",
		ResponseTopic:      "test-response",
		TimeoutSeconds:     5,
		KeepAliveSeconds:   2,
		PingTimeoutSeconds: 1,
	}

	_, _, err = provider.Process(context.TODO(), contexts.ManagerContext{}, v1alpha2.ActivationData{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
		Proxy: &v1alpha2.ProxySpec{
			Config: map[string]interface{}{
				"brokerAddress": proxyProperties.BrokerAddress,
				"requestTopic":  proxyProperties.RequestTopic,
				"responseTopic": proxyProperties.ResponseTopic,
				"timeout":       5,
			},
		},
	})
	assert.Equal(t, err.(v1alpha2.COAError).State, v1alpha2.InternalError)
}
