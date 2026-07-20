/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

// Package scriptutils provides shared HTTP-download and URL-validation helpers used by both
// the target/script and stage/script providers. Centralising these utilities avoids code
// duplication and ensures consistent SSRF/path-traversal protections across both providers.
package scriptutils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("providers.scriptutils")

// PrivateIPNets lists IP ranges that are not routable on the public internet.
// Requests to these ranges are blocked to prevent SSRF attacks.
var PrivateIPNets = func() []*net.IPNet {
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

// ScriptDownloadClient is a dedicated HTTP client used for downloading scripts.
// It enforces a request timeout and re-validates redirect destinations to prevent SSRF.
// The redirect handler uses the deny-list only (no per-provider whitelist) since this
// client is package-level and cannot carry per-instance state.
var ScriptDownloadClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("maximum redirect limit reached")
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return fmt.Errorf("redirect to non-HTTP(S) scheme %q is not permitted", req.URL.Scheme)
		}
		if err := ValidateURLHost(req.URL.Hostname(), nil, false); err != nil {
			return fmt.Errorf("redirect blocked: %w", err)
		}
		return nil
	},
}

// IsRemoteURL returns true when rawURL is an http:// or https:// URL that must be
// downloaded before use. A local filesystem path or empty string returns false.
func IsRemoteURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// ParseIPRanges converts a slice of CIDR strings or plain IP addresses to []*net.IPNet.
// Plain IP addresses are treated as host routes (/32 for IPv4, /128 for IPv6).
func ParseIPRanges(ranges []string) ([]*net.IPNet, error) {
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

// ValidateScriptName ensures the script name is a safe, local filename.
// It rejects path traversal sequences ("..", absolute paths) and directory separators.
func ValidateScriptName(name string) error {
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

// ValidateScriptFolderURL validates that the scriptFolder URL is safe to use.
// It enforces http/https schemes and rejects URLs that resolve to private or loopback addresses,
// unless the address falls within one of the allowedNets ranges.
// When exclusiveMode is true, only IPs explicitly listed in allowedNets are permitted.
func ValidateScriptFolderURL(rawURL string, allowedNets []*net.IPNet, exclusiveMode bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return v1alpha2.NewCOAError(err, "invalid scriptFolder URL", v1alpha2.BadConfig)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return v1alpha2.NewCOAError(nil,
			fmt.Sprintf("invalid URL scheme %q: only http and https are permitted", u.Scheme),
			v1alpha2.BadConfig)
	}
	return ValidateURLHost(u.Hostname(), allowedNets, exclusiveMode)
}

// ValidateURLHost resolves host and rejects loopback, link-local, and private addresses,
// unless the address falls within one of the allowedNets ranges.
// When exclusiveMode is true, only IPs explicitly listed in allowedNets are permitted.
func ValidateURLHost(host string, allowedNets []*net.IPNet, exclusiveMode bool) error {
	if ip := net.ParseIP(host); ip != nil {
		return CheckIPAllowed(ip, host, allowedNets, exclusiveMode)
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
		if err := CheckIPAllowed(ip, host, allowedNets, exclusiveMode); err != nil {
			return err
		}
	}
	return nil
}

// CheckIPAllowed returns an error if the IP is blocked by the current policy.
// IPs in allowedNets are always permitted (they override the deny list).
// When exclusiveMode is true, only IPs in allowedNets are permitted; all others are rejected.
func CheckIPAllowed(ip net.IP, host string, allowedNets []*net.IPNet, exclusiveMode bool) error {
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
	for _, n := range PrivateIPNets {
		if n.Contains(ip) {
			return v1alpha2.NewCOAError(nil,
				fmt.Sprintf("host %q resolves to a private address which is not permitted", host),
				v1alpha2.BadConfig)
		}
	}
	return nil
}

// DownloadFile fetches a single script file from scriptFolder/script and writes it to
// stagingFolder. The script name is validated against path-traversal attacks before use.
func DownloadFile(scriptFolder string, script string, stagingFolder string) error {
	sLog.Debugf("  DownloadFile: scriptFolder=%q, script=%q, stagingFolder=%q", scriptFolder, script, stagingFolder)

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
	if err := ValidateScriptName(rawScript); err != nil {
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

	// 5. Use the raw, clean string for the local file system path.
	//    '$' and '%' are valid filesystem characters; no escaping needed.
	tPath := filepath.Join(stagingFolder, rawScript)
	sLog.Debugf("  DownloadFile: resolved URL=%q, localPath=%q", sPath, tPath)

	// 6. Fetch the script using the dedicated client (enforces timeout and redirect policy).
	resp, err := ScriptDownloadClient.Get(sPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 7. Only write the file on a successful HTTP response.
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
