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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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
		aLog.Errorf("  P (Android ADB Target): expected AdbProviderConfig: %+v", err)
		return err
	}
	return i.Init(config)
}
func (s *AdbProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *AdbProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("Android ADB Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	aLog.InfoCtx(ctx, "  P (Android ADB Target): Init()")

	updateConfig, err := toAdbProviderConfig(config)
	if err != nil {
		aLog.ErrorfCtx(ctx, "  P (Android ADB Target): expected AdbProviderConfig: %+v", err)
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

func (i *AdbProvider) Get(ctx context.Context, reference model.TargetProviderGetReference) ([]model.ComponentSpec, error) {
	ctx, span := observability.StartSpan("Android ADB Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	if reference.Deployment.Instance.Spec == nil {
		err = errors.New("deployment instance spec is nil")
		aLog.ErrorfCtx(ctx, "  P (Android ADB Target): failed to get deployment, error: %+v", err)
		return nil, err
	}
	aLog.InfofCtx(ctx, "  P (Android ADB Target): getting artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)

	re := regexp.MustCompile(`^package:(\w+\.)+\w+$`)

	for _, component := range reference.References {
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
				aLog.ErrorfCtx(ctx, "  P (Android ADB Target): failed to get application %+v, error: %+v", p, err)
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

func (i *AdbProvider) Apply(ctx context.Context, reference model.TargetProviderApplyReference) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Android ADB Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	aLog.InfofCtx(ctx, "  P (Android ADB Target): applying artifacts: %s - %s", reference.Deployment.Instance.Spec.Scope, reference.Deployment.Instance.ObjectMeta.Name)

	components := reference.Step.GetComponents()

	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		aLog.ErrorfCtx(ctx, "  P (Android ADB Target): failed to validate components, error: %v", err)
		return nil, err
	}
	if reference.IsDryRun {
		aLog.DebugCtx(ctx, "  P (Android ADB Target): dryRun is enabled, skipping apply")
		err = nil
		return nil, nil
	}
	ret := reference.Step.PrepareResultMap()
	components = reference.Step.GetUpdatedComponents()
	if len(components) > 0 {
		aLog.InfofCtx(ctx, "  P (Android ADB Target): get updated components: count - %d", len(components))
		for _, component := range components {
			if component.Name != "" {
				if p, ok := component.Properties[model.AppImage]; ok && p != "" {
					if !reference.IsDryRun {
						params := make([]string, 0)
						params = append(params, "install")
						params = append(params, utils.FormatAsString(p))
						cmd := exec.Command("adb", params...)
						err = cmd.Run()
						if err != nil {
							aLog.ErrorfCtx(ctx, "  P (Android ADB Target): failed to install application %+v, error: %+v", p, err)
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
	components = reference.Step.GetDeletedComponents()
	if len(components) > 0 {
		aLog.InfofCtx(ctx, "  P (Android ADB Target): get deleted components: count - %d", len(components))
		for _, component := range components {
			if component.Name != "" {
				if p, ok := component.Properties[model.AppPackage]; ok && p != "" {
					params := make([]string, 0)
					params = append(params, "uninstall")
					params = append(params, utils.FormatAsString(p))

					cmd := exec.Command("adb", params...)
					err = cmd.Run()
					if err != nil {
						aLog.ErrorfCtx(ctx, "  P (Android ADB Target): failed to uninstall application %+v, error: %+v", p, err)
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
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{model.AppPackage, model.AppImage},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
		},
	}
}
