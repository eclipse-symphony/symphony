/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

type HTTPReporterConfig struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

func HTTPReporterConfigFromMap(properties map[string]string) (HTTPReporterConfig, error) {
	ret := HTTPReporterConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["url"]; ok {
		ret.Url = utils.ParseProperty(v)
	}
	return ret, nil
}

func (i *HTTPReporter) InitWithMap(properties map[string]string) error {
	config, err := HTTPReporterConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

type HTTPReporter struct {
	Config  HTTPReporterConfig
	Context *contexts.ManagerContext
}

func (m *HTTPReporter) ID() string {
	return m.Config.Name
}

func (a *HTTPReporter) SetContext(context *contexts.ManagerContext) {
	a.Context = context
}

func (m *HTTPReporter) Init(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toHTTPReporterConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid HTTP reporter config", v1alpha2.BadConfig)
	}
	m.Config = aConfig
	return nil
}

func toHTTPReporterConfig(config providers.IProviderConfig) (HTTPReporterConfig, error) {
	ret := HTTPReporterConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.Url = utils.ParseProperty(ret.Url)
	return ret, err
}

func (m *HTTPReporter) Report(id string, namespace string, group string, kind string, version string, properties map[string]string, overwrite bool) error {
	client := &http.Client{}
	data, _ := json.Marshal(properties)
	req, err := http.NewRequest(http.MethodPost, m.Config.Url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("id", id)
	q.Add("namespace", namespace)
	q.Add("group", group)
	q.Add("kind", kind)
	q.Add("version", version)
	q.Add("overwrite", strconv.FormatBool(overwrite))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return v1alpha2.NewCOAError(nil, string(body), v1alpha2.State(resp.StatusCode))
	}
	return nil
}

func (a *HTTPReporter) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &HTTPReporter{}
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
