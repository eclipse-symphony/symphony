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
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

const (
	loggerName   = "providers.stage.script"
	providerName = "P (Script Stage)"
	script       = "script"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type ScriptStageProviderConfig struct {
	Name          string `json:"name"`
	Script        string `json:"script"`
	ScriptFolder  string `json:"scriptFolder,omitempty"`
	StagingFolder string `json:"stagingFolder,omitempty"`
	ScriptEngine  string `json:"scriptEngine,omitempty"`
}

type ScriptStageProvider struct {
	Config  ScriptStageProviderConfig
	Context *contexts.ManagerContext
}

func ScriptProviderConfigFromMap(properties map[string]string) (ScriptStageProviderConfig, error) {
	ret := ScriptStageProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v
	}
	if v, ok := properties["stagingFolder"]; ok {
		ret.StagingFolder = v
	}
	if v, ok := properties["scriptFolder"]; ok {
		ret.ScriptFolder = v
	}
	if v, ok := properties["script"]; ok {
		ret.Script = v
	} else {
		return ret, v1alpha2.NewCOAError(nil, "invalid script provider config, exptected 'script'", v1alpha2.BadConfig)
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
func (i *ScriptStageProvider) InitWithMap(properties map[string]string) error {
	config, err := ScriptProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *ScriptStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *ScriptStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Script Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Script Stage): Init()")

	updateConfig, err := toScriptStageProviderConfig(config)
	if err != nil {
		err = errors.New("expected ScriptProviderConfig")
		return err
	}
	i.Config = updateConfig

	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		err = downloadFile(i.Config.ScriptFolder, i.Config.Script, i.Config.StagingFolder)
		if err != nil {
			return err
		}
	}
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.Errorf("  P (HTTP Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
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
func toScriptStageProviderConfig(config providers.IProviderConfig) (ScriptStageProviderConfig, error) {
	ret := ScriptStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (i *ScriptStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Script Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Script Stage): start process request")

	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	id := uuid.New().String()
	input := id + ".json"
	output := id + "-output.json"

	staging := filepath.Join(i.Config.StagingFolder, input)
	file, _ := json.MarshalIndent(inputs, "", " ")
	_ = ioutil.WriteFile(staging, file, 0644)

	abs, _ := filepath.Abs(staging)

	defer os.Remove(abs)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.Script))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.Script))
	}

	var o []byte
	o, err = i.runCommand(scriptAbs, abs)
	sLog.DebugfCtx(ctx, "  P (Script Stage): get script output: %s", o)

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Stage): failed to run get script: %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			script,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.ScriptExecutionFailed.String(),
		)
		return nil, false, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	var data []byte
	data, err = ioutil.ReadFile(outputStaging)

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Stage): failed to parse get script output (expected map[string]interface{}): %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			script,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.ScriptResultParsingFailed.String(),
		)
		return nil, false, err
	}

	abs_output, _ := filepath.Abs(outputStaging)

	defer os.Remove(abs_output)

	ret := make(map[string]interface{})
	err = json.Unmarshal(data, &ret)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Stage): failed to parse get script output (expected map[string]interface{}): %+v", err)
		providerOperationMetrics.ProviderOperationErrors(
			script,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.ScriptResultParsingFailed.String(),
		)
		return nil, false, err
	}

	providerOperationMetrics.ProviderOperationLatency(
		processTime,
		script,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)
	return ret, false, nil
}

func (i *ScriptStageProvider) runCommand(scriptAbs string, parameters ...string) ([]byte, error) {
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
