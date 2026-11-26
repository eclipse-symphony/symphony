/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestSolutionVersionQueryMatchesManifest(t *testing.T) {
	repo, err := filepath.Abs(getRepoPath())
	require.NoError(t, err)
	manifest, err := os.Open(filepath.Join(repo, testSolutionVersion_WithEmptyComponentProperties))
	require.NoError(t, err)
	defer manifest.Close()
	decoder := yaml.NewYAMLOrJSONDecoder(manifest, 4096)
	name := ""
	for name == "" {
		var resource struct {
			Kind     string `json:"kind"`
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}
		require.NoError(t, decoder.Decode(&resource))
		if resource.Kind == "SolutionVersion" {
			name = resource.Metadata.Name
		}
	}

	bin := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = apply ]; then exit 0; fi\nif [ \"$1\" = get ] && [ \"$2\" = solutionversion ] && [ \"$3\" = \"$EXPECTED_SOLUTIONVERSION_NAME\" ]; then printf \"'1'\"; else echo 'queried resource does not match manifest' >&2; exit 1; fi\n"
	require.NoError(t, os.WriteFile(filepath.Join(bin, "kubectl"), []byte(script), 0700))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REPO_PATH", repo)
	t.Setenv("EXPECTED_SOLUTIONVERSION_NAME", name)
	executable, err := os.Executable()
	require.NoError(t, err)
	output, err := exec.Command(executable, "-test.run=^TestSolutionVersion_WithEmptyComponentProperties$", "-test.timeout=10s").CombinedOutput()
	require.NoError(t, err, "%s", output)
}
