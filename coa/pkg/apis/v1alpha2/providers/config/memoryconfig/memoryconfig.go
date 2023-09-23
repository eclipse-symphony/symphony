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

package memory

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
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
func (m *MemoryConfigProvider) Read(object string, field string, localContext interface{}) (interface{}, error) {
	if _, ok := m.ConfigData[object]; !ok {
		return "", v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	if _, ok := m.ConfigData[object][field]; !ok {
		return "", v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	return m.ConfigData[object][field], nil
}
func (m *MemoryConfigProvider) ReadObject(object string, localContext interface{}) (map[string]interface{}, error) {
	if _, ok := m.ConfigData[object]; !ok {
		return nil, v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	return m.ConfigData[object], nil
}
func (m *MemoryConfigProvider) Set(object string, field string, value interface{}) error {
	if _, ok := m.ConfigData[object]; !ok {
		m.ConfigData[object] = make(map[string]interface{})
	}
	m.ConfigData[object][field] = value
	return nil
}
func (m *MemoryConfigProvider) SetObject(object string, value map[string]interface{}) error {
	if _, ok := m.ConfigData[object]; !ok {
		m.ConfigData[object] = make(map[string]interface{})
	}
	for k, v := range value {
		m.ConfigData[object][k] = v
	}
	return nil
}
func (m *MemoryConfigProvider) Remove(object string, field string) error {
	if _, ok := m.ConfigData[object]; !ok {
		return v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	if _, ok := m.ConfigData[object][field]; !ok {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
	}
	delete(m.ConfigData[object], field)
	return nil
}
func (m *MemoryConfigProvider) RemoveObject(object string) error {
	if _, ok := m.ConfigData[object]; !ok {
		return v1alpha2.NewCOAError(nil, "object not found", v1alpha2.NotFound)
	}
	delete(m.ConfigData, object)
	return nil
}
