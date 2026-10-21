/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveLocalMQTTContext(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	mqtt := MaestroMqttConfig{
		BrokerAddress: "tcp://127.0.0.1:18883", StatusRequestTopic: "status/request",
		StatusResponseTopic: "status/response", StatusRoute: "report",
	}
	if err := UpdateMaestroConfig("no-k8s", "127.0.0.1", 8082, "/runtime/api.json", mqtt); err != nil {
		t.Fatal(err)
	}
	ctx := GetMaestroConfig("").Contexts["no-k8s"]
	if ctx.Mqtt != mqtt || ctx.Url != "http://127.0.0.1:8082/v1alpha2" || ctx.ConfigFile != "/runtime/api.json" {
		t.Fatalf("context did not round-trip: %+v", ctx)
	}
	info, err := os.Stat(filepath.Join(home, ".symphony", ".config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("context permissions = %o, want 600", info.Mode().Perm())
	}
}
