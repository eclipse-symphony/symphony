/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

type MockManager struct {
	managers.IManager
	stateProvider states.IStateProvider
	pollCnt       int
	reconcilCnt   int
	pollLock      sync.Mutex
	reconcilLock  sync.Mutex
}

func (m *MockManager) Init(context *contexts.VendorContext, config managers.ManagerConfig, providers map[string]providers.IProvider) error {
	stateprovider, err := managers.GetStateProvider(config, providers)
	if err == nil {
		m.stateProvider = stateprovider
	}
	if config.Properties["mockInitError"] == "true" {
		return errors.New("mock manager init error")
	}
	return nil
}

func (m *MockManager) Enabled() bool {
	return true
}

func (m *MockManager) Poll() []error {
	m.pollLock.Lock()
	defer m.pollLock.Unlock()
	m.pollCnt++
	return nil
}

func (m *MockManager) Reconcil() []error {
	m.reconcilLock.Lock()
	defer m.reconcilLock.Unlock()
	m.reconcilCnt++
	return nil
}

func (m *MockManager) GetPollCnt() int {
	m.pollLock.Lock()
	defer m.pollLock.Unlock()
	return m.pollCnt
}

func (m *MockManager) GetReconcilCnt() int {
	m.reconcilLock.Lock()
	defer m.reconcilLock.Unlock()
	return m.reconcilCnt
}

type MockManagerFactory struct {
}

func (m *MockManagerFactory) CreateManager(config managers.ManagerConfig) (managers.IManager, error) {
	var manager managers.IManager
	var err error = nil
	switch config.Type {
	case "managers.symphony.mock":
		manager = &MockManager{}
	case "managers.symphony.error":
		err = errors.New("mock factory create manager error")
	}
	return manager, err
}

func TestInit(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name:       "mock",
				Type:       "managers.symphony.mock",
				Properties: map[string]string{},
				Providers:  map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(v.Managers))
	_, ok := v.Managers[0].(*MockManager)
	assert.True(t, ok)
}

func TestInitFailWithFactoryCreateError(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name:       "error",
				Type:       "managers.symphony.error",
				Properties: map[string]string{},
				Providers:  map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{}, nil)
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.InternalError, coaErr.State)
	assert.Equal(t, "failed to create manager 'error'", coaErr.Message)
}

func TestInitFailWithNoFactoryCanCreateError(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name:       "notexist",
				Type:       "managers.symphony.notexist",
				Properties: map[string]string{},
				Providers:  map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{}, nil)
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, coaErr.State)
	assert.Equal(t, "no manager factories can create manager type 'managers.symphony.notexist'", coaErr.Message)
}

func TestInitWithProviders(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	v := &Vendor{}
	err = v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name: "mock",
				Type: "managers.symphony.mock",
				Properties: map[string]string{
					"providers.state": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{
		"mock": {
			"mem-state": stateProvider,
		},
	}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(v.Managers))
	m, ok := v.Managers[0].(*MockManager)
	assert.True(t, ok)
	assert.NotNil(t, m.stateProvider)
}

func TestInitWithManagerInitError(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name:       "mock",
				Type:       "managers.symphony.mock",
				Properties: map[string]string{"mockInitError": "true"},
				Providers:  map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "mock manager init error", err.Error())
	assert.Equal(t, 0, len(v.Managers))
}

func TestSetEvaluationContext(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{}, map[string]map[string]providers.IProvider{}, nil)
	assert.Nil(t, err)
	v.SetEvaluationContext(&utils.EvaluationContext{})
	assert.NotNil(t, v.Context.EvaluationContext)
}

func TestRunLoop(t *testing.T) {
	v := &Vendor{}
	err := v.Init(VendorConfig{
		Managers: []managers.ManagerConfig{
			{
				Name:       "mock",
				Type:       "managers.symphony.mock",
				Properties: map[string]string{},
				Providers:  map[string]managers.ProviderConfig{},
			},
		},
	}, []managers.IManagerFactroy{&MockManagerFactory{}}, map[string]map[string]providers.IProvider{}, nil)
	assert.Nil(t, err)
	m, ok := v.Managers[0].(*MockManager)
	assert.True(t, ok)
	go func() {
		v.RunLoop(1 * time.Second)
	}()
	time.Sleep(2 * time.Second)
	b1 := m.GetPollCnt() > 1
	b2 := m.GetReconcilCnt() > 1
	assert.True(t, b1)
	assert.True(t, b2)
}
