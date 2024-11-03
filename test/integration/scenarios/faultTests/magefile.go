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
	"github.com/princjef/mageutil/shellcmd"
)

func FaultTests() error {
	fmt.Println("Running fault injection tests")

	// Run fault injection tests
	for _, test := range Faults {
		err := FaultTestHelper(test)
		if err != nil {
			return err
		}
	}
	return nil
}

func FaultTestHelper(test FaultTestCase) error {
	testName := fmt.Sprintf("%s/%s/%s", test.testCase, test.fault, test.faultType)
	fmt.Println("Running ", testName)

	// Step 2.1: setup cluster
	defer testhelpers.Cleanup(testName)
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}
	// Step 2.2: enable port forward on specific pod
	stopChan := make(chan struct{}, 1)
	defer close(stopChan)
	err = testhelpers.EnablePortForward(test.podLabel, LocalPortForward, stopChan)
	if err != nil {
		return err
	}

	InjectCommand := fmt.Sprintf("curl localhost:%s/%s -XPUT -d'%s'", LocalPortForward, test.fault, test.faultType)
	os.Setenv("InjectCommand", InjectCommand)
	os.Setenv("InjectPodLabel", test.podLabel)

	err = Verify(test.testCase)
	return err
}

// Run tests for scenarios/update
func Verify(test string) error {
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	err = shellcmd.Command(fmt.Sprintf("go test -v -timeout %s %s", TEST_TIMEOUT, test)).Run()
	if err != nil {
		return err
	}

	return nil
}
