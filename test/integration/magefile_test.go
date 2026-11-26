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

func TestRunTestScenarioArguments(t *testing.T) {
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "mage"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$MAGE_TEST_ARGUMENTS\"\n"), 0700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, scenario := range []struct {
		name string
		args string
	}{
		{"04.workflow", "test\nfalse\n"},
		{"02.basic", "test\n"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), scenario.name)
			require.NoError(t, os.Mkdir(dir, 0700))
			output := filepath.Join(t.TempDir(), "arguments")
			t.Setenv("MAGE_TEST_ARGUMENTS", output)
			require.NoError(t, RunTest(dir))
			args, err := os.ReadFile(output)
			require.NoError(t, err)
			require.Equal(t, scenario.args, string(args))
		})
	}
}
