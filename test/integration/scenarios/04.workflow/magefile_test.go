//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowCleansUpOnDeploymentFailure(t *testing.T) {
	root := t.TempDir()
	localenv := filepath.Join(root, "localenv")
	scenario := filepath.Join(root, "integration", "scenarios", "04.workflow")
	require.NoError(t, os.MkdirAll(localenv, 0700))
	require.NoError(t, os.MkdirAll(scenario, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(localenv, "symphony-ghcr-values.yaml"), []byte("api:\n  labelKey: \"\"\n  labelValue: \"\"\n  annotationKey: \"\"\n"), 0600))
	t.Chdir(scenario)
	t.Setenv("ENABLE_TLS_OTEL_SETUP", "false")
	t.Setenv("ENABLE_NON_TLS_OTEL_SETUP", "false")

	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "mage"), []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$MAGE_TEST_COMMANDS\"\nif [ \"$1\" = cluster:deploy ]; then exit 37; fi\n"), 0700))
	output := filepath.Join(root, "commands")
	t.Setenv("MAGE_TEST_COMMANDS", output)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	require.ErrorContains(t, Test(false), "exit status 37")
	commands, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, "cluster:deploy\ndumpSymphonyLogsForTest workflow test\ndestroy all,nowait\n", string(commands))
}
