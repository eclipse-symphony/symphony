/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

type MemoryConfigProviderConfig struct {
	Name string `json:"name"`
}

func MockConfigProviderConfigFromMap(properties map[string]string) (MemoryConfigProviderConfig, error) {
	ret := MemoryConfigProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MemoryConfigProvider struct {
	Config     MemoryConfigProviderConfig
	Context    *contexts.ManagerContext
	ConfigData map[string]map[string]interface{}
	Lock       *sync.Mutex
}

func (i *MemoryConfigProvider) InitWithMap(properties map[string]string) error {
	config, err := MockConfigProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *MemoryConfigProvider) ID() string {
	return m.Config.Name
}

func (a *MemoryConfigProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *MemoryConfigProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toMemoryConfigProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid mock config provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	m.ConfigData = make(map[string]map[string]interface{})
	return nil
}

func toMemoryConfigProviderConfig(config providers.IProviderConfig) (MemoryConfigProviderConfig, error) {
	ret := MemoryConfigProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	return ret, err
}
func (m *MemoryConfigProvider) Read(ctx context.Context, object string, field string, localContext interface{}) (interface{}, error) {
	if _, ok := m.ConfigData[object]; !ok {
		return "", v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	if _, ok := m.ConfigData[object][field]; !ok {
		return "", v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	return m.ConfigData[object][field], nil
}
func (m *MemoryConfigProvider) ReadObject(ctx context.Context, object string, localContext interface{}) (map[string]interface{}, error) {
	if _, ok := m.ConfigData[object]; !ok {
		return nil, v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	return m.ConfigData[object], nil
}
func (m *MemoryConfigProvider) Set(ctx context.Context, object string, field string, value interface{}) error {
	if _, ok := m.ConfigData[object]; !ok {
		m.ConfigData[object] = make(map[string]interface{})
	}
	m.ConfigData[object][field] = value
	return nil
}
func (m *MemoryConfigProvider) SetObject(ctx context.Context, object string, value map[string]interface{}) error {
	if _, ok := m.ConfigData[object]; !ok {
		m.ConfigData[object] = make(map[string]interface{})
	}
	for k, v := range value {
		m.ConfigData[object][k] = v
	}
	return nil
}
func (m *MemoryConfigProvider) Remove(ctx context.Context, object string, field string) error {
	if _, ok := m.ConfigData[object]; !ok {
		return v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	if _, ok := m.ConfigData[object][field]; !ok {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
	}
	delete(m.ConfigData[object], field)
	return nil
}
func (m *MemoryConfigProvider) RemoveObject(ctx context.Context, object string) error {
	if _, ok := m.ConfigData[object]; !ok {
		return v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	delete(m.ConfigData, object)
	return nil
}
