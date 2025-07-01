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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/autogen"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/certs/localfile"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
)

var log = logger.NewLogger("coa.runtime")
var (
	ClientCAFile = os.Getenv("CLIENT_CA_FILE")
)

// CertProviderConfig 用于兼容 http.go 的 certProvider 配置结构
// 只实现 certs.localfile

type CertProviderConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type MQTTBindingConfig struct {
	BrokerAddress  string             `json:"brokerAddress"`
	ClientID       string             `json:"clientID"`
	RequestTopic   string             `json:"requestTopic"`
	ResponseTopic  string             `json:"responseTopic"`
	RequestTopics  map[string]string  `json:"requestTopics,omitempty"` // 新增，支持多个client
	ResponseTopics map[string]string  `json:"responseTopics,omitempty"`
	CACert         string             `json:"caCert,omitempty"`
	ClientCert     string             `json:"clientCert,omitempty"`
	ClientKey      string             `json:"clientKey,omitempty"`
	CertProvider   CertProviderConfig `json:"certProvider,omitempty"`
	TLS            bool               `json:"tls"`
}

type MQTTBinding struct {
	MQTTClient   gmqtt.Client
	CertProvider certs.ICertProvider
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
	if ClientCAFile != "" {
		log.Info(fmt.Sprintf("Loading client CA file: %s", ClientCAFile))
		pemData, err := ioutil.ReadFile(ClientCAFile)
		if err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Client CA file '%s' is not read successfully", ClientCAFile), v1alpha2.BadConfig)
		}
		log.InfofCtx(context.Background(), "Loaded CA PEM data (first 200 bytes): %s", string(pemData[:min(200, len(pemData))]))
		certs, err := m.parseCertificates(pemData)
		if err != nil {
			log.Errorf("Failed to parse CA certificates: %v", err)
		} else {
			for i, cert := range certs {
				log.InfofCtx(context.Background(), "CA Cert[%d] Subject: %s, Issuer: %s, NotAfter: %s", i, cert.Subject.String(), cert.Issuer.String(), cert.NotAfter)
			}
		}
		if ok := caCertPool.AppendCertsFromPEM(pemData); !ok {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Failed to append CA cert from file, %s", ClientCAFile), v1alpha2.BadConfig)
		}
	}
	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID(config.ClientID)
	opts.SetKeepAlive(200 * time.Second)
	opts.SetPingTimeout(100 * time.Second)
	opts.CleanSession = false
	// TLS config end
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "10.172.3.39", // 必须和证书CN/SAN一致
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}
	opts.SetTLSConfig(tlsConfig)
	m.MQTTClient = gmqtt.NewClient(opts)
	if token := m.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("failed to connect to MQTT broker: %v", token.Error())
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	}

	// 支持多个 client 的 topic 订阅
	if len(config.RequestTopics) > 0 && len(config.ResponseTopics) > 0 {
		for client, reqTopic := range config.RequestTopics {
			respTopic := config.ResponseTopics[client]
			token := m.MQTTClient.Subscribe(reqTopic, 0, func(clientName string, responseTopic string) func(client gmqtt.Client, msg gmqtt.Message) {
				return func(client gmqtt.Client, msg gmqtt.Message) {
					var request v1alpha2.COARequest
					var response v1alpha2.COAResponse
					if request.Context == nil {
						request.Context = context.TODO()
					}
					contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
					err := json.Unmarshal(msg.Payload(), &request)
					log.InfofCtx(request.Context, "Received request payload: %s", string(msg.Payload()))
					log.InfofCtx(request.Context, "Received request: %s", request)
					log.InfofCtx(request.Context, "Received request Route: %s", request.Route)
					if err != nil {
						response = v1alpha2.COAResponse{
							State:       v1alpha2.BadRequest,
							ContentType: "text/plain",
							Body:        []byte(err.Error()),
						}
					} else {
						response = routeTable[request.Route].Handler(request)
					}
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
							time.Sleep(600 * time.Millisecond)
							log.Errorf("failed to publish response to MQTT: %s", token.Error())
						}
					}()
				}
			}(client, respTopic))
			if token.Wait() && token.Error() != nil {
				if token.Error().Error() != "subscription exists" {
					log.Errorf("  P (MQTT Target): faild to connect to subscribe to request topic - %+v", token.Error())
					log.Errorf("  P (MQTT Target): sleeping 600s for debug, you can exec into the container to check certs and state...")
					time.Sleep(600 * time.Second)
					return v1alpha2.NewCOAError(token.Error(), "failed to subscribe to request topic", v1alpha2.InternalError)
				}
			}
		}
	} else {
		// 兼容单 topic 配置
		requestTopic := config.RequestTopic
		responseTopic := config.ResponseTopic
		token := m.MQTTClient.Subscribe(requestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			var request v1alpha2.COARequest
			var response v1alpha2.COAResponse
			if request.Context == nil {
				request.Context = context.TODO()
			}
			contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
			err := json.Unmarshal(msg.Payload(), &request)
			log.InfofCtx(request.Context, "Received request payload: %s", string(msg.Payload()))
			log.InfofCtx(request.Context, "Received request: %s", request)
			log.InfofCtx(request.Context, "Received request Route: %s", request.Route)
			if err != nil {
				response = v1alpha2.COAResponse{
					State:       v1alpha2.BadRequest,
					ContentType: "text/plain",
					Body:        []byte(err.Error()),
				}
			} else {
				response = routeTable[request.Route].Handler(request)
			}
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
					time.Sleep(600 * time.Millisecond)
					log.Errorf("failed to publish response to MQTT: %s", token.Error())
				}
			}()
		})
		if token.Wait() && token.Error() != nil {
			if token.Error().Error() != "subscription exists" {
				log.Errorf("  P (MQTT Target): faild to connect to subscribe to request topic - %+v", token.Error())
				log.Errorf("  P (MQTT Target): sleeping 600s for debug, you can exec into the container to check certs and state...")
				time.Sleep(600 * time.Second)
				return v1alpha2.NewCOAError(token.Error(), "failed to subscribe to request topic", v1alpha2.InternalError)
			}
		}
	}

	return nil
}
func (m *MQTTBinding) parseCertificates(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	var block *pem.Block
	var rest = pemData
	log.InfoCtx(context.Background(), "Parsing certificates from PEM data")
	log.InfoCtx(context.Background(), fmt.Sprintf("PEM data %+v", pemData))
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// Shutdown stops the MQTT binding
func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	m.MQTTClient.Disconnect(1000)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
