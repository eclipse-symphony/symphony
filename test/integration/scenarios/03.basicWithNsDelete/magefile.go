//go:build mage

/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

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
	TEST_NAME    = "basic with namespace delete"
	TEST_TIMEOUT = "10m"
	NAMESPACE    = "default"
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

	defer testhelpers.Cleanup(TEST_NAME)

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
	testhelpers.SetupCluster()
	err := shellcmd.Command(fmt.Sprintf("kubectl get secret container")).Run()
	if err != nil {
		err := shellcmd.Command("kubectl apply -f manifest/oss/secret.yaml").Run()
		if err != nil {
			return err
		}
	}
	// Deploy the manifests
	for _, manifest := range testManifests {
		fullPath, err := filepath.Abs(fmt.Sprintf(manifest, "oss"))
		if err != nil {
			return err
		}

		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", fullPath, NAMESPACE)).Run()
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
