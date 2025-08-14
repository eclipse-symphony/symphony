/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logger.NewLogger("coa.runtime")

type MQTTBindingConfig struct {
	BrokerAddress string `json:"brokerAddress"`
	ClientID      string `json:"clientID"`
	RequestTopic  string `json:"requestTopic"`
	ResponseTopic string `json:"responseTopic"`
}

type MQTTBinding struct {
	MQTTClient gmqtt.Client
}

var routeTable map[string]v1alpha2.Endpoint

func (m *MQTTBinding) Launch(config MQTTBindingConfig, endpoints []v1alpha2.Endpoint) error {
	routeTable = make(map[string]v1alpha2.Endpoint)
	for _, endpoint := range endpoints {
		route := endpoint.Route
		lastSlash := strings.LastIndex(endpoint.Route, "/")
		if lastSlash > 0 {
			route = strings.TrimPrefix(route, route[:lastSlash+1])
		}
		routeTable[route] = endpoint
	}

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID(config.ClientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.CleanSession = false
	m.MQTTClient = gmqtt.NewClient(opts)
	if token := m.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	if token := m.MQTTClient.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		var response v1alpha2.COAResponse
		if request.Context == nil {
			request.Context = context.TODO()
		}
		// patch correlation id if missing
		contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
		err := json.Unmarshal(msg.Payload(), &request)
		if err != nil {
			response = v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				ContentType: "text/plain",
				Body:        []byte(err.Error()),
			}
		} else {
			//check if the route is in the route table
			if _, ok := routeTable[request.Route]; !ok {
				response = v1alpha2.COAResponse{
					State:       v1alpha2.NotFound,
					ContentType: "text/plain",
					Body:        []byte("route not found"),
				}
			} else {
				response = routeTable[request.Route].Handler(request)
			}
		}

		// needs to carry request-id from request into response
		if request.Metadata != nil {
			if v, ok := request.Metadata["request-id"]; ok {
				if response.Metadata == nil {
					response.Metadata = make(map[string]string)
				}
				response.Metadata["request-id"] = v
			}
			if v, ok := request.Metadata["request-id"]; ok {
				if response.Metadata == nil {
					response.Metadata = make(map[string]string)
				}
				response.Metadata["request-id"] = v
			}
		}

		data, _ := json.Marshal(response)

		go func() {
			if token := client.Publish(config.ResponseTopic, 0, false, data); token.Wait() && token.Error() != nil {
				log.Errorf("failed to handle request from MOTT: %s", token.Error())
			}
		}()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			log.Errorf("  P (MQTT Target): faild to connect to subscribe to request topic - %+v", token.Error())
			return v1alpha2.NewCOAError(token.Error(), "failed to subscribe to request topic", v1alpha2.InternalError)
		}
	}

	return nil
}

// Shutdown stops the MQTT binding
func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	m.MQTTClient.Disconnect(1000)
	return nil
}
