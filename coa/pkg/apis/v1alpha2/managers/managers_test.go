/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package managers

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	memoryconfig "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/memoryconfig"
	mockconfig "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	mockledger "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/probe/rtsp"
	memoryqueue "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/queue/memory"
	mockreference "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	mocksecret "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/uploader/azure/blob"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := Manager{}
	err = manager.Init(nil, ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)
}

func TestGetQueueProvider(t *testing.T) {
	queueProvider := &memoryqueue.MemoryQueueProvider{}
	err := queueProvider.Init(memoryqueue.MemoryQueueProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.queue": "memory-queue",
		},
	}
	pMap := map[string]providers.IProvider{
		"memory-queue": queueProvider,
	}

	p, err := GetQueueProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// queue provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetQueueProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "queue provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// queue provider is not found
	pMap2 := map[string]providers.IProvider{
		"memory-queue-xxx": queueProvider,
	}
	p, err = GetQueueProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "queue provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a queue provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"memory-queue": stateProvider,
	}
	p, err = GetQueueProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a queue provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetStateProvider(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
		},
	}
	pMap := map[string]providers.IProvider{
		"memory-state": stateProvider,
	}

	p, err := GetStateProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// state provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetStateProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "state provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// state provider is not found
	pMap2 := map[string]providers.IProvider{
		"memory-state-xxx": stateProvider,
	}
	p, err = GetStateProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "state provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a state provider
	queueProvider := &memoryqueue.MemoryQueueProvider{}
	err = queueProvider.Init(memoryqueue.MemoryQueueProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"memory-state": queueProvider,
	}
	p, err = GetStateProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a state provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetLedgerProvider(t *testing.T) {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.ledger": "mock-ledger",
		},
	}
	pMap := map[string]providers.IProvider{
		"mock-ledger": ledgerProvider,
	}

	p, err := GetLedgerProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// ledger provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetLedgerProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "ledger provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// ledger provider is not found
	pMap2 := map[string]providers.IProvider{
		"mock-ledger-xxx": ledgerProvider,
	}
	p, err = GetLedgerProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "ledger provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a ledger provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"mock-ledger": stateProvider,
	}
	p, err = GetLedgerProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a ledger provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetConfigProvider(t *testing.T) {
	configProvider := &memoryconfig.MemoryConfigProvider{}
	err := configProvider.Init(memoryconfig.MemoryConfigProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.config": "mock-config",
		},
	}
	pMap := map[string]providers.IProvider{
		"mock-config": configProvider,
	}

	p, err := GetConfigProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// config provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetConfigProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "config provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// config provider is not found
	pMap2 := map[string]providers.IProvider{
		"mock-config-xxx": configProvider,
	}
	p, err = GetConfigProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "config provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a config provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"mock-config": stateProvider,
	}
	p, err = GetConfigProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a config provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetExtConfigProvider(t *testing.T) {
	configProvider := &mockconfig.MockConfigProvider{}
	err := configProvider.Init(mockconfig.MockConfigProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.config": "mock-config",
		},
	}
	pMap := map[string]providers.IProvider{
		"mock-config": configProvider,
	}

	p, err := GetExtConfigProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// config provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetExtConfigProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "config provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// config provider is not found
	pMap2 := map[string]providers.IProvider{
		"mock-config-xxx": configProvider,
	}
	p, err = GetExtConfigProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "config provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a config provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"mock-config": stateProvider,
	}
	p, err = GetExtConfigProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a config provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetSecretProvider(t *testing.T) {
	secretProvider := &mocksecret.MockSecretProvider{}
	err := secretProvider.Init(mocksecret.MockSecretProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.secret": "mock-secret",
		},
	}
	pMap := map[string]providers.IProvider{
		"mock-secret": secretProvider,
	}

	p, err := GetSecretProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// secret provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetSecretProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "secret provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// secret provider is not found
	pMap2 := map[string]providers.IProvider{
		"mock-secret-xxx": secretProvider,
	}
	p, err = GetSecretProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "secret provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a secret provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"mock-secret": stateProvider,
	}
	p, err = GetSecretProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a secret provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetProbeProvider(t *testing.T) {
	probeProvider := &rtsp.RTSPProbeProvider{}
	err := probeProvider.Init(rtsp.RTSPProbeProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.probe": "rtsp-probe",
		},
	}
	pMap := map[string]providers.IProvider{
		"rtsp-probe": probeProvider,
	}

	p, err := GetProbeProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// probe provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetProbeProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "probe provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// probe provider is not found
	pMap2 := map[string]providers.IProvider{
		"rtsp-probe-xxx": probeProvider,
	}
	p, err = GetProbeProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "probe provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a probe provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"rtsp-probe": stateProvider,
	}
	p, err = GetProbeProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a probe provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetUploaderProvider(t *testing.T) {
	uploadProvider := &blob.AzureBlobUploader{}
	err := uploadProvider.Init(blob.AzureBlobUploaderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.uploader": "blob-uploader",
		},
	}
	pMap := map[string]providers.IProvider{
		"blob-uploader": uploadProvider,
	}

	p, err := GetUploaderProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// uploader provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetUploaderProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "uploader provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// uploader provider is not found
	pMap2 := map[string]providers.IProvider{
		"blob-uploader-xxx": uploadProvider,
	}
	p, err = GetUploaderProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "uploader provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a uploader provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"blob-uploader": stateProvider,
	}
	p, err = GetUploaderProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a uploader provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetReferenceProvider(t *testing.T) {
	referenceProvider := &mockreference.MockReferenceProvider{}
	err := referenceProvider.Init(mockreference.MockReferenceProviderConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.reference": "mock-reference",
		},
	}
	pMap := map[string]providers.IProvider{
		"mock-reference": referenceProvider,
	}

	p, err := GetReferenceProvider(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// reference provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetReferenceProvider(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "reference provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// reference provider is not found
	pMap2 := map[string]providers.IProvider{
		"mock-reference-xxx": referenceProvider,
	}
	p, err = GetReferenceProvider(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "reference provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a reference provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"mock-reference": stateProvider,
	}
	p, err = GetReferenceProvider(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a reference provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}

func TestGetReporter(t *testing.T) {
	reporter := &http.HTTPReporter{}
	err := reporter.Init(http.HTTPReporterConfig{})
	assert.Nil(t, err)

	config := ManagerConfig{
		Properties: map[string]string{
			"providers.reporter": "http-reporter",
		},
	}
	pMap := map[string]providers.IProvider{
		"http-reporter": reporter,
	}

	p, err := GetReporter(config, pMap)
	assert.Nil(t, err)
	assert.NotNil(t, p)

	// reporter provider is not configured
	config1 := ManagerConfig{
		Properties: map[string]string{},
	}
	p, err = GetReporter(config1, pMap)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, "reporter provider is not configured", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// reporter provider is not found
	pMap2 := map[string]providers.IProvider{
		"http-reporter-xxx": reporter,
	}
	p, err = GetReporter(config, pMap2)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "reporter provider is not supplied", coaError.Message)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)

	// not a reporter provider
	stateProvider := &memorystate.MemoryStateProvider{}
	err = stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	pMap3 := map[string]providers.IProvider{
		"http-reporter": stateProvider,
	}
	p, err = GetReporter(config, pMap3)
	assert.NotNil(t, err)
	assert.Nil(t, p)
	coaError = err.(v1alpha2.COAError)
	assert.Equal(t, "supplied provider is not a reporter provider", coaError.Message)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
}
