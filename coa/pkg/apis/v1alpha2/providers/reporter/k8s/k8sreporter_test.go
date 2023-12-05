/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8s

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	symphonyDevice := os.Getenv("SYMPHONY_DEVICE")
	if testK8s == "" || symphonyDevice == "" {
		t.Skip("Skipping because TEST_K8S or SYMPHONY_DEVICE enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
	err = provider.Report(symphonyDevice, "default", "fabric.symphony", "devices", "v1", map[string]string{
		"a": "ccc",
		"b": "ddd",
	}, false)
	assert.Nil(t, err)
}
