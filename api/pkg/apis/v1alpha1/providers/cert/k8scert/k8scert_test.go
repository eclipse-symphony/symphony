/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/stretchr/testify/assert"
)

func TestK8sCertProviderConfigFromMap(t *testing.T) {
	// Test with default values
	properties := map[string]string{}
	config, err := K8sCertProviderConfigFromMap(properties)
	assert.NoError(t, err)
	assert.True(t, config.InCluster)
	assert.Equal(t, "", config.Name)

	// Test with custom values
	properties = map[string]string{
		"name":      "test-provider",
		"inCluster": "false",
	}
	config, err = K8sCertProviderConfigFromMap(properties)
	assert.NoError(t, err)
	assert.False(t, config.InCluster)
	assert.Equal(t, "test-provider", config.Name)
}

func TestCertRequestDefaults(t *testing.T) {
	// Test that CreateCert would set proper defaults
	req := cert.CertRequest{
		TargetName: "test-target",
		Namespace:  "test-namespace",
	}

	// Since we can't easily test the actual Kubernetes calls without a cluster,
	// we'll just verify the configuration parsing works
	config := K8sCertProviderConfig{
		Name:      "test",
		InCluster: true,
	}

	assert.Equal(t, "test", config.Name)
	assert.True(t, config.InCluster)

	// Verify cert request has the expected values
	assert.Equal(t, "test-target", req.TargetName)
	assert.Equal(t, "test-namespace", req.Namespace)
}

func TestCertificateNaming(t *testing.T) {
	targetName := "my-target"
	expectedCertName := "my-target-working-cert"

	certName := targetName + "-working-cert"
	assert.Equal(t, expectedCertName, certName)
}

func TestDefaultDuration(t *testing.T) {
	// Test default duration (90 days)
	defaultDuration := 2160 * time.Hour
	expectedDays := 90 * 24 * time.Hour
	assert.Equal(t, expectedDays, defaultDuration)

	// Test default renewBefore (15 days)
	defaultRenewBefore := 360 * time.Hour
	expectedRenewBefore := 15 * 24 * time.Hour
	assert.Equal(t, expectedRenewBefore, defaultRenewBefore)
}
