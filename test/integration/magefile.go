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
	"os/exec"
	"path/filepath"
)

const (
	testFileName  = "test.yaml"
	goTestTimeout = "10m"
	NAMESPACE     = "default"
)

type (
	// TestSuite is a set of tests that run one after the other.
	TestSuite struct {
		Tests []TestEntry `yaml:"tests"`
	}

	// TestEntry is a single test in a suite.
	TestEntry struct {
		Name     string              `yaml:"name"`
		Manifest []string            `yaml:"manifest"`
		Wait     map[string][]string `yaml:"wait"`
		Verify   []string            `yaml:"verify"`
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

// Test runs all integration tests
func Test() error {
	fmt.Println("Searching for integration tests")

	scenariosPath, err := filepath.Abs("scenarios")
	if err != nil {
		return err
	}

	testFiles, err := listTests(scenariosPath)
	if err != nil {
		return err
	}

	for _, testFile := range testFiles {
		fmt.Printf("Running tests in: %s\n", testFile)

		err = RunTest(testFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// Run a test file
// Deploys once at the start and cleans up at the end
func RunTest(testDir string) error {
	absPath, err := filepath.Abs(testDir)
	if err != nil {
		return err
	}

	fmt.Printf("Starting test folder: %s\n", absPath)

	err = shellExec(fmt.Sprintf("cd %s && mage test %s", absPath, conditionalString("azure", "")))
	if err != nil {
		return err
	}

	return nil
}

// Search the scenarios folder for sub folders
func listTests(dir string) ([]string, error) {
	results := make([]string, 0)

	// Read root folder
	subDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Read test subfolders
	for _, entry := range subDirs {
		if entry.IsDir() {
			results = append(results, filepath.Join(dir, entry.Name()))
		}
	}

	return results, nil
}

func Azure() error {
	//this is a hack to get around the fact that mage doesn't support passing args to targets
	return nil
}

// Run a mage command from /localenv
func localenvCmd(mageCmd string, flavor string) error {
	return shellExec(fmt.Sprintf("cd ../localenv && mage %s %s", mageCmd, flavor))
}

// Run a command with | or other things that do not work in shellcmd
func shellExec(cmd string) error {
	fmt.Println("> ", cmd)

	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
