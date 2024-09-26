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
	TEST_NAME    = "Symphony catalog namespace test scenario"
	TEST_TIMEOUT = "4m"
)

var (
	NAMESPACES = []string{
		"default",
		"nondefault",
	}
)

var (
	// catalogs to deploy
	testManifests = []string{
		"manifest/config-container.yaml",
		"manifest/config.yaml",
		"manifest/campaign-container.yaml",
		"manifest/campaign.yaml",
	}

	testActivation = "manifest/activation.yaml"

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

	for _, namespace := range NAMESPACES {
		os.Setenv("NAMESPACE", namespace)
		err := DeployManifests(namespace)
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

// Verify namespace in nondefault namespace
func DeployManifests(namespace string) error {
	// Ensure that namespace is defined
	err := testhelpers.EnsureNamespace(namespace)
	if err != nil {
		return err
	}
	os.Setenv("NAMESPACE", namespace)

	// setup campaign
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, manifest := range testManifests {
		absManifest := filepath.Join(currentPath, manifest)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absManifest, namespace)).Run()
		if err != nil {
			return err
		}
	}

	// setup activation
	absActivation := filepath.Join(currentPath, testActivation)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absActivation, namespace)).Run()
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
	os.Setenv("SYMPHONY_FLAVOR", "oss")
	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
