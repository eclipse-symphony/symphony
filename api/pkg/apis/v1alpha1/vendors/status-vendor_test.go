/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	symmanagers "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/reference"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/status"
	symproviders "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	mqttbinding "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/bindings/mqtt"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	httpreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	k8sreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/k8s"
	mqttreporter "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/mqtt"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func statusVendorConfig(reporterType string, reporterConfig interface{}) vendors.VendorConfig {
	return vendors.VendorConfig{
		Type: "vendors.status", LoopInterval: 1,
		Managers: []managers.ManagerConfig{{
			Name: "status-manager", Type: "managers.symphony.status",
			Properties: map[string]string{
				"poll.enabled": "true", "target.name": "edge-01",
				"providers.reporter": "reporter",
				"report.properties":  `{"agent.site":"$env:SYMPHONY_STATUS_TEST_SITE"}`,
			},
			Providers: map[string]managers.ProviderConfig{
				"reporter": {Type: reporterType, Config: reporterConfig},
			},
		}},
	}
}

func TestStatusVendorFactories(t *testing.T) {
	config := statusVendorConfig("providers.reporter.http", httpreporter.HTTPReporterConfig{Url: "http://localhost"})
	factory := SymphonyVendorFactory{}
	vendor, err := factory.CreateVendor(config)
	require.NoError(t, err)
	require.IsType(t, &StatusVendor{}, vendor)
	ps, err := (symproviders.SymphonyProviderFactory{}).CreateProviders(config)
	require.NoError(t, err)
	require.NoError(t, vendor.Init(config, []managers.IManagerFactroy{&symmanagers.SymphonyManagerFactory{}}, ps, nil))
	assert.Empty(t, vendor.GetEndpoints())
	assert.Equal(t, "Status", vendor.GetInfo().Name)
	require.NoError(t, vendor.Shutdown(context.Background()))
	config.LoopInterval = 0
	assert.Error(t, (&StatusVendor{}).Init(config, nil, nil, nil))
	config.LoopInterval = 1
	config.Managers = nil
	assert.Error(t, (&StatusVendor{}).Init(config, nil, nil, nil))
}

func startStatusTestBroker(t *testing.T) string {
	t.Helper()
	binary, err := exec.LookPath("mosquitto")
	if err != nil {
		t.Skip("MQTT integration test requires mosquitto on PATH")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	configPath := filepath.Join(t.TempDir(), "mosquitto.conf")
	require.NoError(t, os.WriteFile(configPath, []byte(fmt.Sprintf("listener %d 127.0.0.1\nallow_anonymous true\npersistence false\n", port)), 0600))
	command := exec.Command(binary, "-c", configPath)
	require.NoError(t, command.Start())
	t.Cleanup(func() {
		if err := command.Process.Kill(); err != nil {
			t.Errorf("stop test MQTT broker: %v", err)
		}
		_ = command.Wait()
	})
	address := fmt.Sprintf("127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		connection, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err != nil {
			return false
		}
		_ = connection.Close()
		return true
	}, 5*time.Second, 20*time.Millisecond, "test MQTT broker did not become ready")
	return "tcp://" + address
}

func statusTestKubernetesClient(t *testing.T, target *unstructured.Unstructured) dynamic.Interface {
	t.Helper()
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		path := "/apis/fabric.symphony/v1/namespaces/default/targets/edge-01"
		if target == nil || (r.URL.Path != path && r.URL.Path != path+"/status") {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(metav1.Status{Status: "Failure", Reason: metav1.StatusReasonNotFound, Code: http.StatusNotFound})
			return
		}
		switch r.Method {
		case http.MethodGet:
		case http.MethodPut:
			if r.URL.Path != path+"/status" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var update unstructured.Unstructured
			if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
				t.Errorf("decode status update: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			target.Object["status"] = update.Object["status"]
		case http.MethodDelete:
			target = nil
			_ = json.NewEncoder(w).Encode(metav1.Status{Status: "Success"})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewEncoder(w).Encode(target); err != nil {
			t.Errorf("encode target: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := dynamic.NewForConfig(&rest.Config{Host: server.URL})
	require.NoError(t, err)
	return client
}

func TestStatusVendorMQTTEndToEnd(t *testing.T) {
	broker := startStatusTestBroker(t)
	t.Setenv("SYMPHONY_STATUS_TEST_SITE", "factory-a")
	target := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "fabric.symphony/v1", "kind": "Target",
		"metadata": map[string]interface{}{"name": "edge-01", "namespace": "default"},
		"status": map[string]interface{}{
			"properties":         map[string]interface{}{"existing": "preserved"},
			"provisioningStatus": map[string]interface{}{"status": "Succeeded"},
		},
	}}
	client := statusTestKubernetesClient(t, target)
	receiver := &AgentVendor{ReferenceManager: &reference.ReferenceManager{
		Reporter: &k8sreporter.K8sReporter{DynamicClient: client},
	}}
	bindingConfig := mqttbinding.MQTTBindingConfig{
		BrokerAddress: broker, ClientID: "status-control-plane",
		RequestTopic: "status/edge-01/request", ResponseTopic: "status/edge-01/response",
		TimeoutSeconds: 2,
	}
	binding := &mqttbinding.MQTTBinding{}
	require.NoError(t, binding.Launch(bindingConfig, receiver.GetEndpoints()))
	t.Cleanup(func() { require.NoError(t, binding.Shutdown(context.Background())) })
	reporterConfig := mqttreporter.MQTTReporterConfig{MQTTBindingConfig: bindingConfig}
	reporterConfig.ClientID = "status-agent"
	config := statusVendorConfig("providers.reporter.mqtt", reporterConfig)
	ps, err := (symproviders.SymphonyProviderFactory{}).CreateProviders(config)
	require.NoError(t, err)
	r, ok := ps["status-manager"]["reporter"].(*mqttreporter.MQTTReporter)
	require.True(t, ok)
	vendor := &StatusVendor{}
	require.NoError(t, vendor.Init(config, []managers.IManagerFactroy{&symmanagers.SymphonyManagerFactory{}}, ps, nil))
	t.Cleanup(func() { require.NoError(t, vendor.Shutdown(context.Background())) })
	gvr := schema.GroupVersionResource{Group: "fabric.symphony", Version: "v1", Resource: "targets"}
	readProperties := func() map[string]string {
		t.Helper()
		result, err := client.Resource(gvr).Namespace("default").Get(context.Background(), "edge-01", metav1.GetOptions{})
		require.NoError(t, err)
		properties, found, err := unstructured.NestedStringMap(result.Object, "status", "properties")
		require.NoError(t, err)
		require.True(t, found)
		provisioning, _, err := unstructured.NestedString(result.Object, "status", "provisioningStatus", "status")
		require.NoError(t, err)
		assert.Equal(t, "Succeeded", provisioning)
		return properties
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var runErr error
	go func() {
		runErr = vendor.RunLoop(ctx, 20*time.Millisecond)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
			assert.NoError(t, runErr)
		case <-time.After(5 * time.Second):
			t.Error("status vendor did not stop")
		}
	})
	require.Eventually(t, func() bool {
		return readProperties()["agent.site"] == "factory-a"
	}, 5*time.Second, 20*time.Millisecond)
	cancel()
	<-done
	require.NoError(t, runErr)
	properties := readProperties()
	assert.Equal(t, "preserved", properties["existing"])
	_, err = time.Parse(time.RFC3339Nano, properties[status.ReportedAtProperty])
	require.NoError(t, err)
	r.MQTTClient.Disconnect(1000)
	t.Setenv("SYMPHONY_STATUS_TEST_SITE", "factory-b")
	require.Empty(t, vendor.Managers[0].(*status.StatusManager).Poll())
	assert.Equal(t, "factory-b", readProperties()["agent.site"])
	require.NoError(t, client.Resource(gvr).Namespace("default").Delete(context.Background(), "edge-01", metav1.DeleteOptions{}))
	errs := vendor.Managers[0].(*status.StatusManager).Poll()
	require.Len(t, errs, 1, "receiver failures must be returned to the agent")
	assert.Equal(t, v1alpha2.InternalError, v1alpha2.GetErrorState(errs[0]))
}
