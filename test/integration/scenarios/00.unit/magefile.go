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
	"path/filepath"
	"time"

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

	err := Setup()
	if err != nil {
		return err
	}

	err = Verify()
	if err != nil {
		return err
	}

	return nil
}

// Prepare the cluster
// Run this manually to prepare your local environment for testing/debugging
func Setup() error {
	// Deploy symphony
	err := localenvCmd("cluster:deploy", "")
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
	shellExec(fmt.Sprintf("kubectl delete deployment nginx -n default"))
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
