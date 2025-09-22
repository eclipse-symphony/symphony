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

	// Test CreateCert
	req := cert.CertRequest{
		TargetName:  "test-target",
		Namespace:   "test-namespace",
		Duration:    time.Hour * 2160, // 90 days
		RenewBefore: time.Hour * 360,  // 15 days
		CommonName:  "test-service",
		DNSNames:    []string{"test-target", "test-target.test-namespace"},
		IssuerName:  "test-issuer",
		ServiceName: "test-secret",
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
