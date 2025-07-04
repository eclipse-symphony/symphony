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
	ClientCAFile = os.Getenv("CLIENT_CA_FILE")
	MQTTEnabled  = os.Getenv("MQTT_ENABLED") // New environment variable
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
	RequestTopics  map[string]string  `json:"requestTopics,omitempty"` // 支持多个client
	ResponseTopics map[string]string  `json:"responseTopics,omitempty"`
	CACert         string             `json:"caCert,omitempty"`
	ClientCert     string             `json:"clientCert,omitempty"`
	ClientKey      string             `json:"clientKey,omitempty"`
	CertProvider   CertProviderConfig `json:"certProvider,omitempty"`
	TLS            bool               `json:"tls"`
	// TrustedClients 可以保留但不再使用，以保持向后兼容性
	TrustedClients []string `json:"trustedClients,omitempty"`
	TopicPrefix    string   `json:"topicPrefix,omitempty"`
	AutoDiscovery  bool     `json:"autoDiscovery,omitempty"`
	// 新增state provider配置
	StateProvider struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	} `json:"stateProvider,omitempty"`
}

type MQTTBinding struct {
	MQTTClient        gmqtt.Client
	CertProvider      certs.ICertProvider
	subscribedTargets sync.Map // Track subscribed targets
	config            MQTTBindingConfig
	mu                sync.Mutex            // Mutex for subscription operations
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

	log.Infof("MQTT binding configuration: broker=%s, autoDiscovery=%t",
		config.BrokerAddress, config.AutoDiscovery)

	m.config = config

	// 初始化state provider（如果配置了）
	if config.AutoDiscovery && config.StateProvider.Type != "" {
		log.Info("Initializing state provider from config for MQTT binding")
		var stateProvider states.IStateProvider

		switch config.StateProvider.Type {
		case "providers.state.k8s":
			// K8s state provider 初始化
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
	if ClientCAFile != "" {
		log.Info(fmt.Sprintf("Loading client CA file: %s", ClientCAFile))
		pemData, err := ioutil.ReadFile(ClientCAFile)
		if err != nil {
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("Client CA file '%s' is not read successfully", ClientCAFile), v1alpha2.BadConfig)
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
		ServerName:   "10.172.3.39",
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
	}
	opts.SetTLSConfig(tlsConfig)
	m.MQTTClient = gmqtt.NewClient(opts)
	if token := m.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("Failed to connect to MQTT broker: %v", token.Error())
		log.Errorf("MQTT broker address: %s", config.BrokerAddress)
		log.Errorf("MQTT client ID: %s", config.ClientID)
		return v1alpha2.NewCOAError(token.Error(), "failed to connect to MQTT broker", v1alpha2.InternalError)
	} else {
		log.Infof("Successfully connected to MQTT broker: %s", config.BrokerAddress)
	}

	// Set default topic prefix if not specified
	if m.config.TopicPrefix == "" {
		m.config.TopicPrefix = "symphony"
	}

	// 如果配置了TrustedClients，则进行订阅
	if len(config.TrustedClients) > 0 {
		log.Infof("Setting up MQTT subscriptions for %d trusted clients", len(config.TrustedClients))
		for _, client := range config.TrustedClients {
			if err := m.SubscribeToTarget(client); err != nil {
				log.Errorf("Failed to subscribe to target %s: %v", client, err)
			}
		}
	}

	// Handle legacy single topic configuration
	if config.RequestTopic != "" && config.ResponseTopic != "" {
		token := m.MQTTClient.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
			// ...existing message handler...
			var request v1alpha2.COARequest
			var response v1alpha2.COAResponse
			if request.Context == nil {
				request.Context = context.TODO()
			}
			contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
			err := json.Unmarshal(msg.Payload(), &request)
			log.InfofCtx(request.Context, "Received request payload: %s", string(msg.Payload()))
			log.InfofCtx(request.Context, "Received request: %+v", request)
			log.InfofCtx(request.Context, "Received request Route: %s", request.Route)
			// 检查 target 字段
			var targetName string
			if request.Parameters != nil {
				if t, ok := request.Parameters["target"]; ok {
					// t 已经是 string 类型，无需类型断言
					targetName = t
				}
			}
			if targetName != "" && targetName != config.ClientID {
				log.InfofCtx(request.Context, "target mismatch: clientName=%s, target=%s", config.ClientID, targetName)
				response = v1alpha2.COAResponse{
					State:       v1alpha2.BadRequest,
					ContentType: "application/json",
					Body:        []byte(fmt.Sprintf("this client can not handle '%s', this target", targetName)),
				}
			} else if err != nil {
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
				if token := client.Publish(config.ResponseTopic, 0, false, data); token.Wait() && token.Error() != nil {
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

	// Start a reconciliation goroutine that runs every 5 minutes
	go m.reconcileSubscriptions()

	return nil
}

// SubscribeToTarget subscribes to a target's request topic
func (m *MQTTBinding) SubscribeToTarget(targetName string) error {
	if !m.enabled {
		return nil // Skip if MQTT is disabled
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already subscribed - add more detailed logging
	if _, exists := m.subscribedTargets.Load(targetName); exists {
		log.Infof("Already subscribed to target %s (skipping subscription)", targetName)
		return nil
	}

	log.Infof("Not yet subscribed to target %s, will subscribe now", targetName)

	// Create topic names
	requestTopic := fmt.Sprintf("%s/request/%s", m.config.TopicPrefix, strings.ToLower(targetName))
	responseTopic := fmt.Sprintf("%s/response/%s", m.config.TopicPrefix, strings.ToLower(targetName))

	log.Infof("Subscribing to target %s on topic %s", targetName, requestTopic)

	token := m.MQTTClient.Subscribe(requestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		var response v1alpha2.COAResponse
		if request.Context == nil {
			request.Context = context.TODO()
		}
		contexts.GenerateCorrelationIdToParentContextIfMissing(request.Context)
		err := json.Unmarshal(msg.Payload(), &request)
		log.InfofCtx(request.Context, "Received request payload from %s: %s", targetName, string(msg.Payload()))
		log.InfofCtx(request.Context, "Received request: %+v", request)
		log.InfofCtx(request.Context, "Received request Route: %s", request.Route)

		// Check target parameter
		var target string
		if request.Parameters != nil {
			if t, ok := request.Parameters["target"]; ok {
				target = t
			}
		}

		if target != "" && strings.ToLower(target) != strings.ToLower(targetName) {
			log.InfofCtx(request.Context, "target mismatch: clientName=%s, target=%s", targetName, target)
			errObj := map[string]interface{}{
				"error":  fmt.Sprintf("this client can not handle '%s', this target", target),
				"target": target,
				"client": targetName,
			}
			errJson, _ := json.Marshal(errObj)
			response = v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				ContentType: "application/json",
				Body:        errJson,
			}
		} else if err != nil {
			response = v1alpha2.COAResponse{
				State:       v1alpha2.BadRequest,
				ContentType: "text/plain",
				Body:        []byte(err.Error()),
			}
		} else if handler, ok := routeTable[request.Route]; ok {
			response = handler.Handler(request)
		} else {
			log.ErrorfCtx(request.Context, "Route not found: %s", request.Route)
			response = v1alpha2.COAResponse{
				State:       v1alpha2.NotFound,
				ContentType: "text/plain",
				Body:        []byte(fmt.Sprintf("Route not found: %s", request.Route)),
			}
		}

		// Copy request-id to response if present
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
				log.Errorf("Failed to publish response to MQTT topic %s: %s", responseTopic, token.Error())
			}
		}()
	})

	if token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			log.Errorf("Failed to subscribe to target %s: %v", targetName, token.Error())
			return token.Error()
		}
	}

	// Mark as subscribed
	m.subscribedTargets.Store(targetName, struct{}{})
	log.Infof("Successfully subscribed to target %s", targetName)
	return nil
}

// UnsubscribeFromTarget unsubscribes from a target's request topic
func (m *MQTTBinding) UnsubscribeFromTarget(targetName string) error {
	if !m.enabled {
		return nil // Skip if MQTT is disabled
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if subscribed
	if _, exists := m.subscribedTargets.Load(targetName); !exists {
		return nil
	}

	requestTopic := fmt.Sprintf("%s/request/%s", m.config.TopicPrefix, strings.ToLower(targetName))

	token := m.MQTTClient.Unsubscribe(requestTopic)
	if token.Wait() && token.Error() != nil {
		log.Errorf("Failed to unsubscribe from target %s: %v", targetName, token.Error())
		return token.Error()
	}

	// Remove from subscribed targets
	m.subscribedTargets.Delete(targetName)
	log.Infof("Successfully unsubscribed from target %s", targetName)
	return nil
}

// reconcileSubscriptions periodically reconciles MQTT subscriptions with actual targets
func (m *MQTTBinding) reconcileSubscriptions() {
	if !m.enabled {
		log.Info("MQTT binding is disabled, skipping subscription reconciliation")
		return // Skip if MQTT is disabled
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	log.InfofCtx(context.Background(), "Starting MQTT subscription reconciliation every 5 minutes")

	// Run once immediately on startup, don't wait for the first tick
	if m.stateProvider != nil && m.config.AutoDiscovery {
		log.Info("Running initial MQTT target discovery...")
		m.discoverAndReconcileTargets()
	} else {
		if m.stateProvider == nil {
			log.Warn("Cannot run MQTT auto-discovery: state provider is nil")
		}
		if !m.config.AutoDiscovery {
			log.Info("MQTT auto-discovery is disabled in configuration")
		}
	}

	for {
		select {
		case <-ticker.C:
			log.InfofCtx(context.Background(), "Starting MQTT subscription reconciliation")

			// 只使用自动发现，不再考虑 trustedClients
			if m.stateProvider != nil && m.config.AutoDiscovery {
				m.discoverAndReconcileTargets()
			}
		}
	}
}

// discoverAndReconcileTargets finds all targets with remote-agent bindings and reconciles subscriptions
func (m *MQTTBinding) discoverAndReconcileTargets() {
	ctx := context.Background()
	log.Info("Starting MQTT target discovery...")

	// Check if stateProvider is correctly set
	if m.stateProvider == nil {
		log.Error("Cannot discover targets: StateProvider is nil")
		return
	}

	// 这是查询资源的标准模式 - 创建ListRequest指定要查询的资源类型
	listRequest := states.ListRequest{
		Metadata: map[string]interface{}{
			"kind":      "Target",          // 资源种类
			"namespace": "",                // 空字符串表示查询所有命名空间
			"group":     "fabric.symphony", // 资源组
			"version":   "v1",              // API版本
			"resource":  "targets",         // 资源类型
		},
	}

	log.Infof("Using state provider of type: %T", m.stateProvider)

	// 调用List方法执行查询，返回三个值:
	// - result: 包含匹配资源的数组
	// - token: 通常是分页标记
	// - err: 可能的错误
	result, token, err := m.stateProvider.List(ctx, listRequest)
	if err != nil {
		// 处理"资源未找到"错误 - 这在CRD未安装时很常见
		if strings.Contains(err.Error(), "could not find the requested resource") {
			log.Warn("Failed to list targets: The Target CRD might not be installed in the cluster")
			log.Info("MQTT binding will continue but no targets will be auto-discovered")
			result = []states.StateEntry{} // 使用空结果集继续
		} else {
			log.Errorf("Failed to list targets: %v", err)
			return
		}
	}

	log.Infof("Found %d targets in total from state provider (token: %s)", len(result), token)

	// Define remoteTargets variable
	var remoteTargets []string

	// 详细记录每个target
	for _, entry := range result {
		// 打印target信息
		log.Infof("Target found: ID=%s, Type=%T", entry.ID, entry.Body)

		// 检查是否是远程target - add type assertion for entry.Body
		targetBody, ok := entry.Body.(map[string]interface{})
		if !ok {
			// Try to unmarshal if it's a byte array
			if bodyBytes, isByteArray := entry.Body.([]byte); isByteArray {
				var targetMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &targetMap); err == nil {
					targetBody = targetMap
					ok = true
				}
			}

			if !ok {
				log.Warnf("Target body is not a map: %T", entry.Body)
				continue
			}
		}

		if hasRemoteAgentComponent(targetBody) {
			targetName := entry.ID
			remoteTargets = append(remoteTargets, targetName)
			log.Infof("Found remote target with component: %s", targetName)
		} else {
			log.Debugf("Target %s does not have a remote-agent component", entry.ID)
		}
	}

	if len(remoteTargets) > 0 {
		log.Infof("Found %d remote targets during auto-discovery: %v", len(remoteTargets), remoteTargets)
		m.ReconcileWithTargets(remoteTargets)
	} else {
		log.Info("No remote targets found during auto-discovery")
	}
}

// Helper function to check if a target has a remote-agent component
func hasRemoteAgentComponent(target map[string]interface{}) bool {
	spec, ok := target["spec"].(map[string]interface{})
	if !ok {
		log.Debug("Target spec is not a map")
		return false
	}

	// 只检查Components中是否有remote-agent类型的组件
	if components, ok := spec["components"].([]interface{}); ok {
		for _, comp := range components {
			if component, ok := comp.(map[string]interface{}); ok {
				if compType, ok := component["type"].(string); ok && compType == "remote-agent" {
					log.Info("Found remote-agent component in target components")
					return true
				}
			}
		}
	}

	// 不再检查topologies部分
	return false
}

// ReconcileWithTargets ensures the MQTT subscriptions match the provided targets list
func (m *MQTTBinding) ReconcileWithTargets(targets []string) {
	if !m.enabled {
		return // Skip if MQTT is disabled
	}

	log.Info("Reconciling MQTT subscriptions with targets list", "targets", targets)

	// 记录当前已订阅的目标
	currentSubscriptions := []string{}
	m.subscribedTargets.Range(func(key, value interface{}) bool {
		targetName := key.(string)
		currentSubscriptions = append(currentSubscriptions, targetName)
		return true
	})
	log.Infof("Current subscriptions: %v", currentSubscriptions)

	// Track targets that should be subscribed
	shouldBeSubscribed := make(map[string]bool)
	for _, target := range targets {
		log.InfoCtx(context.Background(), fmt.Sprintf("Target to subscribe: %s", target))
		shouldBeSubscribed[target] = true
	}

	// Check current subscriptions and remove any that shouldn't exist
	m.subscribedTargets.Range(func(key, value interface{}) bool {
		targetName := key.(string)
		if !shouldBeSubscribed[targetName] {
			log.Infof("Unsubscribing from target %s as it's no longer needed", targetName)
			if err := m.UnsubscribeFromTarget(targetName); err != nil {
				log.Errorf("Failed to unsubscribe from target %s: %v", targetName, err)
			}
		}
		return true
	})

	// Add any missing subscriptions
	for _, target := range targets {
		if _, exists := m.subscribedTargets.Load(target); !exists {
			log.Infof("Adding missing subscription for target %s", target)
			if err := m.SubscribeToTarget(target); err != nil {
				log.Errorf("Failed to subscribe to target %s: %v", target, err)
			}
		}
	}subscription reconciliation completed")
	var certs []*x509.Certificate}
	var block *pem.Block
	rest := pemDataficates(pemData []byte) ([]*x509.Certificate, error) {
t.Background(), "Parsing certificates from PEM data")
	for {oCtx(context.Background(), fmt.Sprintf("PEM data %+v", pemData))
		block, rest = pem.Decode(rest)
		if block == nil {	var certs []*x509.Certificate
			break
		}mData

		if block.Type != "CERTIFICATE" {	for {
			continue
		} {

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errTE" {
		}	continue
		}
		certs = append(certs, cert)
	}	cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
	return certs, nil
}

func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	// Shutdown stops the MQTT binding
	if m.enabled && m.MQTTClient != nil {
		m.MQTTClient.Disconnect(1000)return certs, nil
	}}



}	return nil
func (m *MQTTBinding) Shutdown(ctx context.Context) error {
	// Shutdown stops the MQTT binding
	if m.enabled && m.MQTTClient != nil {
		m.MQTTClient.Disconnect(1000)
	}
	return nil
}
