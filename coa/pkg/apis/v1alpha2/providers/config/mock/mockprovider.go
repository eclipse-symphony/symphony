/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"encoding/json"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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

func (a *MockConfigProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
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
func (m *MockConfigProvider) Get(ctx context.Context, object string, field string, overrides []string, localContext interface{}) (interface{}, error) {
	return object + "::" + field, nil
}
func (m *MockConfigProvider) GetObject(ctx context.Context, object string, overrides []string, localContext interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{object: object}, nil
}
func (m *MockConfigProvider) Set(object string, field string, value string) error {
	return nil
}
func (m *MockConfigProvider) SetObject(object string, value map[string]interface{}) error {
	return nil
}
func (m *MockConfigProvider) Delete(object string, field string) error {
	return nil
}
func (m *MockConfigProvider) DeleteObject(object string) error {
	return nil
}
func (m *MockConfigProvider) Read(object string, field string, localContext interface{}) (interface{}, error) {
	return object + "::" + field, nil
}
func (m *MockConfigProvider) ReadObject(object string, localContext interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{object: object}, nil
}
