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
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	script     = "script"
	loggerName = "providers.target.script"
)

var (
	sLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

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
		sLog.Errorf("  P (Script Target): expected ScriptProviderConfig: %+v", err)
		return err
	}
	return i.Init(config)
}

func (s *ScriptProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *ScriptProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("Script Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Script Target): Init()")

	updateConfig, err := toScriptProviderConfig(config)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): expected ScriptProviderConfig - %+v", err)
		err = errors.New("expected ScriptProviderConfig")
		return err
	}
	i.Config = updateConfig

	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		err = downloadFile(i.Config.ScriptFolder, i.Config.ApplyScript, i.Config.StagingFolder)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download apply script %s, error: %+v", i.Config.ApplyScript, err)
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.RemoveScript, i.Config.StagingFolder)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download remove script %s, error: %+v", i.Config.RemoveScript, err)
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.GetScript, i.Config.StagingFolder)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download get script %s, error: %+v", i.Config.GetScript, err)
			return err
		}
	}

	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Script Target): failed to create metrics: %+v", err)
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
	ctx, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfofCtx(ctx, "  P (Script Target): getting artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	ret := make([]model.ComponentSpec, 0)
	for _, ref := range references {
		ret = append(ret, ref.Component)
	}

	return ret, nil
}

func (i *ScriptProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	ctx, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Script Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	functionName := observ_utils.GetFunctionName()
	startTime := time.Now().UTC()

	defer providerOperationMetrics.ProviderOperationLatency(
		startTime,
		script,
		metrics.ApplyOperation,
		metrics.ApplyOperationType,
		functionName,
	)

	err = i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components	TODO: is this right?
	if err != nil {
		providerOperationMetrics.ProviderOperationErrors(
			script,
			functionName,
			metrics.ValidateRuleOperation,
			metrics.ApplyOperationType,
			v1alpha2.ValidateFailed.String(),
		)
		return nil, err
	}
	if isDryRun {
		sLog.InfofCtx(ctx, "  P (Proxy Target): dryRun is enabled, skipping apply")
		err = nil
		return nil, nil
	}

	ret := step.PrepareResultMap()
	components := step.GetUpdatedComponents()
	sLog.InfofCtx(ctx, "  P (Script Target): get updated components: count - %d", len(components))
	for _, c := range components {
		path, ok := c.Parameters["path"]
		if !ok {
			sLog.ErrorfCtx(ctx, "  P (Script Target): invalid script provider config, expected 'path'")
			err = v1alpha2.NewCOAError(nil, "  P (Script Target): invalid script component config, expected 'path'", v1alpha2.BadConfig)
			providerOperationMetrics.ProviderOperationErrors(
				script,
				functionName,
				metrics.ApplyScriptOperation,
				metrics.ApplyOperationType,
				v1alpha2.ApplyScriptFailed.String(),
			)
			return nil, err
		}

		var cmd *exec.Cmd
		args, ok := c.Parameters["args"]
		flag, ok := c.Parameters["flag"]
		if flag != "" {
			cmd = exec.Command(path, flag, args)
		} else {
			cmd = exec.Command(path, args)
		}

		output, err := cmd.Output()
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Target): failed to run apply script: %+v", err)
			providerOperationMetrics.ProviderOperationErrors(
				script,
				functionName,
				metrics.ApplyScriptOperation,
				metrics.ApplyOperationType,
				v1alpha2.ApplyScriptFailed.String(),
			)
			return nil, err
		}
		// read the output of the script
		// read the output of the script
		sLog.InfofCtx(ctx, "  P (Script Target): script output: %s", string(output))

		ret[c.Name] = model.ComponentResultSpec{
			Status:  v1alpha2.Updated,
			Message: string(output),
		}
	}

	components = step.GetDeletedComponents()
	for _, c := range components {
		sLog.InfofCtx(ctx, "  P (Script Target): get deleted components: count - %d", len(components))
		ret[c.Name] = model.ComponentResultSpec{
			Status:  v1alpha2.Deleted,
			Message: "deleted",
		}
	}

	for _, v := range ret {
		switch v.Status {
		case v1alpha2.DeleteFailed, v1alpha2.ValidateFailed, v1alpha2.UpdateFailed:
			err := v1alpha2.NewCOAError(errors.New(v.Message), "executing script returned error output", v.Status)
			return ret, err
		}
	}
	return ret, nil
}
func (*ScriptProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{
		AllowSidecar: false,
		ComponentValidationRule: model.ComponentValidationRule{
			RequiredProperties:    []string{},
			OptionalProperties:    []string{},
			RequiredComponentType: "",
			RequiredMetadata:      []string{},
			OptionalMetadata:      []string{},
		},
	}
}
