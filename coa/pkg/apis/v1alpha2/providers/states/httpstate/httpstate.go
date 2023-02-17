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

package httpstate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type HttpStateProviderConfig struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
	PostAsArray       bool   `json:"postAsArray,omitempty"`
	PostNameInPath    bool   `json:"postNameInPath,omitempty"`
	PostBodyKeyName   string `json:"postBodyKeyName,omitempty"`
	PostBodyValueName string `json:"postBodyValueName,omitempty"`
	NotFoundAs204     bool   `json:"notFoundAs204,omitempty"`
}

func HttpStateProviderConfigFromMap(properties map[string]string) (HttpStateProviderConfig, error) {
	ret := HttpStateProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["postBodyKeyName"]; ok {
		ret.PostBodyKeyName = utils.ParseProperty(v)
	}
	if v, ok := properties["postBodyValueName"]; ok {
		ret.PostBodyValueName = utils.ParseProperty(v)
	}
	if v, ok := properties["postAsArray"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'postAsArray' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.PostAsArray = bVal
		}
	}
	if v, ok := properties["postNameInPath"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'postNameInPath' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.PostNameInPath = bVal
		}
	} else {
		ret.PostNameInPath = true
	}
	if v, ok := properties["notFoundAs204"]; ok {
		val := utils.ParseProperty(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'notFoundAs204' setting of Http state provider", v1alpha2.BadConfig)
			}
			ret.NotFoundAs204 = bVal
		}
	}
	if v, ok := properties["url"]; ok {
		ret.Url = utils.ParseProperty(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Http sate provider url is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

type HttpStateProvider struct {
	Config  HttpStateProviderConfig
	Data    map[string]interface{}
	Context *contexts.ManagerContext
}

func (s *HttpStateProvider) ID() string {
	return s.Config.Name
}

func (s *HttpStateProvider) SetContext(ctx *contexts.ManagerContext) error {
	s.Context = ctx
	return nil
}

func (i *HttpStateProvider) InitWithMap(properties map[string]string) error {
	config, err := HttpStateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *HttpStateProvider) Init(config providers.IProviderConfig) error {
	// parameter checks
	stateConfig, err := toHttpStateProviderConfig(config)
	if err != nil {
		return errors.New("expected HttpStateProviderConfig")
	}
	s.Config = stateConfig
	if s.Config.Url == "" {
		return v1alpha2.NewCOAError(nil, "Http sate provider url is not set", v1alpha2.BadConfig)
	}
	s.Data = make(map[string]interface{}, 0)
	return nil
}

func (s *HttpStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	client := &http.Client{}
	rUrl := s.Config.Url
	var err error
	if s.Config.PostNameInPath {
		rUrl, err = url.JoinPath(s.Config.Url, entry.Value.ID)
	}
	if err != nil {
		return "", err
	}
	obj := entry.Value.Body
	if s.Config.PostBodyKeyName != "" && s.Config.PostBodyValueName != "" {
		obj = map[string]interface{}{
			s.Config.PostBodyKeyName:   entry.Value.ID,
			s.Config.PostBodyValueName: obj,
		}
	}
	if s.Config.PostAsArray {
		obj = []interface{}{obj}
	}
	jData, _ := json.Marshal(obj)
	req, err := http.NewRequest("POST", rUrl, bytes.NewBuffer(jData))
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to invoke HTTP state store: [%d]", resp.StatusCode)
	}
	return entry.Value.ID, nil
}

func (s *HttpStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	return nil, "", v1alpha2.NewCOAError(nil, "Http sate store list is not implemented", v1alpha2.NotImplemented)
}

func (s *HttpStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	client := &http.Client{}
	rUrl, err := url.JoinPath(s.Config.Url, request.ID)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", rUrl, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to delete from HTTP state store: [%d]", resp.StatusCode), v1alpha2.InternalError)

	}
	return nil
}

func (s *HttpStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	client := &http.Client{}
	rUrl, err := url.JoinPath(s.Config.Url, request.ID)
	if err != nil {
		return states.StateEntry{}, err
	}
	req, err := http.NewRequest("GET", rUrl, nil)
	if err != nil {
		return states.StateEntry{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return states.StateEntry{}, err
	}
	if resp.StatusCode == 204 && s.Config.NotFoundAs204 {
		return states.StateEntry{}, v1alpha2.NewCOAError(nil, "not found", v1alpha2.NotFound)
	}
	if resp.StatusCode >= 300 {
		if resp.StatusCode == 404 {
			return states.StateEntry{}, v1alpha2.NewCOAError(nil, "not found", v1alpha2.NotFound)
		} else {
			return states.StateEntry{}, v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to invoke HTTP state store: [%d]", resp.StatusCode), v1alpha2.InternalError)
		}

	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return states.StateEntry{}, err
	}
	var obj interface{}
	err = json.Unmarshal(bodyBytes, &obj)
	if err != nil {
		return states.StateEntry{}, err
	}
	return states.StateEntry{
		ID:   request.ID,
		Body: obj,
	}, nil
}

func toHttpStateProviderConfig(config providers.IProviderConfig) (HttpStateProviderConfig, error) {
	ret := HttpStateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (a *HttpStateProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &HttpStateProvider{}
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
