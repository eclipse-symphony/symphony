//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/princjef/mageutil/shellcmd"
)

// Test config
const (
	TEST_NAME    = "basic manifest deploy scenario"
	TEST_TIMEOUT = "10m"
	NAMESPACE    = "default"
)

var (
	// Manifests to deploy
	testManifests = []string{
		"manifest/target.yaml",
		"manifest/instance.yaml",
		"manifest/solution.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	// TODO: enable this once clean up works
	// defer Cleanup()

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
	err := localenvCmd("deploy")
	if err != nil {
		return err
	}

	// Deploy the manifests
	for _, manifest := range testManifests {
		fullPath, err := filepath.Abs(manifest)
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

	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

// Clean up
func Cleanup() {
	localenvCmd("destroy")
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string) error {
	return shellExec(fmt.Sprintf("cd ../../../../localenv && mage %s", mageCmd))
}

// Run a command with | or other things that do not work in shellcmd
func shellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
