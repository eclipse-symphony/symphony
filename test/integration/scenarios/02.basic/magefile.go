//go:build mage

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
	TEST_NAME    = "basic manifest deploy scenario"
	TEST_TIMEOUT = "10m"
	NAMESPACE    = "default"
)

var (
	// Manifests to deploy
	testManifests = []string{
		"manifest/%s/target.yaml",
		"manifest/%s/instance.yaml",
		"manifest/%s/solution.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

func conditionalRun(azureFunc func() error, ossFunc func() error) error {
	if len(os.Args) > 2 && os.Args[len(os.Args)-1] == "azure" {
		return azureFunc()
	}
	return ossFunc()
}
func conditionalString(azureStr string, ossStr string) string {
	if len(os.Args) > 2 && os.Args[len(os.Args)-1] == "azure" {
		return azureStr
	}
	return ossStr
}

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
	err := localenvCmd("cluster:deploy", conditionalString("azure", ""))
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
	// Deploy the manifests
	for _, manifest := range testManifests {
		fullPath, err := filepath.Abs(fmt.Sprintf(manifest, conditionalString("azure", "oss")))
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
	os.Setenv("SYMPHONY_FLAVOR", conditionalString("azure", "oss"))
	for _, verify := range testVerify {
		err := shellcmd.Command(fmt.Sprintf("go test -timeout %s %s", TEST_TIMEOUT, verify)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
func Azure() error {
	return nil
}

// Clean up
func Cleanup() {
	localenvCmd("destroy all", conditionalString("azure", ""))
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string, flavor string) error {
	return shellExec(fmt.Sprintf("cd ../../../../localenv && mage %s %s", mageCmd, flavor))
}

// Run a command with | or other things that do not work in shellcmd
func shellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
