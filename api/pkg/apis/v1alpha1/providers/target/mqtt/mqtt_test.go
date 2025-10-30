/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDoubleIni(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Init(config)
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	configMap := map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
		"requestTopic":  "coa-request",
		"responseTopic": "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestInitWithMapInvalidConfig(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	configMap := map[string]string{
		"name": "me",
	}
	provider := MQTTTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
		"requestTopic":  "coa-request",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":           "me",
		"brokerAddress":  "tcp://127.0.0.1:1883",
		"clientID":       "coa-test2",
		"requestTopic":   "coa-request",
		"responseTopic":  "coa-response",
		"timeoutSeconds": "abcd",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":             "me",
		"brokerAddress":    "tcp://127.0.0.1:1883",
		"clientID":         "coa-test2",
		"requestTopic":     "coa-request",
		"responseTopic":    "coa-response",
		"timeoutSeconds":   "2",
		"keepAliveSeconds": "abc",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "me",
		"brokerAddress":      "tcp://127.0.0.1:1883",
		"clientID":           "coa-test2",
		"requestTopic":       "coa-request",
		"responseTopic":      "coa-response",
		"keepAliveSeconds":   "2",
		"pingTimeoutSeconds": "abc",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "me",
		"brokerAddress":      "tcp://127.0.0.1:1883",
		"clientID":           "coa-test2",
		"requestTopic":       "coa-request",
		"responseTopic":      "coa-response",
		"timeoutSeconds":     "2",
		"keepAliveSeconds":   "2",
		"pingTimeoutSeconds": "2",
	}
	err = provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	// Connect with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Connect()
		if tok.Wait() && tok.Error() != nil {
			if attempts == 9 {
				t.Fatalf("failed to connect mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	// Wait until connected
	for i := 0; i < 25 && !c.IsConnected(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if !c.IsConnected() {
		t.Fatalf("mqtt responder not connected")
	}
	// Subscribe with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			var request v1alpha2.COARequest
			err := json.Unmarshal(msg.Payload(), &request)
			assert.Nil(t, err)
			var response v1alpha2.COAResponse
			ret := make([]model.ComponentSpec, 0)
			data, _ := json.Marshal(ret)
			response.State = v1alpha2.OK
			response.Metadata = make(map[string]string)
			response.Metadata["request-id"] = request.Metadata["request-id"]
			response.Body = data
			data, _ = json.Marshal(response)
			token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
			token.Wait()

		})
		if tok.Wait() && tok.Error() != nil {
			if tok.Error().Error() == "subscription exists" {
				break
			}
			if attempts == 9 {
				t.Fatalf("failed to subscribe mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}
func TestGetBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	// Connect with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Connect()
		if tok.Wait() && tok.Error() != nil {
			if attempts == 9 {
				t.Fatalf("failed to connect mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	// Wait until connected
	for i := 0; i < 25 && !c.IsConnected(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if !c.IsConnected() {
		t.Fatalf("mqtt responder not connected")
	}
	// Subscribe with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			var request v1alpha2.COARequest
			err := json.Unmarshal(msg.Payload(), &request)
			assert.Nil(t, err)
			var response v1alpha2.COAResponse
			response.State = v1alpha2.InternalError
			response.Metadata = make(map[string]string)
			response.Metadata["request-id"] = request.Metadata["request-id"]
			response.Body = []byte("didn't get response to Get() call over MQTT")
			data, _ := json.Marshal(response)
			token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
			token.Wait()

		})
		if tok.Wait() && tok.Error() != nil {
			if tok.Error().Error() == "subscription exists" {
				break
			}
			if attempts == 9 {
				t.Fatalf("failed to subscribe mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}

	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.NotNil(t, err)
	assert.Equal(t, "Internal Error: didn't get response to Get() call over MQTT", err.Error())
}
func TestApply(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT enviornment variable is not set")
	}

	const (
		MQTTName          string = "me"
		MQTTBrokerAddress string = "tcp://localhost:1883"
		MQTTClientID      string = "coa-test2"
		MQTTRequestTopic  string = "coa-request"
		MQTTResponseTopic string = "coa-response"

		TestTargetSuccessMessage string = ""
	)

	config := MQTTTargetProviderConfig{
		Name:          MQTTName,
		BrokerAddress: MQTTBrokerAddress,
		ClientID:      MQTTClientID,
		RequestTopic:  MQTTRequestTopic,
		ResponseTopic: MQTTResponseTopic,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	// Connect with simple retry to avoid transient broker readiness issues
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect mqtt responder: %v", token.Error())
	}
	// Subscribe with simple retry, tolerating existing subscription
	token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		summarySpec := model.SummarySpec{
			TargetCount:  1,
			SuccessCount: 1,
			TargetResults: map[string]model.TargetResultSpec{
				"test-target": {
					Status: v1alpha2.OK.String(),
					ComponentResults: map[string]model.ComponentResultSpec{
						"test-component": {
							Status:  v1alpha2.Updated,
							Message: TestTargetSuccessMessage,
						},
					},
				},
			},
			Skipped:             false,
			IsRemoval:           false,
			AllAssignedDeployed: true,
		}
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Body, _ = json.Marshal(summarySpec)
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	})
	if token.Wait() && token.Error() != nil {
		t.Fatalf("failed to subscribe mqtt responder: %v", token.Error())
	}

	deploymentSpec := model.DeploymentSpec{
		SolutionName: "test-solution",
		Solution:     model.SolutionState{},
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-instance",
			},
			Spec: &model.InstanceSpec{
				DisplayName: "test-instance",
				Solution:    "test-solution",
				Target: model.TargetSelector{
					Name: "test-target",
				},
			},
		},
		Targets: map[string]model.TargetState{
			"test-target": {
				Spec: &model.TargetSpec{
					DisplayName: "test-target",
					Components: []model.ComponentSpec{
						{
							Name: "test-component",
							Type: "test-component",
						},
					},
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "test-target",
									Provider: "providers.target.mqtt",
									Config: map[string]string{
										"name":          MQTTName,
										"brokerAddress": MQTTBrokerAddress,
										"clientID":      MQTTClientID,
										"requestTopic":  MQTTRequestTopic,
										"responseTopic": MQTTResponseTopic,
									},
								},
							},
						},
					},
				},
			},
		},
		Assignments: map[string]string{
			"test-component": "{test-target}",
		},
	}

	stepSpec := model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}

	ret, err := provider.Apply(context.Background(), deploymentSpec, stepSpec, false)

	assert.Nil(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, ret["test-component"].Status, v1alpha2.Untouched)
	assert.Equal(t, ret["test-component"].Message, "No error. test-component is untouched")
	assert.Equal(t, ret["test-target"].Status, v1alpha2.Updated)
	assert.Equal(t, ret["test-target"].Message, TestTargetSuccessMessage)
}
func TestApplyBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)

	assert.NotNil(t, err)
}

func TestARemove(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	// Connect with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Connect()
		if tok.Wait() && tok.Error() != nil {
			if attempts == 9 {
				t.Fatalf("failed to connect mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	// Wait until connected
	for i := 0; i < 25 && !c.IsConnected(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if !c.IsConnected() {
		t.Fatalf("mqtt responder not connected")
	}
	// Subscribe with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			var request v1alpha2.COARequest
			err := json.Unmarshal(msg.Payload(), &request)
			assert.Nil(t, err)
			var response v1alpha2.COAResponse
			response.State = v1alpha2.OK
			response.Metadata = make(map[string]string)
			response.Metadata["request-id"] = request.Metadata["request-id"]
			data, _ := json.Marshal(response)
			token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
			token.Wait()

		})
		if tok.Wait() && tok.Error() != nil {
			if tok.Error().Error() == "subscription exists" {
				break
			}
			if attempts == 9 {
				t.Fatalf("failed to subscribe mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)
	assert.Nil(t, err)
}
func TestARemoveBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)

	assert.NotNil(t, err)
}
func TestGetApply(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	// Connect with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Connect()
		if tok.Wait() && tok.Error() != nil {
			if attempts == 9 {
				t.Fatalf("failed to connect mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}
	// Wait until connected
	for i := 0; i < 25 && !c.IsConnected(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if !c.IsConnected() {
		t.Fatalf("mqtt responder not connected")
	}
	// Subscribe with retry
	for attempts := 0; attempts < 10; attempts++ {
		tok := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			var request v1alpha2.COARequest
			json.Unmarshal(msg.Payload(), &request)
			var response v1alpha2.COAResponse
			response.Metadata = make(map[string]string)
			response.Metadata["request-id"] = request.Metadata["request-id"]
			if request.Method == "GET" {
				ret := make([]model.ComponentSpec, 0)
				data, _ := json.Marshal(ret)
				response.State = v1alpha2.OK
				response.Body = data
			} else {
				response.State = v1alpha2.OK
			}

			data, _ := json.Marshal(response)
			token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
			token.Wait()

		})
		if tok.Wait() && tok.Error() != nil {
			if tok.Error().Error() == "subscription exists" {
				break
			}
			if attempts == 9 {
				t.Fatalf("failed to subscribe mqtt responder: %v", tok.Error())
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		break
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))

	err = provider.Init(config)
	assert.Nil(t, err)

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)
	assert.Nil(t, err)
}

func TestLocalApplyGet(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:           "me",
		BrokerAddress:  "tcp://127.0.0.1:1883",
		ClientID:       "coa-test2",
		RequestTopic:   "coa-request",
		ResponseTopic:  "coa-response",
		TimeoutSeconds: 8,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	ctx := context.TODO()
	correlationId := uuid.New().String()
	resourceId := uuid.New().String()
	ctx = contexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)

	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		var request v1alpha2.COARequest
		json.Unmarshal(msg.Payload(), &request)

		assert.NotEqual(t, ctx, request.Context)
		assert.NotNil(t, request.Context)
		diagCtx, ok := request.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		assert.True(t, ok)
		assert.NotNil(t, diagCtx)
		assert.Equal(t, correlationId, diagCtx.GetCorrelationId())
		assert.Equal(t, resourceId, diagCtx.GetResourceId())
		response.Metadata["request-id"] = request.Metadata["request-id"]
		if request.Method == "GET" {
			ret := make([]model.ComponentSpec, 0)
			data, _ := json.Marshal(ret)
			response.State = v1alpha2.OK
			response.Body = data
		} else {
			response.State = v1alpha2.OK
		}
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data)
		token.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(ctx, model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{}, false)
	assert.Nil(t, err)
	arr, err := provider.Get(ctx, model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

func TestInitFailed(t *testing.T) {
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://8.8.8.8:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &MQTTTargetProvider{}
	_ = provider.Init(MQTTTargetProviderConfig{})
	// assert.Nil(t, err) okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}

// --- TLS/mTLS unit tests ---

// generateSelfSignedCert creates a temporary self-signed certificate and key.
// Returns paths to cert and key files and the certificate bytes.
func generateSelfSignedCert(t *testing.T) (string, string, []byte) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	tmpl := x509.Certificate{
		SerialNumber:          bigIntOne(t),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privKey.PublicKey, privKey)
	assert.Nil(t, err)

	// Write cert
	certFile, err := os.CreateTemp("", "mtls-cert-*.pem")
	assert.Nil(t, err)
	defer certFile.Close()
	assert.Nil(t, pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}))

	// Write key
	keyFile, err := os.CreateTemp("", "mtls-key-*.pem")
	assert.Nil(t, err)
	defer keyFile.Close()
	assert.Nil(t, pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}))

	return certFile.Name(), keyFile.Name(), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

func bigIntOne(t *testing.T) *big.Int {
	t.Helper()
	return big.NewInt(1)
}

func TestCreateTLSConfig_InvalidCAPath(t *testing.T) {
	provider := &MQTTTargetProvider{Config: MQTTTargetProviderConfig{
		UseTLS:     true,
		CACertPath: filepath.Join(os.TempDir(), "non-existent-ca.pem"),
	}}
	_, err := provider.createTLSConfig(context.Background())
	assert.NotNil(t, err)
}

func TestCreateTLSConfig_InvalidCAPEM(t *testing.T) {
	caFile, err := os.CreateTemp("", "invalid-ca-*.pem")
	assert.Nil(t, err)
	defer os.Remove(caFile.Name())
	defer caFile.Close()
	_, _ = caFile.Write([]byte("not a pem"))

	provider := &MQTTTargetProvider{Config: MQTTTargetProviderConfig{
		UseTLS:     true,
		CACertPath: caFile.Name(),
	}}
	_, cfgErr := provider.createTLSConfig(context.Background())
	assert.NotNil(t, cfgErr)
}

func TestCreateTLSConfig_ClientCertWithoutKey(t *testing.T) {
	certPath, _, _ := generateSelfSignedCert(t)
	defer os.Remove(certPath)

	provider := &MQTTTargetProvider{Config: MQTTTargetProviderConfig{
		UseTLS:         true,
		ClientCertPath: certPath,
		// missing key path
	}}
	_, err := provider.createTLSConfig(context.Background())
	assert.NotNil(t, err)
}

func TestCreateTLSConfig_ClientCertAndKey_Success(t *testing.T) {
	certPath, keyPath, caBytes := generateSelfSignedCert(t)
	defer os.Remove(certPath)
	defer os.Remove(keyPath)

	// Use the same self-signed cert as CA to exercise RootCAs path
	caFile, err := os.CreateTemp("", "ca-*.pem")
	assert.Nil(t, err)
	defer os.Remove(caFile.Name())
	defer caFile.Close()
	_, _ = caFile.Write(caBytes)

	provider := &MQTTTargetProvider{Config: MQTTTargetProviderConfig{
		UseTLS:         true,
		CACertPath:     caFile.Name(),
		ClientCertPath: certPath,
		ClientKeyPath:  keyPath,
	}}
	cfg, err := provider.createTLSConfig(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, len(cfg.Certificates) == 1)
}

// Optional integration-style test to actually run MQTT with mTLS against a live broker.
// Requires environment variables:
// - TEST_MQTT_MTLS=1 (enables the test)
// - TEST_MQTT_MTLS_BROKER (e.g., ssl://127.0.0.1:8883)
// - TEST_MQTT_MTLS_CA, TEST_MQTT_MTLS_CERT, TEST_MQTT_MTLS_KEY (paths to PEM files)
// - TEST_MQTT_MTLS_REQUEST_TOPIC, TEST_MQTT_MTLS_RESPONSE_TOPIC
func TestGet_mTLS(t *testing.T) {
	if os.Getenv("TEST_MQTT_MTLS") == "" {
		t.Skip("Skipping mTLS test; set TEST_MQTT_MTLS and related env vars to enable")
	}
	broker := os.Getenv("TEST_MQTT_MTLS_BROKER")
	ca := os.Getenv("TEST_MQTT_MTLS_CA")
	cert := os.Getenv("TEST_MQTT_MTLS_CERT")
	key := os.Getenv("TEST_MQTT_MTLS_KEY")
	reqTopic := os.Getenv("TEST_MQTT_MTLS_REQUEST_TOPIC")
	respTopic := os.Getenv("TEST_MQTT_MTLS_RESPONSE_TOPIC")
	if broker == "" || ca == "" || cert == "" || key == "" || reqTopic == "" || respTopic == "" {
		t.Skip("Skipping mTLS test; missing required TEST_MQTT_MTLS_* env vars")
	}

	provider := &MQTTTargetProvider{}
	err := provider.Init(MQTTTargetProviderConfig{
		Name:           "mtls-test",
		BrokerAddress:  broker,
		ClientID:       "mtls-provider",
		RequestTopic:   reqTopic,
		ResponseTopic:  respTopic,
		UseTLS:         true,
		CACertPath:     ca,
		ClientCertPath: cert,
		ClientKeyPath:  key,
	})
	assert.Nil(t, err)

	// Separate client to respond to requests, also using mTLS
	respTLS := newTLSConfigFromFiles(t, ca, cert, key)

	opts := gmqtt.NewClientOptions().AddBroker(broker).SetClientID("mtls-responder")
	opts.SetTLSConfig(respTLS)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect mtls responder: %v", token.Error())
	}
	if token := c.Subscribe(reqTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		_ = json.Unmarshal(msg.Payload(), &request)
		var response v1alpha2.COAResponse
		ret := make([]model.ComponentSpec, 0)
		data, _ := json.Marshal(ret)
		response.State = v1alpha2.OK
		response.Metadata = map[string]string{"request-id": request.Metadata["request-id"]}
		response.Body = data
		data, _ = json.Marshal(response)
		tok := c.Publish(respTopic, 0, false, data)
		tok.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			t.Fatalf("subscribe failed: %v", token.Error())
		}
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{Instance: model.InstanceState{Spec: &model.InstanceSpec{}}}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

// TLS server-auth only (no client cert). Requires a TLS listener without mTLS on the broker.
// Env vars:
// - TEST_MQTT_TLS=1 (enables the test)
// - TEST_MQTT_TLS_BROKER (e.g., ssl://127.0.0.1:8883)
// - TEST_MQTT_TLS_CA (path to broker CA cert)
// - TEST_MQTT_TLS_REQUEST_TOPIC, TEST_MQTT_TLS_RESPONSE_TOPIC
func TestGet_TLS(t *testing.T) {
	if os.Getenv("TEST_MQTT_TLS") == "" {
		t.Skip("Skipping TLS test; set TEST_MQTT_TLS and related env vars to enable")
	}
	broker := os.Getenv("TEST_MQTT_TLS_BROKER")
	ca := os.Getenv("TEST_MQTT_TLS_CA")
	reqTopic := os.Getenv("TEST_MQTT_TLS_REQUEST_TOPIC")
	respTopic := os.Getenv("TEST_MQTT_TLS_RESPONSE_TOPIC")
	if broker == "" || ca == "" || reqTopic == "" || respTopic == "" {
		t.Skip("Skipping TLS test; missing required TEST_MQTT_TLS_* env vars")
	}

	provider := &MQTTTargetProvider{}
	err := provider.Init(MQTTTargetProviderConfig{
		Name:          "tls-test",
		BrokerAddress: broker,
		ClientID:      "tls-provider",
		RequestTopic:  reqTopic,
		ResponseTopic: respTopic,
		UseTLS:        true,
		CACertPath:    ca,
	})
	assert.Nil(t, err)

	// TLS responder without client certificate
	caBytes, err := os.ReadFile(ca)
	assert.Nil(t, err)
	pool := x509.NewCertPool()
	assert.True(t, pool.AppendCertsFromPEM(caBytes))
	tlsCfg := &tls.Config{RootCAs: pool}

	opts := gmqtt.NewClientOptions().AddBroker(broker).SetClientID("tls-responder")
	opts.SetTLSConfig(tlsCfg)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetAutoReconnect(false)
	opts.SetPingTimeout(10 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect tls responder: %v", token.Error())
	}
	if token := c.Subscribe(reqTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		_ = json.Unmarshal(msg.Payload(), &request)
		var response v1alpha2.COAResponse
		ret := make([]model.ComponentSpec, 0)
		data, _ := json.Marshal(ret)
		response.State = v1alpha2.OK
		response.Metadata = map[string]string{"request-id": request.Metadata["request-id"]}
		response.Body = data
		data, _ = json.Marshal(response)
		tok := c.Publish(respTopic, 0, false, data)
		tok.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			t.Fatalf("subscribe failed: %v", token.Error())
		}
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{Instance: model.InstanceState{Spec: &model.InstanceSpec{}}}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

func newTLSConfigFromFiles(t *testing.T, caPath, certPath, keyPath string) *tls.Config {
	t.Helper()
	caBytes, err := os.ReadFile(caPath)
	assert.Nil(t, err)
	pool := x509.NewCertPool()
	assert.True(t, pool.AppendCertsFromPEM(caBytes))

	crt, err := tls.LoadX509KeyPair(certPath, keyPath)
	assert.Nil(t, err)

	return &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{crt}}
}
