//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"fmt"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	TEST_NAME    = "naming test"
	TEST_TIMEOUT = "10m"
)

var (

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer testhelpers.Cleanup(TEST_NAME)
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
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func PrepareManifests() error {
	// Prepare manifests
	err := testhelpers.ReplacePlaceHolderInManifest("manifest/instance.yaml", "${PLACEHOLDER_TARGET}", "${PLACEHOLDER_SOLUTIONCONTAINER}", "${PLACEHOLDER_SOLUTION}", "${PLACEHOLDER_INSTANCE}", "")
	if err != nil {
		return err
	}
	err = testhelpers.ReplacePlaceHolderInManifest("manifest/solution.yaml", "${PLACEHOLDER_TARGET}", "${PLACEHOLDER_SOLUTIONCONTAINER}", "${PLACEHOLDER_SOLUTION}", "${PLACEHOLDER_INSTANCE}", "")
	if err != nil {
		return err
	}
	err = testhelpers.ReplacePlaceHolderInManifest("manifest/solution-container.yaml", "${PLACEHOLDER_TARGET}", "${PLACEHOLDER_SOLUTIONCONTAINER}", "${PLACEHOLDER_SOLUTION}", "${PLACEHOLDER_INSTANCE}", "")
	if err != nil {
		return err
	}
	err = testhelpers.ReplacePlaceHolderInManifest("manifest/target.yaml", "${PLACEHOLDER_TARGET}", "${PLACEHOLDER_SOLUTIONCONTAINER}", "${PLACEHOLDER_SOLUTION}", "${PLACEHOLDER_INSTANCE}", "")
	if err != nil {
		return err
	}
	err = testhelpers.ReplacePlaceHolderInManifest("manifest/target01.yaml", "${PLACEHOLDER_TARGET}", "${PLACEHOLDER_SOLUTIONCONTAINER}", "${PLACEHOLDER_SOLUTION}", "${PLACEHOLDER_INSTANCE}", "${PLACEHOLDER_INSTANCEHISTORY}")
	return nil
}
