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
	"os/exec"
	"time"

	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	TEST_NAME    = "scenario/update"
	TEST_TIMEOUT = "15m"
)

var (
	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
	TestNamespaces = []string{
		"default",
		"nondefault",
	}
)

// Entry point for running the tests, including setup, verify and cleanup
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer Cleanup()
	err := Setup()
	if err != nil {
		return err
	}

	// Wait a few secs for symphony cert to be ready;
	// otherwise we will see error when creating symphony manifests in the cluster
	// <Error from server (InternalError): error when creating
	// "/mnt/vss/_work/1/s/test/integration/scenarios/basic/manifest/target.yaml":
	// Internal error occurred: failed calling webhook "mtarget.kb.io": failed to
	// call webhook: Post
	// "https://symphony-webhook-service.default.svc:443/mutate-symphony-microsoft-com-v1-target?timeout=10s":
	// x509: certificate signed by unknown authority>
	time.Sleep(time.Second * 10)
	for _, namespace := range TestNamespaces {

		os.Setenv("NAMESPACE", namespace)
		err = Verify()
		if err != nil {
			return err
		}
	}

	return nil
}

// Deploy Symphony to the cluster
func Setup() error {
	// Deploy symphony
	return localenvCmd("cluster:deploy", "")
}

// Run tests for scenarios/update
func Verify() error {
	err := shellcmd.Command("go clean -testcache").Run()
	if err != nil {
		return err
	}
	os.Setenv("SYMPHONY_FLAVOR", "oss")
	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -v -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Clean up Symphony release and temporary test files
func Cleanup() {

	_ = shellcmd.Command("rm -rf ./manifestForTestingOnly/oss").Run()

	localenvCmd(fmt.Sprintf("dumpSymphonyLogsForTest '%s'", TEST_NAME), "")
	localenvCmd("destroy all", "")
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string, flavor string) error {
	return shellExec(fmt.Sprintf("cd ../../../localenv && mage %s %s", mageCmd, flavor))
}

// Run a command with | or other things that do not work in shellcmd
func shellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
