/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func testRuntimeConfig(t *testing.T, text string) map[string]interface{} {
	t.Helper()
	cfg, err := parseRuntimeConfig([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func testCLIHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestRuntimeHTTPPort(t *testing.T) {
	cfg := testRuntimeConfig(t, `{"bindings":[
		{"type":"bindings.mqtt","config":{}},
		{"type":"bindings.http","config":{"port":8443,"tls":true}},
		{"type":"bindings.http","config":{"port":8082}}]}`)
	port, err := runtimeHTTPPort(cfg)
	if err != nil || port != 8082 {
		t.Fatalf("port=%d error=%v", port, err)
	}
	for _, text := range []string{
		`{}`, `{"bindings":[]}`, `{"bindings":[null]}`,
		`{"bindings":[{"type":"bindings.http","config":{"port":0}}]}`,
		`{"bindings":[{"type":"bindings.http","config":{"port":1.5}}]}`,
		`{"bindings":[{"type":"bindings.http","config":{"port":65536}}]}`,
	} {
		if _, err := runtimeHTTPPort(testRuntimeConfig(t, text)); err == nil {
			t.Errorf("expected invalid configuration: %s", text)
		}
	}
	if _, err := parseRuntimeConfig([]byte("null")); err == nil {
		t.Error("null configuration accepted")
	}
}

func TestConfigureLocalMQTT(t *testing.T) {
	cfg := testRuntimeConfig(t, `{"custom":"preserve","api":{"vendors":[{"type":"vendors.targets"}]},"bindings":[{"type":"bindings.http","config":{"port":8082}}]}`)
	mqtt, err := configureLocalMQTT(cfg, "tcp://127.0.0.1:18883")
	if err != nil {
		t.Fatal(err)
	}
	if mqtt.StatusRoute != "report" || mqtt.StatusRequestTopic != localStatusRequestTopic || mqtt.RequestTopic != "coa-request" {
		t.Fatalf("unexpected context: %+v", mqtt)
	}
	if cfg["custom"] != "preserve" || len(cfg["bindings"].([]interface{})) != 2 {
		t.Fatal("unrelated configuration was lost")
	}
	mqtt, err = configureLocalMQTT(cfg, "tcp://127.0.0.1:18884")
	if err != nil || len(cfg["bindings"].([]interface{})) != 2 {
		t.Fatalf("binding injection was not idempotent: %v", err)
	}
	binding := cfg["bindings"].([]interface{})[1].(map[string]interface{})["config"].(map[string]interface{})
	if binding["brokerAddress"] != mqtt.BrokerAddress {
		t.Fatal("broker override not applied")
	}
	cfg = testRuntimeConfig(t, `{"api":{"vendors":[]},"bindings":[]}`)
	if _, err := configureLocalMQTT(cfg, "tcp://localhost:1883"); err == nil {
		t.Error("missing Target receiver accepted")
	}
}

func TestWriteRuntimeConfigPreservesSource(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.json")
	original := `{"custom":{"value":"keep"},"bindings":[{"type":"bindings.http","config":{"port":8082}}]}`
	if err := os.WriteFile(source, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := testRuntimeConfig(t, original)
	output := filepath.Join(dir, "runtime", "effective.json")
	if err := writeRuntimeConfig(output, cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil || string(data) != original {
		t.Fatalf("source changed: %v", err)
	}
	port, err := extractPortFromConfig(output)
	if err != nil || port != 8082 {
		t.Fatalf("effective configuration unusable: %v", err)
	}
}

func TestLocalProcessHelper(t *testing.T) {
	if os.Getenv("MAESTRO_TEST_CHILD") != "true" {
		return
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestLocalProcessLifecycle(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestLocalProcessHelper$")
	command.Env = append(os.Environ(), "MAESTRO_TEST_CHILD=true")
	process, err := startLocalProcess(command)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := process.stop(); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err = waitLocalReady(ctx, process, func() error { return errors.New("not ready") })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("readiness did not honor cancellation: %v", err)
	}
	if err := process.stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-process.done:
	default:
		t.Fatal("child process was not reaped")
	}
	if _, err := startLocalProcess(exec.Command(filepath.Join(t.TempDir(), "missing"))); err == nil {
		t.Error("missing executable accepted")
	}
}
