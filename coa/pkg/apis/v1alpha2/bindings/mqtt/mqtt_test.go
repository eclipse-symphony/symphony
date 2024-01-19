/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestMQTTEcho(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	sig := make(chan int)
	config := MQTTBindingConfig{
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	binding := MQTTBinding{}
	endpoints := []v1alpha2.Endpoint{
		{
			Methods: []string{"GET"},
			Route:   "greetings",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				return v1alpha2.COAResponse{
					Body: []byte("Hi there!!"),
				}
			},
		},
	}
	err := binding.Launch(config, endpoints)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	c := gmqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe(config.ResponseTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var response v1alpha2.COAResponse
		err := json.Unmarshal(msg.Payload(), &response)
		assert.Nil(t, err)
		assert.Equal(t, string(response.Body), "Hi there!!")
		sig <- 1
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}
	request := v1alpha2.COARequest{
		Route:  "greetings",
		Method: "GET",
	}
	data, _ := json.Marshal(request)
	token := c.Publish(config.RequestTopic, 0, false, data) //sending COARequest directly doesn't seem to work
	token.Wait()
	<-sig
}
