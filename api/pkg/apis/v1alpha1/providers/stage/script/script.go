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
	"fmt"
	"io"
	"net"
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
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
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

// privateIPNets lists IP ranges that are not routable on the public internet.
// Requests to these ranges are blocked to prevent SSRF attacks.
var privateIPNets = func() []*net.IPNet {
	var nets []*net.IPNet
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10", // shared address space (RFC 6598)
		"fc00::/7",      // IPv6 unique local (RFC 4193)
	} {
		_, n, err := net.ParseCIDR(cidr)
		if err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}()

// scriptDownloadClient is a dedicated HTTP client used for downloading scripts.
// It enforces a request timeout and re-validates redirect destinations to prevent SSRF.
var scriptDownloadClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("maximum redirect limit reached")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("redirect to non-HTTP(S) scheme %q is not permitted", req.URL.Scheme)
		}
		if err := validateURLHost(req.URL.Hostname()); err != nil {
			return fmt.Errorf("redirect blocked: %w", err)
		}
		return nil
	},
}

// validateScriptName ensures the script name is a safe, local filename.
// It rejects path traversal sequences ("..", absolute paths) and directory separators.
func validateScriptName(name string) error {
	if !filepath.IsLocal(name) {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid script name %q: must be a local relative path with no '..' components", name),
			v1alpha2.BadConfig)
	}
	if strings.ContainsAny(name, "/\\") {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid script name %q: directory separators are not allowed", name),
			v1alpha2.BadConfig)
	}
	return nil
}

// validateScriptFolderURL validates that the scriptFolder URL is safe to use.
// It enforces http/https schemes and rejects URLs that resolve to private or loopback addresses.
func validateScriptFolderURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return v1alpha2.NewCOAError(err, "invalid scriptFolder URL", v1alpha2.BadConfig)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid URL scheme %q: only http and https are permitted", u.Scheme),
			v1alpha2.BadConfig)
	}
	return validateURLHost(u.Hostname())
}

// validateURLHost resolves host and rejects loopback, link-local, and private addresses.
func validateURLHost(host string) error {
	if ip := net.ParseIP(host); ip != nil {
		return checkIPAllowed(ip, host)
	}
	addrs, err := net.LookupHost(host)
	if err != nil {
		return v1alpha2.NewCOAError(err, fmt.Sprintf("cannot resolve host %q", host), v1alpha2.BadConfig)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if err := checkIPAllowed(ip, host); err != nil {
			return err
		}
	}
	return nil
}

// checkIPAllowed returns an error if the IP is a loopback, link-local, or private address.
func checkIPAllowed(ip net.IP, host string) error {
	if ip.IsLoopback() {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("host %q resolves to a loopback address which is not permitted", host),
			v1alpha2.BadConfig)
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("host %q resolves to a link-local address which is not permitted", host),
			v1alpha2.BadConfig)
	}
	for _, n := range privateIPNets {
		if n.Contains(ip) {
			return v1alpha2.NewCOAError(nil,
				fmt.Sprintf("host %q resolves to a private address which is not permitted", host),
				v1alpha2.BadConfig)
		}
	}
	return nil
}

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
		err = validateScriptFolderURL(i.Config.ScriptFolder)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Stage): scriptFolder URL validation failed: %+v", err)
			return err
		}
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
	sLog.Debugf("  downloadFile: scriptFolder=%q, script=%q, stagingFolder=%q", scriptFolder, script, stagingFolder)

	// 1. Normalize: unescape first to get a clean raw string.
	//    If the input is already unescaped (e.g. "deploy$1.sh"), PathUnescape is a no-op.
	//    If it was pre-encoded (e.g. "deploy%241.sh"), this yields "deploy$1.sh".
	rawScript, err := url.PathUnescape(script)
	if err != nil {
		// Fallback: if unescaping fails (e.g. bare "%" like "deploy%test.sh"),
		// use the original string as-is.
		rawScript = script
	}

	// 2. Validate the script name to prevent path traversal attacks.
	if err := validateScriptName(rawScript); err != nil {
		return err
	}

	// 3. Escape the script name for the URL path.
	//    url.PathEscape handles spaces (%20), percent (%25), etc. but does NOT
	//    escape RFC 3986 sub-delimiters ($, &, +, =). We must encode them manually
	//    to ensure the download URL is unambiguous for all HTTP servers.
	escapedScript := url.PathEscape(rawScript)
	escapedScript = utils2.EncodeSubDelimiters(escapedScript)

	// 4. Normalize and encode sub-delimiters in the scriptFolder URL path.
	escapedFolder := utils2.EscapeURLPathSubDelims(scriptFolder)

	sPath, err := url.JoinPath(escapedFolder, escapedScript)
	if err != nil {
		return err
	}

	// 5. Use the raw, clean string for the local file system path.
	//    '$' and '%' are valid filesystem characters; no escaping needed.
	tPath := filepath.Join(stagingFolder, rawScript)
	sLog.Debugf("  downloadFile: resolved URL=%q, localPath=%q", sPath, tPath)

	// 6. Fetch the script using the dedicated client (enforces timeout and redirect policy).
	resp, err := scriptDownloadClient.Get(sPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return v1alpha2.NewCOAError(
			nil,
			"Response body content: "+string(body),
			v1alpha2.State(resp.StatusCode),
		)
	}

	out, err := os.Create(tPath)
	if err != nil {
		return err
	}
	defer out.Close()

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
	err = utils2.UnmarshalJson(data, &ret)
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
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		script,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)
	id := uuid.New().String()
	input := id + ".json"
	output := id + "-output.json"

	staging := filepath.Join(i.Config.StagingFolder, input)
	file, _ := json.MarshalIndent(inputs, "", " ")
	_ = os.WriteFile(staging, file, 0644)

	abs, _ := filepath.Abs(staging)

	defer os.Remove(abs)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.Script))
	observ_utils.EmitUserAuditsLogs(ctx, "  P (Script Stage): Start to run script %s", i.Config.Script)
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
	data, err = os.ReadFile(outputStaging)

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
	err = utils2.UnmarshalJson(data, &ret)
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
