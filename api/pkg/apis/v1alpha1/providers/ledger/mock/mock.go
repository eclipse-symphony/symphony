/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type MockLedgerProviderConfig struct {
	Name string `json:"name"`
}
type MockLedgerProvider struct {
	Config     MockLedgerProviderConfig
	Context    *contexts.ManagerContext
	LedgerData []model.Trail
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

func (a *MockLedgerProvider) SetContext(context *contexts.ManagerContext) error {
	a.Context = context
	return nil
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
func (i *MockLedgerProvider) Append(ctx context.Context, trails []model.Trail) error {
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
