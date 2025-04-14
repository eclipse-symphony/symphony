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
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	TEST_NAME    = "basic manifest deploy scenario"
	TEST_TIMEOUT = "10m"
)

var (
	NAMESPACES = []string{
		"default",
		"nondefault",
	}
)

var (
	// Manifests to deploy
	testManifests = []string{
		"manifest/%s/solution-container.yaml",
		"manifest/%s/target.yaml",
		"manifest/%s/solution.yaml",
		"manifest/%s/instance.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	// defer testhelpers.Cleanup(TEST_NAME)
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}
	for _, namespace := range NAMESPACES {
		os.Setenv("NAMESPACE", namespace)
		err = Verify()
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}

// Run tests
func Verify() error {
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	os.Setenv("SYMPHONY_FLAVOR", "oss")
	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
