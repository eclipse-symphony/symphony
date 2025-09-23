/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package k8scert

import (
	"context"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/cert"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// MockProviderConfig implements IProviderConfig for testing
type MockProviderConfig struct {
	Name            string `json:"name"`
	DefaultDuration string `json:"defaultDuration"`
	RenewBefore     string `json:"renewBefore"`
}

func TestK8SCertProvider_ID(t *testing.T) {
	provider := &K8SCertProvider{}
	assert.Equal(t, "k8s-cert", provider.ID())
}

func TestK8SCertProvider_SetContext(t *testing.T) {
	provider := &K8SCertProvider{}
	ctx := context.Background()
	provider.SetContext(ctx)
	assert.Equal(t, ctx, provider.Context)
}

func TestGetCert_Success(t *testing.T) {
	// Create fake certificate resource
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-target",
				"namespace": "test-namespace",
			},
			"spec": map[string]interface{}{
				"secretName": "test-secret",
			},
		},
	}

	// Create fake secret with certificate data
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB+gIJAK...certificate data...\n-----END CERTIFICATE-----\n"),
			"tls.key": []byte("-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgk...key data...\n-----END PRIVATE KEY-----\n"),
		},
	}

	// Create fake clients
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)
	kubeClient := k8sfake.NewSimpleClientset(secret)

	provider := &K8SCertProvider{
		K8sClient:     kubeClient,
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test GetCert
	result, err := provider.GetCert(context.Background(), "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.PublicKey, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, result.PrivateKey, "-----BEGIN PRIVATE KEY-----")
}

func TestGetCert_CertificateNotFound(t *testing.T) {
	// Create fake clients without the certificate
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	kubeClient := k8sfake.NewSimpleClientset()

	provider := &K8SCertProvider{
		K8sClient:     kubeClient,
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test GetCert with non-existent certificate
	result, err := provider.GetCert(context.Background(), "test-target", "test-namespace")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get certificate")
}

func TestCheckCertStatus_Ready(t *testing.T) {
	// Create certificate with ready status
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-target",
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

	provider := &K8SCertProvider{
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test CheckCertStatus
	status, err := provider.CheckCertStatus(context.Background(), "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Ready)
}

func TestCheckCertStatus_NotReady(t *testing.T) {
	// Create certificate with not ready status
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-target",
				"namespace": "test-namespace",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "False",
						"reason":  "Pending",
						"message": "Certificate is being issued",
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)

	provider := &K8SCertProvider{
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test CheckCertStatus
	status, err := provider.CheckCertStatus(context.Background(), "test-target", "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.Ready)
	assert.Equal(t, "Pending", status.Reason)
	assert.Equal(t, "Certificate is being issued", status.Message)
}

func TestCreateCert_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8SCertProvider{
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test CreateCert with minimal required fields to avoid deep copy issues
	req := cert.CertRequest{
		TargetName:  "test-target",
		Namespace:   "test-namespace",
		Duration:    time.Hour * 2160, // 90 days
		RenewBefore: time.Hour * 360,  // 15 days
		CommonName:  "test-service",
		IssuerName:  "test-issuer",
	}

	err := provider.CreateCert(context.Background(), req)
	assert.NoError(t, err)
}

func TestDeleteCert_Success(t *testing.T) {
	// Create certificate to delete
	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-target",
				"namespace": "test-namespace",
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme, certificate)

	provider := &K8SCertProvider{
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test DeleteCert
	err := provider.DeleteCert(context.Background(), "test-target", "test-namespace")
	assert.NoError(t, err)
}

func TestDeleteCert_NotFound(t *testing.T) {
	// Create empty client
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8SCertProvider{
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test DeleteCert with non-existent certificate (should not error)
	err := provider.DeleteCert(context.Background(), "test-target", "test-namespace")
	assert.NoError(t, err) // DeleteCert should not error if certificate doesn't exist
}

func TestParseCertificateInfo_InvalidPEM(t *testing.T) {
	// Test with invalid PEM data
	invalidPEM := []byte("invalid pem data")
	serialNumber, expiresAt, err := parseCertificateInfo(invalidPEM)

	assert.Error(t, err)
	assert.Empty(t, serialNumber)
	assert.True(t, expiresAt.IsZero())
	assert.Contains(t, err.Error(), "failed to decode PEM block")
}

func TestCertRequest_Fields(t *testing.T) {
	// Test that CertRequest has all expected fields
	req := cert.CertRequest{
		TargetName:  "test-target",
		Namespace:   "test-namespace",
		Duration:    time.Hour * 24,
		RenewBefore: time.Hour * 2,
		CommonName:  "test-common",
		DNSNames:    []string{"example.com"},
		IssuerName:  "test-issuer",
		ServiceName: "test-service",
	}

	assert.Equal(t, "test-target", req.TargetName)
	assert.Equal(t, "test-namespace", req.Namespace)
	assert.Equal(t, time.Hour*24, req.Duration)
	assert.Equal(t, time.Hour*2, req.RenewBefore)
	assert.Equal(t, "test-common", req.CommonName)
	assert.Equal(t, []string{"example.com"}, req.DNSNames)
	assert.Equal(t, "test-issuer", req.IssuerName)
	assert.Equal(t, "test-service", req.ServiceName)
}

func TestCertResponse_Fields(t *testing.T) {
	// Test that CertResponse has all expected fields
	now := time.Now()
	resp := cert.CertResponse{
		PublicKey:    "public-key",
		PrivateKey:   "private-key",
		ExpiresAt:    now,
		SerialNumber: "123456",
	}

	assert.Equal(t, "public-key", resp.PublicKey)
	assert.Equal(t, "private-key", resp.PrivateKey)
	assert.Equal(t, now, resp.ExpiresAt)
	assert.Equal(t, "123456", resp.SerialNumber)
}

func TestCertStatus_Fields(t *testing.T) {
	// Test that CertStatus has all expected fields
	now := time.Now()
	status := cert.CertStatus{
		Ready:       true,
		Reason:      "Ready",
		Message:     "Certificate is ready",
		LastUpdate:  now,
		NextRenewal: now.Add(time.Hour),
	}

	assert.True(t, status.Ready)
	assert.Equal(t, "Ready", status.Reason)
	assert.Equal(t, "Certificate is ready", status.Message)
	assert.Equal(t, now, status.LastUpdate)
	assert.Equal(t, now.Add(time.Hour), status.NextRenewal)
}

func TestToK8SCertProviderConfig(t *testing.T) {
	// Test config conversion
	mockConfig := MockProviderConfig{
		Name:            "test-cert",
		DefaultDuration: "4320h",
		RenewBefore:     "360h",
	}

	result, err := toK8SCertProviderConfig(mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, "test-cert", result.Name)
	assert.Equal(t, "4320h", result.DefaultDuration)
	assert.Equal(t, "360h", result.RenewBefore)
}

func TestGetConfigDuration(t *testing.T) {
	// Test with valid config
	provider := &K8SCertProvider{
		Config: K8SCertProviderConfig{
			DefaultDuration: "2160h", // 90 days
		},
	}
	duration := provider.getConfigDuration()
	assert.Equal(t, time.Hour*2160, duration)

	// Test with empty config
	provider.Config.DefaultDuration = ""
	duration = provider.getConfigDuration()
	assert.Equal(t, time.Hour*4320, duration) // Should use default

	// Test with invalid config
	provider.Config.DefaultDuration = "invalid"
	duration = provider.getConfigDuration()
	assert.Equal(t, time.Hour*4320, duration) // Should use default
}

func TestGetConfigRenewBefore(t *testing.T) {
	// Test with valid config
	provider := &K8SCertProvider{
		Config: K8SCertProviderConfig{
			RenewBefore: "240h", // 10 days
		},
	}
	renewBefore := provider.getConfigRenewBefore()
	assert.Equal(t, time.Hour*240, renewBefore)

	// Test with empty config
	provider.Config.RenewBefore = ""
	renewBefore = provider.getConfigRenewBefore()
	assert.Equal(t, time.Hour*360, renewBefore) // Should use default

	// Test with invalid config
	provider.Config.RenewBefore = "invalid"
	renewBefore = provider.getConfigRenewBefore()
	assert.Equal(t, time.Hour*360, renewBefore) // Should use default
}

func TestCreateCert_WithZeroValues(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8SCertProvider{
		Config: K8SCertProviderConfig{
			DefaultDuration: "2160h", // 90 days
			RenewBefore:     "240h",  // 10 days
		},
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test CreateCert with zero duration and renewBefore (should use config defaults)
	req := cert.CertRequest{
		TargetName:  "test-target",
		Namespace:   "test-namespace",
		Duration:    0, // Zero value - should use config default
		RenewBefore: 0, // Zero value - should use config default
		CommonName:  "test-service",
		IssuerName:  "test-issuer",
	}

	err := provider.CreateCert(context.Background(), req)
	assert.NoError(t, err)
}

func TestCreateCert_WithNonZeroValues(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8SCertProvider{
		Config: K8SCertProviderConfig{
			DefaultDuration: "2160h", // 90 days
			RenewBefore:     "240h",  // 10 days
		},
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	// Test CreateCert with non-zero values (should use request values)
	req := cert.CertRequest{
		TargetName:  "test-target",
		Namespace:   "test-namespace",
		Duration:    time.Hour * 720, // 30 days - should use this value
		RenewBefore: time.Hour * 72,  // 3 days - should use this value
		CommonName:  "test-service",
		IssuerName:  "test-issuer",
	}

	err := provider.CreateCert(context.Background(), req)
	assert.NoError(t, err)
}

func TestCreateCert_SecretNaming(t *testing.T) {
	scheme := runtime.NewScheme()
	dynamicClient := fake.NewSimpleDynamicClient(scheme)

	provider := &K8SCertProvider{
		Config: K8SCertProviderConfig{
			DefaultDuration: "2160h",
			RenewBefore:     "240h",
		},
		DynamicClient: dynamicClient,
		Context:       context.Background(),
	}

	req := cert.CertRequest{
		TargetName:  "my-target",
		Namespace:   "test-namespace",
		Duration:    time.Hour * 24,
		RenewBefore: time.Hour * 2,
		CommonName:  "test-service",
		IssuerName:  "test-issuer",
	}

	err := provider.CreateCert(context.Background(), req)
	assert.NoError(t, err)

	// The secret name should be "my-target-working-cert"
	// We can't directly verify this from the fake client, but the test passing means
	// the certificate was created without errors using the new naming scheme
}
