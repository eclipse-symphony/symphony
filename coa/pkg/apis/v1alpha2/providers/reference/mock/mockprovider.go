/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type MockReferenceProviderConfig struct {
	Name   string                 `json:"name"`
	Values map[string]interface{} `json:"values"`
}

func MockReferenceProviderConfigFromMap(properties map[string]string) (MockReferenceProviderConfig, error) {
	ret := MockReferenceProviderConfig{}
	for k, v := range properties {
		if k == "name" {
			ret.Name = utils.ParseProperty(v)
		} else {
			if ret.Values == nil {
				ret.Values = make(map[string]interface{})
			}
			ret.Values[k] = utils.ParseProperty(v)
		}
	}
	return ret, nil
}

type MockReferenceProvider struct {
	Config  MockReferenceProviderConfig
	Context *contexts.ManagerContext
}

func (m *MockReferenceProvider) ID() string {
	return m.Config.Name
}

func (m *MockReferenceProvider) TargetID() string {
	return "mock-target"
}

func (m *MockReferenceProvider) ReferenceType() string {
	return "mock"
}

func (a *MockReferenceProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *MockReferenceProvider) Reconfigure(config providers.IProviderConfig) error {
	return nil
}

func (m *MockReferenceProvider) Init(config providers.IProviderConfig) error {
	aConfig, err := toMockReferenceProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid mock config provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toMockReferenceProviderConfig(config providers.IProviderConfig) (MockReferenceProviderConfig, error) {
	ret := MockReferenceProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	// ret.Name = providers.LoadEnv(ret.Name)
	// for k, v := range ret.Values {
	// 	str, ok := v.(string)
	// 	if ok {
	// 		ret.Values[k] = providers.LoadEnv(str)
	// 	}
	// }
	return ret, err
}
func (m *MockReferenceProvider) List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	return nil, nil
}
func (m *MockReferenceProvider) Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	if id == "timestamp" {
		return time.Now(), nil
	}
	if val, ok := m.Config.Values[id]; ok {
		return val, nil
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("configuraion key '%s' is not found", id), v1alpha2.NotFound)
}

func (a *MockReferenceProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &MockReferenceProvider{}
	if config == nil {
		err := ret.Init(a.Config)
		if err != nil {
			return nil, err
		}
	} else {
		err := ret.Init(config)
		if err != nil {
			return nil, err
		}
	}
	if a.Context != nil {
		ret.Context = a.Context
	}
	return ret, nil
}
