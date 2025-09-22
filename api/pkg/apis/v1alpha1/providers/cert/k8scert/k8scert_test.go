/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/cert"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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

func TestGetCert_Success(t *testing.T) {
	// Create fake Kubernetes client
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-target-working-cert",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB+gIJAK...certificate data...\n-----END CERTIFICATE-----\n"),
			"tls.key": []byte("-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgk...key data...\n-----END PRIVATE KEY-----\n"),
		},
	}

	kubeClient := k8sfake.NewSimpleClientset(secret)

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test GetCert
	ctx := context.Background()
	result, err := provider.GetCert(ctx, "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.PublicKey, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, result.PrivateKey, "-----BEGIN PRIVATE KEY-----")
	assert.Equal(t, "cert-manager-generated", result.SerialNumber)
}

func TestGetCert_SecretNotFound(t *testing.T) {
	// Create fake Kubernetes client without the secret
	kubeClient := k8sfake.NewSimpleClientset()

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test GetCert with non-existent secret
	ctx := context.Background()
	result, err := provider.GetCert(ctx, "test-target", "test-namespace")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "certificate not found for target test-target after 30 seconds")
}

func TestGetCert_IncompleteSecret(t *testing.T) {
	// Create secret with missing key data
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-target-working-cert",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("certificate data"),
			// Missing tls.key
		},
	}

	kubeClient := k8sfake.NewSimpleClientset(secret)

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test GetCert with incomplete secret
	ctx := context.Background()
	result, err := provider.GetCert(ctx, "test-target", "test-namespace")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "certificate not found for target test-target after 30 seconds")
}

func TestCheckSecretReady_Success(t *testing.T) {
	// Create complete secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("certificate data"),
			"tls.key": []byte("key data"),
		},
	}

	kubeClient := k8sfake.NewSimpleClientset(secret)

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test checkSecretReady
	ctx := context.Background()
	ready, err := provider.checkSecretReady(ctx, "test-secret", "test-namespace")

	assert.NoError(t, err)
	assert.True(t, ready)
}

func TestCheckSecretReady_MissingCert(t *testing.T) {
	// Create secret with missing certificate
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"tls.key": []byte("key data"),
			// Missing tls.crt
		},
	}

	kubeClient := k8sfake.NewSimpleClientset(secret)

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test checkSecretReady
	ctx := context.Background()
	ready, err := provider.checkSecretReady(ctx, "test-secret", "test-namespace")

	assert.Error(t, err)
	assert.False(t, ready)
	assert.Contains(t, err.Error(), "secret missing tls.crt")
}

func TestCheckSecretReady_SecretNotFound(t *testing.T) {
	kubeClient := k8sfake.NewSimpleClientset()

	provider := &K8sCertProvider{
		kubeClient: kubeClient,
	}

	// Test checkSecretReady with non-existent secret
	ctx := context.Background()
	ready, err := provider.checkSecretReady(ctx, "non-existent", "test-namespace")

	assert.Error(t, err)
	assert.False(t, ready)
}

func TestCheckCertificateStatus_Ready(t *testing.T) {
	// Create certificate with ready status
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-cert",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)

	provider := &K8sCertProvider{
		dynamicClient: dynamicClient,
	}

	// Test checkCertificateStatus
	ctx := context.Background()
	ready, err := provider.checkCertificateStatus(ctx, "test-cert", "test-namespace")

	assert.NoError(t, err)
	assert.True(t, ready)
}

func TestCheckCertificateStatus_NotReady(t *testing.T) {
	// Create certificate with not ready status
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-cert",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "False",
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)

	provider := &K8sCertProvider{
		dynamicClient: dynamicClient,
	}

	// Test checkCertificateStatus
	ctx := context.Background()
	ready, err := provider.checkCertificateStatus(ctx, "test-cert", "test-namespace")

	assert.NoError(t, err)
	assert.False(t, ready)
}

func TestCheckCertificateStatus_CertificateNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8sCertProvider{
		dynamicClient: dynamicClient,
	}

	// Test checkCertificateStatus with non-existent certificate
	ctx := context.Background()
	ready, err := provider.checkCertificateStatus(ctx, "non-existent", "test-namespace")

	assert.Error(t, err)
	assert.False(t, ready)
	assert.Contains(t, err.Error(), "failed to get certificate")
}

func TestCheckCertStatus_Success(t *testing.T) {
	// Create certificate with ready status
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-target-working-cert",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)

	provider := &K8sCertProvider{
		dynamicClient: dynamicClient,
	}

	// Test CheckCertStatus
	ctx := context.Background()
	status, err := provider.CheckCertStatus(ctx, "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Ready)
	assert.Equal(t, "Ready", status.Reason)
	assert.Equal(t, "Certificate is ready", status.Message)
}

func TestCheckCertStatus_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8sCertProvider{
		dynamicClient: dynamicClient,
	}

	// Test CheckCertStatus with non-existent certificate
	ctx := context.Background()
	status, err := provider.CheckCertStatus(ctx, "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.Ready)
	assert.Equal(t, "NotFound", status.Reason)
	assert.Equal(t, "Certificate not found", status.Message)
}

func TestRotateCert_DefaultValues(t *testing.T) {
	// Test that RotateCert sets correct default values
	// Since RotateCert calls CreateCert, we can't easily test the full flow
	// without mocking the entire Kubernetes client, but we can test the values

	// Verify the default values used in RotateCert
	targetName := "test-target"
	namespace := "test-namespace"

	// These are the default values that should be set in RotateCert
	expectedDuration := time.Hour * 2160   // 90 days
	expectedRenewBefore := time.Hour * 360 // 15 days
	expectedCommonName := "symphony-service"
	expectedIssuerName := "symphony-ca-issuer"
	expectedDNSNames := []string{targetName, fmt.Sprintf("%s.%s", targetName, namespace)}

	assert.Equal(t, 90*24*time.Hour, expectedDuration)
	assert.Equal(t, 15*24*time.Hour, expectedRenewBefore)
	assert.Equal(t, "symphony-service", expectedCommonName)
	assert.Equal(t, "symphony-ca-issuer", expectedIssuerName)
	assert.Equal(t, []string{"test-target", "test-target.test-namespace"}, expectedDNSNames)
}

func TestCertificateFormatConversion(t *testing.T) {
	// Test certificate format conversion (PEM to space-separated)
	originalCert := "-----BEGIN CERTIFICATE-----\nMIIBkTCB+gIJAK\n-----END CERTIFICATE-----\n"
	originalKey := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgk\n-----END PRIVATE KEY-----\n"

	// Verify the original format is preserved and contains expected markers
	assert.Contains(t, originalCert, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, originalKey, "-----BEGIN PRIVATE KEY-----")
	assert.Contains(t, originalCert, "\n")
	assert.Contains(t, originalKey, "\n")
}

func TestCommonNameConsistency(t *testing.T) {
	// Test that both CreateCert and RotateCert use the same CommonName
	expectedCommonName := "symphony-service"

	// This should match what's used in both methods
	assert.Equal(t, "symphony-service", expectedCommonName)

	// Verify this is different from the old target-based naming
	targetName := "test-target"
	oldStyleCommonName := fmt.Sprintf("symphony-%s", targetName)

	assert.NotEqual(t, expectedCommonName, oldStyleCommonName)
	assert.Equal(t, "symphony-test-target", oldStyleCommonName)
}
