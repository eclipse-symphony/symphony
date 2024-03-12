/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package customvision

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var log = logger.NewLogger("coa.runtime")

type CustomVisionReferenceProviderConfig struct {
	Name          string `json:"name"`
	APIKey        string `json:"key"`
	Retries       int    `json:"retries,omitempty"`
	RetryInterval int    `json:"retryInterval,omitempty"`
	TargetID      string `json:"target"`
}

func CustomVisionReferenceProviderConfigFromMap(properties map[string]string) (CustomVisionReferenceProviderConfig, error) {
	ret := CustomVisionReferenceProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["key"]; ok {
		ret.APIKey = utils.ParseProperty(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Custom Vision reference provider key is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["retries"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'retries' setting of Custom Vision reference provider", v1alpha2.BadConfig)
			}
			ret.Retries = n
		} else {
			ret.Retries = 3
		}
	}
	if v, ok := properties["retryInterval"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'retryInterval' setting of Custom Vision reference provider", v1alpha2.BadConfig)
			}
			ret.RetryInterval = n
		} else {
			ret.RetryInterval = 5
		}
	}
	return ret, nil
}

type CustomVisionReferenceProvider struct {
	Config  CustomVisionReferenceProviderConfig
	Context *contexts.ManagerContext
}

func (m *CustomVisionReferenceProvider) ID() string {
	return m.Config.Name
}
func (m *CustomVisionReferenceProvider) TargetID() string {
	return m.Config.TargetID
}

func (m *CustomVisionReferenceProvider) Reconfigure(config providers.IProviderConfig) error {
	var err error
	aConfig, err := toCustomVisionReferenceProviderConfig(config)
	if err != nil {
		return v1alpha2.NewCOAError(nil, "provided config is not a valid custom vision reference provider config", v1alpha2.BadConfig)
	}

	m.Config = aConfig

	return nil
}

func (i *CustomVisionReferenceProvider) InitWithMap(properties map[string]string) error {
	config, err := CustomVisionReferenceProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (m *CustomVisionReferenceProvider) ReferenceType() string {
	return "v1alpha2.CustomVision"
}

func (a *CustomVisionReferenceProvider) SetContext(context *contexts.ManagerContext) {
	a.Context = context

}

func (m *CustomVisionReferenceProvider) Init(config providers.IProviderConfig) error {
	return m.Reconfigure(config)
}

func toCustomVisionReferenceProviderConfig(config providers.IProviderConfig) (CustomVisionReferenceProviderConfig, error) {
	ret := CustomVisionReferenceProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	if ret.Retries == 0 {
		ret.Retries = 3
	}
	if ret.RetryInterval == 0 {
		ret.RetryInterval = 5
	}
	ret.Name = utils.ParseProperty(ret.Name)
	ret.APIKey = utils.ParseProperty(ret.APIKey)
	return ret, err
}

func (m *CustomVisionReferenceProvider) Get(id string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	for i := 0; i < m.Config.Retries; i++ {
		url, err := m.localGet(version, namespace, id, group, kind)
		if err == nil && url != "" {
			return url, nil
		}
		time.Sleep(time.Second * time.Duration(m.Config.RetryInterval))
	}
	return "", v1alpha2.NewCOAError(nil, "failed to get Custom Vision export after retries", v1alpha2.InternalError)
}
func (m *CustomVisionReferenceProvider) List(labelSelector string, fieldSelector string, namespace string, group string, kind string, version string, ref string) (interface{}, error) {
	return nil, nil
}
func (m *CustomVisionReferenceProvider) localGet(iteration string, endpoint string, project string, platform string, flavor string) (interface{}, error) {
	log.Debugf("localGet parms - iteration: %s", iteration)
	log.Debugf("localGet parms - endpoint: %s", endpoint)
	log.Debugf("localGet parms - project: %s", project)
	log.Debugf("localGet parms - platform: %s", platform)
	log.Debugf("localGet parms - flavor: %s", flavor)
	log.Debugf("API key = %s", m.Config.APIKey)
	client := &http.Client{}
	url := fmt.Sprintf("https://%s/customvision/v3.3/training/projects/%s/iterations/%s/export", endpoint, project, iteration)
	log.Debugf("request url: %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("failed to get Custom Vision export: %v", err)
		return "", v1alpha2.NewCOAError(err, "failed to get Custom Vision export", v1alpha2.InternalError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Training-Key", m.Config.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to get Custom Vision export: %v", err)
		return "", v1alpha2.NewCOAError(err, "failed to get Custom Vision export", v1alpha2.InternalError)
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("failed to get Custom Vision export: %v", resp)
		return "", v1alpha2.NewCOAError(nil, "failed to get Custom Vision export", v1alpha2.InternalError) //TODO: carry over HTTP status code
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to get Custom Vision export: %v", err)
		return "", v1alpha2.NewCOAError(err, "failed to get Custom Vision export", v1alpha2.InternalError)
	}
	exports := make([]Export, 0)
	err = json.Unmarshal(bodyBytes, &exports)
	if err != nil {
		log.Errorf("failed to get Custom Vision export: %v", err)
		return "", v1alpha2.NewCOAError(err, "failed to get Custom Vision export", v1alpha2.InternalError)
	}
	if len(exports) == 0 {
		url := fmt.Sprintf("https://%s/customvision/v3.3/training/projects/%s/iterations/%s/export?platform=%s&flavor=%s", endpoint, project, iteration, platform, flavor)
		log.Debugf("request url: %s", url)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Errorf("failed to create Custom Vision export: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to create Custom Vision export", v1alpha2.InternalError)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Training-Key", m.Config.APIKey)
		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("failed to create Custom Vision export: %v", err)
			return "", v1alpha2.NewCOAError(err, "failed to create Custom Vision export", v1alpha2.InternalError)
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("failed to create Custom Vision export: %v", resp)
			return "", v1alpha2.NewCOAError(nil, "failed to create Custom Vision export", v1alpha2.InternalError) //TODO: carry over HTTP status code
		}
		return "", nil
	} else {
		return exports, nil
		// for _, e := range exports {
		// 	if e.Status == "Done" {
		// 		return e.DownloadUri, nil
		// 	}
		// }
	}
	//return "", nil
}

type Export struct {
	Platform            string `json:"platform"`
	Status              string `json:"status"`
	DownloadUri         string `json:"downloadUri"`
	Flavor              string `json:"flavor"`
	NewVersionAvailable bool   `json:"newerVersionAvailable"`
}
