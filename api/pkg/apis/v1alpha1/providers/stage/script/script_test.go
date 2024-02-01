/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
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
