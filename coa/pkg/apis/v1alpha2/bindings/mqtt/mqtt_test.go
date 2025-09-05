/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestMQTTEcho(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	sig := make(chan int)
	config := MQTTBindingConfig{
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coabinding-test2",
		RequestTopic:  "coabinding-request2",
		ResponseTopic: "coabinding-response2",
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

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender2")
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

func TestMQTTConnectFail(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTBindingConfig{
		BrokerAddress: "tcp://169.254.1.1:1883",
		ClientID:      "coabinding-test",
		RequestTopic:  "coabinding-request",
		ResponseTopic: "coabinding-response",
	}
	binding := MQTTBinding{}
	err := binding.Launch(config, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.InternalError, coaError.State)
	assert.Contains(t, coaError.Error(), "failed to connect to MQTT broker")
}

func TestMQTT_CannotParseCOARequest(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	sig := make(chan int)
	config := MQTTBindingConfig{
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coabinding-test3",
		RequestTopic:  "coabinding-request3",
		ResponseTopic: "coabinding-response3",
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

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender3")
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
		assert.Equal(t, v1alpha2.BadRequest, response.State)
		sig <- 1
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}
	// error Request
	data := []byte("This is not a COARequest")
	token := c.Publish(config.RequestTopic, 0, false, data) //sending COARequest directly doesn't seem to work
	token.Wait()
	<-sig
}

func TestMQTTEchoWithCallContext(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	sig := make(chan int)
	config := MQTTBindingConfig{
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coabinding-test4",
		RequestTopic:  "coabinding-request4",
		ResponseTopic: "coabinding-response4",
	}
	binding := MQTTBinding{}
	endpoints := []v1alpha2.Endpoint{
		{
			Methods: []string{"GET"},
			Route:   "greetings",
			Handler: func(c v1alpha2.COARequest) v1alpha2.COAResponse {
				if c.Metadata != nil {
					if v, ok := c.Metadata["request-id"]; ok {
						return v1alpha2.COAResponse{
							Body: []byte(v),
						}
					}
				}
				return v1alpha2.COAResponse{
					Body: []byte("Hi there!!"),
				}
			},
		},
	}
	err := binding.Launch(config, endpoints)
	assert.Nil(t, err)

	opts := gmqtt.NewClientOptions().AddBroker(config.BrokerAddress).SetClientID("test-sender4")
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
		assert.Equal(t, string(response.Body), "request-1")
		sig <- 1
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}
	request := v1alpha2.COARequest{
		Route:  "greetings",
		Method: "GET",
		Metadata: map[string]string{
			"request-id": "request-1",
		},
	}
	data, _ := json.Marshal(request)
	token := c.Publish(config.RequestTopic, 0, false, data) //sending COARequest directly doesn't seem to work
	token.Wait()
	<-sig
}

// Tests for invalid CA certificate handling
func TestCreateTLSConfig_InvalidCA_NoPEMHeader(t *testing.T) {
	t.Parallel()
	path := createTempFile(t, "not a certificate")
	defer os.Remove(path)

	m := &MQTTBinding{}
	cfg := MQTTBindingConfig{CACertPath: path, UseTLS: "true"}
	if _, err := m.createTLSConfig(cfg); err == nil {
		t.Fatalf("expected error for invalid CA without PEM header, got nil")
	}
}

func TestCreateTLSConfig_InvalidCA_BadPEM(t *testing.T) {
	t.Parallel()
	// use an empty PEM block so x509.ParseCertificate will definitely fail
	badPEM := "-----BEGIN CERTIFICATE-----\n\n-----END CERTIFICATE-----\n"
	path := createTempFile(t, badPEM)
	defer os.Remove(path)

	m := &MQTTBinding{}
	cfg := MQTTBindingConfig{CACertPath: path, UseTLS: "true"}
	_, err := m.createTLSConfig(cfg)
	if err == nil {
		t.Fatalf("expected error for invalid PEM CA, got nil")
	}
	if !strings.Contains(err.Error(), "does not contain a valid X.509 certificate") {
		t.Fatalf("expected X.509 certificate error, got: %v", err)
	}
}

func TestCreateTLSConfig_ValidCA_ReturnsTLSConfig(t *testing.T) {
	t.Parallel()
	// generate a self-signed CA certificate
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	path := createTempFile(t, string(pemBytes))
	defer os.Remove(path)

	m := &MQTTBinding{}
	cfg := MQTTBindingConfig{CACertPath: path, UseTLS: "true"}
	tlsCfg, err := m.createTLSConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error for valid CA, got %v", err)
	}
	if tlsCfg == nil || tlsCfg.RootCAs == nil {
		t.Fatalf("expected non-nil TLS config with RootCAs")
	}
	if len(tlsCfg.RootCAs.Subjects()) == 0 {
		t.Fatalf("expected RootCAs to contain at least one certificate")
	}
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "mqtt-test-ca-*.pem")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}
