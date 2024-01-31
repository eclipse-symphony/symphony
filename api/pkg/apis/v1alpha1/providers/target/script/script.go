/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

var sLog = logger.NewLogger("coa.runtime")

type ScriptProviderConfig struct {
	Name          string `json:"name"`
	ApplyScript   string `json:"applyScript"`
	RemoveScript  string `json:"removeScript"`
	GetScript     string `json:"getScript"`
	ScriptFolder  string `json:"scriptFolder,omitempty"`
	StagingFolder string `json:"stagingFolder,omitempty"`
	ScriptEngine  string `json:"scriptEngine,omitempty"`
}

type ScriptProvider struct {
	Config  ScriptProviderConfig
	Context *contexts.ManagerContext
}

func ScriptProviderConfigFromMap(properties map[string]string) (ScriptProviderConfig, error) {
	ret := ScriptProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["stagingFolder"]; ok {
		ret.StagingFolder = v
	}
	if v, ok := properties["scriptFolder"]; ok {
		ret.ScriptFolder = v
	}
	if v, ok := properties["applyScript"]; ok {
		ret.ApplyScript = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid script provider config, exptected 'applyScript'", v1alpha2.BadConfig)
	}
	if v, ok := properties["removeScript"]; ok {
		ret.RemoveScript = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid script provider config, exptected 'removeScript'", v1alpha2.BadConfig)
	}
	if v, ok := properties["getScript"]; ok {
		ret.GetScript = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid script provider config, exptected 'getScript'", v1alpha2.BadConfig)
	}
	if v, ok := properties["scriptEngine"]; ok {
		ret.ScriptEngine = v
	} else {
		ret.ScriptEngine = "bash"
	}
	if ret.ScriptEngine != "bash" && ret.ScriptEngine != "powershell" {
		return ret, v1alpha2.NewCOAError(nil, "invalid script engine, exptected 'bash' or 'powershell'", v1alpha2.BadConfig)
	}
	return ret, nil
}
func (i *ScriptProvider) InitWithMap(properties map[string]string) error {
	config, err := ScriptProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *ScriptProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *ScriptProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Script Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Info("  P (Script Target): Init()")

	updateConfig, err := toScriptProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Script Target): expected ScriptProviderConfig - %+v", err)
		err = errors.New("expected ScriptProviderConfig")
		return err
	}
	i.Config = updateConfig

	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		err = downloadFile(i.Config.ScriptFolder, i.Config.ApplyScript, i.Config.StagingFolder)
		if err != nil {
			sLog.Errorf("  P (Script Target): failed to download apply script %s, error: %+v", i.Config.ApplyScript, err)
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.RemoveScript, i.Config.StagingFolder)
		if err != nil {
			sLog.Errorf("  P (Script Target): failed to download remove script %s, error: %+v", i.Config.RemoveScript, err)
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.GetScript, i.Config.StagingFolder)
		if err != nil {
			sLog.Errorf("  P (Script Target): failed to download get script %s, error: %+v", i.Config.GetScript, err)
			return err
		}
	}

	return nil
}
func downloadFile(scriptFolder string, script string, stagingFolder string) error {
	sPath, err := url.JoinPath(scriptFolder, script)
	if err != nil {
		return err
	}
	tPath := filepath.Join(stagingFolder, script)

	out, err := os.Create(tPath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(sPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return os.Chmod(tPath, 0755)
}
func toScriptProviderConfig(config providers.IProviderConfig) (ScriptProviderConfig, error) {
	ret := ScriptProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *ScriptProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Infof("  P (Script Target): getting artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	id := uuid.New().String()
	input := id + ".json"
	input_ref := id + "-ref.json"
	output := id + "-output.json"

	staging := filepath.Join(i.Config.StagingFolder, input)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = ioutil.WriteFile(staging, file, 0644)

	staging_ref := filepath.Join(i.Config.StagingFolder, input_ref)
	file_ref, _ := json.MarshalIndent(references, "", " ")
	_ = ioutil.WriteFile(staging_ref, file_ref, 0644)

	abs, _ := filepath.Abs(staging)
	abs_ref, _ := filepath.Abs(staging_ref)

	defer os.Remove(abs)
	defer os.Remove(abs_ref)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.GetScript))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.GetScript))
	}

	o, err := i.runCommand(scriptAbs, abs, abs_ref)
	sLog.Debugf("  P (Script Target): get script output: %s, traceId: %s", o, span.SpanContext().TraceID().String())

	if err != nil {
		sLog.Errorf("  P (Script Target): failed to run get script: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	data, err := ioutil.ReadFile(outputStaging)

	if err != nil {
		sLog.Errorf("  P (Script Target): failed to read output file: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}

	abs_output, _ := filepath.Abs(outputStaging)

	defer os.Remove(abs_output)

	ret := make([]model.ComponentSpec, 0)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		sLog.Errorf("  P (Script Target): failed to parse get script output (expected []ComponentSpec): %+v, traceId: %s", err, span.SpanContext().TraceID().String())
		return nil, err
	}
	return ret, nil
}
func (i *ScriptProvider) runScriptOnComponents(deployment model.DeploymentSpec, components []model.ComponentSpec, isRemove bool) (map[string]model.ComponentResultSpec, error) {
	id := uuid.New().String()
	deploymentId := id + ".json"
	currenRefId := id + "-ref.json"
	output := id + "-output.json"

	stagingDeployment := filepath.Join(i.Config.StagingFolder, deploymentId)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = ioutil.WriteFile(stagingDeployment, file, 0644)

	stagingRef := filepath.Join(i.Config.StagingFolder, currenRefId)
	file, _ = json.MarshalIndent(components, "", " ")
	_ = ioutil.WriteFile(stagingRef, file, 0644)

	absDeployment, _ := filepath.Abs(stagingDeployment)
	absRef, _ := filepath.Abs(stagingRef)

	var scriptAbs = ""
	if isRemove {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.RemoveScript))
		if strings.HasPrefix(i.Config.ScriptFolder, "http") {
			scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.RemoveScript))
		}
	} else {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.ApplyScript))
		if strings.HasPrefix(i.Config.ScriptFolder, "http") {
			scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.ApplyScript))
		}
	}
	o, err := i.runCommand(scriptAbs, absDeployment, absRef)
	sLog.Debugf("  P (Script Target): apply script output: %s", o)

	defer os.Remove(absDeployment)
	defer os.Remove(absRef)

	if err != nil {
		sLog.Errorf("  P (Script Target): failed to run apply script: %+v", err)
		return nil, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	data, err := ioutil.ReadFile(outputStaging)

	if err != nil {
		sLog.Errorf("  P (Script Target): failed to pread output file: %+v", err)
		return nil, err
	}

	abs_output, _ := filepath.Abs(outputStaging)

	defer os.Remove(abs_output)

	ret := make(map[string]model.ComponentResultSpec)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		sLog.Errorf("  P (Script Target): failed to parse get script output (expected map[string]model.ComponentResultSpec): %+v", err)
		return nil, err
	}
	return ret, nil
}
func (i *ScriptProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	sLog.Infof("  P (Script Target): applying artifacts: %s - %s, traceId: %s", deployment.Instance.Scope, deployment.Instance.Name, span.SpanContext().TraceID().String())

	err = i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components	TODO: is this right?
	if err != nil {
		return nil, err
	}
	if isDryRun {
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components := step.GetUpdatedComponents()
	if len(components) > 0 {
		var retU map[string]model.ComponentResultSpec
		retU, err = i.runScriptOnComponents(deployment, components, false)
		if err != nil {
			sLog.Errorf("  P (Script Target): failed to run apply script: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return nil, err
		}
		for k, v := range retU {
			ret[k] = v
		}
	}
	components = step.GetDeletedComponents()
	if len(components) > 0 {
		var retU map[string]model.ComponentResultSpec
		retU, err = i.runScriptOnComponents(deployment, components, true)
		if err != nil {
			sLog.Errorf("  P (Script Target): failed to run remove script: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return nil, err
		}
		for k, v := range retU {
			ret[k] = v
		}
	}
	return ret, nil
}
func (*ScriptProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		RequiredProperties:    []string{},
		OptionalProperties:    []string{},
		RequiredComponentType: "",
		RequiredMetadata:      []string{},
		OptionalMetadata:      []string{},
	}
}

func (i *ScriptProvider) runCommand(scriptAbs string, parameters ...string) ([]byte, error) {
	// Sanitize input to prevent command injection
	scriptAbs = strings.ReplaceAll(scriptAbs, "|", "")
	scriptAbs = strings.ReplaceAll(scriptAbs, "&", "")
	for idx, param := range parameters {
		parameters[idx] = strings.ReplaceAll(param, "|", "")
		parameters[idx] = strings.ReplaceAll(param, "&", "")
	}

	var err error
	var out []byte
	params := make([]string, 0)
	if i.Config.ScriptEngine == "" || i.Config.ScriptEngine == "bash" {
		params = append(params, parameters...)
		out, err = exec.Command(scriptAbs, params...).Output()
	} else {
		params = append(params, scriptAbs)
		params = append(params, parameters...)
		out, err = exec.Command("powershell", params...).Output()
	}
	return out, err
}
