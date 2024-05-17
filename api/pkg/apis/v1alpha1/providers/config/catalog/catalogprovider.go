/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var log = logger.NewLogger("coa.runtime")

type CatalogConfigProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type CatalogConfigProvider struct {
	Config    CatalogConfigProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func (s *CatalogConfigProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toCatalogConfigProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}
func (s *CatalogConfigProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
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
	if utils.ShouldUseUserCreds() {
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
	}
	return ret, nil
}
func (m *CatalogConfigProvider) unwindOverrides(override string, field string, namespace string) (string, error) {
	catalog, err := m.ApiClient.GetCatalog(context.TODO(), override, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		return "", err
	}
	if v, ok := catalog.Spec.Properties[field]; ok {
		if vstring, ok := v.(string); ok {
			return vstring, nil
		} else {
			return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' doesn't has a valid value in configuration'%s'", field, override), v1alpha2.BadConfig)
		}
	}
	if catalog.Spec.ParentName != "" {
		return m.unwindOverrides(catalog.Spec.ParentName, field, namespace)
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, override), v1alpha2.NotFound)
}
func (m *CatalogConfigProvider) Read(object string, field string, localcontext interface{}) (interface{}, error) {
	namespace := m.getNamespaceFromContext(localcontext)
	log.Infof("P (Catalog Config): Read %v, %d", object, namespace)

	catalog, err := m.ApiClient.GetCatalog(context.TODO(), object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		return "", err
	}

	if v, ok := catalog.Spec.Properties[field]; ok {
		return m.traceValue(v, localcontext)
	}

	if catalog.Spec.ParentName != "" {
		overrid, err := m.unwindOverrides(catalog.Spec.ParentName, field, namespace)
		if err != nil {
			return "", err
		} else {
			return overrid, nil
		}
	}

	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
}

func (m *CatalogConfigProvider) ReadObject(object string, localcontext interface{}) (map[string]interface{}, error) {
	namespace := m.getNamespaceFromContext(localcontext)
	log.Infof("P (Catalog Config): ReadObject %v, %d", object, namespace)

	catalog, err := m.ApiClient.GetCatalog(context.TODO(), object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{}
	for k, v := range catalog.Spec.Properties {
		tv, err := m.traceValue(v, localcontext)
		if err != nil {
			return nil, err
		}
		// line 189-196 extracts the returned map and merge the keys with the parent
		// this allows a referenced configuration to be overriden by local values
		if tmap, ok := tv.(map[string]interface{}); ok {
			for tk, tv := range tmap {
				if _, ok := ret[tk]; !ok {
					ret[tk] = tv
				}
			}
			continue
		}
		ret[k] = tv
	}
	return ret, nil
}

func (m *CatalogConfigProvider) getNamespaceFromContext(localContext interface{}) string {
	if localContext != nil {
		if ltx, ok := localContext.(coa_utils.EvaluationContext); ok {
			return ltx.Namespace
		}
	}
	return ""
}

func (m *CatalogConfigProvider) traceValue(v interface{}, localcontext interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := m.Context.VencorContext.EvaluationContext.Clone()
		context.DeploymentSpec = m.Context.VencorContext.EvaluationContext.DeploymentSpec
		if localcontext != nil {
			if ltx, ok := localcontext.(coa_utils.EvaluationContext); ok {
				context.Inputs = ltx.Inputs
				context.Outputs = ltx.Outputs
				context.Value = ltx.Value
				context.Properties = ltx.Properties
				context.Component = ltx.Component
				context.Namespace = ltx.Namespace
				if ltx.DeploymentSpec != nil {
					context.DeploymentSpec = ltx.DeploymentSpec
				}
			}
		}
		v, err := parser.Eval(*context)
		if err != nil {
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return m.traceValue(v, localcontext)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := m.traceValue(v, localcontext)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := m.traceValue(v, localcontext)
			if err != nil {
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
}

// TODO: IConfigProvider interface methods should be enhanced to accept namespace as a parameter
// so we can get rid of getCatalogInDefaultNamespace.
func (m *CatalogConfigProvider) Set(object string, field string, value interface{}) error {
	catalog, err := m.getCatalogInDefaultNamespace(context.TODO(), object)
	if err != nil {
		return err
	}
	catalog.Spec.Properties[field] = value
	data, _ := json.Marshal(catalog)
	return m.ApiClient.UpsertCatalog(context.TODO(), object, data, m.Config.User, m.Config.Password)
}
func (m *CatalogConfigProvider) SetObject(object string, value map[string]interface{}) error {
	catalog, err := m.getCatalogInDefaultNamespace(context.TODO(), object)
	if err != nil {
		return err
	}
	catalog.Spec.Properties = map[string]interface{}{}
	for k, v := range value {
		catalog.Spec.Properties[k] = v
	}
	data, _ := json.Marshal(catalog)
	return m.ApiClient.UpsertCatalog(context.TODO(), object, data, m.Config.User, m.Config.Password)
}
func (m *CatalogConfigProvider) Remove(object string, field string) error {
	catlog, err := m.getCatalogInDefaultNamespace(context.TODO(), object)
	if err != nil {
		return err
	}
	if _, ok := catlog.Spec.Properties[field]; !ok {
		return v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	delete(catlog.Spec.Properties, field)
	data, _ := json.Marshal(catlog)
	return m.ApiClient.UpsertCatalog(context.TODO(), object, data, m.Config.User, m.Config.Password)
}
func (m *CatalogConfigProvider) RemoveObject(object string) error {
	return m.ApiClient.DeleteCatalog(context.TODO(), object, "", m.Config.User, m.Config.Password)
}

func (m *CatalogConfigProvider) getCatalogInDefaultNamespace(context context.Context, catalog string) (model.CatalogState, error) {
	return m.ApiClient.GetCatalog(context, catalog, "", m.Config.User, m.Config.Password)
}
