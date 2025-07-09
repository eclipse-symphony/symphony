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
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/states/k8s"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/localfile"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logger.NewLogger("coa.runtime")
var (
	// SERVER_CA_FILE is the file containing the CA cert that signed the MQTT broker's certificate
	ServerCAFile = os.Getenv("REMOTE_CA_FILE") // Rename in code but keep env var name for backward compatibility
	MQTTEnabled  = os.Getenv("MQTT_ENABLED")
)

type CertProviderConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type MQTTBindingConfig struct {
	BrokerAddress string             `json:"brokerAddress"`
	ClientID      string             `json:"clientID"`
	RequestTopic  string             `json:"requestTopic,omitempty"`
	ResponseTopic string             `json:"responseTopic,omitempty"`
	CACert        string             `json:"caCert,omitempty"`
	ClientCert    string             `json:"clientCert,omitempty"`
	ClientKey     string             `json:"clientKey,omitempty"`
	CertProvider  CertProviderConfig `json:"certProvider,omitempty"`
	TLS           bool               `json:"tls"`
	TopicPrefix   string             `json:"topicPrefix,omitempty"`
	StateProvider struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	} `json:"stateProvider,omitempty"`
}

type MQTTBinding struct {
	MQTTClient        gmqtt.Client
	CertProvider      certs.ICertProvider
	subscribedTargets sync.Map
	config            MQTTBindingConfig
	mu                sync.Mutex
	stateProvider     states.IStateProvider // Added for querying targets
	enabled           bool                  // Whether MQTT is enabled
}

var routeTable map[string]v1alpha2.Endpoint

// SetStateProvider allows setting a state provider for target discovery
func (m *MQTTBinding) SetStateProvider(provider states.IStateProvider) {
	m.stateProvider = provider
	log.Info("State provider set for MQTT binding")
}

func (m *MQTTBinding) Launch(config MQTTBindingConfig, endpoints []v1alpha2.Endpoint) error {
	// Check if MQTT is enabled via environment variable
	log.InfoCtx(context.Background(), "Checking MQTT binding enabled status...")
	m.enabled = false
	if MQTTEnabled == "true" {
		m.enabled = true
		log.Info("MQTT binding is enabled via MQTT_ENABLED=true")
	} else {
		log.Info("MQTT binding is disabled via environment variable (MQTT_ENABLED is not 'true')")
		return nil // Skip initialization if MQTT is disabled
	}

	m.config = config

	if config.StateProvider.Type != "" {
		log.Info("Initializing state provider from config for MQTT binding")
		var stateProvider states.IStateProvider

		switch config.StateProvider.Type {
		case "providers.state.k8s":
			k8sProvider := &k8s.K8sStateProvider{}
			err := k8sProvider.Init(config.StateProvider.Config)
			if err != nil {
				log.Errorf("Failed to initialize K8s state provider: %v", err)
				return err
			}
			stateProvider = k8sProvider
		default:
			log.Errorf("Unsupported state provider type: %s", config.StateProvider.Type)
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported state provider type: %s", config.StateProvider.Type), v1alpha2.BadConfig)
		}

		m.stateProvider = stateProvider
		log.Info("State provider initialized for MQTT binding")
	}

	routeTable = make(map[string]v1alpha2.Endpoint)
	for _, endpoint := range endpoints {
		route := endpoint.Route
		lastSlash := strings.LastIndex(endpoint.Route, "/")
		if lastSlash > 0 {
			route = strings.TrimPrefix(route, route[:lastSlash+1])
		}
		routeTable[route] = endpoint
		log.InfofCtx(context.Background(), "Registering route: %s Endpoint: %s", route, endpoint)
	}

	cert := tls.Certificate{}
	caCertPool := x509.NewCertPool()
	if config.TLS {
		switch config.CertProvider.Type {
		case "certs.autogen":
			m.CertProvider = &autogen.AutoGenCertProvider{}
		case "certs.localfile":
			m.CertProvider = &localfile.LocalCertFileProvider{}
			localConfig := &localfile.LocalCertFileProviderConfig{}
			data, err := json.Marshal(config.CertProvider.Config)
			if err != nil {
				log.Errorf("B (HTTP): failed to marshall config %+v", err)
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("B (HTTP): failed to marshall config"), v1alpha2.BadConfig)
			}
			err = json.Unmarshal(data, &localConfig)
			if err != nil {
				log.Errorf("B (HTTP): failed to unmarshall config %+v", err)
				return v1alpha2.NewCOAError(nil, fmt.Sprintf("B (HTTP): failed to unmarshall config"), v1alpha2.BadConfig)
			}

			cert, err = tls.LoadX509KeyPair(localConfig.CertFile, localConfig.KeyFile)
			if err != nil {
				log.Errorf("B (HTTP): failed to load cert/key pair: %v", err)
				return v1alpha2.NewCOAError(err, fmt.Sprintf("B (MQTT): failed to load cert/key pair"), v1alpha2.BadConfig)

			}
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("cert provider type '%s' is not recognized", config.CertProvider.Type), v1alpha2.BadConfig)
		}
		err := m.CertProvider.Init(config.CertProvider.Config)
		if err != nil {
			return err
		}
	}
	log.InfoCtx(context.Background(), "MQTT binding is launching...")

	// Load the MQTT server's CA certificate to validate the server
	if ServerCAFile != "" {
		log.Info(fmt.Sprintf("Loading MQTT server's CA certificate from: %s", ServerCAFile))
		pemData, err := ioutil.ReadFile(ServerCAFile)
		if err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("MQTT server CA file '%s' could not be read", ServerCAFile), v1alpha2.BadConfig)
		}
		if ok := caCertPool.AppendCertsFromPEM(pemData); !ok {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Failed to append MQTT server CA cert from file %s", ServerCAFile), v1alpha2.BadConfig)
		}
		log.Info("Successfully loaded MQTT server's CA certificate")
	} else {
		log.Warn("No MQTT server CA certificate provided. TLS verification may fail if self-signed certificates are used.")
	}

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID(config.ClientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.CleanSession = false

	// Configure TLS with the loaded certificates
	serverHost := config.BrokerAddress
	if strings.HasPrefix(serverHost, "tcp://") {
		serverHost = strings.TrimPrefix(serverHost, "tcp://")
	}
	if idx := strings.Index(serverHost, ":"); idx > 0 {
		serverHost = serverHost[:idx]
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert}, // Client certificate for client authentication
		RootCAs:      caCertPool,              // Server CA certificate to validate the MQTT broker
		ServerName:   serverHost,              // Must match the MQTT broker's certificate CN/SAN (no port)
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}
	opts.SetTLSConfig(tlsConfig)
	m.MQTTClient = gmqtt.NewClient(opts)
	if token := m.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	} else {
		log.Infof("Successfully connected to MQTT broker: %s", config.BrokerAddress)
	}

	// Set default topic prefix if not specified
	if m.config.TopicPrefix == "" {
		m.config.TopicPrefix = "symphony"
	}

	// Handle legacy single topic configuration
	if config.RequestTopic != "" && config.ResponseTopic != "" {
		if token := m.MQTTClient.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			// ...existing message handler...
			var request v1alpha2.COARequest
			var response v1alpha2.COAResponse
			if request.Context == nil {
				request.Context = context.TODO()
			}
			contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
			err := json.Unmarshal(msg.Payload(), &request)
			if err != nil {
				log.Errorf("Failed to parse COARequest from MQTT message: %v", err)
				response = v1alpha2.COAResponse{
					State:       v1alpha2.BadRequest,
					ContentType: "application/json",
					Body:        []byte(fmt.Sprintf("Failed to parse COARequest: %s", err.Error())),
				}
			}
			response = routeTable[request.Route].Handler(request)
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
				if token := client.Publish(config.ResponseTopic, 0, false, data); token.Wait() && token.Error() != nil {
					log.Errorf("failed to handle request from MOTT: %s", token.Error())
				}
			}()
		}); token.Wait() && token.Error() != nil {
			log.Errorf("Failed to subscribe to request topic '%s': %v", config.RequestTopic, token.Error())
			return v1alpha2.NewCOAError(token.Error(), fmt.Sprintf("failed to subscribe to request topic '%s'", config.RequestTopic), v1alpha2.InternalError)
		}
	}
	return nil
}

func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	// Shutdown stops the MQTT binding
	if m.enabled && m.MQTTClient != nil {
		m.MQTTClient.Disconnect(1000)
	}
	return nil
}
