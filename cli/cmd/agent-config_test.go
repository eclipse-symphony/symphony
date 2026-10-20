/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/spf13/cobra"
)

func agentTestOptions(t *testing.T) config.MaestroContext {
	t.Helper()
	oldTarget, oldNamespace := targetName, targetNamespace
	oldBroker, oldRequest, oldResponse, oldClient := mqttBrokerAddress, mqttRequestTopic, mqttResponseTopic, mqttClientId
	oldConfig, oldStatus, oldInterval := agentConfigFile, reportStatus, reportInterval
	oldBag, oldProperties := statusBagFile, statusProperties
	t.Cleanup(func() {
		targetName, targetNamespace = oldTarget, oldNamespace
		mqttBrokerAddress, mqttRequestTopic, mqttResponseTopic, mqttClientId = oldBroker, oldRequest, oldResponse, oldClient
		agentConfigFile, reportStatus, reportInterval = oldConfig, oldStatus, oldInterval
		statusBagFile, statusProperties = oldBag, oldProperties
	})
	targetName, targetNamespace = "edge-01", "factory"
	mqttBrokerAddress, mqttRequestTopic, mqttResponseTopic, mqttClientId = "tcp://127.0.0.1:18883", "deployment/request", "deployment/response", "edge-01"
	agentConfigFile, reportStatus, reportInterval = "custom-template.json", true, time.Second
	statusBagFile, statusProperties = "", nil
	return config.MaestroContext{Url: "http://localhost:8082/v1alpha2", User: "admin", Mqtt: localMQTTContext(mqttBrokerAddress)}
}

func TestAgentRuntimePreservesCustomConfiguration(t *testing.T) {
	ctx := agentTestOptions(t)
	statusProperties = []string{"agent.site=$env:SITE_NAME", "agent.answer=${{ 6 * 7 }}", "agent.value=a=b"}
	cfg := testRuntimeConfig(t, `{
	  "custom":{"preserve":true},
	  "api":{"vendors":[
	    {"type":"vendors.echo","extra":"keep"},
	    {"type":"vendors.solutionversion","managers":[
	      {"type":"managers.symphony.solutionversion","properties":{"custom":"keep"},
	       "providers":{"instance":{"type":"providers.target.script","config":{"getScript":"keep.sh"}}}}
	    ]}]},
	  "bindings":[
	    {"type":"bindings.http","config":{"port":9090}},
	    {"type":"bindings.mqtt","config":{"useTLS":"true","caCertPath":"keep.pem"}}
	  ]}`)
	if err := configureAgentRuntime(cfg, ctx); err != nil {
		t.Fatal(err)
	}
	if cfg["custom"].(map[string]interface{})["preserve"] != true {
		t.Fatal("custom fields lost")
	}
	bindings := cfg["bindings"].([]interface{})
	if bindings[0].(map[string]interface{})["config"].(map[string]interface{})["port"] != float64(9090) {
		t.Fatal("non-MQTT binding modified")
	}
	mqtt := bindings[1].(map[string]interface{})["config"].(map[string]interface{})
	if mqtt["caCertPath"] != "keep.pem" || mqtt["brokerAddress"] != mqttBrokerAddress {
		t.Fatal("MQTT settings lost or not updated")
	}
	vendors := cfg["api"].(map[string]interface{})["vendors"].([]interface{})
	deployment := vendors[1].(map[string]interface{})["managers"].([]interface{})[0].(map[string]interface{})
	if deployment["providers"].(map[string]interface{})["instance"].(map[string]interface{})["type"] != "providers.target.script" {
		t.Fatal("custom deployment provider replaced")
	}
	if deployment["properties"].(map[string]interface{})["targetNames"] != "edge-01" {
		t.Fatal("target identity not updated")
	}
	status := vendors[2].(map[string]interface{})
	manager := status["managers"].([]interface{})[0].(map[string]interface{})
	properties := manager["properties"].(map[string]interface{})
	bag, err := parseStatusBag([]byte(properties["report.properties"].(string)))
	if err != nil || bag["agent.site"] != "$env:SITE_NAME" || bag["agent.answer"] != "${{ 6 * 7 }}" || bag["agent.value"] != "a=b" {
		t.Fatalf("bag expressions were changed: %+v %v", bag, err)
	}
	reporter := manager["providers"].(map[string]interface{})["mqtt-status-reporter"].(map[string]interface{})["config"].(map[string]interface{})
	if reporter["route"] != "report" || reporter["requestTopic"] != localStatusRequestTopic || reporter["clientID"] == mqttClientId {
		t.Fatalf("incorrect reporter configuration: %+v", reporter)
	}
	if err := configureAgentRuntime(cfg, ctx); err != nil {
		t.Fatal(err)
	}
	if len(cfg["api"].(map[string]interface{})["vendors"].([]interface{})) != 3 {
		t.Fatal("status vendor duplicated")
	}
}

func TestAgentConfigurationOutputAndValidation(t *testing.T) {
	ctx := agentTestOptions(t)
	home := testCLIHome(t)
	agentConfigFile = filepath.Join(home, "template.json")
	template := `{"custom":"preserve","api":{"vendors":[]},"bindings":[]}`
	if err := os.WriteFile(agentConfigFile, []byte(template), 0600); err != nil {
		t.Fatal(err)
	}
	statusBagFile = filepath.Join(home, "bag.json")
	if err := os.WriteFile(statusBagFile, []byte(`{"agent.site":"$env:SITE_NAME"}`), 0600); err != nil {
		t.Fatal(err)
	}
	path, err := updateAgentConfig(ctx)
	if err != nil || !strings.Contains(path, "factory-edge-01") {
		t.Fatalf("generated path=%q error=%v", path, err)
	}
	data, err := os.ReadFile(agentConfigFile)
	if err != nil || string(data) != template {
		t.Fatal("template was modified")
	}
	agentConfigFile = path
	generatedSource, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	newPath, err := updateAgentConfig(ctx)
	if err != nil || newPath == path {
		t.Fatalf("generated template would be overwritten: path=%q error=%v", newPath, err)
	}
	unchanged, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(generatedSource, unchanged) {
		t.Fatal("generated source configuration was modified")
	}
	for _, invalid := range []string{"null", "[]", `{"value":null}`, `{"value":1}`, `{"agent.reportedAt":"override"}`} {
		if _, err := parseStatusBag([]byte(invalid)); err == nil {
			t.Errorf("invalid bag accepted: %s", invalid)
		}
	}
	reportInterval = 1500 * time.Millisecond
	if _, err := updateAgentConfig(ctx); err == nil {
		t.Error("fractional reporting interval accepted")
	}
	targetName = "../escape"
	if err := resolveMqttSettingsFromContext(ctx); err == nil {
		t.Error("unsafe Target name accepted")
	}
}

func TestRegisterTargetNamespaceAndClientID(t *testing.T) {
	settings := agentTestOptions(t)
	var registered atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/users/auth" {
			_, _ = w.Write([]byte(`{"accessToken":"test"}`))
			return
		}
		if r.URL.Path != "/targets/registry/edge-01" || r.URL.Query().Get("namespace") != "factory" {
			t.Errorf("incorrect registration URL: %s", r.URL)
		}
		var target Target
		if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		binding := target.Spec.Topologies[0].Bindings[0]
		if binding.Config["clientID"] == "" || binding.Config["clientID"] == mqttClientId || binding.Config["clientId"] != "" {
			t.Errorf("incorrect MQTT client identity: %+v", binding.Config)
		}
		if target.Metadata.Namespace != "factory" {
			t.Errorf("missing metadata namespace: %+v", target.Metadata)
		}
		registered.Store(true)
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()
	settings.Url = server.URL
	if err := createTarget(settings); err != nil {
		t.Fatal(err)
	}
	if !registered.Load() {
		t.Fatal("Target was not registered")
	}
}

func TestWatchAgentStatusDistinctReports(t *testing.T) {
	settings := agentTestOptions(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/users/auth" {
			_, _ = w.Write([]byte(`{"accessToken":"test"}`))
			return
		}
		if r.URL.Query().Get("namespace") != "factory" {
			t.Error("watch omitted namespace")
		}
		timestamp := "first"
		if calls.Add(1) > 2 {
			timestamp = "second"
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": map[string]interface{}{"properties": map[string]string{"agent.reportedAt": timestamp}},
		})
	}))
	defer server.Close()
	settings.Url = server.URL
	var out bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&out)
	command.SetErr(&bytes.Buffer{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := watchAgentStatus(ctx, command, settings, time.Millisecond, 2); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "agent.reportedAt") != 2 || calls.Load() != 3 {
		t.Fatalf("duplicate reports emitted: %s", out.String())
	}
}
