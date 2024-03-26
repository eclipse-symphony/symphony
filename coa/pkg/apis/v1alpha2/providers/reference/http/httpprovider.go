/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

type HTTPReferenceProviderConfig struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	TargetID string `json:"target"`
}

func HTTPReferenceProviderConfigFromMap(properties map[string]string) (HTTPReferenceProviderConfig, error) {
	ret := HTTPReferenceProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["url"]; ok {
		ret.Url = utils.ParseProperty(v)
	}
	if v, ok := properties["target"]; ok {
		ret.TargetID = utils.ParseProperty(v)
	}
	return ret, nil
}

func (i *HTTPReferenceProvider) InitWithMap(properties map[string]string) error {
	config, err := HTTPReferenceProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

type HTTPReferenceProvider struct {
	Config  HTTPReferenceProviderConfig
	Context *contexts.ManagerContext
}

func (m *HTTPReferenceProvider) ID() string {
	return m.Config.Name
}
func (m *HTTPReferenceProvider) TargetID() string {
	return m.Config.TargetID
}

func (m *HTTPReferenceProvider) ReferenceType() string {
	return "v1alpha2.ReferenceHTTP"
}

func (m *HTTPReferenceProvider) Reconfigure(config providers.IProviderConfig) error {
	return nil
}

func (a *HTTPReferenceProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *HTTPReferenceProvider) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toHTTPReferenceProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid HTTP reference provider config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toHTTPReferenceProviderConfig(config providers.IProviderConfig) (HTTPReferenceProviderConfig, error) {
	ret := HTTPReferenceProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.Url = utils.ParseProperty(ret.Url)
	ret.TargetID = utils.ParseProperty(ret.TargetID)
	return ret, err
}

func (m *HTTPReferenceProvider) Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, m.Config.Url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("id", id)
	q.Add("namespace", namespace)
	q.Add("scope", namespace)
	q.Add("group", group)
	q.Add("kind", kind)
	q.Add("version", version)
	if ref != "" {
		q.Add("ref", ref)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ret interface{}
	json.Unmarshal(responseBody, &ret)
	return ret, nil
}

func (m *HTTPReferenceProvider) List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, m.Config.Url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("label-selector", labelSelector)
	q.Add("field-selector", fieldSelector)
	q.Add("namespace", namespace)
	q.Add("scope", namespace)
	q.Add("group", group)
	q.Add("kind", kind)
	q.Add("version", version)
	if ref != "" {
		q.Add("ref", ref)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, 0)
	json.Unmarshal(responseBody, &ret)
	return ret, nil
}

func (a *HTTPReferenceProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &HTTPReferenceProvider{}
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
