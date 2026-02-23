/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mqtt

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/conformance"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	gmqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const mqttMTLSEnvVar = "TEST_MQTT_MTLS_ENABLED"

type mtlsBroker struct {
	address         string
	caCertPath      string
	clientCertPath  string
	clientKeyPath   string
	clientTLSConfig *tls.Config
	cleanup         func()
}

func requireMQTTMTLS(t *testing.T) {
	if os.Getenv(mqttMTLSEnvVar) == "" {
		t.Skip("Skipping because TEST_MQTT_MTLS_ENABLED environment variable is not set")
	}
	if _, err := exec.LookPath("mosquitto"); err != nil {
		t.Skip("Skipping because mosquitto is not installed")
	}
}

func startMTLSBroker(t *testing.T) *mtlsBroker {
	t.Helper()
	dir := t.TempDir()
	port := allocateFreePort(t)

	caCertPath, serverCertPath, serverKeyPath, clientCertPath, clientKeyPath := generateMTLSCerts(t, dir)
	configPath := filepath.Join(dir, "mosquitto.conf")
	config := fmt.Sprintf(`per_listener_settings true
listener %d 127.0.0.1
allow_anonymous true
cafile %s
certfile %s
keyfile %s
require_certificate true
use_identity_as_username true
`, port, caCertPath, serverCertPath, serverKeyPath)
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatalf("failed to write mosquitto config: %v", err)
	}

	cmd := exec.Command("mosquitto", "-c", configPath, "-v")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start mosquitto: %v", err)
	}

	address := fmt.Sprintf("127.0.0.1:%d", port)
	waitForTCP(t, address, 5*time.Second)

	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		t.Fatalf("failed to read CA cert: %v", err)
	}
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		t.Fatalf("failed to load client key pair: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		t.Fatalf("failed to append CA cert to pool")
	}

	cleanup := func() {
		if cmd.Process == nil {
			return
		}
		_ = cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case <-time.After(3 * time.Second):
			_ = cmd.Process.Kill()
		case <-done:
		}
	}

	return &mtlsBroker{
		address:         fmt.Sprintf("tls://%s", address),
		caCertPath:      caCertPath,
		clientCertPath:  clientCertPath,
		clientKeyPath:   clientKeyPath,
		clientTLSConfig: &tls.Config{RootCAs: certPool, Certificates: []tls.Certificate{clientCert}},
		cleanup:         cleanup,
	}
}

func allocateFreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForTCP(t *testing.T, address string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for mosquitto to listen on %s", address)
}

func generateMTLSCerts(t *testing.T, dir string) (string, string, string, string, string) {
	t.Helper()
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          randomSerial(t),
		Subject:               pkix.Name{CommonName: "mqtt-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create CA cert: %v", err)
	}
	caCertPath := filepath.Join(dir, "ca.crt")
	if err := writePEMFile(caCertPath, "CERTIFICATE", caDER); err != nil {
		t.Fatalf("failed to write CA cert: %v", err)
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate server key: %v", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: randomSerial(t),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create server cert: %v", err)
	}
	serverCertPath := filepath.Join(dir, "server.crt")
	serverKeyPath := filepath.Join(dir, "server.key")
	if err := writePEMFile(serverCertPath, "CERTIFICATE", serverDER); err != nil {
		t.Fatalf("failed to write server cert: %v", err)
	}
	if err := writePEMFile(serverKeyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey)); err != nil {
		t.Fatalf("failed to write server key: %v", err)
	}

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate client key: %v", err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: randomSerial(t),
		Subject:      pkix.Name{CommonName: "mqtt-test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caTemplate, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create client cert: %v", err)
	}
	clientCertPath := filepath.Join(dir, "client.crt")
	clientKeyPath := filepath.Join(dir, "client.key")
	if err := writePEMFile(clientCertPath, "CERTIFICATE", clientDER); err != nil {
		t.Fatalf("failed to write client cert: %v", err)
	}
	if err := writePEMFile(clientKeyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(clientKey)); err != nil {
		t.Fatalf("failed to write client key: %v", err)
	}

	return caCertPath, serverCertPath, serverKeyPath, clientCertPath, clientKeyPath
}

func randomSerial(t *testing.T) *big.Int {
	t.Helper()
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	if err != nil {
		t.Fatalf("failed to generate serial: %v", err)
	}
	return serial
}

func writePEMFile(path string, blockType string, der []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	return pem.Encode(file, &pem.Block{Type: blockType, Bytes: der})
}

func newMTLSClient(t *testing.T, broker *mtlsBroker, clientID string) gmqtt.Client {
	t.Helper()
	opts := gmqtt.NewClientOptions().AddBroker(broker.address).SetClientID(clientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetTLSConfig(broker.clientTLSConfig)
	client := gmqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect mtls client: %v", token.Error())
	}
	return client
}

func newLocalClient(t *testing.T, brokerAddress string) gmqtt.Client {
	t.Helper()
	opts := gmqtt.NewClientOptions().AddBroker(brokerAddress).
		SetClientID(fmt.Sprintf("test-sender-%s", uuid.NewString()))
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	client := gmqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		t.Fatalf("failed to connect local client: %v", token.Error())
	}
	return client
}

func TestCreateTLSConfigMultiCertBundle(t *testing.T) {
	dir := t.TempDir()
	caCertPath, _, _, clientCertPath, _ := generateMTLSCerts(t, dir)

	caCert, err := os.ReadFile(caCertPath)
	assert.Nil(t, err)
	clientCert, err := os.ReadFile(clientCertPath)
	assert.Nil(t, err)

	bundlePath := filepath.Join(dir, "bundle.pem")
	bundle := append([]byte("# test bundle\n"), caCert...)
	bundle = append(bundle, []byte("\n# extra cert\n")...)
	bundle = append(bundle, clientCert...)
	assert.Nil(t, os.WriteFile(bundlePath, bundle, 0600))

	provider := MQTTTargetProvider{
		Config: MQTTTargetProviderConfig{
			CACertPath: bundlePath,
		},
	}
	tlsConfig, err := provider.createTLSConfig(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, tlsConfig)
	assert.NotNil(t, tlsConfig.RootCAs)
	assert.Greater(t, len(tlsConfig.RootCAs.Subjects()), 0)
}

func TestDoubleInit(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)
	err = provider.Init(config)
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	configMap := map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
		"requestTopic":  "coa-request",
		"responseTopic": "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestInitWithMapInvalidConfig(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	configMap := map[string]string{
		"name": "me",
	}
	provider := MQTTTargetProvider{}
	err := provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":          "me",
		"brokerAddress": "tcp://127.0.0.1:1883",
		"clientID":      "coa-test2",
		"requestTopic":  "coa-request",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":           "me",
		"brokerAddress":  "tcp://127.0.0.1:1883",
		"clientID":       "coa-test2",
		"requestTopic":   "coa-request",
		"responseTopic":  "coa-response",
		"timeoutSeconds": "abcd",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":             "me",
		"brokerAddress":    "tcp://127.0.0.1:1883",
		"clientID":         "coa-test2",
		"requestTopic":     "coa-request",
		"responseTopic":    "coa-response",
		"timeoutSeconds":   "2",
		"keepAliveSeconds": "abc",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "me",
		"brokerAddress":      "tcp://127.0.0.1:1883",
		"clientID":           "coa-test2",
		"requestTopic":       "coa-request",
		"responseTopic":      "coa-response",
		"keepAliveSeconds":   "2",
		"pingTimeoutSeconds": "abc",
	}
	err = provider.InitWithMap(configMap)
	assert.NotNil(t, err)

	configMap = map[string]string{
		"name":               "me",
		"brokerAddress":      "tcp://127.0.0.1:1883",
		"clientID":           "coa-test2",
		"requestTopic":       "coa-request",
		"responseTopic":      "coa-response",
		"timeoutSeconds":     "2",
		"keepAliveSeconds":   "2",
		"pingTimeoutSeconds": "2",
	}
	err = provider.InitWithMap(configMap)
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://127.0.0.1:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		ret := make([]model.ComponentSpec, 0)
		data, _ := json.Marshal(ret)
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		response.Body = data
		data, _ = json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

func TestGetMTLS(t *testing.T) {
	requireMQTTMTLS(t)
	broker := startMTLSBroker(t)
	defer broker.cleanup()

	config := MQTTTargetProviderConfig{
		Name:           "me",
		BrokerAddress:  broker.address,
		ClientID:       "coa-test-mtls",
		RequestTopic:   "coa-request-mtls",
		ResponseTopic:  "coa-response-mtls",
		UseTLS:         true,
		CACertPath:     broker.caCertPath,
		ClientCertPath: broker.clientCertPath,
		ClientKeyPath:  broker.clientKeyPath,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	client := newMTLSClient(t, broker, "test-sender-mtls")
	defer client.Disconnect(250)

	if token := client.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		ret := make([]model.ComponentSpec, 0)
		data, _ := json.Marshal(ret)
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		response.Body = data
		data, _ = json.Marshal(response)
		token := client.Publish(config.ResponseTopic, 0, false, data)
		token.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

func TestApplyMTLS(t *testing.T) {
	requireMQTTMTLS(t)
	broker := startMTLSBroker(t)
	defer broker.cleanup()

	const (
		MQTTName          string = "me"
		MQTTClientID      string = "coa-test-mtls"
		MQTTRequestTopic  string = "coa-request-mtls"
		MQTTResponseTopic string = "coa-response-mtls"

		TestTargetSuccessMessage string = "Success"
	)

	config := MQTTTargetProviderConfig{
		Name:           MQTTName,
		BrokerAddress:  broker.address,
		ClientID:       MQTTClientID,
		RequestTopic:   MQTTRequestTopic,
		ResponseTopic:  MQTTResponseTopic,
		UseTLS:         true,
		CACertPath:     broker.caCertPath,
		ClientCertPath: broker.clientCertPath,
		ClientKeyPath:  broker.clientKeyPath,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	client := newMTLSClient(t, broker, "test-sender-mtls-apply")
	defer client.Disconnect(250)

	if token := client.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		summarySpec := model.SummarySpec{
			TargetCount:  1,
			SuccessCount: 1,
			TargetResults: map[string]model.TargetResultSpec{
				"test-target": {
					Status: v1alpha2.OK.String(),
					ComponentResults: map[string]model.ComponentResultSpec{
						"test-component": {
							Status:  v1alpha2.Updated,
							Message: TestTargetSuccessMessage,
						},
					},
				},
			},
			Skipped:             false,
			IsRemoval:           false,
			AllAssignedDeployed: true,
		}
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Body, _ = json.Marshal(summarySpec)
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := client.Publish(config.ResponseTopic, 0, false, data)
		token.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	deploymentSpec := model.DeploymentSpec{
		SolutionName: "test-solution",
		Solution:     model.SolutionState{},
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-instance",
			},
			Spec: &model.InstanceSpec{
				DisplayName: "test-instance",
				Solution:    "test-solution",
				Target: model.TargetSelector{
					Name: "test-target",
				},
			},
		},
		Targets: map[string]model.TargetState{
			"test-target": {
				Spec: &model.TargetSpec{
					DisplayName: "test-target",
					Components: []model.ComponentSpec{
						{
							Name: "test-component",
							Type: "test-component",
						},
					},
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "test-target",
									Provider: "providers.target.mqtt",
									Config: map[string]string{
										"name":          MQTTName,
										"brokerAddress": broker.address,
										"clientID":      MQTTClientID,
										"requestTopic":  MQTTRequestTopic,
										"responseTopic": MQTTResponseTopic,
										"useTLS":        "true",
										"caCertPath":    broker.caCertPath,
										"clientCertPath": broker.clientCertPath,
										"clientKeyPath":  broker.clientKeyPath,
									},
								},
							},
						},
					},
				},
			},
		},
		Assignments: map[string]string{
			"test-component": "{test-target}",
		},
	}

	stepSpec := model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}

	ret, err := provider.Apply(context.Background(), deploymentSpec, stepSpec, false)

	assert.Nil(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, ret["test-component"].Status, v1alpha2.Untouched)
	assert.Equal(t, ret["test-component"].Message, "No error. test-component is untouched")
	assert.Equal(t, ret["test-target"].Status, v1alpha2.Updated)
	assert.Equal(t, ret["test-target"].Message, TestTargetSuccessMessage)
}
func TestGetBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		response.Body = []byte("BAD!!")
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.NotNil(t, err)
	assert.Equal(t, "Internal Error: BAD!!", err.Error())
}
func TestApply(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}

	const (
		MQTTName          string = "me"
		MQTTBrokerAddress string = "tcp://localhost:1883"
		MQTTClientID      string = "coa-test2"
		MQTTRequestTopic  string = "coa-request"
		MQTTResponseTopic string = "coa-response"

		TestTargetSuccessMessage string = "Success"
	)

	config := MQTTTargetProviderConfig{
		Name:          MQTTName,
		BrokerAddress: MQTTBrokerAddress,
		ClientID:      MQTTClientID,
		RequestTopic:  MQTTRequestTopic,
		ResponseTopic: MQTTResponseTopic,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		summarySpec := model.SummarySpec{
			TargetCount:  1,
			SuccessCount: 1,
			TargetResults: map[string]model.TargetResultSpec{
				"test-target": {
					Status: v1alpha2.OK.String(),
					ComponentResults: map[string]model.ComponentResultSpec{
						"test-component": {
							Status:  v1alpha2.Updated,
							Message: TestTargetSuccessMessage,
						},
					},
				},
			},
			Skipped:             false,
			IsRemoval:           false,
			AllAssignedDeployed: true,
		}
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Body, _ = json.Marshal(summarySpec)
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	deploymentSpec := model.DeploymentSpec{
		SolutionName: "test-solution",
		Solution:     model.SolutionState{},
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "test-instance",
			},
			Spec: &model.InstanceSpec{
				DisplayName: "test-instance",
				Solution:    "test-solution",
				Target: model.TargetSelector{
					Name: "test-target",
				},
			},
		},
		Targets: map[string]model.TargetState{
			"test-target": {
				Spec: &model.TargetSpec{
					DisplayName: "test-target",
					Components: []model.ComponentSpec{
						{
							Name: "test-component",
							Type: "test-component",
						},
					},
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "test-target",
									Provider: "providers.target.mqtt",
									Config: map[string]string{
										"name":          MQTTName,
										"brokerAddress": MQTTBrokerAddress,
										"clientID":      MQTTClientID,
										"requestTopic":  MQTTRequestTopic,
										"responseTopic": MQTTResponseTopic,
									},
								},
							},
						},
					},
				},
			},
		},
		Assignments: map[string]string{
			"test-component": "{test-target}",
		},
	}

	stepSpec := model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}

	ret, err := provider.Apply(context.Background(), deploymentSpec, stepSpec, false)

	assert.Nil(t, err)
	assert.NotNil(t, ret)
	assert.Equal(t, ret["test-component"].Status, v1alpha2.Untouched)
	assert.Equal(t, ret["test-component"].Message, "No error. test-component is untouched")
	assert.Equal(t, ret["test-target"].Status, v1alpha2.Updated)
	assert.Equal(t, ret["test-target"].Message, TestTargetSuccessMessage)
}
func TestApplyBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "update",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)

	assert.NotNil(t, err)
}

func TestRemove(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)
	assert.Nil(t, err)
}
func TestRemoveBad(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		err := json.Unmarshal(msg.Payload(), &request)
		assert.Nil(t, err)
		var response v1alpha2.COAResponse
		response.State = v1alpha2.InternalError
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)

	assert.NotNil(t, err)
}
func TestGetApply(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TEST_MQTT_LOCAL_ENABLED enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://localhost:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)
	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var request v1alpha2.COARequest
		json.Unmarshal(msg.Payload(), &request)
		var response v1alpha2.COAResponse
		response.Metadata = make(map[string]string)
		response.Metadata["request-id"] = request.Metadata["request-id"]
		if request.Method == "GET" {
			ret := make([]model.ComponentSpec, 0)
			data, _ := json.Marshal(ret)
			response.State = v1alpha2.OK
			response.Body = data
		} else {
			response.State = v1alpha2.OK
		}

		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data) //sending COARequest directly doesn't seem to work
		token.Wait()

	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	arr, err := provider.Get(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))

	err = provider.Init(config)
	assert.Nil(t, err)

	_, err = provider.Apply(context.Background(), model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{
		Target: "test-target",
		Components: []model.ComponentStep{{
			Action: "delete",
			Component: model.ComponentSpec{
				Name: "test-component",
				Type: "test-component",
			},
		}},
	}, false)
	assert.Nil(t, err)
}

func TestLocalApplyGet(t *testing.T) {
	testMQTT := os.Getenv("TEST_MQTT_LOCAL_ENABLED")
	if testMQTT == "" {
		t.Skip("Skipping because TES_MQTT enviornment variable is not set")
	}
	config := MQTTTargetProviderConfig{
		Name:           "me",
		BrokerAddress:  "tcp://127.0.0.1:1883",
		ClientID:       "coa-test2",
		RequestTopic:   "coa-request",
		ResponseTopic:  "coa-response",
		TimeoutSeconds: 8,
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.Nil(t, err)

	c := newLocalClient(t, config.BrokerAddress)
	defer c.Disconnect(250)

	ctx := context.TODO()
	correlationId := uuid.New().String()
	resourceId := uuid.New().String()
	ctx = contexts.PopulateResourceIdAndCorrelationIdToDiagnosticLogContext(correlationId, resourceId, ctx)

	if token := c.Subscribe(config.RequestTopic, 0, func(client gmqtt.Client, msg gmqtt.Message) {
		var response v1alpha2.COAResponse
		response.State = v1alpha2.OK
		response.Metadata = make(map[string]string)
		var request v1alpha2.COARequest
		json.Unmarshal(msg.Payload(), &request)

		assert.NotEqual(t, ctx, request.Context)
		assert.NotNil(t, request.Context)
		diagCtx, ok := request.Context.Value(contexts.DiagnosticLogContextKey).(*contexts.DiagnosticLogContext)
		assert.True(t, ok)
		assert.NotNil(t, diagCtx)
		assert.Equal(t, correlationId, diagCtx.GetCorrelationId())
		assert.Equal(t, resourceId, diagCtx.GetResourceId())
		response.Metadata["request-id"] = request.Metadata["request-id"]
		if request.Method == "GET" {
			ret := make([]model.ComponentSpec, 0)
			data, _ := json.Marshal(ret)
			response.State = v1alpha2.OK
			response.Body = data
		} else {
			response.State = v1alpha2.OK
		}
		data, _ := json.Marshal(response)
		token := c.Publish(config.ResponseTopic, 0, false, data)
		token.Wait()
	}); token.Wait() && token.Error() != nil {
		if token.Error().Error() != "subscription exists" {
			panic(token.Error())
		}
	}

	_, err = provider.Apply(ctx, model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, model.DeploymentStep{}, false)
	assert.Nil(t, err)
	arr, err := provider.Get(ctx, model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
	}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(arr))
}

func TestInitFailed(t *testing.T) {
	config := MQTTTargetProviderConfig{
		Name:          "me",
		BrokerAddress: "tcp://8.8.8.8:1883",
		ClientID:      "coa-test2",
		RequestTopic:  "coa-request",
		ResponseTopic: "coa-response",
	}
	provider := MQTTTargetProvider{}
	err := provider.Init(config)
	assert.NotNil(t, err)
}

// Conformance: you should call the conformance suite to ensure provider conformance
func TestConformanceSuite(t *testing.T) {
	provider := &MQTTTargetProvider{}
	_ = provider.Init(MQTTTargetProviderConfig{})
	// assert.Nil(t, err) okay if provider is not fully initialized
	conformance.ConformanceSuite(t, provider)
}
