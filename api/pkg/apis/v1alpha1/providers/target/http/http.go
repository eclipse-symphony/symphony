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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type HttpTargetProviderConfig struct {
	Name string `json:"name"`
}

type HttpTargetProvider struct {
	Config HttpTargetProviderConfig
}

func HttpTargetProviderConfigFromMap(properties map[string]string) (HttpTargetProviderConfig, error) {
	ret := HttpTargetProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	return ret, nil
}

func (i *HttpTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := HttpTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (i *HttpTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Http Target Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("~~~ Http Target Provider ~~~ : Init()")

	updateConfig, err := toHttpTargetProviderConfig(config)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Http Target Provider ~~~ : expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func toHttpTargetProviderConfig(config providers.IProviderConfig) (HttpTargetProviderConfig, error) {
	ret := HttpTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *HttpTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Infof("~~~ Http Target Provider ~~~ : getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	// This provider doesn't remember what it does, so it always return nil when asked

	observ_utils.CloseSpanWithError(span, nil)
	return nil, nil
}
func (i *HttpTargetProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	ctx, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("~~~ Http Target Provider ~~~ : deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	// This provider doesn't remove anything

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
func (i *HttpTargetProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	// This essentially always return true. This means the HTTP trigger needs to idempotent
	return !model.SlicesCover(desired, current)
}
func (i *HttpTargetProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	// This essentially always return true. This means the HTTP trigger needs to idempotent
	return model.SlicesAny(desired, current)
}

func (i *HttpTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec) error {
	_, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("~~~ Http Target Provider ~~~ : applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Stages[0].Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := deployment.GetComponentSlice()

	for _, component := range components {

		body := model.ReadProperty(component.Properties, "http.body", injections)
		url := model.ReadProperty(component.Properties, "http.url", injections)
		method := model.ReadProperty(component.Properties, "http.method", injections)

		if url == "" {
			err := errors.New("component doesn't have a http.url property")
			observ_utils.CloseSpanWithError(span, err)
			sLog.Error("~~~ HTML Target Provider ~~~ : +%v", err)
			return err
		}
		if method == "" {
			method = "POST"
		}

		jsonData := []byte(body)
		request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Error("~~~ HTML Target Provider ~~~ : +%v", err)
			return err
		}
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			observ_utils.CloseSpanWithError(span, err)
			sLog.Error("~~~ HTML Target Provider ~~~ : +%v", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			err = errors.New("HTTP request didn't respond 200 OK")
			observ_utils.CloseSpanWithError(span, err)
			sLog.Error("~~~ HTML Target Provider ~~~ : +%v", err)
			return err
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}
