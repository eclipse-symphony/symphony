/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	// Script download is now lazy: happens on first ensureScriptReady call, not in Init.
	err = provider.ensureScriptReady(context.Background())
	assert.Nil(t, err)
	_, err = os.Stat("staging/go1.21.6.src.tar.gz")
	assert.Nil(t, err)
	os.Remove("staging/go1.21.6.src.tar.gz")
}

func TestShellScriptNotFoundOnline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	vc := &contexts.VendorContext{
		SecurityPolicy: &contexts.SecurityPolicy{
			AllowedIPRanges: []string{"127.0.0.1/32"},
		},
	}
	mc := &contexts.ManagerContext{}
	require.NoError(t, mc.Init(vc, nil))

	provider := ScriptStageProvider{}
	require.NoError(t, provider.Init(ScriptStageProviderConfig{
		Name:          "test",
		Script:        "test.ps1",
		ScriptEngine:  "powershell",
		ScriptFolder:  server.URL,
		StagingFolder: t.TempDir(),
	}))
	provider.SetContext(mc)

	err := provider.ensureScriptReady(context.Background())

	require.Error(t, err)
	var coaErr v1alpha2.COAError
	require.ErrorAs(t, err, &coaErr)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
}

// TestShellScriptNotFoundOnline is the last test in this file that references provider behaviour.
// URL-validation, IP-allow-list, ParseIPRanges, and DownloadFile tests live in
// api/pkg/apis/v1alpha1/providers/scriptutils/scriptutils_test.go.

// TestEnsureScriptReadyWithSecurityPolicy verifies that ensureScriptReady reads
// the SecurityPolicy from ManagerContext (populated by SecurityPolicyVendor)
// rather than from the provider's own config.
func TestEnsureScriptReadyWithSecurityPolicy(t *testing.T) {
	vc := &contexts.VendorContext{}
	vc.SecurityPolicy = &contexts.SecurityPolicy{
		AllowedIPRanges:    []string{"10.0.0.0/8"},
		AllowListExclusive: false,
	}
	mc := &contexts.ManagerContext{}
	_ = mc.Init(vc, nil)

	t.Run("private IP whitelisted via context policy", func(t *testing.T) {
		p := &ScriptStageProvider{}
		err := p.Init(ScriptStageProviderConfig{
			Script:        "run.sh",
			ScriptFolder:  "http://10.0.0.1/scripts",
			StagingFolder: "/tmp",
		})
		require.NoError(t, err)
		p.SetContext(mc)
		err = p.ensureScriptReady(context.Background())
		if err != nil {
			assert.NotContains(t, err.Error(), "private address which is not permitted",
				"URL validation should pass for whitelisted IP")
		}
	})

	t.Run("nil context uses deny-list only", func(t *testing.T) {
		p := &ScriptStageProvider{}
		err := p.Init(ScriptStageProviderConfig{
			Script:        "run.sh",
			ScriptFolder:  "http://10.0.0.1/scripts",
			StagingFolder: "/tmp",
		})
		require.NoError(t, err)
		err = p.ensureScriptReady(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "private address which is not permitted")
	})
}
