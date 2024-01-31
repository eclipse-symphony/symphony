/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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
	_, span := observability.StartSpan("Android ADB Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	aLog.Info("  P (Android ADB): Init()")

	updateConfig, err := toAdbProviderConfig(config)
	if err != nil {
		aLog.Errorf("  P (Android ADB): expected AdbProviderConfig: %+v", err)
		return errors.New("expected AdbProviderConfig")
	}
	i.Config = updateConfig
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
	_, span := observability.StartSpan("Android ADB Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	aLog.Infof("  P (Android ADB): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

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
			var out []byte
			out, err = exec.Command("adb", params...).Output()

			if err != nil {
				aLog.Errorf("  P (Android ADB): failed to get application %+v, error: %+v, traceId: %s", p, err, span.SpanContext().TraceID().String())
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
	return ret, nil
}

func (i *AdbProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Android ADB Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	aLog.Infof("  P (Android ADB Provider): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()

	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		aLog.Errorf("  P (Android ADB Provider): failed to validate components, error: %v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	if isDryRun {
		err = nil
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
						err = cmd.Run()
						if err != nil {
							aLog.Errorf("  P (Android ADB): failed to install application %+v, error: %+v, traceId: %s", p, err, span.SpanContext().TraceID().String())
							ret[component.Name] = model.ComponentResultSpec{
								Status:  v1alpha2.UpdateFailed,
								Message: err.Error(),
							}
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
					err = cmd.Run()
					if err != nil {
						aLog.Errorf("  P (Android ADB): failed to uninstall application %+v, error: %+v, traceId: %s", p, err, span.SpanContext().TraceID().String())
						ret[component.Name] = model.ComponentResultSpec{
							Status:  v1alpha2.DeleteFailed,
							Message: err.Error(),
						}
						return ret, err
					}
				}
			}
		}
	}
	err = nil
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
