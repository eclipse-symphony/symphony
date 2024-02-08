/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sideload

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")

type Win10SideLoadProviderConfig struct {
	Name                string `json:"name"`
	IPAddress           string `json:"ipAddress"`
	Pin                 string `json:"pin,omitempty"`
	WinAppDeployCmdPath string `json:"winAppDeployCmdPath"`
	NetworkUser         string `json:"networkUser,omitempty"`
	NetworkPassword     string `json:"networkPassword,omitempty"`
	Silent              bool   `json:"silent,omitempty"`
}

type Win10SideLoadProvider struct {
	Config  Win10SideLoadProviderConfig
	Context *contexts.ManagerContext
}

func Win10SideLoadProviderConfigFromMap(properties map[string]string) (Win10SideLoadProviderConfig, error) {
	ret := Win10SideLoadProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["ipAddress"]; ok {
		ret.IPAddress = v
	} else {
		ret.IPAddress = "localhost"
	}
	if v, ok := properties["pin"]; ok {
		ret.Pin = v
	}
	if v, ok := properties["winAppDeployCmdPath"]; ok {
		ret.WinAppDeployCmdPath = v
	} else {
		ret.WinAppDeployCmdPath = "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.19041.0\\x86\\WinAppDeployCmd.exe"
	}
	if v, ok := properties["networkUser"]; ok {
		ret.NetworkUser = v
	}
	if v, ok := properties["networkPassword"]; ok {
		ret.NetworkPassword = v
	}
	if v, ok := properties["silent"]; ok {
		bVal, err := strconv.ParseBool(v)
		if err != nil {
			ret.Silent = false
		} else {
			ret.Silent = bVal
		}
	}
	return ret, nil
}
func (i *Win10SideLoadProvider) InitWithMap(properties map[string]string) error {
	config, err := Win10SideLoadProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *Win10SideLoadProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *Win10SideLoadProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Win 10 Sideload Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (Win10Sideload Target): Init()")

	updateConfig, err := toWin10SideLoadProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Win10Sideload Target): expected Win10SideLoadProviderConfig - %+v", err)
		err = errors.New("expected Win10SideLoadProviderConfig")
		return err
	}
	i.Config = updateConfig

	return nil
}
func toWin10SideLoadProviderConfig(config providers.IProviderConfig) (Win10SideLoadProviderConfig, error) {
	ret := Win10SideLoadProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *Win10SideLoadProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Win 10 Sideload Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Win10Sideload Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	params := make([]string, 0)
	params = append(params, "list")
	params = append(params, "-ip")
	params = append(params, i.Config.IPAddress)
	if i.Config.Pin != "" {
		params = append(params, "-pin")
		params = append(params, i.Config.Pin)
	}

	out, err := exec.Command(i.Config.WinAppDeployCmdPath, params...).Output()

	if err != nil {
		sLog.Errorf("  P (Win10Sideload Target): failed to run deploy cmd %s, error: %+v, traceId: %s", i.Config.WinAppDeployCmdPath, err, span.SpanContext().TraceID().String())
		return nil, err
	}
	str := string(out)
	lines := strings.Split(str, "\r\n")

	desired := deployment.GetComponentSlice()

	re := regexp.MustCompile(`^(\w+\.)+\w+$`)
	ret := make([]model.ComponentSpec, 0)
	for _, line := range lines {
		if re.Match([]byte(line)) {
			mLine := line
			if strings.LastIndex(line, "__") > 0 {
				mLine = line[:strings.LastIndex(line, "__")]
			}
			for _, component := range desired {
				if component.Name == mLine {
					ret = append(ret, model.ComponentSpec{
						Name: line,
						Type: "win.uwp",
					})
				}
			}
		}
	}

	return ret, nil
}
func (i *Win10SideLoadProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Win 10 Sideload Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Win10Sideload Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	components := step.GetComponents()
	err = i.GetValidationRule(ctx).Validate(components)
	if err != nil {
		sLog.Errorf("  P (Win10Sideload Target): failed to validate components, error: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
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
			if path, ok := component.Properties["app.package.path"].(string); ok {
				params := make([]string, 0)
				params = append(params, "install")
				params = append(params, "-ip")
				params = append(params, i.Config.IPAddress)
				if i.Config.Pin != "" {
					params = append(params, "-pin")
					params = append(params, i.Config.Pin)
				}
				params = append(params, "-file")
				params = append(params, path)

				cmd := exec.Command(i.Config.WinAppDeployCmdPath, params...)
				err = cmd.Run()
				if err != nil {
					sLog.Errorf("  P (Win10Sideload Target): failed to install application %s, error: %+v, traceId: %s", path, err, span.SpanContext().TraceID().String())
					ret[component.Name] = model.ComponentResultSpec{
						Status:  v1alpha2.UpdateFailed,
						Message: err.Error(),
					}
					if i.Config.Silent {
						return ret, nil
					} else {
						return ret, err
					}
				}
			}
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		for _, component := range components {
			if component.Name != "" {
				params := make([]string, 0)
				params = append(params, "uninstall")
				params = append(params, "-ip")
				params = append(params, i.Config.IPAddress)
				if i.Config.Pin != "" {
					params = append(params, "-pin")
					params = append(params, i.Config.Pin)
				}
				params = append(params, "-package")

				name := component.Name

				// TODO: this is broken due to the refactor, the current reference is no longer available
				// for _, ref := range currentRef {
				// 	if ref.Name == name || strings.HasPrefix(ref.Name, name) {
				// 		name = ref.Name
				// 		break
				// 	}
				// }

				params = append(params, name)

				cmd := exec.Command(i.Config.WinAppDeployCmdPath, params...)
				err = cmd.Run()
				if err != nil {
					sLog.Errorf("  P (Win10Sideload Target): failed to uninstall application %s, error: %+v, traceId: %s", name, err, span.SpanContext().TraceID().String())
					if i.Config.Silent {
						return ret, nil
					} else {
						return ret, err
					}
				}

			}
		}
	}
	err = nil
	return ret, nil
}

func (i *Win10SideLoadProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, d := range desired {
		found := false
		for _, c := range current {
			if c.Name == d.Name || strings.HasPrefix(c.Name, d.Name) {
				found = true
			}
		}
		if !found {
			return true
		}
	}
	return false
}
func (i *Win10SideLoadProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	for _, d := range desired {
		for _, c := range current {
			if c.Name == d.Name || strings.HasPrefix(c.Name, d.Name) {
				return true
			}
		}
	}
	return false
}

func (*Win10SideLoadProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
		ChangeDetectionProperties: []model.PropertyDesc{
			{Name: "", IsComponentName: true, IgnoreCase: true, PrefixMatch: true},
		},
	}
}
