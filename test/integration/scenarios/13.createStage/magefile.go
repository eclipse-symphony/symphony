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
	"gopkg.in/yaml.v2"
)

// Test config
const (
	TEST_NAME    = "createStage test"
	TEST_TIMEOUT = "8m"
)

var (
	NAMESPACES = []string{
		"default",
	}
)

var (
	// Tests to run
	testVerify = []string{
		"./verify/...",
	}
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	//defer Cleanup()
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}

	err = Verify()
	if err != nil {
		fmt.Printf("Failed to run %s tests: %v\n", TEST_NAME, err)
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

// Clean up
func Cleanup() {
	err := modifyYAML("", "")
	if err != nil {
		fmt.Printf("Failed to set up the %s. Please make sure the labelKey and labelValue is set to null.\n", getGhcrValueFileName())
	}
	testhelpers.Cleanup(TEST_NAME)
}

func writeYamlStringsToFile(yamlString string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(yamlString))
	if err != nil {
		return err
	}

	return nil
}

func enableTlsOtelSetup() bool {
	return os.Getenv("ENABLE_TLS_OTEL_SETUP") == "true"
}

func enableNonTlsOtelSetup() bool {
	return os.Getenv("ENABLE_NON_TLS_OTEL_SETUP") == "true"
}

func getGhcrValueFileName() string {
	if enableTlsOtelSetup() {
		return "symphony-ghcr-values.otel.yaml"
	} else if enableNonTlsOtelSetup() {
		return "symphony-ghcr-values.otel.non-tls.yaml"
	} else {
		return "symphony-ghcr-values.yaml"
	}
}

func modifyYAML(v string, annotationKey string) error {
	// Read the YAML file
	ghcrValueFilePath := fmt.Sprintf("../../../localenv/%s", getGhcrValueFileName())
	data, err := os.ReadFile(ghcrValueFilePath)
	if err != nil {
		return err
	}

	// Unmarshal the YAML data into a map
	var values map[string]interface{}
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return err
	}

	// Modify the 'api' fields
	if api, ok := values["api"].(map[interface{}]interface{}); ok {
		api["labelKey"] = v
		api["labelValue"] = v
		api["annotationKey"] = annotationKey
	} else {
		return fmt.Errorf("'api' field is not a map")
	}

	// Marshal the map back into YAML
	data, err = yaml.Marshal(values)
	if err != nil {
		return err
	}

	// Write the modified YAML data back to the file
	err = os.WriteFile(ghcrValueFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
