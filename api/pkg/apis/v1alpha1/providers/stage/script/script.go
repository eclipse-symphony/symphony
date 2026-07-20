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
// The redirect handler uses the deny-list only (no per-provider whitelist) since this
// client is package-level and cannot carry per-instance state.
var scriptDownloadClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("maximum redirect limit reached")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("redirect to non-HTTP(S) scheme %q is not permitted", req.URL.Scheme)
		}
		if err := validateURLHost(req.URL.Hostname(), nil, false); err != nil {
			return fmt.Errorf("redirect blocked: %w", err)
		}
		return nil
	},
}

// parseIPRanges converts a slice of CIDR strings or plain IP addresses to []*net.IPNet.
// Plain IP addresses are treated as host routes (/32 for IPv4, /128 for IPv6).
func parseIPRanges(ranges []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(ranges))
	for _, r := range ranges {
		if strings.Contains(r, "/") {
			_, n, err := net.ParseCIDR(r)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR range %q: %w", r, err)
			}
			nets = append(nets, n)
		} else {
			ip := net.ParseIP(r)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP address %q in allowedIPRanges", r)
			}
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			_, n, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip.String(), bits))
			if err != nil {
				return nil, fmt.Errorf("invalid IP address %q: %w", r, err)
			}
			nets = append(nets, n)
		}
	}
	return nets, nil
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
// It enforces http/https schemes and rejects URLs that resolve to private or loopback addresses,
// unless the address falls within one of the allowedNets ranges.
// When exclusiveMode is true, only IPs explicitly listed in allowedNets are permitted.
func validateScriptFolderURL(rawURL string, allowedNets []*net.IPNet, exclusiveMode bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return v1alpha2.NewCOAError(err, "invalid scriptFolder URL", v1alpha2.BadConfig)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid URL scheme %q: only http and https are permitted", u.Scheme),
			v1alpha2.BadConfig)
	}
	return validateURLHost(u.Hostname(), allowedNets, exclusiveMode)
}

// validateURLHost resolves host and rejects loopback, link-local, and private addresses,
// unless the address falls within one of the allowedNets ranges.
// When exclusiveMode is true, only IPs explicitly listed in allowedNets are permitted.
func validateURLHost(host string, allowedNets []*net.IPNet, exclusiveMode bool) error {
	if ip := net.ParseIP(host); ip != nil {
		return checkIPAllowed(ip, host, allowedNets, exclusiveMode)
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
		if err := checkIPAllowed(ip, host, allowedNets, exclusiveMode); err != nil {
			return err
		}
	}
	return nil
}

// checkIPAllowed returns an error if the IP is blocked by the current policy.
// IPs in allowedNets are always permitted (they override the deny list).
// When exclusiveMode is true, only IPs in allowedNets are permitted; all others are rejected.
func checkIPAllowed(ip net.IP, host string, allowedNets []*net.IPNet, exclusiveMode bool) error {
	// Whitelist check: allowedNets override the deny list.
	for _, n := range allowedNets {
		if n.Contains(ip) {
			return nil
		}
	}
	// In exclusive mode, reject anything not in the whitelist.
	if exclusiveMode {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("host %q resolves to an address not in the configured allowedIPRanges", host),
			v1alpha2.BadConfig)
	}
	// Standard deny-list checks.
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
	Config         ScriptStageProviderConfig
	Context        *contexts.ManagerContext
	scriptsReady   bool
	scriptsReadyMu sync.Mutex
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

// ensureScriptReady validates the scriptFolder URL against the server-side SecurityPolicy
// (from the SecurityPolicyVendor via ManagerContext) and downloads the script on first call.
func (i *ScriptStageProvider) ensureScriptReady(ctx context.Context) error {
	i.scriptsReadyMu.Lock()
	defer i.scriptsReadyMu.Unlock()
	if i.scriptsReady {
		return nil
	}

	if !strings.HasPrefix(i.Config.ScriptFolder, "http") {
		i.scriptsReady = true
		return nil
	}

	var allowedNets []*net.IPNet
	exclusiveMode := false
	if policy := i.Context.GetSecurityPolicy(); policy != nil {
		var err error
		allowedNets, err = parseIPRanges(policy.AllowedIPRanges)
		if err != nil {
			return v1alpha2.NewCOAError(err, "invalid allowedIPRanges in security policy", v1alpha2.BadConfig)
		}
		exclusiveMode = policy.AllowListExclusive
	}

	if err := validateScriptFolderURL(i.Config.ScriptFolder, allowedNets, exclusiveMode); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Stage): scriptFolder URL validation failed: %+v", err)
		return err
	}

	if err := downloadFile(i.Config.ScriptFolder, i.Config.Script, i.Config.StagingFolder); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Stage): failed to download script %s, error: %+v", i.Config.Script, err)
		return err
	}

	i.scriptsReady = true
	return nil
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

	if err = i.ensureScriptReady(ctx); err != nil {
		return nil, false, err
	}

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
