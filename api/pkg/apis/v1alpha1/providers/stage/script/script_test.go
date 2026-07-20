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
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptInitWithMap(t *testing.T) {
	provider := ScriptStageProvider{}
	input := map[string]string{
		"name":          "test",
		"script":        "test.sh",
		"scriptEngine":  "bash",
		"scriptFolder":  "staging",
		"stagingFolder": "staging",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
}
func TestShellScript(t *testing.T) {
	provider := ScriptStageProvider{}
	err := provider.Init(ScriptStageProviderConfig{
		Name:          "test",
		Script:        "test.sh",
		ScriptEngine:  "bash",
		ScriptFolder:  "staging",
		StagingFolder: "staging",
	})
	assert.Nil(t, err)
	output, paused, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	assert.Nil(t, err)
	assert.False(t, paused)
	assert.Equal(t, "VALUE1", output["key1"])
	assert.Equal(t, "VALUE2", output["key2"])
}

func TestShellScriptOnline(t *testing.T) {
	provider := ScriptStageProvider{}
	err := provider.Init(ScriptStageProviderConfig{
		Name:          "test",
		Script:        "go1.21.6.src.tar.gz",
		ScriptEngine:  "gz",
		ScriptFolder:  "https://golang.google.cn/dl/",
		StagingFolder: "staging",
	})
	assert.Nil(t, err)
	_, err = os.Stat("staging/go1.21.6.src.tar.gz")
	assert.Nil(t, err)
	os.Remove("staging/go1.21.6.src.tar.gz")
}

func TestShellScriptNotFoundOnline(t *testing.T) {
	provider := ScriptStageProvider{}
	err := provider.Init(ScriptStageProviderConfig{
		Name:          "test",
		Script:        "test.ps1",
		ScriptEngine:  "powershell",
		ScriptFolder:  "https://bing.com",
		StagingFolder: "staging",
	})
	assert.NotNil(t, err)
	assert.IsType(t, v1alpha2.COAError{}, err)
}

// TestValidateScriptNameStage checks that path traversal and unsafe names are rejected.
func TestValidateScriptNameStage(t *testing.T) {
	cases := []struct {
		name      string
		script    string
		wantError bool
	}{
		{"plain filename", "script.sh", false},
		{"filename with dash", "my-script.sh", false},
		{"dot-dot traversal", "../../../etc/passwd", true},
		{"dot-dot encoded as %2e%2e", "%2e%2e/etc/passwd", true}, // after PathUnescape → "../etc/passwd"
		{"absolute path", "/etc/passwd", true},
		{"forward slash in name", "subdir/script.sh", true},
		{"backslash in name", "subdir\\script.sh", true},
		{"just dot-dot", "..", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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

// TestValidateScriptFolderURLStage checks URL scheme and IP validation.
// Public-IP cases use numeric IPs directly to avoid DNS lookups in restricted environments.
func TestValidateScriptFolderURLStage(t *testing.T) {
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
		{"loopback IPv4", "http://127.0.0.1/scripts", true},
		{"link-local IPv4 (IMDS)", "http://169.254.169.254/latest/meta-data/", true},
		{"RFC1918 10.x", "http://10.0.0.1/scripts", true},
		{"RFC1918 172.16.x", "http://172.16.0.1/scripts", true},
		{"RFC1918 192.168.x", "http://192.168.1.1/scripts", true},
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

// TestValidateScriptFolderURLStageWithWhitelist checks that allowedIPRanges can permit
// otherwise-blocked addresses, and that allowListExclusive enforces exclusive access.
func TestValidateScriptFolderURLStageWithWhitelist(t *testing.T) {
	_, allowedNet, _ := net.ParseCIDR("10.0.0.0/8")
	_, otherNet, _ := net.ParseCIDR("192.168.0.0/16")

	cases := []struct {
		name          string
		rawURL        string
		allowedNets   []*net.IPNet
		exclusiveMode bool
		wantError     bool
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
			name:          "public IP blocked in exclusive mode when not whitelisted",
			rawURL:        "http://8.8.8.8/scripts",
			allowedNets:   []*net.IPNet{otherNet},
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

// TestCheckIPAllowedStage verifies that loopback, link-local, and private IPs are rejected.
func TestCheckIPAllowedStage(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		wantAllow bool
	}{
		{"public IPv4", "8.8.8.8", true},
		{"loopback 127.0.0.1", "127.0.0.1", false},
		{"link-local 169.254.169.254", "169.254.169.254", false},
		{"RFC1918 10.0.0.1", "10.0.0.1", false},
		{"RFC1918 172.16.0.1", "172.16.0.1", false},
		{"RFC1918 192.168.0.1", "192.168.0.1", false},
		{"shared space 100.64.0.1", "100.64.0.1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip)
			err := checkIPAllowed(ip, tc.ip, nil, false)
			if tc.wantAllow {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestCheckIPAllowedStageWithWhitelist verifies whitelist and exclusive mode behavior.
func TestCheckIPAllowedStageWithWhitelist(t *testing.T) {
	_, privateNet, _ := net.ParseCIDR("10.0.0.0/8")

	cases := []struct {
		name          string
		ip            string
		allowedNets   []*net.IPNet
		exclusiveMode bool
		wantAllow     bool
	}{
		{
			name:          "private IP allowed when whitelisted",
			ip:            "10.0.0.1",
			allowedNets:   []*net.IPNet{privateNet},
			exclusiveMode: false,
			wantAllow:     true,
		},
		{
			name:          "public IP blocked in exclusive mode",
			ip:            "8.8.8.8",
			allowedNets:   []*net.IPNet{privateNet},
			exclusiveMode: true,
			wantAllow:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip)
			err := checkIPAllowed(ip, tc.ip, tc.allowedNets, tc.exclusiveMode)
			if tc.wantAllow {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
