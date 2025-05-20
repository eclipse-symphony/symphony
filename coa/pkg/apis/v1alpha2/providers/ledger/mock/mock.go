/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

type MockLedgerProviderConfig struct {
	Name string `json:"name"`
}
type MockLedgerProvider struct {
	Config     MockLedgerProviderConfig
	Context    *contexts.ManagerContext
	LedgerData []v1alpha2.Trail
	Lock       *sync.Mutex
}

func (m *MockLedgerProvider) Init(config providers.IProviderConfig) error {
	m.Lock = &sync.Mutex{}
	mockConfig, err := toMockLedgerProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	return nil
}
func toMockLedgerProviderConfig(config providers.IProviderConfig) (MockLedgerProviderConfig, error) {
	ret := MockLedgerProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (m *MockLedgerProvider) ID() string {
	return m.Config.Name
}

func (a *MockLedgerProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (i *MockLedgerProvider) InitWithMap(properties map[string]string) error {
	config, err := MockLedgerProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockLedgerProviderConfigFromMap(properties map[string]string) (MockLedgerProviderConfig, error) {
	ret := MockLedgerProviderConfig{}
	ret.Name = properties["name"]
	return ret, nil
}
func (i *MockLedgerProvider) Append(ctx context.Context, trails []v1alpha2.Trail) error {
	i.Lock.Lock()
	defer i.Lock.Unlock()

	jData, _ := json.Marshal(trails)
	fmt.Printf("MOCK LEDGER IS RECORDING: %s\n", jData)

	i.LedgerData = append(i.LedgerData, trails...)
	//trim the ledger data
	if len(i.LedgerData) > 100 {
		i.LedgerData = i.LedgerData[len(i.LedgerData)-100:]
	}
	return nil
}
