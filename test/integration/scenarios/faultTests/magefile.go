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

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/eclipse-symphony/symphony/test/integration/scenarios/faultTests/utils"
	"github.com/princjef/mageutil/shellcmd"
)

// Entry point for running the tests
func FaultTests() error {
	fmt.Println("Running fault injection tests")

	// Run fault injection tests
	for _, test := range utils.Faults {
		err := FaultTestHelper(test)
		if err != nil {
			return err
		}
	}
	return nil
}

func FaultTestHelper(test utils.FaultTestCase) error {
	testName := fmt.Sprintf("%s/%s/%s", test.TestCase, test.Fault, test.FaultType)
	fmt.Println("Running ", testName)

	// Step 2.1: setup cluster
	defer testhelpers.Cleanup(testName)
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}
	// Step 2.2: pass the fault info to the test process. Each injection dials
	// a fresh port-forward to the live pod; a long-lived port-forward set up
	// here does not survive the pod restarting after a panic fault.
	os.Setenv(utils.PodEnvKey, test.PodLabel)
	os.Setenv(utils.FaultNameEnvKey, test.Fault)
	os.Setenv(utils.FaultTypeEnvKey, test.FaultType)

	err = Verify(test.TestCase)
	return err
}

func Verify(test string) error {
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	err = shellcmd.Command(fmt.Sprintf("go test -v -timeout %s %s", utils.TEST_TIMEOUT, test)).Run()
	if err != nil {
		return err
	}

	return nil
}
