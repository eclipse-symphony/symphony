/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package mqtt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/logger"
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
		request.Context = context.Background()
		err := json.Unmarshal(msg.Payload(), &request)
		if err != nil {
			response = v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				ContentType: "application/text",
				Body:        []byte(err.Error()),
			}
		} else {
			response = routeTable[request.Route].Handler(request)
		}

		// needs to carry call-context from request into response
		if request.Metadata != nil {
			if v, ok := request.Metadata["call-context"]; ok {
				if response.Metadata == nil {
					response.Metadata = make(map[string]string)
				}
				response.Metadata["call-context"] = v
			}
		}

		data, _ := json.Marshal(response)

		if token := client.Publish(config.ResponseTopic, 0, false, data); token.Wait() && token.Error() != nil {
			log.Errorf("failed to handle request from MOTT: %s", token.Error())
		}
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			log.Errorf("  P (MQTT Target): faild to connect to subscribe to request topic - %+v", token.Error())
			return v1alpha2.NewCOAError(token.Error(), "failed to subscribe to request topic", v1alpha2.InternalError)
		}
	}

	return nil
}
