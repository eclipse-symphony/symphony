//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderTestEnvironment(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is required to flatten the test kubeconfig")
	}
	sourceDir := t.TempDir()
	for _, name := range []string{"ca", "cert", "key"} {
		require.NoError(t, os.WriteFile(filepath.Join(sourceDir, name), []byte("test-"+name), 0600))
	}
	source := []byte(fmt.Sprintf(`{
		"apiVersion":"v1","kind":"Config",
		"current-context":"test",
		"contexts":[{"name":"test","context":{"cluster":"test","user":"test"}}],
		"clusters":[{"name":"test","cluster":{"server":"https://127.0.0.1:6443","certificate-authority":%q}},
		            {"name":"unused","cluster":{"server":"https://unused.invalid","certificate-authority":"must-not-be-read"}}],
		"users":[{"name":"test","user":{"client-certificate":%q,"client-key":%q}},
		         {"name":"unused","user":{"tokenFile":"must-not-be-read"}}]
	}`, filepath.Join(sourceDir, "ca"), filepath.Join(sourceDir, "cert"), filepath.Join(sourceDir, "key")))
	sourcePath := filepath.Join(sourceDir, "config")
	require.NoError(t, os.WriteFile(sourcePath, source, 0600))
	t.Setenv("KUBECONFIG", sourcePath)
	originalHome := os.Getenv("HOME")
	home := t.TempDir()
	env, err := providerTestEnvironment(home)
	require.NoError(t, err)
	require.Equal(t, originalHome, os.Getenv("HOME"))
	require.Equal(t, sourcePath, os.Getenv("KUBECONFIG"))
	unchanged, err := os.ReadFile(sourcePath)
	require.NoError(t, err)
	require.Equal(t, source, unchanged)

	path := filepath.Join(home, ".kube", "config")
	require.Contains(t, env, "HOME="+home)
	require.Contains(t, env, "KUBECONFIG="+path)
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var config struct {
		Clusters []struct {
			Cluster map[string]interface{} `json:"cluster"`
		} `json:"clusters"`
		Users []struct {
			User map[string]interface{} `json:"user"`
		} `json:"users"`
	}
	require.NoError(t, json.Unmarshal(data, &config))
	require.Len(t, config.Clusters, 1)
	require.Len(t, config.Users, 1)
	require.NotContains(t, config.Clusters[0].Cluster, "certificate-authority")
	require.NotContains(t, config.Users[0].User, "client-certificate")
	require.NotContains(t, config.Users[0].User, "client-key")
	for field, name := range map[string]string{
		"certificate-authority-data": "ca",
		"client-certificate-data":    "cert",
		"client-key-data":            "key",
	} {
		var actual interface{}
		if name == "ca" {
			actual = config.Clusters[0].Cluster[field]
		} else {
			actual = config.Users[0].User[field]
		}
		require.Equal(t, base64.StdEncoding.EncodeToString([]byte("test-"+name)), actual)
	}
	for _, entry := range env[2:] {
		key, value, ok := strings.Cut(entry, "=")
		require.True(t, ok)
		output, err := exec.Command("go", "env", key).Output()
		require.NoError(t, err)
		require.Equal(t, strings.TrimSpace(string(output)), value)
	}
}

func TestProviderTestEnvironmentMissingCredentials(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is required to flatten the test kubeconfig")
	}
	sourcePath := filepath.Join(t.TempDir(), "config")
	require.NoError(t, os.WriteFile(sourcePath, []byte(`{
		"apiVersion":"v1","kind":"Config","current-context":"test",
		"contexts":[{"name":"test","context":{"cluster":"test","user":"test"}}],
		"clusters":[{"name":"test","cluster":{"server":"https://127.0.0.1:6443"}}],
		"users":[{"name":"test","user":{"client-certificate":"missing-certificate","client-key":"missing-key"}}]
	}`), 0600))
	t.Setenv("KUBECONFIG", sourcePath)
	env, err := providerTestEnvironment(t.TempDir())
	require.ErrorContains(t, err, "preparing embedded provider-test kubeconfig")
	require.Nil(t, env)
}
