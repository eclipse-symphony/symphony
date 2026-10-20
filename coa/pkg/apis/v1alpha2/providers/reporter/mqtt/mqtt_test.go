/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testToken struct {
	gmqtt.Token
	err     error
	timeout bool
}

func (t testToken) WaitTimeout(time.Duration) bool { return !t.timeout }
func (t testToken) Error() error                   { return t.err }

type testMessage struct {
	gmqtt.Message
	data []byte
}

func (m testMessage) Payload() []byte { return m.data }

type testClient struct {
	gmqtt.Client
	open       bool
	connects   int
	connectErr error
	subErr     error
	publishErr error
	handler    gmqtt.MessageHandler
	request    v1alpha2.COARequest
	state      v1alpha2.State
	noResponse bool
}

func (c *testClient) IsConnectionOpen() bool { return c.open }
func (c *testClient) Connect() gmqtt.Token {
	c.connects++
	c.open = c.connectErr == nil
	return testToken{err: c.connectErr}
}
func (c *testClient) Subscribe(topic string, qos byte, handler gmqtt.MessageHandler) gmqtt.Token {
	c.handler = handler
	return testToken{err: c.subErr}
}
func (c *testClient) Publish(topic string, qos byte, retained bool, payload interface{}) gmqtt.Token {
	if err := json.Unmarshal(payload.([]byte), &c.request); err != nil {
		return testToken{err: err}
	}
	if !c.noResponse && c.publishErr == nil {
		data, err := json.Marshal(v1alpha2.COAResponse{
			State: c.state, Body: []byte("report result"), Metadata: c.request.Metadata,
		})
		if err != nil {
			return testToken{err: err}
		}
		c.handler(c, testMessage{data: data})
	}
	return testToken{err: c.publishErr}
}
func (c *testClient) Disconnect(uint) { c.open = false }

func reporterConfig() map[string]interface{} {
	return map[string]interface{}{
		"name": "test-reporter", "brokerAddress": "tcp://localhost:1883",
		"clientID": "status-agent", "requestTopic": "status/request",
		"responseTopic": "status/response", "timeoutSeconds": 1,
	}
}

func TestMQTTReporterConfiguration(t *testing.T) {
	r := &MQTTReporter{}
	require.NoError(t, r.Init(reporterConfig()))
	assert.Equal(t, "test-reporter", r.ID())
	assert.Equal(t, "references", r.Config.Route)
	clone, err := r.Clone(nil)
	require.NoError(t, err)
	assert.Equal(t, r.Config, clone.(*MQTTReporter).Config)
	assert.NotSame(t, r.MQTTClient, clone.(*MQTTReporter).MQTTClient)
	for _, field := range []string{"brokerAddress", "clientID", "requestTopic", "responseTopic"} {
		t.Run(field, func(t *testing.T) {
			c := reporterConfig()
			delete(c, field)
			assert.Error(t, (&MQTTReporter{}).Init(c))
		})
	}
	for key, value := range map[string]interface{}{
		"timeoutSeconds": -1, "keepAliveSeconds": -1, "pingTimeoutSeconds": -1,
		"requestTopic": "wildcard/+", "responseTopic": "status/request",
		"useTLS": "invalid", "insecureSkipVerify": "invalid",
	} {
		t.Run(key, func(t *testing.T) {
			c := reporterConfig()
			c[key] = value
			assert.Error(t, (&MQTTReporter{}).Init(c))
		})
	}
	c := reporterConfig()
	c["useTLS"] = "true"
	c["caCertPath"] = t.TempDir() + "/missing.pem"
	assert.Error(t, r.Init(c))
	assert.Error(t, r.InitWithMap(map[string]string{"timeoutSeconds": "bad"}))
	require.NoError(t, r.InitWithMap(map[string]string{
		"brokerAddress": "tcp://localhost:1883", "clientID": "map-reporter",
		"requestTopic": "status/request", "responseTopic": "status/response",
		"timeoutSeconds": "0",
	}))
	assert.Equal(t, 8, r.Config.TimeoutSeconds)
	assert.Equal(t, 2, r.Config.KeepAliveSeconds)
	assert.Equal(t, 1, r.Config.PingTimeoutSeconds)
	t.Setenv("SYMPHONY_REPORTER_BROKER", "tcp://localhost:1883")
	c = reporterConfig()
	c["brokerAddress"] = "$env:SYMPHONY_REPORTER_BROKER"
	require.NoError(t, r.Init(c))
	assert.Equal(t, "tcp://localhost:1883", r.Config.BrokerAddress)
}

func TestMQTTReporterRequestAndReconnect(t *testing.T) {
	r := &MQTTReporter{}
	require.NoError(t, r.Init(reporterConfig()))
	client := &testClient{state: v1alpha2.OK}
	r.MQTTClient = client
	properties := map[string]string{"site": "factory-a"}
	require.NoError(t, r.Report("edge", "default", "fabric.symphony", "targets", "v1", properties, false))
	assert.Equal(t, "POST", client.request.Method)
	assert.Equal(t, "references", client.request.Route)
	assert.Equal(t, "application/json", client.request.ContentType)
	assert.Equal(t, map[string]string{
		"id": "edge", "namespace": "default", "group": "fabric.symphony",
		"kind": "targets", "version": "v1", "overwrite": "false",
	}, client.request.Parameters)
	var received map[string]string
	require.NoError(t, json.Unmarshal(client.request.Body, &received))
	assert.Equal(t, properties, received)
	firstID := client.request.Metadata["request-id"]
	assert.NotEmpty(t, firstID)
	require.NoError(t, r.Report("edge", "default", "fabric.symphony", "targets", "v1", properties, true))
	assert.Equal(t, 1, client.connects)
	assert.NotEqual(t, firstID, client.request.Metadata["request-id"])
	assert.Equal(t, "true", client.request.Parameters["overwrite"])
	client.Disconnect(0)
	require.NoError(t, r.Report("edge", "default", "fabric.symphony", "targets", "v1", properties, false))
	assert.Equal(t, 2, client.connects)
	assert.Empty(t, responseEntries(r))
	require.NoError(t, r.Shutdown(context.Background()))
	assert.False(t, client.open)
	assert.Error(t, r.Report("edge", "default", "fabric.symphony", "targets", "v1", properties, false))
}

func responseEntries(r *MQTTReporter) []interface{} {
	var keys []interface{}
	r.responses.Range(func(key, value interface{}) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

func TestMQTTReporterFailures(t *testing.T) {
	for name, client := range map[string]*testClient{
		"connect":   {connectErr: errors.New("connect failed")},
		"subscribe": {subErr: errors.New("subscribe failed")},
		"publish":   {publishErr: errors.New("publish failed")},
		"remote":    {state: v1alpha2.InternalError},
		"timeout":   {noResponse: true},
	} {
		t.Run(name, func(t *testing.T) {
			r := &MQTTReporter{}
			require.NoError(t, r.Init(reporterConfig()))
			r.MQTTClient = client
			assert.Error(t, r.Report("edge", "default", "fabric.symphony", "targets", "v1", nil, false))
			assert.Empty(t, responseEntries(r))
		})
	}
	r := &MQTTReporter{}
	require.NoError(t, r.Init(reporterConfig()))
	r.MQTTClient = &testClient{}
	assert.Error(t, r.Report("", "default", "fabric.symphony", "targets", "v1", nil, false))
	assert.Error(t, r.wait(testToken{timeout: true}, "publish"))
}

func TestMQTTReporterIgnoresUnrelatedAndDuplicateResponses(t *testing.T) {
	r := &MQTTReporter{}
	responses := make(chan v1alpha2.COAResponse, 1)
	r.responses.Store("ours", responses)
	r.onResponse(nil, testMessage{data: []byte("invalid")})
	r.onResponse(nil, testMessage{data: []byte(`{"state":200,"metadata":{"request-id":"other"}}`)})
	assert.Empty(t, responses)
	data := []byte(`{"state":200,"metadata":{"request-id":"ours"}}`)
	r.onResponse(nil, testMessage{data: data})
	r.onResponse(nil, testMessage{data: data})
	assert.Len(t, responses, 1)
}
