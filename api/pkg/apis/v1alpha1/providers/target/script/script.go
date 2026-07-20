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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/google/uuid"
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
	Config         ScriptProviderConfig
	Context        *contexts.ManagerContext
	scriptsReady   bool
	scriptsReadyMu sync.Mutex
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

// ensureScriptsReady validates the scriptFolder URL against the server-side SecurityPolicy
// (obtained from ManagerContext, which is populated by the SecurityPolicyVendor) and
// downloads all scripts into the staging folder on first call. Subsequent calls are no-ops.
// This is called lazily from Apply/Get so that the SecurityPolicy is always available.
func (i *ScriptProvider) ensureScriptsReady(ctx context.Context) error {
	i.scriptsReadyMu.Lock()
	defer i.scriptsReadyMu.Unlock()
	if i.scriptsReady {
		return nil
	}

	if !strings.HasPrefix(i.Config.ScriptFolder, "http") {
		i.scriptsReady = true
		return nil
	}

	// Obtain allow-list policy from the SecurityPolicyVendor via ManagerContext.
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
		sLog.ErrorfCtx(ctx, "  P (Script Target): scriptFolder URL validation failed: %+v", err)
		return err
	}

	if err := downloadFile(i.Config.ScriptFolder, i.Config.ApplyScript, i.Config.StagingFolder); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download apply script %s, error: %+v", i.Config.ApplyScript, err)
		return err
	}
	if err := downloadFile(i.Config.ScriptFolder, i.Config.RemoveScript, i.Config.StagingFolder); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download remove script %s, error: %+v", i.Config.RemoveScript, err)
		return err
	}
	if err := downloadFile(i.Config.ScriptFolder, i.Config.GetScript, i.Config.StagingFolder); err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to download get script %s, error: %+v", i.Config.GetScript, err)
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
	escapedScript = coa_utils.EncodeSubDelimiters(escapedScript)

	// 4. Normalize and encode sub-delimiters in the scriptFolder URL path.
	escapedFolder := coa_utils.EscapeURLPathSubDelims(scriptFolder)

	sPath, err := url.JoinPath(escapedFolder, escapedScript)
	if err != nil {
		return err
	}

	tPath := filepath.Join(stagingFolder, rawScript)
	sLog.Debugf("  downloadFile: resolved URL=%q, localPath=%q", sPath, tPath)

	// 5. Fetch the script using the dedicated client (enforces timeout and redirect policy).
	resp, err := scriptDownloadClient.Get(sPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 6. Only write the file on a successful HTTP response.
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
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

	if err = i.ensureScriptsReady(ctx); err != nil {
		return nil, err
	}

	id := uuid.New().String()
	input := id + ".json"
	input_ref := id + "-ref.json"
	output := id + "-get-output.json"

	staging := filepath.Join(i.Config.StagingFolder, input)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = os.WriteFile(staging, file, 0644)

	staging_ref := filepath.Join(i.Config.StagingFolder, input_ref)
	file_ref, _ := json.MarshalIndent(references, "", " ")
	_ = os.WriteFile(staging_ref, file_ref, 0644)

	abs, _ := filepath.Abs(staging)
	abs_ref, _ := filepath.Abs(staging_ref)

	defer os.Remove(abs)
	defer os.Remove(abs_ref)

	scriptAbs, _ := filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.GetScript))
	if strings.HasPrefix(i.Config.ScriptFolder, "http") {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.GetScript))
	}

	o, err := i.runCommand(scriptAbs, abs, abs_ref)
	sLog.DebugfCtx(ctx, "  P (Script Target): get script output: %s", string(o))

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to run get script: %+v", err)
		return nil, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	data, err := os.ReadFile(outputStaging)

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to read output file: %+v", err)
		return nil, err
	}

	abs_output, _ := filepath.Abs(outputStaging)

	defer os.Remove(abs_output)

	ret := make([]model.ComponentSpec, 0)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to parse get script output (expected []ComponentSpec): %+v", err)
		return nil, err
	}
	return ret, nil
}
func (i *ScriptProvider) runScriptOnComponents(ctx context.Context, deployment model.DeploymentSpec, components []model.ComponentSpec, isRemove bool) (map[string]model.ComponentResultSpec, error) {
	id := uuid.New().String()
	deploymentId := id + ".json"
	currenRefId := id + "-ref.json"
	output := id + "-output.json"

	stagingDeployment := filepath.Join(i.Config.StagingFolder, deploymentId)
	file, _ := json.MarshalIndent(deployment, "", " ")
	_ = os.WriteFile(stagingDeployment, file, 0644)

	stagingRef := filepath.Join(i.Config.StagingFolder, currenRefId)
	file, _ = json.MarshalIndent(components, "", " ")
	_ = os.WriteFile(stagingRef, file, 0644)

	absDeployment, _ := filepath.Abs(stagingDeployment)
	absRef, _ := filepath.Abs(stagingRef)

	var scriptAbs = ""
	if isRemove {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.RemoveScript))
		observ_utils.EmitUserAuditsLogs(ctx, "  P (Script Target): Start to run remove script - %s", i.Config.RemoveScript)
		if strings.HasPrefix(i.Config.ScriptFolder, "http") {
			scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.RemoveScript))
		}
	} else {
		scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.ScriptFolder, i.Config.ApplyScript))
		observ_utils.EmitUserAuditsLogs(ctx, "  P (Script Target): Start to run apply script - %s", i.Config.ApplyScript)
		if strings.HasPrefix(i.Config.ScriptFolder, "http") {
			scriptAbs, _ = filepath.Abs(filepath.Join(i.Config.StagingFolder, i.Config.ApplyScript))
		}
	}
	o, err := i.runCommand(scriptAbs, absDeployment, absRef)
	sLog.DebugfCtx(ctx, "  P (Script Target): apply script output: %s", o)

	defer os.Remove(absDeployment)
	defer os.Remove(absRef)

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to run apply script: %+v", err)
		return nil, err
	}

	outputStaging := filepath.Join(i.Config.StagingFolder, output)

	data, err := os.ReadFile(outputStaging)

	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to parse apply script output (expected map[string]model.ComponentResultSpec): %+v", err)
		return nil, err
	}

	abs_output, _ := filepath.Abs(outputStaging)

	defer os.Remove(abs_output)

	ret := make(map[string]model.ComponentResultSpec)
	err = json.Unmarshal(data, &ret)
	if err != nil {
		sLog.ErrorfCtx(ctx, "  P (Script Target): failed to parse get script output (expected map[string]model.ComponentResultSpec): %+v", err)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	sLog.InfofCtx(ctx, "  P (Script Target): applying artifacts: %s - %s", deployment.Instance.Spec.Scope, deployment.Instance.ObjectMeta.Name)

	if err = i.ensureScriptsReady(ctx); err != nil {
		return nil, err
	}

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
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Script Target): get updated components: count - %d", len(components))
		var retU map[string]model.ComponentResultSpec
		retU, err = i.runScriptOnComponents(ctx, deployment, components, false)
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
		for k, v := range retU {
			ret[k] = v
		}
	}

	components = step.GetDeletedComponents()
	if len(components) > 0 {
		sLog.InfofCtx(ctx, "  P (Script Target): get deleted components: count - %d", len(components))
		var retU map[string]model.ComponentResultSpec
		retU, err = i.runScriptOnComponents(ctx, deployment, components, true)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Script Target): failed to run remove script: %+v", err)
			providerOperationMetrics.ProviderOperationErrors(
				script,
				functionName,
				metrics.ApplyScriptOperation,
				metrics.ApplyOperationType,
				v1alpha2.RemoveScriptFailed.String(),
			)
			return nil, err
		}
		for k, v := range retU {
			ret[k] = v
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
