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

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
	"github.com/google/uuid"
)

var sLog = logger.NewLogger("coa.runtime")

type ScriptProviderConfig struct {
	Name          string `json:"name"`
	ApplyScript   string `json:"applyScript"`
	RemoveScript  string `json:"removeScript"`
	GetScript     string `json:"getScript"`
	NeedsUpdate   string `json:"needsUpdate,omitempty"`
	NeedsRemove   string `json:"needsRemove,omitempty"`
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
	if v, ok := properties["needsUpdate"]; ok {
		ret.NeedsUpdate = v
	}
	if v, ok := properties["needsRemove"]; ok {
		ret.NeedsRemove = v
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

func (i *ScriptProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan("Script Provider", context.Background(), &map[string]string{
		"method": "Init",
	})
	sLog.Info("~~~ Script Provider ~~~ : Init()")

	updateConfig, err := toScriptProviderConfig(config)
	if err != nil {
		return errors.New("expected ScriptProviderConfig")
	}
	i.Config = updateConfig

	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		err = downloadFile(i.Config.ScriptFolder, i.Config.ApplyScript, i.Config.StagingFolder)
		if err != nil {
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.RemoveScript, i.Config.StagingFolder)
		if err != nil {
			return err
		}
		err = downloadFile(i.Config.ScriptFolder, i.Config.GetScript, i.Config.StagingFolder)
		if err != nil {
			return err
		}
		if i.Config.NeedsUpdate != "" {
			err = downloadFile(i.Config.ScriptFolder, i.Config.NeedsUpdate, i.Config.StagingFolder)
			if err != nil {
				return err
			}
		}
		if i.Config.NeedsRemove != "" {
			err = downloadFile(i.Config.ScriptFolder, i.Config.NeedsRemove, i.Config.StagingFolder)
			if err != nil {
				return err
			}
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
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

func (i *ScriptProvider) Get(ctx context.Context, deployment model.DeploymentSpec) ([]model.ComponentSpec, error) {
	_, span := observability.StartSpan("Script Provider", context.Background(), &map[string]string{
		"method": "Get",
	})
	sLog.Infof("  P (Script Target): getting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	id := uuid.New().String()
	input := id + ".json"
	output := id + "-output.json"

	staging := filepath.Join(i.Config.StagingFolder, input)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = ioutil.WriteFile(staging, file, 0644)

	abs, _ := filepath.Abs(staging)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.GetScript))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.GetScript))
	}

	_, err := i.runCommand(scriptAbs, abs)

	os.Remove(abs)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Script Target): failed to run get script: %+v", err)
		return nil, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	data, err := ioutil.ReadFile(outputStaging)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("  P (Script Target): failed to parse get script output (expected []ComponentSpec): %+v", err)
		return nil, err
	}

	abs, _ = filepath.Abs(outputStaging)

	os.Remove(abs)

	ret := make([]model.ComponentSpec, 0)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Script Provider ~~~ : failed to parse get script output (expected []ComponentSpec): %+v", err)
		return nil, err
	}
	observ_utils.CloseSpanWithError(span, nil)
	return ret, nil
}

func (i *ScriptProvider) NeedsUpdate(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Script Provider", context.Background(), &map[string]string{
		"method": "NeedsUpdate",
	})
	sLog.Info("~~~ Script Provider ~~~ : needs update")

	if i.Config.NeedsUpdate == "" {
		observ_utils.CloseSpanWithError(span, nil)
		return !model.SlicesCover(desired, current)
	}

	currentId := uuid.New().String() + "-current.json"
	desiredId := uuid.New().String() + "-desired.json"

	stagingCurrent := filepath.Join(i.Config.StagingFolder, currentId)
	file, _ := json.MarshalIndent(current, "", " ")
	_ = ioutil.WriteFile(stagingCurrent, file, 0644)

	stagingDesired := filepath.Join(i.Config.StagingFolder, desiredId)
	file, _ = json.MarshalIndent(desired, "", " ")
	_ = ioutil.WriteFile(stagingDesired, file, 0644)

	absCurrent, _ := filepath.Abs(stagingCurrent)
	absDesired, _ := filepath.Abs(stagingDesired)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.NeedsUpdate))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.NeedsUpdate))
	}

	out, err := i.runCommand(scriptAbs, absCurrent, absDesired)

	os.Remove(absCurrent)
	os.Remove(absDesired)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Script Provider ~~~ : failed to run needsupdate script: %+v", err)
		return false
	}
	str := string(out)

	if str != "" {
		s := strings.ToLower(strings.TrimSpace(str))
		observ_utils.CloseSpanWithError(span, nil)
		return s == "1" || s == "true"
	} else { // The script opts to return nothing, return false
		observ_utils.CloseSpanWithError(span, nil)
		return false
	}
}
func (i *ScriptProvider) NeedsRemove(ctx context.Context, desired []model.ComponentSpec, current []model.ComponentSpec) bool {
	_, span := observability.StartSpan("Script Provider", context.Background(), &map[string]string{
		"method": "NeedsRemove",
	})
	sLog.Info("~~~ Script Provider ~~~ : needs remove")

	if i.Config.NeedsRemove == "" {
		observ_utils.CloseSpanWithError(span, nil)
		return model.SlicesAny(desired, current)
	}

	currentId := uuid.New().String() + ".json"
	desiredId := uuid.New().String() + ".json"

	stagingCurrent := filepath.Join(i.Config.StagingFolder, currentId)
	file, _ := json.MarshalIndent(current, "", " ")
	_ = ioutil.WriteFile(stagingCurrent, file, 0644)

	stagingDesired := filepath.Join(i.Config.StagingFolder, desiredId)
	file, _ = json.MarshalIndent(desired, "", " ")
	_ = ioutil.WriteFile(stagingDesired, file, 0644)

	absCurrent, _ := filepath.Abs(stagingCurrent)
	absDesired, _ := filepath.Abs(stagingDesired)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.NeedsRemove))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.NeedsRemove))
	}

	out, err := i.runCommand(scriptAbs, absCurrent, absDesired)

	os.Remove(absCurrent)
	os.Remove(absDesired)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Script Provider ~~~ : failed to run needsupdate script: %+v", err)
		return false
	}
	str := string(out)
	if str != "" {
		s := strings.ToLower(strings.TrimSpace(str))
		observ_utils.CloseSpanWithError(span, nil)
		return s == "1" || s == "true"
	} else { // The script opts to return nothing, return false
		observ_utils.CloseSpanWithError(span, nil)
		return false
	}
}

func (i *ScriptProvider) Remove(ctx context.Context, deployment model.DeploymentSpec, currentRef []model.ComponentSpec) error {
	_, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	sLog.Infof("~~~ Script Provider ~~~ : deleting artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	deploymentId := uuid.New().String() + ".json"
	currenRefId := uuid.New().String() + ".json"

	stagingDeployment := filepath.Join(i.Config.StagingFolder, deploymentId)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = ioutil.WriteFile(stagingDeployment, file, 0644)

	stagingRef := filepath.Join(i.Config.StagingFolder, currenRefId)
	file, _ = json.MarshalIndent(currentRef, "", " ")
	_ = ioutil.WriteFile(stagingRef, file, 0644)

	absDeployment, _ := filepath.Abs(stagingDeployment)
	absRef, _ := filepath.Abs(stagingRef)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.RemoveScript))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.RemoveScript))
	}

	_, err := i.runCommand(scriptAbs, absDeployment, absRef)

	os.Remove(absDeployment)
	os.Remove(absRef)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Script Provider ~~~ : failed to run remove script: %+v", err)
		return err
	}
	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (i *ScriptProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, isDryRun bool) error {
	_, span := observability.StartSpan("Script Provider", ctx, &map[string]string{
		"method": "Apply",
	})
	sLog.Infof("~~~ Script Provider ~~~ : applying artifacts: %s - %s", deployment.Instance.Scope, deployment.Instance.Name)

	err := i.GetValidationRule(ctx).Validate([]model.ComponentSpec{}) //this provider doesn't handle any components
	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		return err
	}
	if isDryRun {
		observ_utils.CloseSpanWithError(span, nil)
		return nil
	}

	id := uuid.New().String() + ".json"

	staging := filepath.Join(i.Config.StagingFolder, id)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = ioutil.WriteFile(staging, file, 0644)

	abs, _ := filepath.Abs(staging)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.ApplyScript))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.ApplyScript))
	}

	_, err = i.runCommand(scriptAbs, abs)

	os.Remove(abs)

	if err != nil {
		observ_utils.CloseSpanWithError(span, err)
		sLog.Errorf("~~~ Script Provider ~~~ : failed to run apply script: %+v", err)
		return err
	}

	observ_utils.CloseSpanWithError(span, nil)
	return nil
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
