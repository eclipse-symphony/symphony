/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	mqttbinding "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var log = logger.NewLogger("providers.reporter.mqtt")

type MQTTReporterConfig struct {
	mqttbinding.MQTTBindingConfig
	Name  string `json:"name"`
	Route string `json:"route,omitempty"`
}

type MQTTReporter struct {
	Config     MQTTReporterConfig
	Context    *contexts.ManagerContext
	MQTTClient gmqtt.Client
	mu         sync.Mutex
	responses  sync.Map
	closed     bool
}

func (r *MQTTReporter) ID() string {
	return r.Config.Name
}

func (r *MQTTReporter) SetContext(ctx *contexts.ManagerContext) {
	r.Context = ctx
}

func (r *MQTTReporter) InitWithMap(properties map[string]string) error {
	values := make(map[string]interface{}, len(properties))
	for key, value := range properties {
		switch key {
		case "timeoutSeconds", "keepAliveSeconds", "pingTimeoutSeconds":
			n, err := strconv.Atoi(value)
			if err != nil {
				return v1alpha2.NewCOAError(err, key+" must be an integer", v1alpha2.BadConfig)
			}
			values[key] = n
		default:
			values[key] = value
		}
	}
	return r.Init(values)
}

func (r *MQTTReporter) Init(config providers.IProviderConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var c MQTTReporterConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return v1alpha2.NewCOAError(err, "invalid MQTT reporter config", v1alpha2.BadConfig)
	}
	c.BrokerAddress = utils.ParseProperty(c.BrokerAddress)
	c.ClientID = utils.ParseProperty(c.ClientID)
	c.RequestTopic = utils.ParseProperty(c.RequestTopic)
	c.ResponseTopic = utils.ParseProperty(c.ResponseTopic)
	c.Username = utils.ParseProperty(c.Username)
	c.Password = utils.ParseProperty(c.Password)
	c.CACertPath = utils.ParseProperty(c.CACertPath)
	c.ClientCertPath = utils.ParseProperty(c.ClientCertPath)
	c.ClientKeyPath = utils.ParseProperty(c.ClientKeyPath)
	for _, field := range []struct{ name, value string }{
		{"brokerAddress", c.BrokerAddress}, {"clientID", c.ClientID},
		{"requestTopic", c.RequestTopic}, {"responseTopic", c.ResponseTopic},
	} {
		if strings.TrimSpace(field.value) == "" {
			return v1alpha2.NewCOAError(nil, field.name+" is required in MQTT reporter config", v1alpha2.BadConfig)
		}
	}
	if c.RequestTopic == c.ResponseTopic || strings.ContainsAny(c.RequestTopic+c.ResponseTopic, "+#") {
		return v1alpha2.NewCOAError(nil, "MQTT reporter requires distinct, non-wildcard request and response topics", v1alpha2.BadConfig)
	}
	if c.Route == "" {
		c.Route = "references"
	}
	if c.TimeoutSeconds < 0 || c.KeepAliveSeconds < 0 || c.PingTimeoutSeconds < 0 {
		return v1alpha2.NewCOAError(nil, "MQTT reporter timeouts must not be negative", v1alpha2.BadConfig)
	}
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = 8
	}
	if c.KeepAliveSeconds == 0 {
		c.KeepAliveSeconds = 2
	}
	if c.PingTimeoutSeconds == 0 {
		c.PingTimeoutSeconds = 1
	}
	timeout := time.Duration(c.TimeoutSeconds) * time.Second
	opts := gmqtt.NewClientOptions().AddBroker(c.BrokerAddress).SetClientID(c.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(false)
	opts.SetConnectTimeout(timeout)
	opts.SetWriteTimeout(timeout)
	opts.SetKeepAlive(time.Duration(c.KeepAliveSeconds) * time.Second)
	opts.SetPingTimeout(time.Duration(c.PingTimeoutSeconds) * time.Second)
	opts.SetUsername(c.Username)
	opts.SetPassword(c.Password)
	if c.UseTLS != "" && c.UseTLS != "true" && c.UseTLS != "false" {
		return v1alpha2.NewCOAError(nil, "useTLS must be true or false", v1alpha2.BadConfig)
	}
	if c.InsecureSkipVerify != "" && c.InsecureSkipVerify != "true" && c.InsecureSkipVerify != "false" {
		return v1alpha2.NewCOAError(nil, "insecureSkipVerify must be true or false", v1alpha2.BadConfig)
	}
	if c.UseTLS == "true" {
		tlsConfig, err := mqttbinding.CreateTLSConfig(c.MQTTBindingConfig)
		if err != nil {
			return v1alpha2.NewCOAError(err, "invalid MQTT reporter TLS config", v1alpha2.BadConfig)
		}
		opts.SetTLSConfig(tlsConfig)
	}
	r.Config = c
	r.MQTTClient = gmqtt.NewClient(opts)
	return nil
}

func (r *MQTTReporter) wait(token gmqtt.Token, operation string) error {
	if !token.WaitTimeout(time.Duration(r.Config.TimeoutSeconds) * time.Second) {
		return fmt.Errorf("MQTT reporter %s timed out", operation)
	}
	if err := token.Error(); err != nil {
		return fmt.Errorf("MQTT reporter %s: %w", operation, err)
	}
	return nil
}

func (r *MQTTReporter) onResponse(_ gmqtt.Client, message gmqtt.Message) {
	var response v1alpha2.COAResponse
	if err := json.Unmarshal(message.Payload(), &response); err != nil {
		log.Errorf("MQTT reporter: invalid response: %v", err)
		return
	}
	if value, ok := r.responses.Load(response.Metadata["request-id"]); ok {
		select {
		case value.(chan v1alpha2.COAResponse) <- response:
		default:
		}
	}
}

func (r *MQTTReporter) Report(id, namespace, group, kind, version string, properties map[string]string, overwrite bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.MQTTClient == nil {
		return fmt.Errorf("MQTT reporter is closed or not initialized")
	}
	if id == "" || namespace == "" || group == "" || kind == "" || version == "" {
		return fmt.Errorf("MQTT reporter requires complete resource identity")
	}
	if !r.MQTTClient.IsConnectionOpen() {
		if err := r.wait(r.MQTTClient.Connect(), "connect"); err != nil {
			r.MQTTClient.Disconnect(1000)
			return err
		}
	}
	if err := r.wait(r.MQTTClient.Subscribe(r.Config.ResponseTopic, 1, r.onResponse), "subscribe"); err != nil {
		r.MQTTClient.Disconnect(1000)
		return err
	}
	body, err := json.Marshal(properties)
	if err != nil {
		return err
	}
	requestID := uuid.NewString()
	responses := make(chan v1alpha2.COAResponse, 1)
	r.responses.Store(requestID, responses)
	defer r.responses.Delete(requestID)
	request := v1alpha2.COARequest{
		Method:      "POST",
		Route:       r.Config.Route,
		ContentType: "application/json",
		Body:        body,
		Metadata:    map[string]string{"request-id": requestID},
		Parameters: map[string]string{
			"id": id, "namespace": namespace, "group": group,
			"kind": kind, "version": version, "overwrite": strconv.FormatBool(overwrite),
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	if err := r.wait(r.MQTTClient.Publish(r.Config.RequestTopic, 1, false, data), "publish"); err != nil {
		return err
	}
	timer := time.NewTimer(time.Duration(r.Config.TimeoutSeconds) * time.Second)
	defer timer.Stop()
	select {
	case response := <-responses:
		if response.State != v1alpha2.OK {
			return v1alpha2.NewCOAError(nil, string(response.Body), response.State)
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("MQTT reporter timed out waiting for status acknowledgement")
	}
}

func (r *MQTTReporter) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	if config == nil {
		config = r.Config
	}
	clone := &MQTTReporter{}
	if err := clone.Init(config); err != nil {
		return nil, err
	}
	clone.Context = r.Context
	return clone, nil
}

func (r *MQTTReporter) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.MQTTClient != nil {
		r.MQTTClient.Disconnect(1000)
	}
	return nil
}
