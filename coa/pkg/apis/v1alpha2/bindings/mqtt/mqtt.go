/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
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
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
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

	// Set default values
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 8
	}
	if config.KeepAliveSeconds <= 0 {
		config.KeepAliveSeconds = 2
	}
	if config.PingTimeoutSeconds <= 0 {
		config.PingTimeoutSeconds = 1
	}

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID(config.ClientID)
	opts.SetKeepAlive(time.Duration(config.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(config.PingTimeoutSeconds) * time.Second)
	if config.TimeoutSeconds > 0 {
		timeout := time.Duration(config.TimeoutSeconds) * time.Second
		opts.SetConnectTimeout(timeout)
		opts.SetWriteTimeout(timeout)
	}
	opts.CleanSession = false

	// Configure authentication
	if config.Username != "" {
		opts.SetUsername(config.Username)
	}
	if config.Password != "" {
		opts.SetPassword(config.Password)
	}

	// Configure TLS if enabled
	if config.UseTLS == "true" {
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
		
		// Provide specific guidance for common TLS errors
		if strings.Contains(connErr.Error(), "certificate signed by unknown authority") {
			log.Errorf("MQTT Binding: TLS certificate verification failed. Common solutions:")
			log.Errorf("MQTT Binding: 1. Set 'caCertPath' to the path of your broker's CA certificate")
			log.Errorf("MQTT Binding: 2. Set 'insecureSkipVerify' to 'true' for testing (not recommended for production)")
			log.Errorf("MQTT Binding: 3. Ensure your broker certificate is issued by a trusted CA")
		} else if strings.Contains(connErr.Error(), "tls:") {
			log.Errorf("MQTT Binding: TLS connection error. Check your TLS configuration:")
			log.Errorf("MQTT Binding: - Broker address should use 'ssl://' or 'tls://' prefix for TLS connections")
			log.Errorf("MQTT Binding: - Verify CA certificate path and format")
			log.Errorf("MQTT Binding: - Check client certificate and key paths if using mutual TLS")
		}
		
		return v1alpha2.NewCOAError(connErr, "failed to connect to MQTT broker", v1alpha2.InternalError)
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

// createTLSConfig creates a TLS configuration for MQTT client authentication
func (m *MQTTBinding) createTLSConfig(config MQTTBindingConfig) (*tls.Config, error) {
	insecureSkipVerify := config.InsecureSkipVerify == "true"
	
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Load CA certificate if provided
	if config.CACertPath != "" {
		log.Infof("MQTT Binding: attempting to load CA certificate from %s", config.CACertPath)

		caCert, err := os.ReadFile(config.CACertPath)
		if err != nil {
			log.Errorf("MQTT Binding: failed to read CA certificate - %+v", err)
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		// Verify the CA cert content
		log.Infof("MQTT Binding: CA certificate file size: %d bytes", len(caCert))
		if len(caCert) == 0 {
			return nil, fmt.Errorf("CA certificate file is empty")
		}

		// Validate that the file contains valid PEM data
		if !isCertificatePEM(caCert) {
			log.Errorf("MQTT Binding: CA certificate file does not contain valid PEM data")
			return nil, fmt.Errorf("CA certificate file does not contain valid PEM data")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			log.Errorf("MQTT Binding: failed to parse CA certificate - invalid PEM format or corrupted certificate")
			return nil, fmt.Errorf("failed to parse CA certificate - invalid PEM format or corrupted certificate")
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

	return tlsConfig, nil
}

// isCertificatePEM checks if the given data contains valid PEM formatted certificate data
func isCertificatePEM(data []byte) bool {
	// Check if the data contains PEM headers
	dataStr := string(data)
	if !strings.Contains(dataStr, "-----BEGIN CERTIFICATE-----") || 
	   !strings.Contains(dataStr, "-----END CERTIFICATE-----") {
		return false
	}
	
	// Try to decode the PEM block
	block, _ := pem.Decode(data)
	return block != nil && block.Type == "CERTIFICATE"
}

// Shutdown stops the MQTT binding
func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	m.MQTTClient.Disconnect(1000)
	return nil
}
