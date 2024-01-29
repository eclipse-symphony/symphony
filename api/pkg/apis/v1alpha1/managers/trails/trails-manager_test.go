/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package trails

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	mockledger "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger/mock"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)
	providers := make(map[string]providers.IProvider)
	providers["MockLedgerProvider"] = ledgerProvider
	manager := TrailsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{},
	}
	err = manager.Init(nil, config, providers)
	assert.Nil(t, err)
}

func TestAppend(t *testing.T) {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)
	providers := make(map[string]providers.IProvider)
	providers["MockLedgerProvider"] = ledgerProvider
	manager := TrailsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{},
	}
	err = manager.Init(nil, config, providers)
	assert.Nil(t, err)
	err = manager.Append(context.Background(), nil)
	assert.Nil(t, err)
}

func TestAppendFail(t *testing.T) {
	ledgerProvider := &MockLedgerProviderFail{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)
	providers := make(map[string]providers.IProvider)
	providers["MockLedgerProviderFail"] = ledgerProvider
	manager := TrailsManager{}
	config := managers.ManagerConfig{
		Properties: map[string]string{},
	}
	err = manager.Init(nil, config, providers)
	assert.Nil(t, err)
	err = manager.Append(context.Background(), nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.InternalError, coaError.State)
	assert.Equal(t, assert.AnError.Error()+";", coaError.Message)
}

type MockLedgerProviderFail struct {
}

func (m *MockLedgerProviderFail) Init(config providers.IProviderConfig) error {
	return nil
}

func (m *MockLedgerProviderFail) Append(ctx context.Context, trails []v1alpha2.Trail) error {
	// always return error
	return assert.AnError
}
