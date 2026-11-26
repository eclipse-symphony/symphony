//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	TEST_NAME = "Symphony Provider test scenario"
)

var (
	// Manifests to deploy
	testPackage = []string{
		"../../../../api/pkg/apis/v1alpha1/providers/target/helm",
		"../../../../api/pkg/apis/v1alpha1/providers/target/kubectl",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer Cleanup()

	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}

	err = Verify()
	if err != nil {
		return err
	}

	return nil
}

// Run tests
func Verify() (result error) {
	home, err := os.MkdirTemp("", "symphony-provider-tests-")
	if err != nil {
		return err
	}
	defer func() {
		result = errors.Join(result, os.RemoveAll(home))
	}()
	env, err := providerTestEnvironment(home)
	if err != nil {
		return err
	}
	// Set env vars for helm provider testing
	os.Setenv("TEST_HELM_CHART", "true")
	os.Setenv("TEST_SYMPHONY_HELM_VERSION", "true")
	// Set env vars for kubectl provider testing
	os.Setenv("TEST_KUBECTL", "true")
	err = shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	os.Setenv("SYMPHONY_FLAVOR", "oss")
	for _, testFile := range testPackage {
		fullPath, err := filepath.Abs(testFile)
		if err != nil {
			return err
		}

		cmd := exec.Command("go", "-C", fullPath, "test", ".")
		cmd.Env = append(os.Environ(), env...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("> go -C %q test .\n", fullPath)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func providerTestEnvironment(home string) ([]string, error) {
	configDir := filepath.Join(home, ".kube")
	if err := os.Mkdir(configDir, 0700); err != nil {
		return nil, err
	}
	configPath := filepath.Join(configDir, "config")
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	// Only the trusted local test setup may read certificate files to embed them.
	cmd := exec.Command("kubectl", "config", "view", "--minify", "--flatten", "--raw", "-o=json")
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	if err := errors.Join(cmd.Run(), file.Close()); err != nil {
		return nil, fmt.Errorf("preparing embedded provider-test kubeconfig: %w", err)
	}
	// Isolate the default kubeconfig without moving Go's caches or toolchain settings.
	output, err := exec.Command("go", "env", "-json", "GOPATH", "GOMODCACHE", "GOCACHE", "GOENV").Output()
	if err != nil {
		return nil, fmt.Errorf("reading Go environment for provider tests: %w", err)
	}
	var goEnv map[string]string
	if err := json.Unmarshal(output, &goEnv); err != nil {
		return nil, err
	}
	env := []string{"HOME=" + home, "KUBECONFIG=" + configPath}
	for _, key := range []string{"GOPATH", "GOMODCACHE", "GOCACHE", "GOENV"} {
		env = append(env, key+"="+goEnv[key])
	}
	return env, nil
}

// Clean up
func Cleanup() {
	testhelpers.ShellExec(fmt.Sprintf("kubectl delete deployment nginx -n default"))
	testhelpers.Cleanup(TEST_NAME)
}
