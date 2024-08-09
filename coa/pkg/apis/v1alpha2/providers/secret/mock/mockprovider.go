/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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

func (a *MockSecretProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
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
func (m *MockSecretProvider) Read(object string, field string, localContext interface{}) (string, error) {
	return object + ">>" + field, nil
}
func (m *MockSecretProvider) Get(object string, field string, localContext interface{}) (string, error) {
	return object + ">>" + field, nil
}
