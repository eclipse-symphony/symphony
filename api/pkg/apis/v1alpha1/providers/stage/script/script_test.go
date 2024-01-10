/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package script

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

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
