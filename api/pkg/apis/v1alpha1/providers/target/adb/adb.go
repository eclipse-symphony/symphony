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

package adb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var aLog = logger.NewLogger("coa.runtime")

type AdbProviderConfig struct {
	Name string `json:"name"`
}

type AdbProvider struct {
	Config  AdbProviderConfig
	Context *contexts.ManagerContext
}

func AdbProviderConfigFromMap(properties map[string]string) (AdbProviderConfig, error) {
	ret := AdbProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	return ret, nil
}

func (i *AdbProvider) InitWithMap(properties map[string]string) error {
	config, err := AdbProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func (s *AdbProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *AdbProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Android ADB Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	aLog.Info("  P (Android ADB): Init()")

	updateConfig, err := toAdbProviderConfig(config)
	if err != nil {
		return errors.New("expected AdbProviderConfig")
	}
	i.Config = updateConfig

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func toAdbProviderConfig(config providers.IProviderConfig) (AdbProviderConfig, error) {
	ret := AdbProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *AdbProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Android ADB Provider", context.Background(), &map[string]string{
		"method": "Get",
	})
	aLog.Infof("  P (Android ADB): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	ret := make([]model.ComponentSpec, 0)

	re := regexp.MustCompile(`^package:(\w+\.)+\w+$`)

	for _, component := range references {
		if p, ok := component.Component.Properties[model.AppPackage]; ok {
			params := make([]string, 0)
			params = append(params, "shell")
			params = append(params, "pm")
			params = append(params, "list")
			params = append(params, "packages")
			params = append(params, fmt.Sprintf("%v", p))
			out, err := exec.Command("adb", params...).Output()

			if err != nil {
				observ_utils.CloseSpanWithError(span, err)
				return nil, err
			}
			str := string(out)
			lines := strings.Split(str, "\r\n")
			for _, line := range lines {
				if re.Match([]byte(line)) {
					ret = append(ret, model.ComponentSpec{
						Name: line[8:],
						Type: model.AppPackage,
					})
				}
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (i *AdbProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	_, span := observability.StartSpan("Android ADB Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	aLog.Infof("  P (Android ADB Provider): applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	components := step.GetComponents()

	err := i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return nil, err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil, nil
	}
	ret := step.PrepareResultMap()
	components = step.GetUpdatedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Name != "" {
				if p, ok := component.Properties[model.AppImage]; ok && p != "" {
					if !isDryRun {
						params := make([]string, 0)
						params = append(params, "install")
						params = append(params, p.(string))
						cmd := exec.Command("adb", params...)
						err := cmd.Run()
						if err != nil {
							ret[component.Name] = model.ComponentResultSpec{
								Status:  v1alpha2.UpdateFailed,
								Message: err.Error(),
							}
							observ_utils.CloseSpanWithError(span, err)
							return ret, err
						}
					}
				}
			}
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Name != "" {
				if p, ok := component.Properties[model.AppPackage]; ok && p != "" {
					params := make([]string, 0)
					params = append(params, "uninstall")
					params = append(params, p.(string))

					cmd := exec.Command("adb", params...)
					err := cmd.Run()
					if err != nil {
						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.DeleteFailed,
							Message: err.Error(),
						}
						observ_utils.CloseSpanWithError(span, err)
						return ret, err
					}
				}
			}
		}
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (*AdbProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{model.AppPackage, model.AppImage},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}
