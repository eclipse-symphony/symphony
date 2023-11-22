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
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
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
	Config  HttpTargetProviderConfig
	Context *contexts.ManagerContext
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

func (s *HttpTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *HttpTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Http Target Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P(HTTP Target): Init()")

	updateConfig, err := toHttpTargetProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P(HTTP Target): expected HttpTargetProviderConfig: %+v", err)
		return err
	}
	i.Config = updateConfig

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
func (i *HttpTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P(HTTP Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	// This provider doesn't remember what it does, so it always return nil when asked
	return nil, nil
}

func (i *HttpTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Http Target Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P(HTTP Target): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	injections := &model.ValueInjections{
		InstanceId: deployment.Instance.Name,
		SolutionId: deployment.Instance.Solution,
		TargetId:   deployment.ActiveTarget,
	}

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	for _, component := range step.Components {
		if component.Action == "update" {
			body := model.ReadPropertyCompat(component.Component.Properties, "http.body", injections)
			url := model.ReadPropertyCompat(component.Component.Properties, "http.url", injections)
			method := model.ReadPropertyCompat(component.Component.Properties, "http.method", injections)

			if url == "" {
				err = errors.New("component doesn't have a http.url property")
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P(HTTP Target): %v", err)
				return ret, err
			}
			if method == "" {
				method = "POST"
			}
			jsonData := []byte(body)
			var request *http.Request
			request, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
			if err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P(HTTP Target): %v", err)
				return ret, err
			}
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")

			client := &http.Client{}
			var resp *http.Response
			resp, err = client.Do(request)
			if err != nil {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				sLog.Errorf("  P(HTTP Target): %v", err)
				return ret, err
			}
			if resp.StatusCode != http.StatusOK {
				ret[component.Component.Name] = model.ComponentResultSpec{
					Status:  v1alpha2.UpdateFailed,
					Message: err.Error(),
				}
				err = errors.New("HTTP request didn't respond 200 OK")
				sLog.Errorf("  P(HTTP Target): %v", err)
				return ret, err
			}
		}
	}
	return ret, nil
}
func (*HttpTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{"http.url"},
		OptionalProperties:    []string{"http.method", "http.body"},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
