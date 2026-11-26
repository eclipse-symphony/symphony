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

func TestSetupStopsOnClusterDeploymentFailure(t *testing.T) {
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "mage"), []byte("#!/bin/sh\nexit 37\n"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(bin, "kubectl"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" >> \"$KUBECTL_TEST_ARGUMENTS\"\n"), 0700))
	output := filepath.Join(t.TempDir(), "arguments")
	t.Setenv("KUBECTL_TEST_ARGUMENTS", output)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := Setup()

	require.ErrorContains(t, err, "exit status 37")
	_, err = os.Stat(output)
	require.ErrorIs(t, err, os.ErrNotExist)
}
