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

	err = shellExec(fmt.Sprintf("cd %s && mage test", absPath))
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
			dirPath := filepath.Join(dir, entry.Name())
			filePath := filepath.Join(dirPath, "magefile.go")
			if _, err := os.Stat(filePath); err == nil {
				results = append(results, dirPath)
			}
		}
	}

	return results, nil
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
