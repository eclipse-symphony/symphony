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
package catalog

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type CatalogConfigProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type CatalogConfigProvider struct {
	Config CatalogConfigProviderConfig
}

func (s *CatalogConfigProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toCatalogConfigProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func toCatalogConfigProviderConfig(config providers.IProviderConfig) (CatalogConfigProviderConfig, error) {
	ret := CatalogConfigProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *CatalogConfigProvider) InitWithMap(properties map[string]string) error {
	config, err := CatalogConfigProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func CatalogConfigProviderConfigFromMap(properties map[string]string) (CatalogConfigProviderConfig, error) {
	ret := CatalogConfigProviderConfig{}
	baseUrl, err := utils.GetString(properties, "baseUrl")
	if err != nil {
		return ret, err
	}
	ret.BaseUrl = baseUrl
	if ret.BaseUrl == "" {
		return ret, v1alpha2.NewCOAError(nil, "baseUrl is required", v1alpha2.BadConfig)
	}
	user, err := utils.GetString(properties, "user")
	if err != nil {
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
	}
	password, err := utils.GetString(properties, "password")
	if err != nil {
		return ret, err
	}
	ret.Password = password
	return ret, nil
}
func (m *CatalogConfigProvider) Read(object string, field string) (string, error) {
	catalog, err := utils.GetCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
	if err != nil {
		return "", err
	}
	if v, ok := catalog.Spec.Properties[field]; ok {
		return v.(string), nil
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
}
func (m *CatalogConfigProvider) ReadObject(object string) (map[string]string, error) {
	catalog, err := utils.GetCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
	if err != nil {
		return nil, err
	}
	ret := map[string]string{}
	for k, v := range catalog.Spec.Properties {
		ret[k] = v.(string)
	}
	return ret, nil
}
func (m *CatalogConfigProvider) Set(object string, field string, value string) error {
	catalog, err := utils.GetCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
	if err != nil {
		return err
	}
	catalog.Spec.Properties[field] = value
	data, _ := json.Marshal(catalog.Spec)
	return utils.UpsertCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password, data)
}
func (m *CatalogConfigProvider) SetObject(object string, value map[string]string) error {
	catalog, err := utils.GetCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
	if err != nil {
		return err
	}
	catalog.Spec.Properties = map[string]interface{}{}
	for k, v := range value {
		catalog.Spec.Properties[k] = v
	}
	data, _ := json.Marshal(catalog.Spec)
	return utils.UpsertCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password, data)
}
func (m *CatalogConfigProvider) Remove(object string, field string) error {
	catlog, err := utils.GetCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
	if err != nil {
		return err
	}
	if _, ok := catlog.Spec.Properties[field]; !ok {
		return v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	delete(catlog.Spec.Properties, field)
	data, _ := json.Marshal(catlog.Spec)
	return utils.UpsertCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password, data)
}
func (m *CatalogConfigProvider) RemoveObject(object string) error {
	return utils.DeleteCatalog(m.Config.BaseUrl, object, m.Config.User, m.Config.Password)
}
