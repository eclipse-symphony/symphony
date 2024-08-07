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
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"gopkg.in/yaml.v2"
)

// Test config
const (
	TEST_NAME    = "workflow test"
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
		"test/integration/scenarios/04.workflow/manifest/catalog-catalog-container.yaml",
		"test/integration/scenarios/04.workflow/manifest/instance-catalog-container.yaml",
		"test/integration/scenarios/04.workflow/manifest/solution-catalog-container.yaml",
		"test/integration/scenarios/04.workflow/manifest/target-catalog-container.yaml",

		"test/integration/scenarios/04.workflow/manifest/catalog-catalog.yaml",
		"test/integration/scenarios/04.workflow/manifest/instance-catalog.yaml",
		"test/integration/scenarios/04.workflow/manifest/solution-catalog.yaml",
		"test/integration/scenarios/04.workflow/manifest/target-catalog.yaml",
	}

	testCampaign = []string{
		"test/integration/scenarios/04.workflow/manifest/campaign-container.yaml",
		"test/integration/scenarios/04.workflow/manifest/campaign.yaml",
	}

	testActivations = []string{
		"test/integration/scenarios/04.workflow/manifest/activation.yaml",
	}

	// Tests to run
	testVerify = []string{
		"./verify/...",
	}

	CampaignNotExistActivation = "test/integration/scenarios/04.workflow/manifest/activation-campaignnotexist.yaml"

	WithStageActivation = "test/integration/scenarios/04.workflow/manifest/activation-stage.yaml"
)

// Entry point for running the tests
func Test(labeling bool) error {
	fmt.Println("Running ", TEST_NAME)

	if labeling {
		err := modifyYAML("localtest", "management.azure.com/azureName")
		if err != nil {
			return err
		}
		os.Setenv("labelingEnabled", "true")
	}
	defer Cleanup()
	err := testhelpers.SetupCluster()
	if err != nil {
		return err
	}
	for _, namespace := range NAMESPACES {
		os.Setenv("NAMESPACE", namespace)
		err := DeployManifests(namespace)
		if err != nil {
			return err
		}
		err = Verify()
		if err != nil {
			return err
		}
		err = FaultTest(namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeployManifests(namespace string) error {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../"
	}
	if namespace != "default" {
		// Create non-default namespace if not exist
		err := shellcmd.Command(fmt.Sprintf("kubectl get namespace %s", namespace)).Run()
		if err != nil {
			// Better to check err message here but command only returns "exit status 1" for non-exisiting namespace
			err = shellcmd.Command(fmt.Sprintf("kubectl create namespace %s", namespace)).Run()
			if err != nil {
				return err
			}
		}
	}
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

	for _, campaign := range testCampaign {
		absCampaign := filepath.Join(repoPath, campaign)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absCampaign, namespace)).Run()
		if err != nil {
			return err
		}
	}

	// wait for 5 seconds to make sure campaign is created
	time.Sleep(time.Second * 5)
	for _, activation := range testActivations {
		absActivation := filepath.Join(repoPath, activation)
		err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", absActivation, namespace)).Run()
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

func FaultTest(namespace string) error {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../"
	}
	var err error
	CampaignNotExistActivationAbs := filepath.Join(repoPath, CampaignNotExistActivation)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", CampaignNotExistActivationAbs, namespace)).Run()
	if err == nil {
		return fmt.Errorf("fault test failed for non-existing campaign")
	}
	WithStageActivationAbs := filepath.Join(repoPath, WithStageActivation)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s -n %s", WithStageActivationAbs, namespace)).Run()
	if err == nil {
		return fmt.Errorf("fault test failed for non-existing campaign")
	}
	return nil
}

// Clean up
func Cleanup() {
	err := modifyYAML("", "")
	if err != nil {
		fmt.Printf("Failed to set up the symphony-ghcr-values.yaml. Please make sure the labelKey and labelValue is set to null.\n")
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
	data, err := os.ReadFile("../../../localenv/symphony-ghcr-values.yaml")
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
	err = os.WriteFile("../../../localenv/symphony-ghcr-values.yaml", data, 0644)
	if err != nil {
		return err
	}

	return nil
}
