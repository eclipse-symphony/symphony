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

type MockSecretProviderConfig struct {
	Name string `json:"name"`
}

func MockSecretProviderConfigFromMap(properties map[string]string) (MockSecretProviderConfig, error) {
	ret := MockSecretProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MockSecretProvider struct {
	Config  MockSecretProviderConfig
	Context *contexts.ManagerContext
}

func (i *MockSecretProvider) InitWithMap(properties map[string]string) error {
	config, err := MockSecretProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *MockSecretProvider) ID() string {
	return m.Config.Name
}

func (a *MockSecretProvider) SetContext(context *contexts.ManagerContext) error {
	a.Context = context
	return nil
}

func (m *MockSecretProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toMockSecretProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid mock config provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toMockSecretProviderConfig(config providers.IProviderConfig) (MockSecretProviderConfig, error) {
	ret := MockSecretProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	return ret, err
}
func (m *MockSecretProvider) Get(object string, field string) (string, error) {
	return object + ">>" + field, nil
}
