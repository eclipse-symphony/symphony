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
package list

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type ListStageProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ListStageProvider struct {
	Config ListStageProviderConfig
}

func (s *ListStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toListStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func toListStageProviderConfig(config providers.IProviderConfig) (ListStageProviderConfig, error) {
	ret := ListStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *ListStageProvider) InitWithMap(properties map[string]string) error {
	config, err := ListStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func ListStageProviderConfigFromMap(properties map[string]string) (ListStageProviderConfig, error) {
	ret := ListStageProviderConfig{}
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
func (i *ListStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	objectType := inputs["objectType"].(string)
	namesOnly := false
	if v, ok := inputs["namesOnly"]; ok {
		if v.(bool) {
			namesOnly = v.(bool)
		}
	}
	switch objectType {
	case "instance":
		instances, err := utils.GetInstances(i.Config.BaseUrl, i.Config.User, i.Config.Password)
		if err != nil {
			return nil, err
		}
		if namesOnly {
			names := make([]string, 0)
			for _, instance := range instances {
				names = append(names, instance.Spec.Name)
			}
			outputs["items"] = names
		} else {
			outputs["items"] = instances
		}
	case "sites":
		sites, err := utils.GetSites(i.Config.BaseUrl, i.Config.User, i.Config.Password)
		if err != nil {
			return nil, err
		}
		if namesOnly {
			names := make([]string, 0)
			for _, site := range sites {
				names = append(names, site.Spec.Name)
			}
			outputs["items"] = names
		} else {
			outputs["items"] = sites
		}
	}
	outputs["objectType"] = objectType
	outputs["status"] = "OK"
	return outputs, nil
}
