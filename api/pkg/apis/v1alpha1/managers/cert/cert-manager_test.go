/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package cert

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func createCertManager(t *testing.T) *CertManager {
	stateProvider := memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.NoError(t, err)

	secretProvider := mock.MockSecretProvider{}
	err = secretProvider.Init(mock.MockSecretProviderConfig{Name: "test-secret"})
	assert.NoError(t, err)

	certManager := &CertManager{}

	config := managers.ManagerConfig{
		Properties: map[string]string{
			"workingCertDuration":       "2160h",
			"workingCertRenewBefore":    "360h",
			"providers.persistentstate": "state",
			"providers.secret":          "secret",
		},
		Providers: map[string]managers.ProviderConfig{
			"state": {
				Type:   "providers.state.memory",
				Config: memorystate.MemoryStateProviderConfig{},
			},
			"secret": {
				Type:   "providers.secret.mock",
				Config: mock.MockSecretProviderConfig{Name: "test-secret"},
			},
		},
	}

	providers := map[string]providers.IProvider{
		"state":  &stateProvider,
		"secret": &secretProvider,
	}

	vendorContext := &contexts.VendorContext{}
	err = certManager.Init(vendorContext, config, providers)
	assert.NoError(t, err)

	return certManager
}

func TestCertManagerInit(t *testing.T) {
	stateProvider := memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.NoError(t, err)

	secretProvider := mock.MockSecretProvider{}
	err = secretProvider.Init(mock.MockSecretProviderConfig{Name: "test-secret"})
	assert.NoError(t, err)

	certManager := &CertManager{}

	config := managers.ManagerConfig{
		Properties: map[string]string{
			"workingCertDuration":       "1440h",
			"workingCertRenewBefore":    "240h",
			"providers.persistentstate": "state",
			"providers.secret":          "secret",
		},
		Providers: map[string]managers.ProviderConfig{
			"state": {
				Type:   "providers.state.memory",
				Config: memorystate.MemoryStateProviderConfig{},
			},
			"secret": {
				Type:   "providers.secret.mock",
				Config: mock.MockSecretProviderConfig{Name: "test-secret"},
			},
		},
	}

	providers := map[string]providers.IProvider{
		"state":  &stateProvider,
		"secret": &secretProvider,
	}

	vendorContext := &contexts.VendorContext{}
	err = certManager.Init(vendorContext, config, providers)
	assert.NoError(t, err)

	// Verify configuration
	assert.Equal(t, "1440h", certManager.Config.WorkingCertDuration)
	assert.Equal(t, "240h", certManager.Config.WorkingCertRenewBefore)
	assert.NotNil(t, certManager.StateProvider)
	assert.NotNil(t, certManager.SecretProvider)
}

func TestCertManagerInitWithDefaults(t *testing.T) {
	stateProvider := memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.NoError(t, err)

	secretProvider := mock.MockSecretProvider{}
	err = secretProvider.Init(mock.MockSecretProviderConfig{Name: "test-secret"})
	assert.NoError(t, err)

	certManager := &CertManager{}

	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "state",
			"providers.secret":          "secret",
		},
		Providers: map[string]managers.ProviderConfig{
			"state": {
				Type:   "providers.state.memory",
				Config: memorystate.MemoryStateProviderConfig{},
			},
			"secret": {
				Type:   "providers.secret.mock",
				Config: mock.MockSecretProviderConfig{Name: "test-secret"},
			},
		},
	}

	providers := map[string]providers.IProvider{
		"state":  &stateProvider,
		"secret": &secretProvider,
	}

	vendorContext := &contexts.VendorContext{}
	err = certManager.Init(vendorContext, config, providers)
	assert.NoError(t, err)

	// Verify default values
	assert.Equal(t, "2160h", certManager.Config.WorkingCertDuration)
	assert.Equal(t, "360h", certManager.Config.WorkingCertRenewBefore)
}

func TestCreateWorkingCert(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "test-target"
	namespace := "test-namespace"

	// Test creating a certificate
	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Verify certificate was created in StateProvider
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	entry, err := certManager.StateProvider.Get(ctx, getRequest)
	assert.NoError(t, err)
	assert.NotNil(t, entry.Body)

	// Verify certificate structure
	cert := entry.Body.(map[string]interface{})
	assert.Equal(t, targetName, cert["metadata"].(map[string]interface{})["name"])
	assert.Equal(t, namespace, cert["metadata"].(map[string]interface{})["namespace"])

	spec := cert["spec"].(map[string]interface{})
	assert.Equal(t, "test-target-tls", spec["secretName"])
	assert.Equal(t, "2160h", spec["duration"])
	assert.Equal(t, "360h", spec["renewBefore"])
}

func TestCreateWorkingCertAlreadyExists(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "test-target"
	namespace := "test-namespace"

	// Create certificate first time
	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Create again - should not error
	err = certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)
}

func TestDeleteWorkingCert(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "test-target"
	namespace := "test-namespace"

	// Create certificate first
	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Delete certificate
	err = certManager.DeleteWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Verify certificate was deleted
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	_, err = certManager.StateProvider.Get(ctx, getRequest)
	assert.Error(t, err)
	assert.True(t, v1alpha2.IsNotFound(err))
}

func TestDeleteWorkingCertNotFound(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "non-existent-target"
	namespace := "test-namespace"

	// Try to delete non-existent certificate
	err := certManager.DeleteWorkingCert(ctx, targetName, namespace)
	assert.Error(t, err)
}

func TestGetWorkingCert(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "test-target"
	namespace := "test-namespace"

	// Create certificate first
	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Get certificate
	public, private, err := certManager.GetWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)
	assert.NotEmpty(t, public)
	assert.NotEmpty(t, private)

	// Verify MockSecretProvider format: "secretName>>fieldName"
	expectedSecretName := "test-target-tls"
	assert.Equal(t, expectedSecretName+">>tls.crt", public)
	assert.Equal(t, expectedSecretName+">>tls.key", private)
}

func TestGetWorkingCertNotFound(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "non-existent-target"
	namespace := "test-namespace"

	// Try to get non-existent certificate
	_, _, err := certManager.GetWorkingCert(ctx, targetName, namespace)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "working certificate not found")
}

func TestCheckCertificateReady(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "test-target"
	namespace := "test-namespace"

	// Create certificate first
	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Certificate without Ready status should return false
	ready, err := certManager.CheckCertificateReady(ctx, targetName, namespace)
	assert.NoError(t, err)
	assert.False(t, ready)

	// Update certificate with Ready status
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	entry, err := certManager.StateProvider.Get(ctx, getRequest)
	assert.NoError(t, err)

	cert := entry.Body.(map[string]interface{})
	cert["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Ready",
				"status": "True",
			},
		},
	}

	upsertRequest := states.UpsertRequest{
		Value: states.StateEntry{
			ID:   targetName,
			Body: cert,
		},
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	_, err = certManager.StateProvider.Upsert(ctx, upsertRequest)
	assert.NoError(t, err)

	// Now certificate should be ready
	ready, err = certManager.CheckCertificateReady(ctx, targetName, namespace)
	assert.NoError(t, err)
	assert.True(t, ready)
}

func TestCheckCertificateReadyNotFound(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "non-existent-target"
	namespace := "test-namespace"

	// Check non-existent certificate
	_, err := certManager.CheckCertificateReady(ctx, targetName, namespace)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get certificate")
}

func TestCheckSecretReady(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	secretName := "test-secret"
	namespace := "test-namespace"

	// MockSecretProvider always returns data, so secret should be ready
	ready, err := certManager.checkSecretReady(ctx, secretName, namespace)
	assert.NoError(t, err)
	assert.True(t, ready)
}

func TestGetConfigValue(t *testing.T) {
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"testKey": "testValue",
		},
	}

	// Test existing key
	value := getConfigValue(config, "testKey", "defaultValue")
	assert.Equal(t, "testValue", value)

	// Test non-existent key
	value = getConfigValue(config, "nonExistentKey", "defaultValue")
	assert.Equal(t, "defaultValue", value)

	// Test empty value
	config.Properties["emptyKey"] = ""
	value = getConfigValue(config, "emptyKey", "defaultValue")
	assert.Equal(t, "defaultValue", value)
}

func TestCertManagerWithCustomSubject(t *testing.T) {
	certManager := createCertManager(t)
	ctx := context.Background()
	targetName := "custom-target"
	namespace := "custom-namespace"

	// Set custom service name for testing
	originalServiceName := ServiceName
	ServiceName = "test-service"
	defer func() {
		ServiceName = originalServiceName
	}()

	err := certManager.CreateWorkingCert(ctx, targetName, namespace)
	assert.NoError(t, err)

	// Verify certificate was created with custom subject
	getRequest := states.GetRequest{
		ID: targetName,
		Metadata: map[string]interface{}{
			"namespace": namespace,
			"group":     "cert-manager.io",
			"version":   "v1",
			"resource":  "certificates",
			"kind":      "Certificate",
		},
	}

	entry, err := certManager.StateProvider.Get(ctx, getRequest)
	assert.NoError(t, err)

	cert := entry.Body.(map[string]interface{})
	spec := cert["spec"].(map[string]interface{})
	expectedCommonName := "CN=custom-namespace-custom-target.test-service"
	assert.Equal(t, expectedCommonName, spec["commonName"])

	dnsNames := spec["dnsNames"].([]interface{})
	assert.Contains(t, dnsNames, expectedCommonName)
}

func TestCertManagerWithNilProviders(t *testing.T) {
	certManager := &CertManager{}

	config := managers.ManagerConfig{
		Properties: map[string]string{},
	}

	// Test with nil providers
	providers := map[string]providers.IProvider{}

	vendorContext := &contexts.VendorContext{}
	err := certManager.Init(vendorContext, config, providers)
	assert.Error(t, err)
}
