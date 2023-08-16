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
package wait

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var mwLock sync.Mutex

type WaitStageProviderConfig struct {
	BaseUrl      string `json:"baseUrl"`
	User         string `json:"user"`
	Password     string `json:"password"`
	WaitInterval int    `json:"wait.interval,omitempty"`
	WaitCount    int    `json:"wait.count,omitempty"`
}

type WaitStageProvider struct {
	Config WaitStageProviderConfig
}

func (s *WaitStageProvider) Init(config providers.IProviderConfig) error {
	mwLock.Lock()
	defer mwLock.Unlock()
	mockConfig, err := toWaitStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func toWaitStageProviderConfig(config providers.IProviderConfig) (WaitStageProviderConfig, error) {
	ret := WaitStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *WaitStageProvider) InitWithMap(properties map[string]string) error {
	config, err := WaitStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func WaitStageProviderConfigFromVendorMap(properties map[string]string) (WaitStageProviderConfig, error) {
	ret := make(map[string]string)
	for k, v := range properties {
		if strings.HasPrefix(k, "wait.") {
			ret[k[5:]] = v
		}
	}
	return WaitStageProviderConfigFromMap(ret)
}
func WaitStageProviderConfigFromMap(properties map[string]string) (WaitStageProviderConfig, error) {
	ret := WaitStageProviderConfig{}
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
	if v, ok := properties["wait.interval"]; ok {
		interval, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait interval %v", v), v1alpha2.BadConfig)
		}
		ret.WaitInterval = interval
	}
	if v, ok := properties["wait.count"]; ok {
		count, err := strconv.Atoi(v)
		if err != nil {
			return ret, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait count %v", v), v1alpha2.BadConfig)
		}
		ret.WaitCount = count
	}
	return ret, nil
}
func (i *WaitStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	objectType := inputs["objectType"].(string)
	objects := inputs["names"].([]interface{})
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		prefixedNames[i] = fmt.Sprintf("%v-%v", inputs["__origin"], object)
	}
	counter := 0
	for counter < i.Config.WaitCount || i.Config.WaitCount == 0 {
		foundCount := 0
		switch objectType {
		case "instance":
			instances, err := utils.GetInstances(i.Config.BaseUrl, i.Config.User, i.Config.Password)
			if err != nil {
				return nil, false, err
			}
			for _, instance := range instances {
				for _, object := range prefixedNames {
					if instance.Spec.Name == object {
						foundCount++
					}
				}
			}
		case "sites":
			sites, err := utils.GetSites(i.Config.BaseUrl, i.Config.User, i.Config.Password)
			if err != nil {
				return nil, false, err
			}
			for _, site := range sites {
				for _, object := range prefixedNames {
					if site.Spec.Name == object {
						foundCount++
					}
				}
			}
		case "catalogs":
			catalogs, err := utils.GetCatalogs(i.Config.BaseUrl, i.Config.User, i.Config.Password)
			if err != nil {
				return nil, false, err
			}
			for _, catalog := range catalogs {
				for _, object := range prefixedNames {
					if catalog.Spec.Name == object {
						foundCount++
					}
				}
			}
		}
		if foundCount == len(objects) {
			outputs["objectType"] = objectType
			outputs["status"] = "OK"
			return outputs, false, nil
		}
		counter++
		if i.Config.WaitInterval > 0 {
			time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
		}
	}

	outputs["objectType"] = objectType
	outputs["status"] = "Failed"
	return outputs, false, v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to find %v %v", objectType, objects), v1alpha2.NotFound)
}
