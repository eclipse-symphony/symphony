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
	TEST_NAME = "Symphony sample test scenario"
	TEST_TIMEOUT = "10m"
)

var (
	// Manifests to deploy
	testSamples = map[string][]string{
        "sample-hello-world":  {
			"../../../../docs/samples/k8s/hello-world/solution-container.yaml",
			"../../../../docs/samples/k8s/hello-world/solution.yaml",
			"../../../../docs/samples/k8s/hello-world/target.yaml",
			"../../../../docs/samples/k8s/hello-world/instance.yaml",
		},
        "sample-staged": {
			"../../../../docs/samples/k8s/staged/solution-container.yaml",
			"../../../../docs/samples/k8s/staged/solution.yaml",
			"../../../../docs/samples/k8s/staged/target.yaml",
			"../../../../docs/samples/k8s/staged/instance.yaml",
		},
    }

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

	// Deploy solution, target and instance
	for namespace, manifests := range testSamples {
		os.Setenv("NAMESPACE", namespace)
		err := DeployManifests(namespace, manifests)
		if err != nil {
			return err
		}

		err = Verify()
		if err != nil {
			return err
		}
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

// Deploy solution, target and instance
func DeployManifests(namespace string, testManifests []string) error {
	// Ensure that namespace is defined
	err := testhelpers.EnsureNamespace(namespace)
	if err != nil {
		return err
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}

	// Deploy the manifests
	for _, manifest := range testManifests {
		manifestPath := filepath.Join(currentPath, manifest)
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", manifestPath, namespace)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}