/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logger.NewLogger("coa.runtime")

type MQTTBindingConfig struct {
	BrokerAddress      string `json:"brokerAddress"`
	ClientID           string `json:"clientID"`
	RequestTopic       string `json:"requestTopic"`
	ResponseTopic      string `json:"responseTopic"`
	TimeoutSeconds     int    `json:"timeoutSeconds,omitempty"`
	KeepAliveSeconds   int    `json:"keepAliveSeconds,omitempty"`
	PingTimeoutSeconds int    `json:"pingTimeoutSeconds,omitempty"`
	// TLS/Certificate configuration fields
	UseTLS             string `json:"useTLS,omitempty"`
	CACertPath         string `json:"caCertPath,omitempty"`
	ClientCertPath     string `json:"clientCertPath,omitempty"`
	ClientKeyPath      string `json:"clientKeyPath,omitempty"`
	InsecureSkipVerify string `json:"insecureSkipVerify,omitempty"`
}

type MQTTBinding struct {
	MQTTClient      gmqtt.Client
	subscribedTopic map[string]struct{}
	lock            sync.RWMutex
	Handler         gmqtt.MessageHandler
	config          MQTTBindingConfig
}

var routeTable map[string]v1alpha2.Endpoint

func (m *MQTTBinding) createHandlerWithResponseTopic(responseTopic string) gmqtt.MessageHandler {
	return func(client gmqtt.Client, msg gmqtt.Message) {
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
			response = routeTable[request.Route].Handler(request)
		}

		// needs to carry request-id from request into response
		if request.Metadata != nil {
			if v, ok := request.Metadata["request-id"]; ok {
				if response.Metadata == nil {
					response.Metadata = make(map[string]string)
				}
				response.Metadata["request-id"] = v
			}
		}

		data, _ := json.Marshal(response)

		go func() {
			if token := client.Publish(responseTopic, 0, false, data); token.Wait() && token.Error() != nil {
				log.Errorf("failed to handle request from MQTT: %s", token.Error())
			}
		}()
	}
}

func (m *MQTTBinding) defaultHandler(client gmqtt.Client, msg gmqtt.Message) {
	m.createHandlerWithResponseTopic(m.config.ResponseTopic)(client, msg)
}

// todo improve this function to handle more complex request/response patterns
func (m *MQTTBinding) generateResponseTopic(requestTopic string) string {
	// if requestTopic matches the expected format, generate a response topic
	if strings.HasPrefix(requestTopic, "symphony/request/") {
		targetName := strings.TrimPrefix(requestTopic, "symphony/request/")
		return "symphony/response/" + targetName
	}
	// If it doesn't match the expected format, return the default response topic
	if m.config.ResponseTopic != "" {
		log.Infof("MQTT Binding: request topic '%s' does not match expected format, using default response topic '%s'", requestTopic, m.config.ResponseTopic)
		return m.config.ResponseTopic
	}
	// If no default response topic is configured, log an error and return empty string
	log.Errorf("MQTT Binding: cannot generate response topic for '%s' - no default response topic configured and topic doesn't match 'symphony/request/' pattern", requestTopic)
	return ""
}

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

	// Set default values
	if config.KeepAliveSeconds <= 0 {
		config.KeepAliveSeconds = 2
	}
	if config.PingTimeoutSeconds <= 0 {
		config.PingTimeoutSeconds = 1
	}

	m.config = config

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID(config.ClientID)
	opts.SetKeepAlive(time.Duration(config.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(config.PingTimeoutSeconds) * time.Second)
	opts.CleanSession = false

	// Configure TLS if enabled
	if strings.ToLower(config.UseTLS) == "true" {
		tlsConfig, err := m.createTLSConfig(config)
		if err != nil {
			log.Errorf("MQTT Binding: failed to create TLS config - %+v", err)
			return v1alpha2.NewCOAError(err, "failed to create TLS config", v1alpha2.InternalError)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	m.MQTTClient = gmqtt.NewClient(opts)
	if token := m.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		connErr := token.Error()
		log.Errorf("MQTT Binding: failed to connect to MQTT broker - %+v", connErr)
		return v1alpha2.NewCOAError(connErr, "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	m.Handler = m.defaultHandler
	if config.RequestTopic != "" && config.ResponseTopic != "" {
		if token := m.MQTTClient.Subscribe(config.RequestTopic, 0, m.Handler); token.Wait() && token.Error() != nil {
			if token.Error().Error() != "subscription exists" {
				log.Errorf("  P (MQTT Target): failed to connect to subscribe to request topic - %+v", token.Error())
				return v1alpha2.NewCOAError(token.Error(), "failed to subscribe to request topic", v1alpha2.InternalError)
			}
		}
	}
	return nil
}

// createTLSConfig creates a TLS configuration for MQTT client authentication
func (m *MQTTBinding) createTLSConfig(config MQTTBindingConfig) (*tls.Config, error) {
	insecureSkipVerify := strings.ToLower(config.InsecureSkipVerify) == "true"

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Load CA certificate if provided
	if config.CACertPath != "" {
		log.Infof("MQTT Binding: attempting to load CA certificate from %s", config.CACertPath)

		caCertPool, err := bindings.LoadCACertPool(config.CACertPath)
		if err != nil {
			log.Errorf("MQTT Binding: failed to load/validate CA certificate - %+v", err)
			return nil, err
		}
		tlsConfig.RootCAs = caCertPool
		log.Infof("MQTT Binding: successfully loaded CA certificate from %s", config.CACertPath)
	} else {
		if !insecureSkipVerify {
			log.Warn("MQTT Binding: no CA certificate path provided - using system CA pool. If connection fails with 'certificate signed by unknown authority', either provide a CA certificate or set insecureSkipVerify to true")
		} else {
			log.Infof("MQTT Binding: TLS certificate verification disabled (insecureSkipVerify=true)")
		}
	}

	// Load client certificate and key if provided
	if config.ClientCertPath != "" && config.ClientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientKeyPath)
		if err != nil {
			log.Errorf("MQTT Binding: failed to load client certificate and key - %+v", err)
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		log.Infof("MQTT Binding: loaded client certificate from %s and key from %s",
			config.ClientCertPath, config.ClientKeyPath)
	} else if config.ClientCertPath != "" || config.ClientKeyPath != "" {
		// Both cert and key must be provided together
		return nil, fmt.Errorf("both clientCertPath and clientKeyPath must be provided for client certificate authentication")
	}

	// Set ServerName for proper certificate validation
	if !insecureSkipVerify && config.BrokerAddress != "" {
		host := config.BrokerAddress
		// Remove protocol prefix if present
		if strings.HasPrefix(host, "ssl://") || strings.HasPrefix(host, "tls://") {
			host = strings.SplitN(host, "://", 2)[1]
		}
		// Remove port if present
		if idx := strings.Index(host, ":"); idx > 0 {
			host = host[:idx]
		}
		tlsConfig.ServerName = host
	}

	return tlsConfig, nil
}

// SubscribeTopic
func (m *MQTTBinding) SubscribeTopic(topic string) error {
	if m.IsSubscribed(topic) {
		return nil
	}

	// Need to subscribe, acquire write lock
	m.lock.Lock()
	defer m.lock.Unlock()

	// Double-check after acquiring write lock (in case another goroutine already subscribed)
	if m.subscribedTopic == nil {
		m.subscribedTopic = make(map[string]struct{})
	}
	if _, ok := m.subscribedTopic[topic]; ok {
		return nil
	}

	log.Infof("MQTT Binding: subscribing to topic %s", topic)

	// generate response topic based on request topic
	responseTopic := m.generateResponseTopic(topic)
	if responseTopic == "" {
		log.Errorf("MQTT Binding: no response topic generated for request topic %s, cannot subscribe", topic)
		return fmt.Errorf("no response topic available for request topic %s", topic)
	}
	handler := m.createHandlerWithResponseTopic(responseTopic)

	token := m.MQTTClient.Subscribe(topic, 0, handler)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	m.subscribedTopic[topic] = struct{}{}
	return nil
}

// IsSubscribed returns true if the topic is already subscribed (thread-safe, read-only operation)
func (m *MQTTBinding) IsSubscribed(topic string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.subscribedTopic == nil {
		return false
	}
	_, ok := m.subscribedTopic[topic]
	return ok
}

// UnsubscribeTopic
func (m *MQTTBinding) UnsubscribeTopic(topic string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.subscribedTopic == nil {
		return nil
	}
	if _, ok := m.subscribedTopic[topic]; !ok {
		return nil
	}
	token := m.MQTTClient.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	delete(m.subscribedTopic, topic)
	return nil
}

// Shutdown stops the MQTT binding
func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	m.MQTTClient.Disconnect(1000)
	return nil
}
