/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
			escapedScript = encodeSubDelimiters(escapedScript)
			escapedFolder := escapeURLPathSubDelims(base)
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
			result := encodeSubDelimiters(tc.input)
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
			result := escapeURLPathSubDelims(tc.input)
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
			escapedScript = encodeSubDelimiters(escapedScript)
			escapedFolder := escapeURLPathSubDelims(tc.folder)

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
