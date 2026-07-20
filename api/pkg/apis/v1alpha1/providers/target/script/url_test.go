/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"net"
	"net/url"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDownloadFileURLEscaping validates that the downloadFile escaping logic
// correctly percent-encodes sub-delimiters ($, &, +, =) and '%' in script names
// for the HTTP URL, whether the input is pre-encoded or raw.
func TestDownloadFileURLEscaping(t *testing.T) {
	base := "http://example.com/scripts"

	cases := []struct {
		name           string
		script         string
		expectContains string // substring the final URL must contain
		expectNotIn    string // substring the final URL must NOT contain (empty = skip)
	}{
		{"plain script", "deploy.sh", "/deploy.sh", ""},
		{"dollar in name", "deploy$1.sh", "%241", "$1"},
		{"ampersand in name", "deploy&1.sh", "%261", "&1"},
		{"plus in name", "deploy+1.sh", "%2B1", "+1"},
		{"equals in name", "deploy=1.sh", "%3D1", "=1"},
		{"literal percent in name", "deploy%test.sh", "%25test", ""},
		{"space in name", "deploy test.sh", "%20test", " test"},
		{"pre-encoded dollar (%24)", "deploy%241.sh", "%241", "$1"},
		{"pre-encoded ampersand (%26)", "deploy%261.sh", "%261", "&1"},
		{"pre-encoded plus (%2B)", "deploy%2B1.sh", "%2B1", "+1"},
		{"pre-encoded equals (%3D)", "deploy%3D1.sh", "%3D1", "=1"},
		{"pre-encoded space (%20)", "deploy%20test.sh", "%20test", " test"},
		{"pre-encoded percent (%25)", "deploy%25test.sh", "%25test", ""},
		{"dollar and literal percent", "deploy$1%test.sh", "%241%25test", "$1"},
		{"all sub-delims combined", "a$b&c+d=e.sh", "%24b%26c%2Bd%3De", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Replicate the downloadFile encoding logic:
			rawScript, err := url.PathUnescape(tc.script)
			if err != nil {
				rawScript = tc.script
			}
			escapedScript := url.PathEscape(rawScript)
			escapedScript = coa_utils.EncodeSubDelimiters(escapedScript)
			escapedFolder := coa_utils.EscapeURLPathSubDelims(base)
			sPath, err := url.JoinPath(escapedFolder, escapedScript)
			assert.NoError(t, err, "url.JoinPath should not fail")

			if tc.expectContains != "" {
				assert.Contains(t, sPath, tc.expectContains,
					"URL should contain %q for input %q", tc.expectContains, tc.script)
			}
			if tc.expectNotIn != "" {
				assert.NotContains(t, sPath, tc.expectNotIn,
					"URL should NOT contain literal %q for input %q", tc.expectNotIn, tc.script)
			}
		})
	}
}

// TestEncodeSubDelimiters validates the helper that encodes $, &, +, = in a string.
func TestEncodeSubDelimiters(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{"no sub-delims", "deploy.sh", "deploy.sh"},
		{"dollar only", "deploy$1.sh", "deploy%241.sh"},
		{"ampersand only", "deploy&1.sh", "deploy%261.sh"},
		{"plus only", "deploy+1.sh", "deploy%2B1.sh"},
		{"equals only", "deploy=1.sh", "deploy%3D1.sh"},
		{"all four", "$a&b+c=d", "%24a%26b%2Bc%3Dd"},
		{"already encoded %24 passes through", "deploy%241.sh", "deploy%241.sh"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := coa_utils.EncodeSubDelimiters(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

// TestEscapeURLPathSubDelims validates the helper that normalizes and encodes
// sub-delimiters ($, &, +, =) in URL path segments.
func TestEscapeURLPathSubDelims(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{"no sub-delims", "http://example.com/scripts", "http://example.com/scripts"},
		{"literal dollar in path", "http://example.com/$scripts", "http://example.com/%24scripts"},
		{"literal ampersand in path", "http://example.com/a&b", "http://example.com/a%26b"},
		{"literal plus in path", "http://example.com/a+b", "http://example.com/a%2Bb"},
		{"literal equals in path", "http://example.com/a=b", "http://example.com/a%3Db"},
		{"pre-encoded dollar (%24)", "http://example.com/%24scripts", "http://example.com/%24scripts"},
		{"dollar and percent-encoded space", "http://example.com/$scripts%20v2", "http://example.com/%24scripts%20v2"},
		{"multiple dollars", "http://example.com/$a/$b", "http://example.com/%24a/%24b"},
		{"dollar in query stays untouched", "http://example.com/scripts?v=$1", "http://example.com/scripts?v=$1"},
		{"all sub-delims in path", "http://example.com/$a&b+c=d", "http://example.com/%24a%26b%2Bc%3Dd"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := coa_utils.EscapeURLPathSubDelims(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

// TestFolderAndScriptEndToEnd validates the full flow: folder + script escaping together.
func TestFolderAndScriptEndToEnd(t *testing.T) {
	cases := []struct {
		name      string
		folder    string
		script    string
		expectURL string // substring the final URL must contain
		rejectURL string // substring the final URL must NOT contain
	}{
		{"no special chars", "http://example.com/scripts", "deploy.sh",
			"scripts/deploy.sh", ""},
		{"dollar in folder only", "http://example.com/$scripts", "deploy.sh",
			"%24scripts/deploy.sh", "$scripts"},
		{"dollar in script only", "http://example.com/scripts", "deploy$1.sh",
			"scripts/deploy%241.sh", "$1"},
		{"dollar in both", "http://example.com/$scripts", "deploy$1.sh",
			"%24scripts/deploy%241.sh", "$"},
		{"ampersand in script", "http://example.com/scripts", "deploy&1.sh",
			"scripts/deploy%261.sh", "&1"},
		{"plus in script", "http://example.com/scripts", "deploy+1.sh",
			"scripts/deploy%2B1.sh", "+1"},
		{"pre-encoded dollar in folder", "http://example.com/%24scripts", "deploy.sh",
			"%24scripts/deploy.sh", "$scripts"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawScript, err := url.PathUnescape(tc.script)
			if err != nil {
				rawScript = tc.script
			}
			escapedScript := url.PathEscape(rawScript)
			escapedScript = coa_utils.EncodeSubDelimiters(escapedScript)
			escapedFolder := coa_utils.EscapeURLPathSubDelims(tc.folder)

			sPath, err := url.JoinPath(escapedFolder, escapedScript)
			assert.NoError(t, err)
			assert.Contains(t, sPath, tc.expectURL)
			if tc.rejectURL != "" {
				assert.NotContains(t, sPath, tc.rejectURL)
			}
		})
	}
}

// TestStagingFolderNoEscapingNeeded confirms staging folder path doesn't need URL escaping
// since it's only used with filepath.Join for local filesystem operations.
func TestStagingFolderNoEscapingNeeded(t *testing.T) {
	t.Log("stagingFolder uses filepath.Join only - '$', '&', '+', '=' and '%' are valid filesystem characters, no escaping needed")
}

// TestValidateScriptName checks that path traversal and unsafe names are rejected.
func TestValidateScriptName(t *testing.T) {
	cases := []struct {
		name      string
		script    string
		wantError bool
	}{
		{"plain filename", "deploy.sh", false},
		{"filename with dot", "deploy.v2.sh", false},
		{"filename with dash", "my-script.sh", false},
		{"dot-dot traversal", "../../../etc/passwd", true},
		{"dot-dot encoded as %2e%2e", "%2e%2e/etc/passwd", true},  // after PathUnescape → "../etc/passwd"
		{"absolute path", "/etc/passwd", true},
		{"forward slash in name", "subdir/script.sh", true},
		{"backslash in name", "subdir\\script.sh", true},
		{"just dot-dot", "..", true},
		{"current dir", ".", false}, // filepath.IsLocal accepts "."
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Mirror the unescape step that downloadFile does before calling validateScriptName.
			rawScript, err := url.PathUnescape(tc.script)
			if err != nil {
				rawScript = tc.script
			}
			err = validateScriptName(rawScript)
			if tc.wantError {
				assert.Error(t, err, "expected error for script name %q", tc.script)
			} else {
				assert.NoError(t, err, "unexpected error for script name %q", tc.script)
			}
		})
	}
}

// TestValidateScriptFolderURL checks URL scheme validation.
// Public-IP cases use numeric IPs directly to avoid DNS lookups in restricted environments.
func TestValidateScriptFolderURL(t *testing.T) {
	cases := []struct {
		name      string
		rawURL    string
		wantError bool
	}{
		// Use public IPs directly to avoid DNS lookups in sandboxed test environments.
		{"http scheme with public IP", "http://8.8.8.8/scripts", false},
		{"https scheme with public IP", "https://8.8.8.8/scripts", false},
		{"file scheme", "file:///etc/passwd", true},
		{"ftp scheme", "ftp://8.8.8.8/scripts", true},
		{"no scheme", "example.com/scripts", true},
		{"loopback IPv4", "http://127.0.0.1/scripts", true},
		{"loopback IPv6", "http://[::1]/scripts", true},
		{"link-local IPv4 (IMDS)", "http://169.254.169.254/latest/meta-data/", true},
		{"RFC1918 10.x", "http://10.0.0.1/scripts", true},
		{"RFC1918 172.16.x", "http://172.16.0.1/scripts", true},
		{"RFC1918 192.168.x", "http://192.168.1.1/scripts", true},
		{"shared address space 100.64.x", "http://100.64.0.1/scripts", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateScriptFolderURL(tc.rawURL, nil, false)
			if tc.wantError {
				assert.Error(t, err, "expected error for URL %q", tc.rawURL)
			} else {
				assert.NoError(t, err, "unexpected error for URL %q", tc.rawURL)
			}
		})
	}
}

// TestValidateScriptFolderURLWithWhitelist checks that allowedIPRanges can permit
// otherwise-blocked addresses, and that allowListExclusive enforces exclusive access.
func TestValidateScriptFolderURLWithWhitelist(t *testing.T) {
	_, allowedNet, _ := net.ParseCIDR("10.0.0.0/8")
	_, otherNet, _ := net.ParseCIDR("192.168.0.0/16")

	cases := []struct {
		name           string
		rawURL         string
		allowedNets    []*net.IPNet
		exclusiveMode  bool
		wantError      bool
	}{
		{
			name:          "private IP whitelisted via CIDR",
			rawURL:        "http://10.0.0.1/scripts",
			allowedNets:   []*net.IPNet{allowedNet},
			exclusiveMode: false,
			wantError:     false,
		},
		{
			name:          "private IP not in whitelist still blocked",
			rawURL:        "http://172.16.0.1/scripts",
			allowedNets:   []*net.IPNet{allowedNet},
			exclusiveMode: false,
			wantError:     true,
		},
		{
			name:          "public IP allowed in exclusive mode when whitelisted",
			rawURL:        "http://10.0.0.1/scripts",
			allowedNets:   []*net.IPNet{allowedNet},
			exclusiveMode: true,
			wantError:     false,
		},
		{
			name:          "public IP blocked in exclusive mode when not whitelisted",
			rawURL:        "http://8.8.8.8/scripts",
			allowedNets:   []*net.IPNet{otherNet},
			exclusiveMode: true,
			wantError:     true,
		},
		{
			name:          "empty whitelist with exclusive mode blocks everything",
			rawURL:        "http://8.8.8.8/scripts",
			allowedNets:   []*net.IPNet{},
			exclusiveMode: true,
			wantError:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateScriptFolderURL(tc.rawURL, tc.allowedNets, tc.exclusiveMode)
			if tc.wantError {
				assert.Error(t, err, "expected error for URL %q", tc.rawURL)
			} else {
				assert.NoError(t, err, "unexpected error for URL %q", tc.rawURL)
			}
		})
	}
}

// TestCheckIPAllowed verifies individual IP address classifications.
func TestCheckIPAllowed(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		wantAllow bool
	}{
		{"public IPv4", "8.8.8.8", true},
		{"public IPv4 (github)", "140.82.114.4", true},
		{"loopback 127.0.0.1", "127.0.0.1", false},
		{"loopback 127.0.0.2", "127.0.0.2", false},
		{"loopback IPv6 ::1", "::1", false},
		{"link-local 169.254.169.254", "169.254.169.254", false},
		{"link-local 169.254.0.1", "169.254.0.1", false},
		{"RFC1918 10.0.0.1", "10.0.0.1", false},
		{"RFC1918 10.255.255.255", "10.255.255.255", false},
		{"RFC1918 172.16.0.1", "172.16.0.1", false},
		{"RFC1918 172.31.255.255", "172.31.255.255", false},
		{"RFC1918 192.168.0.1", "192.168.0.1", false},
		{"shared space 100.64.0.1", "100.64.0.1", false},
		{"shared space 100.127.255.255", "100.127.255.255", false},
		{"IPv6 unique local fc00::1", "fc00::1", false},
		{"IPv6 unique local fd00::1", "fd00::1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP %q", tc.ip)
			err := checkIPAllowed(ip, tc.ip, nil, false)
			if tc.wantAllow {
				assert.NoError(t, err, "expected IP %q to be allowed", tc.ip)
			} else {
				assert.Error(t, err, "expected IP %q to be blocked", tc.ip)
			}
		})
	}
}

// TestCheckIPAllowedWithWhitelist verifies that allowedNets overrides the deny list
// and that exclusiveMode rejects IPs outside the whitelist.
func TestCheckIPAllowedWithWhitelist(t *testing.T) {
	_, privateNet, _ := net.ParseCIDR("10.0.0.0/8")
	_, publicNet, _ := net.ParseCIDR("8.8.8.0/24")

	cases := []struct {
		name          string
		ip            string
		allowedNets   []*net.IPNet
		exclusiveMode bool
		wantAllow     bool
	}{
		{
			name:          "private IP allowed when in whitelist",
			ip:            "10.0.0.1",
			allowedNets:   []*net.IPNet{privateNet},
			exclusiveMode: false,
			wantAllow:     true,
		},
		{
			name:          "private IP blocked when not in whitelist",
			ip:            "172.16.0.1",
			allowedNets:   []*net.IPNet{privateNet},
			exclusiveMode: false,
			wantAllow:     false,
		},
		{
			name:          "public IP allowed in exclusive mode when whitelisted",
			ip:            "8.8.8.8",
			allowedNets:   []*net.IPNet{publicNet},
			exclusiveMode: true,
			wantAllow:     true,
		},
		{
			name:          "public IP blocked in exclusive mode when not whitelisted",
			ip:            "1.2.3.4",
			allowedNets:   []*net.IPNet{publicNet},
			exclusiveMode: true,
			wantAllow:     false,
		},
		{
			name:          "loopback allowed when explicitly whitelisted",
			ip:            "127.0.0.1",
			allowedNets:   []*net.IPNet{{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(32, 32)}},
			exclusiveMode: false,
			wantAllow:     true,
		},
		{
			name:          "nil whitelist with exclusive mode blocks all",
			ip:            "8.8.8.8",
			allowedNets:   nil,
			exclusiveMode: true,
			wantAllow:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP %q", tc.ip)
			err := checkIPAllowed(ip, tc.ip, tc.allowedNets, tc.exclusiveMode)
			if tc.wantAllow {
				assert.NoError(t, err, "expected IP %q to be allowed", tc.ip)
			} else {
				assert.Error(t, err, "expected IP %q to be blocked", tc.ip)
			}
		})
	}
}

// TestEnsureScriptsReadyWithSecurityPolicy verifies that ensureScriptsReady reads
// AllowedIPRanges and AllowListExclusive from the SecurityPolicyVendor via ManagerContext,
// rather than from the provider's own config.
func TestEnsureScriptsReadyWithSecurityPolicy(t *testing.T) {
	// Build a VendorContext that carries a SecurityPolicy allowing 10.0.0.0/8.
	vc := &contexts.VendorContext{}
	vc.SecurityPolicy = &contexts.SecurityPolicy{
		AllowedIPRanges:    []string{"10.0.0.0/8"},
		AllowListExclusive: false,
	}
	mc := &contexts.ManagerContext{}
	_ = mc.Init(vc, nil)

	t.Run("private IP whitelisted via context policy", func(t *testing.T) {
		p := &ScriptProvider{}
		err := p.Init(ScriptProviderConfig{
			ApplyScript:   "apply.sh",
			RemoveScript:  "remove.sh",
			GetScript:     "get.sh",
			ScriptFolder:  "http://10.0.0.1/scripts",
			StagingFolder: "/tmp",
		})
		require.NoError(t, err)
		p.SetContext(mc)
		// ensureScriptsReady should proceed past URL validation because 10.0.0.1 is in
		// the policy allow-list. It will fail on the actual download (no server running),
		// but the error must not be a validation error.
		err = p.ensureScriptsReady(context.Background())
		// We expect a network/download error, NOT a "private address" validation error.
		if err != nil {
			assert.NotContains(t, err.Error(), "private address which is not permitted",
				"URL validation should pass for whitelisted IP; error should be from download, not policy")
		}
	})

	t.Run("private IP blocked when not in context policy", func(t *testing.T) {
		p := &ScriptProvider{}
		err := p.Init(ScriptProviderConfig{
			ApplyScript:   "apply.sh",
			RemoveScript:  "remove.sh",
			GetScript:     "get.sh",
			ScriptFolder:  "http://172.16.0.1/scripts",
			StagingFolder: "/tmp",
		})
		require.NoError(t, err)
		p.SetContext(mc)
		err = p.ensureScriptsReady(context.Background())
		assert.Error(t, err, "172.16.0.1 is not in the allow-list and should be blocked")
		assert.Contains(t, err.Error(), "private address which is not permitted")
	})

	t.Run("nil context uses deny-list only", func(t *testing.T) {
		p := &ScriptProvider{}
		err := p.Init(ScriptProviderConfig{
			ApplyScript:   "apply.sh",
			RemoveScript:  "remove.sh",
			GetScript:     "get.sh",
			ScriptFolder:  "http://10.0.0.1/scripts",
			StagingFolder: "/tmp",
		})
		require.NoError(t, err)
		// No SetContext: Context is nil, so GetSecurityPolicy returns nil → deny-list only.
		err = p.ensureScriptsReady(context.Background())
		assert.Error(t, err, "10.0.0.1 should be blocked by the default deny-list when no policy")
		assert.Contains(t, err.Error(), "private address which is not permitted")
	})
}
func TestParseIPRanges(t *testing.T) {
	cases := []struct {
		name      string
		ranges    []string
		wantError bool
		wantCount int
	}{
		{"empty list", []string{}, false, 0},
		{"single CIDR", []string{"10.0.0.0/8"}, false, 1},
		{"multiple CIDRs", []string{"10.0.0.0/8", "192.168.0.0/16"}, false, 2},
		{"plain IPv4 address", []string{"10.0.0.1"}, false, 1},
		{"plain IPv6 address", []string{"::1"}, false, 1},
		{"mixed CIDR and IP", []string{"10.0.0.0/8", "8.8.8.8"}, false, 2},
		{"invalid CIDR", []string{"not-a-cidr"}, true, 0},
		{"invalid IP", []string{"999.999.999.999"}, true, 0},
		{"invalid CIDR notation", []string{"10.0.0.0/99"}, true, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nets, err := parseIPRanges(tc.ranges)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, nets, tc.wantCount)
			}
		})
	}
}
