/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/eclipse-symphony/symphony/cli/config"
	"github.com/spf13/cobra"
)

const (
	localStatusRequestTopic  = "symphony-status-request"
	localStatusResponseTopic = "symphony-status-response"
)

func parseRuntimeConfig(data []byte) (map[string]interface{}, error) {
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("runtime configuration must be a JSON object")
	}
	return cfg, nil
}

func runtimeHTTPPort(cfg map[string]interface{}) (int, error) {
	bindings, ok := cfg["bindings"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("configuration requires bindings")
	}
	for _, entry := range bindings {
		binding, ok := entry.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("invalid binding")
		}
		if binding["type"] != "bindings.http" {
			continue
		}
		settings, ok := binding["config"].(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("HTTP binding requires config")
		}
		if tls, _ := settings["tls"].(bool); tls {
			continue
		}
		port, ok := settings["port"].(float64)
		if !ok || port < 1 || port > 65535 || port != float64(int(port)) {
			return 0, fmt.Errorf("HTTP binding port must be an integer between 1 and 65535")
		}
		return int(port), nil
	}
	return 0, fmt.Errorf("local control plane requires a non-TLS HTTP binding")
}

func runtimeVendors(cfg map[string]interface{}) ([]interface{}, error) {
	api, ok := cfg["api"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("configuration requires api")
	}
	vendors, ok := api["vendors"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("configuration requires api.vendors")
	}
	return vendors, nil
}

func configureLocalMQTT(cfg map[string]interface{}, address string) (config.MaestroMqttConfig, error) {
	vendors, err := runtimeVendors(cfg)
	if err != nil {
		return config.MaestroMqttConfig{}, err
	}
	hasTargets := false
	for _, entry := range vendors {
		if vendor, ok := entry.(map[string]interface{}); ok && vendor["type"] == "vendors.targets" {
			hasTargets = true
		}
	}
	if !hasTargets {
		return config.MaestroMqttConfig{}, fmt.Errorf("MQTT status reporting requires vendors.targets in the API configuration")
	}
	bindings, ok := cfg["bindings"].([]interface{})
	if !ok {
		return config.MaestroMqttConfig{}, fmt.Errorf("configuration requires bindings")
	}
	for _, entry := range bindings {
		binding, ok := entry.(map[string]interface{})
		if !ok || binding["type"] != "bindings.mqtt" {
			continue
		}
		settings, ok := binding["config"].(map[string]interface{})
		if !ok {
			return config.MaestroMqttConfig{}, fmt.Errorf("MQTT binding requires config")
		}
		if settings["name"] == "maestro-status" {
			settings["brokerAddress"] = address
			return localMQTTContext(address), nil
		}
		if settings["requestTopic"] == localStatusRequestTopic || settings["responseTopic"] == localStatusResponseTopic {
			return config.MaestroMqttConfig{}, fmt.Errorf("existing MQTT binding conflicts with Maestro status topics")
		}
	}
	cfg["bindings"] = append(bindings, map[string]interface{}{
		"type": "bindings.mqtt",
		"config": map[string]interface{}{
			"name": "maestro-status", "brokerAddress": address, "clientID": "maestro-status-control-plane",
			"requestTopic": localStatusRequestTopic, "responseTopic": localStatusResponseTopic,
			"timeoutSeconds": 8,
		},
	})
	return localMQTTContext(address), nil
}

func localMQTTContext(address string) config.MaestroMqttConfig {
	return config.MaestroMqttConfig{
		BrokerAddress: address, RequestTopic: "coa-request", ResponseTopic: "coa-response",
		StatusRequestTopic: localStatusRequestTopic, StatusResponseTopic: localStatusResponseTopic,
		StatusRoute: "report",
	}
}

func resolveLocalBinary(override string) (string, error) {
	if override == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		name := "symphony-api"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		override = filepath.Join(home, ".symphony", name)
	}
	binary, err := exec.LookPath(override)
	if err != nil {
		return "", fmt.Errorf("find Symphony binary %q (use --api-binary or --agent-binary): %w", override, err)
	}
	return filepath.Abs(binary)
}

func writeRuntimeConfig(path string, cfg map[string]interface{}) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

type localProcess struct {
	command *exec.Cmd
	done    chan struct{}
	err     error
}

func startLocalProcess(command *exec.Cmd) (*localProcess, error) {
	if err := command.Start(); err != nil {
		return nil, err
	}
	process := &localProcess{command: command, done: make(chan struct{})}
	go func() {
		process.err = command.Wait()
		close(process.done)
	}()
	return process, nil
}

func (p *localProcess) stop() error {
	select {
	case <-p.done:
		return nil
	default:
	}
	if err := p.command.Process.Signal(os.Interrupt); err != nil && !errors.Is(err, os.ErrProcessDone) {
		if err := p.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
		return nil
	case <-timer.C:
		if err := p.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		<-p.done
		return nil
	}
}

func waitLocalReady(ctx context.Context, process *localProcess, probe func() error) error {
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var lastError error
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-process.done:
			return fmt.Errorf("%s exited before becoming ready: %v", process.command.Path, process.err)
		default:
		}
		if lastError = probe(); lastError == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-process.done:
			return fmt.Errorf("%s exited before becoming ready: %v", process.command.Path, process.err)
		case <-timer.C:
			return fmt.Errorf("process did not become ready: %w", lastError)
		case <-ticker.C:
		}
	}
}

func requireFreeLocalPort(port int) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
	if err != nil {
		return fmt.Errorf("local port %d is unavailable: %w", port, err)
	}
	return listener.Close()
}

func runLocalControlPlane(cmd *cobra.Command) (result error) {
	if noRestApi {
		return fmt.Errorf("--no-rest-api is only supported for Kubernetes installations")
	}
	if !withMqttBroker && (cmd.Flags().Changed("mqtt-port") || cmd.Flags().Changed("mqtt-binary")) {
		return fmt.Errorf("--mqtt-port and --mqtt-binary require --with-mqtt-broker")
	}
	binary, err := resolveLocalBinary(apiBinary)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	source := apiConfigFile
	if source == "" {
		source = filepath.Join(home, ".symphony", "symphony-api-no-k8s.json")
	}
	source, err = filepath.Abs(source)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read API configuration (use --api-config): %w", err)
	}
	cfg, err := parseRuntimeConfig(data)
	if err != nil {
		return err
	}
	port, err := runtimeHTTPPort(cfg)
	if err != nil {
		return err
	}
	if err := requireFreeLocalPort(port); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	effective := source
	mqtt := config.MaestroMqttConfig{}
	var broker *localProcess
	if withMqttBroker {
		if mqttPort < 1 || mqttPort > 65535 || mqttPort == port {
			return fmt.Errorf("--mqtt-port must be between 1 and 65535 and differ from the HTTP port")
		}
		if err := requireFreeLocalPort(mqttPort); err != nil {
			return err
		}
		brokerBinary, err := exec.LookPath(mqttBinary)
		if err != nil {
			return fmt.Errorf("local MQTT demo requires Mosquitto on PATH or --mqtt-binary: %w", err)
		}
		mqtt, err = configureLocalMQTT(cfg, fmt.Sprintf("tcp://127.0.0.1:%d", mqttPort))
		if err != nil {
			return err
		}
		runtimeDir := filepath.Join(home, ".symphony", "runtime")
		if err := os.MkdirAll(runtimeDir, 0700); err != nil {
			return err
		}
		file, err := os.CreateTemp(runtimeDir, "api-*.json")
		if err != nil {
			return err
		}
		effective = file.Name()
		if err := file.Close(); err != nil {
			return err
		}
		if err := writeRuntimeConfig(effective, cfg); err != nil {
			return err
		}
		brokerFile, err := os.CreateTemp(runtimeDir, "mosquitto-*.conf")
		if err != nil {
			return err
		}
		brokerConfig := brokerFile.Name()
		defer func() { result = errors.Join(result, os.Remove(brokerConfig)) }()
		_, writeErr := fmt.Fprintf(brokerFile, "listener %d 127.0.0.1\nallow_anonymous true\npersistence false\n", mqttPort)
		closeErr := brokerFile.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return err
		}
		command := exec.Command(brokerBinary, "-c", brokerConfig)
		command.Stdout, command.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
		broker, err = startLocalProcess(command)
		if err != nil {
			return fmt.Errorf("start MQTT broker: %w", err)
		}
		defer func() { result = errors.Join(result, broker.stop()) }()
		if err := waitLocalReady(ctx, broker, func() error {
			connection, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", mqttPort), 200*time.Millisecond)
			if err != nil {
				return err
			}
			return connection.Close()
		}); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Test MQTT broker: %s (loopback only, stopped with this command)\n", mqtt.BrokerAddress)
	}
	apiURL := fmt.Sprintf("http://127.0.0.1:%d/v1alpha2", port)
	command := exec.Command(binary, "-c", effective, "-l", "Debug")
	command.Env = append(os.Environ(), "SYMPHONY_API_URL="+apiURL+"/", "USE_SERVICE_ACCOUNT_TOKENS=false")
	command.Stdin, command.Stdout, command.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
	api, err := startLocalProcess(command)
	if err != nil {
		return fmt.Errorf("start Symphony API: %w", err)
	}
	defer func() { result = errors.Join(result, api.stop()) }()
	client := &http.Client{Timeout: 500 * time.Millisecond}
	if err := waitLocalReady(ctx, api, func() error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/greetings", nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body)
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("readiness endpoint returned HTTP %d", response.StatusCode)
		}
		return nil
	}); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	if err := config.UpdateMaestroConfig("no-k8s", "127.0.0.1", port, effective, mqtt); err != nil {
		return fmt.Errorf("save local Maestro context: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Symphony API ready: %s\nConfiguration: %s\nContext: no-k8s\nPress Ctrl+C to stop.\n", apiURL, effective)
	if broker == nil {
		select {
		case <-ctx.Done():
			return nil
		case <-api.done:
			return api.err
		}
	}
	select {
	case <-ctx.Done():
		return nil
	case <-api.done:
		return api.err
	case <-broker.done:
		return fmt.Errorf("MQTT broker exited while the API was running: %v", broker.err)
	}
}
