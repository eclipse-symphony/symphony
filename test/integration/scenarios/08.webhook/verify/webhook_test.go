/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/assert"
)

var (
	testSolutionContainer = "test/integration/scenarios/01.update/manifestTemplates/oss/solution-container.yaml"
	testSolution          = "test/integration/scenarios/01.update/manifestTemplates/oss/solution.yaml"
	testTarget            = "test/integration/scenarios/01.update/manifestTemplates/oss/target.yaml"
	testInstance          = "test/integration/scenarios/01.update/manifestTemplates/oss/instance.yaml"
	testCampaign          = "test/integration/scenarios/04.workflow/manifest/campaign.yaml"
	testCampaignContainer = "test/integration/scenarios/04.workflow/manifest/campaign-container.yaml"

	testCampaignWithWrongStage     = "test/integration/scenarios/08.webhook/manifest/campaignWithWrongStages.yaml"
	testCampaignWithLongRunning    = "test/integration/scenarios/08.webhook/manifest/campaignLongRunning.yaml"
	testActivationsWithWrongStage  = "test/integration/scenarios/08.webhook/manifest/activationWithWrongStage.yaml"
	testActivationsWithLongRunning = "test/integration/scenarios/08.webhook/manifest/activationLongRunning.yaml"

	testCatalog          = "test/integration/scenarios/05.catalogNconfigmap/manifests/config.yaml"
	testCatalogContainer = "test/integration/scenarios/05.catalogNconfigmap/manifests/config-container.yaml"
	testSchema           = "test/integration/scenarios/05.catalogNconfigmap/manifests/schema.yaml"
	testSchemaContainer  = "test/integration/scenarios/05.catalogNconfigmap/manifests/schema-container.yaml"
	testWrongSchema      = "test/integration/scenarios/05.catalogNconfigmap/manifests/wrongconfig.yaml"
	testChildCatalog     = "test/integration/scenarios/08.webhook/manifest/childCatalog.yaml"

	testCircularParentContainer = "test/integration/scenarios/08.webhook/manifest/parent-container.yaml"
	testCircularParent          = "test/integration/scenarios/08.webhook/manifest/parent-config.yaml"
	testCircularParentUpdate    = "test/integration/scenarios/08.webhook/manifest/parent-update.yaml"
	testCircularChildContainer  = "test/integration/scenarios/08.webhook/manifest/child-container.yaml"
	testCircularChild           = "test/integration/scenarios/08.webhook/manifest/child-config.yaml"
	testNoParentChild           = "test/integration/scenarios/08.webhook/manifest/child-noparent.yaml"

	diagnostic_01_WithoutEdgeLocation     = "test/integration/scenarios/08.webhook/manifest/diagnostic_01.WithoutEdgeLocation.yaml"
	diagnostic_02_WithCorrectEdgeLocation = "test/integration/scenarios/08.webhook/manifest/diagnostic_02.WithCorrectEdgeLocation.yaml"
	diagnostic_03_WithAnotherNS           = "test/integration/scenarios/08.webhook/manifest/diagnostic_03.WithAnotherNS.yaml"
	diagnostic_04_WithAnotherName         = "test/integration/scenarios/08.webhook/manifest/diagnostic_04.WithAnotherName.yaml"
	diagnostic_05_UpdateOtherAnnotations  = "test/integration/scenarios/08.webhook/manifest/diagnostic_05.UpdateOtherAnnotations.yaml"

	historyCreate         = "test/integration/scenarios/08.webhook/manifest/history.yaml"
	historyUpdate         = "test/integration/scenarios/08.webhook/manifest/history-update.yaml"
	historyTarget         = "test/integration/scenarios/08.webhook/manifest/history-target.yaml"
	historySolution       = "test/integration/scenarios/08.webhook/manifest/history-solution.yaml"
	historyInstance       = "test/integration/scenarios/08.webhook/manifest/history-instance.yaml"
	historySolutionUpdate = "test/integration/scenarios/08.webhook/manifest/history-solution-update.yaml"
	historyInstanceUpdate = "test/integration/scenarios/08.webhook/manifest/history-instance-update.yaml"
)

// Define a struct to parse the JSON output
type HistoryList struct {
	Items []map[string]interface{} `json:"items"`
}

func TestCreateSolutionWithoutContainer(t *testing.T) {
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testSolution)).CombinedOutput()
	assert.Contains(t, string(output), "rootResource must be a valid container")
	assert.NotNil(t, err, "solution creation without container should fail")
}

func TestInstanceWithoutSolution(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testTarget))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testInstance)).CombinedOutput()
	assert.Contains(t, string(output), "solution does not exist")
	assert.NotNil(t, err, "instance creation without solution should fail")
	err = shellcmd.Command("kubectl delete targets.fabric.symphony self").Run()
	assert.Nil(t, err)
}

func TestInstanceWithoutTarget(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolutionContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolution))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testInstance)).CombinedOutput()
	assert.Contains(t, string(output), "target does not exist")
	assert.NotNil(t, err, "instance creation without target should fail")
	output, err = exec.Command("kubectl", "delete", "solutioncontainers.solution.symphony", "mysol").CombinedOutput()
	assert.Contains(t, string(output), "nested resources with root resource 'mysol' are not empty")
	assert.NotNil(t, err, "solution container deletion with solution should fail")
	err = shellcmd.Command("kubectl delete solutions.solution.symphony mysol-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutioncontainers.solution.symphony mysol").Run()
	assert.Nil(t, err)
}

func TestTargetSolutionDeletionWithInstance(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testTarget))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolutionContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolution))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testInstance))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "delete", "solutions.solution.symphony", "mysol-v-version1").CombinedOutput()
	assert.Contains(t, string(output), "Solution has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err)
	output, err = exec.Command("kubectl", "delete", "targets.fabric.symphony", "self").CombinedOutput()
	assert.Contains(t, string(output), "Target has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err, "target deletion with instance should fail")
	err = shellcmd.Command("kubectl delete instances.solution.symphony instance").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete targets.fabric.symphony self").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutions.solution.symphony mysol-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutioncontainers.solution.symphony mysol").Run()
	assert.Nil(t, err)
}

func TestCreateActivationWithoutCampaign(t *testing.T) {
	// Already covered in 04.workflow tests
}

func TestCreateActivationWithWrongFirstStage(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaignContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaign))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testActivationsWithWrongStage)).CombinedOutput()
	assert.Contains(t, string(output), "spec.stage must be a valid stage in the campaign")
	assert.NotNil(t, err, "activation creation with wrong stage should fail")
	output, err = exec.Command("kubectl", "delete", "campaigncontainers.workflow.symphony", "04campaign").CombinedOutput()
	assert.Contains(t, string(output), "nested resources with root resource '04campaign' are not empty")
	assert.NotNil(t, err, "campaign container deletion with campaign should fail")
	err = shellcmd.Command("kubectl delete campaigns.workflow.symphony 04campaign-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete campaigncontainers.workflow.symphony 04campaign").Run()
	assert.Nil(t, err)
}

func TestCreateCampaignWithWrongStages(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaignContainer))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCampaignWithWrongStage)).CombinedOutput()
	assert.Contains(t, string(output), "stageSelector must be one of the stages in the stages list")
	assert.NotNil(t, err, "campaign creation with wrong stages should fail")
	err = shellcmd.Command("kubectl delete campaigncontainers.workflow.symphony 04campaign").Run()
	assert.Nil(t, err)
}

func TestDeleteCampaignWithRunningActivation(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaignContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaignWithLongRunning))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testActivationsWithLongRunning))).Run()
	assert.Nil(t, err)
	start := time.Now()
	for {
		output, err := exec.Command("kubectl", "get", "activation", "activationlongrunning", "-o", "jsonpath={.status.statusMessage}").CombinedOutput()
		if err != nil {
			assert.Fail(t, "failed to get activation %s state: %s", "activationlongrunning", err.Error())
		}
		state := string(output)
		if state == "Running" {
			break
		}
		if time.Since(start) > time.Second*30 {
			assert.Fail(t, "timed out waiting for activation %s to reach state %s", "activationlongrunning", "Running")
		}
		time.Sleep(5 * time.Second)
	}
	output, err := exec.Command("kubectl", "delete", "campaigns.workflow.symphony", "04campaign-v-version3").CombinedOutput()
	assert.Contains(t, string(output), "Campaign has one or more running activations. Update or Deletion is not allowed")
	assert.NotNil(t, err, "campaign deletion with running activation should fail")
	time.Sleep(15 * time.Second)
	// Campaign can be deleted once the activation is DONE
	output, err = exec.Command("kubectl", "delete", "campaigncontainers.workflow.symphony", "04campaign").CombinedOutput()
	assert.NotNil(t, err, "campaign container deletion with campaign should fail")
	assert.Contains(t, string(output), "nested resources with root resource '04campaign' are not empty")

	err = shellcmd.Command("kubectl delete campaigns.workflow.symphony 04campaign-v-version3").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete activations.workflow.symphony activationlongrunning").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete campaigncontainers.workflow.symphony 04campaign").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogWithoutContainer(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCatalog)).CombinedOutput()
	assert.Contains(t, string(output), "rootResource must be a valid container")
	assert.NotNil(t, err, "catalog creation without container should fail")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	output, err = exec.Command("kubectl", "delete", "catalogcontainer.federation.symphony", "config").CombinedOutput()
	assert.NotNil(t, err, "catalog container deletion with catalog should fail")
	assert.Contains(t, string(output), "nested resources with root resource 'config' are not empty")
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogWithoutSchema(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogContainer))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCatalog)).CombinedOutput()
	assert.NotNil(t, err, "catalog creation without schema should fail")
	assert.Contains(t, string(output), "could not find the required schema")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)

	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogWithWrongSchema(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testWrongSchema)).CombinedOutput()
	assert.NotNil(t, err, "catalog creation without schema should fail")
	assert.Contains(t, string(output), "property does not match pattern: <email>")
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogWithoutParent(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testChildCatalog)).CombinedOutput()
	assert.Contains(t, string(output), "parent catalog not found")
	assert.NotNil(t, err, "catalog creation without parent should fail")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testChildCatalog))).Run()
	assert.Nil(t, err)
	output, err = exec.Command("kubectl", "delete", "catalog.federation.symphony", "config-v-version1").CombinedOutput()
	assert.Contains(t, string(output), "Catalog has one or more child catalogs. Update or Deletion is not allowed")
	assert.NotNil(t, err, "catalog deletion with child catalog should fail")
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-version3").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestUpdateCatalogWithCircularParentDependency(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularParentContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularParent))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularChildContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularChild))).Run()
	assert.Nil(t, err)

	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCircularParentUpdate)).CombinedOutput()
	assert.Contains(t, string(output), "parent catalog has circular dependency")
	assert.NotNil(t, err, "catalog upsert with circular parent dependency should fail")
}

func TestUpdateCatalogRemoveParentLabel(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testNoParentChild))).Run()
	assert.Nil(t, err)

	// Should be able to delete parent catalog, because child catalog has updated without parent catalog
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony parent-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony parent").Run()
	assert.Nil(t, err)

	err = shellcmd.Command("kubectl delete catalogs.federation.symphony child-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony child").Run()
	assert.Nil(t, err)
}

func IsTestInAzure() bool {
	// Check if the environment variable is set to "true" (case-insensitive)
	return strings.EqualFold(os.Getenv("AZURE_TEST"), "true")
}

func TestDiagnosticWithoutEdgeLocation(t *testing.T) {
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_01_WithoutEdgeLocation)).CombinedOutput()
	if IsTestInAzure() {
		assert.Contains(t, string(output), "metadata.annotations.management.azure.com/customLocation: Required value: Azure Edge Location is required")
		assert.NotNil(t, err, "diagnostic creation without edge location should fail")

		output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_02_WithCorrectEdgeLocation)).CombinedOutput()
		assert.Contains(t, string(output), "created")
		assert.Nil(t, err, "diagnostic creation with correct edge location should pass")
	} else {
		assert.Contains(t, string(output), "created")
		assert.Nil(t, err, "diagnostic creation without edge location should pass")
	}

	output, err = exec.Command("kubectl", "create", "ns", "default2").CombinedOutput()
	// ignore error if ns already exists
	if err != nil {
		assert.Contains(t, string(output), "already exists")
	} else {
		assert.Nil(t, err)
	}

	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_03_WithAnotherNS)).CombinedOutput()
	assert.Contains(t, string(output), "resource already exists in this cluster")
	assert.NotNil(t, err, "diagnostic creation with another namespace should fail")

	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_04_WithAnotherName)).CombinedOutput()
	assert.Contains(t, string(output), "resource already exists in this cluster")
	assert.NotNil(t, err, "diagnostic creation with name should fail")

	err = shellcmd.Command("kubectl delete diagnostics.monitor.symphony default").Run()
	assert.Nil(t, err)

	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_04_WithAnotherName)).CombinedOutput()
	assert.Contains(t, string(output), "created")
	assert.Nil(t, err)

	err = shellcmd.Command("kubectl delete diagnostics.monitor.symphony default2").Run()
	assert.Nil(t, err)

	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_03_WithAnotherNS)).CombinedOutput()
	assert.Contains(t, string(output), "created")
	assert.Nil(t, err)

	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_05_UpdateOtherAnnotations)).CombinedOutput()
	assert.Contains(t, string(output), "configured")
	assert.Nil(t, err)
}

func TestUpdateInstanceCreateInstanceHistory(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyTarget))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historySolution))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyInstance))).Run()
	assert.Nil(t, err)

	// wait until instance deployed
	for {
		output, err := exec.Command("kubectl", "get", "instances.solution.symphony", "history-instance", "-o", "jsonpath={.status.status}").CombinedOutput()
		if err != nil {
			assert.Fail(t, "failed to get instance %s state: %s", "history-instance", err.Error())
		}
		err = shellcmd.Command("kubectl get instances.solution.symphony history-instance").Run()
		assert.Nil(t, err)
		status := string(output)
		if status == "Succeeded" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historySolutionUpdate))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyInstanceUpdate))).Run()
	assert.Nil(t, err)

	// check instance history result
	output, err := exec.Command("kubectl", "get", "instancehistory", "-o", "json").CombinedOutput()
	if err != nil {
		assert.Fail(t, "failed to get instance %s state: %s", "history-instance", err.Error())
	}

	var historyList HistoryList
	err = json.Unmarshal(output, &historyList)
	if err != nil {
		assert.Fail(t, "failed to parse instance history", err.Error())
	}

	assert.Equal(t, 1, len(historyList.Items))
	history := historyList.Items[0]
	metadata := history["metadata"].(map[string]interface{})
	spec := history["spec"].(map[string]interface{})
	status := history["status"].(map[string]interface{})
	assert.True(t, strings.HasPrefix(metadata["name"].(string), "history-instance-v-"))
	assert.Equal(t, "history-instance", spec["rootResource"].(string))
	assert.Equal(t, "history-solution:version1", spec["solutionId"].(string))
	assert.Equal(t, "Succeeded", status["status"])
}

func TestUpdateInstanceHistory(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyCreate))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), historyUpdate)).CombinedOutput()
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(string(output), "Cannot update instance history spec because it is readonly"))
}

func getRepoPath() string {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../../"
	}
	return repoPath
}
