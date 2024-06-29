//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"fmt"
	"os"
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
func Verify() error {
	// Set env vars for helm provider testing
	os.Setenv("TEST_HELM_CHART", "true")
	os.Setenv("TEST_SYMPHONY_HELM_VERSION", "true")
	// Set env vars for kubectl provider testing
	os.Setenv("TEST_KUBECTL", "true")
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	os.Setenv("SYMPHONY_FLAVOR", "oss")
	for _, testFile := range testPackage {
		fullPath, err := filepath.Abs(testFile)
		if err != nil {
			return err
		}

		err = shellcmd.Command(fmt.Sprintf("go test %s", fullPath)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Clean up
func Cleanup() {
	testhelpers.ShellExec(fmt.Sprintf("kubectl delete deployment nginx -n default"))
	testhelpers.Cleanup(TEST_NAME)
}
