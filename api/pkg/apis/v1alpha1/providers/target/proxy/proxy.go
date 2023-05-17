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

package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type ProxyUpdateProviderConfig struct {
	Name      string `json:"name"`
	ServerURL string `json:"serverUrl"`
}

type ProxyUpdateProvider struct {
	Config  ProxyUpdateProviderConfig
	Context *contexts.ManagerContext
}

func ProxyUpdateProviderConfigFromMap(properties map[string]string) (ProxyUpdateProviderConfig, error) {
	ret := ProxyUpdateProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	if v, ok := properties["serverUrl"]; ok {
		ret.ServerURL = utils.ParseProperty(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "proxy update provider server url is not set", v1alpha2.BadConfig)
	}
	return ret, nil
}

func (i *ProxyUpdateProvider) InitWithMap(properties map[string]string) error {
	config, err := ProxyUpdateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (i *ProxyUpdateProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("~~~ Proxy Provider ~~~ : Init()")

	updateConfig, err := toProxyUpdateProviderConfig(config)
	if err != nil {
		return errors.New("expected ProxyUpdateProviderConfig")
	}
	i.Config = updateConfig

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toProxyUpdateProviderConfig(config providers.IProviderConfig) (ProxyUpdateProviderConfig, error) {
	ret := ProxyUpdateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	ret.Name = utils.ParseProperty(ret.Name)
	ret.ServerURL = utils.ParseProperty(ret.ServerURL)
	return ret, err
}

func (a *ProxyUpdateProvider) callRestAPI(route string, method string, payload []byte) ([]byte, error) {
	client := &http.Client{}
	url := a.Config.ServerURL + route
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", err), v1alpha2.InternalError)
	}
	if resp.StatusCode >= 300 {
		return nil, v1alpha2.NewCOAError(err, fmt.Sprintf("failed to invoke Percept API: %v", string(bodyBytes)), v1alpha2.InternalError)
	}
	return bodyBytes, nil
}

func (i *ProxyUpdateProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "Get",
	})
	sLog.Infof("~~~ Proxy Provider ~~~ : getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	data, _ := json.Marshal(deployment)
	payload, err := i.callRestAPI("instances", "GET", data)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return nil, err
	}
	ret := make([]model.ComponentSpec, 0)
	err = json.Unmarshal(payload, &ret)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return nil, err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (i *ProxyUpdateProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("~~~ Proxy Provider ~~~ : removing artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	data, _ := json.Marshal(deployment)
	_, err := i.callRestAPI("instances", "DELETE", data)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (i *ProxyUpdateProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, isDryRun bool) error {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("~~~ Proxy Provider ~~~ : applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	err := i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil
	}

	data, _ := json.Marshal(deployment)

	_, err = i.callRestAPI("instances", "POST", data)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (i *ProxyUpdateProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "NeedsUpdate",
	})
	sLog.Info("~~~ Proxy Provider ~~~ : needs update")

	data, _ := json.Marshal(TwoComponentSlices{Current: current, Desired: desired})
	_, err := i.callRestAPI("needsupdate", "GET", data)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return false
	}
	observ_utils.CloseSpanWithError(span, nil)
	return true
}
func (i *ProxyUpdateProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Proxy Provider", context.Background(), &map[string]string{
		"method": "NeedsRemove",
	})
	sLog.Info("~~~ Proxy Provider ~~~ : needs remove")

	data, _ := json.Marshal(TwoComponentSlices{Current: current, Desired: desired})

	_, err := i.callRestAPI("needsremove", "GET", data)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return false
	}
	observ_utils.CloseSpanWithError(span, nil)
	return true
}
func (*ProxyUpdateProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}

type TwoComponentSlices struct {
	Current []model.ComponentSpec `json:"current"`
	Desired []model.ComponentSpec `json:"desired"`
}
