//go:build mage

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"gopkg.in/yaml.v2"
)

// Test config
const (
	TEST_NAME    = "Symphony Catalog Evaluation test scenario"
	TEST_TIMEOUT = "4m"
)

var (
	NAMESPACES = []string{
		"default",
		"nondefault",
	}
)

var (
	// catalogs to deploy
	testCatalogs = []string{
		"test/integration/scenarios/10.catalogEval/manifest/catalog-catalog-container.yaml",
		"test/integration/scenarios/10.catalogEval/manifest/catalog-catalog.yaml",
		"test/integration/scenarios/10.catalogEval/manifest/catalog-catalog2.yaml",
		"test/integration/scenarios/10.catalogEval/manifest/catalog-catalog3.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}

	testEval = "test/integration/scenarios/10.catalogEval/manifest/eval.yaml"

	testWrongEval = "test/integration/scenarios/10.catalogEval/manifest/wrongEval.yaml"

	testEvalUpdate = "test/integration/scenarios/10.catalogEval/manifest/evalUpdate.yaml"

	testEval03 = "test/integration/scenarios/10.catalogEval/manifest/eval03.yaml"
)

// Entry point for running the tests
func Test() error {
	fmt.Println("Running ", TEST_NAME)

	defer Cleanup()
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}
	err = DeployManifests()
	if err != nil {
		return err
	}
	err = Verify()
	if err != nil {
		return err
	}

	return nil
}

func DeployManifests() error {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../"
	}
	namespace := "default"
	// Deploy the catalogs
	for _, catalog := range testCatalogs {
		absCatalog := filepath.Join(repoPath, catalog)

		data, err := os.ReadFile(absCatalog)
		if err != nil {
			return err
		}
		stringYaml := string(data)
		stringYaml = strings.ReplaceAll(stringYaml, "SCOPENAME", namespace)

		err = writeYamlStringsToFile(stringYaml, "./test.yaml")
		if err != nil {
			return err
		}
		err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
		if err != nil {
			return err
		}
		os.Remove("./test.yaml")
	}

	// wait for 2 seconds to make sure catalog is created
	time.Sleep(time.Second * 2)
	// deploy eval catalog evaluateevalcatalog01
	evalC := filepath.Join(repoPath, testEval)
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", evalC, namespace)).Run()
	if err != nil {
		return err
	}

	// wait for 2 seconds to make sure evaluateevalcatalog01 is created
	time.Sleep(time.Second * 2)
	// update eval catalog evaluateevalcatalog01
	evalUpdateC := filepath.Join(repoPath, testEvalUpdate)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", evalUpdateC, namespace)).Run()
	if err == nil {
		return errors.New("Update should not be successful")
	}

	// create eval catalog evaluateevalcatalog02
	evalWrongC := filepath.Join(repoPath, testWrongEval)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", evalWrongC, namespace)).Run()
	if err != nil {
		return err
	}

	// create eval catalog evaluateevalcatalog03
	eval03C := filepath.Join(repoPath, testEval03)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", eval03C, namespace)).Run()
	if err != nil {
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
