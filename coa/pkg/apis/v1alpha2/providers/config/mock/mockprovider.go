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
	"encoding/json"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type MockConfigProviderConfig struct {
	Name string `json:"name"`
}

func MockConfigProviderConfigFromMap(properties map[string]string) (MockConfigProviderConfig, error) {
	ret := MockConfigProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MockConfigProvider struct {
	Config  MockConfigProviderConfig
	Context *contexts.ManagerContext
}

func (i *MockConfigProvider) InitWithMap(properties map[string]string) error {
	config, err := MockConfigProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *MockConfigProvider) ID() string {
	return m.Config.Name
}

func (a *MockConfigProvider) SetContext(context *contexts.ManagerContext) error {
	a.Context = context
	return nil
}

func (m *MockConfigProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toMockConfigProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid mock config provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toMockConfigProviderConfig(config providers.IProviderConfig) (MockConfigProviderConfig, error) {
	ret := MockConfigProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	return ret, err
}
func (m *MockConfigProvider) Get(object string, field string, overrides []string) (string, error) {
	return object + "::" + field, nil
}
func (m *MockConfigProvider) GetObject(object string, overrides []string) (map[string]string, error) {
	return map[string]string{object: object}, nil
}
func (m *MockConfigProvider) Set(object string, field string, value string) error {
	return nil
}
func (m *MockConfigProvider) SetObject(object string, value map[string]string) error {
	return nil
}
func (m *MockConfigProvider) Delete(object string, field string) error {
	return nil
}
func (m *MockConfigProvider) DeleteObject(object string) error {
	return nil
}
func (m *MockConfigProvider) Read(object string, field string) (string, error) {
	return object + "::" + field, nil
}
func (m *MockConfigProvider) ReadObject(object string) (map[string]string, error) {
	return map[string]string{object: object}, nil
}
