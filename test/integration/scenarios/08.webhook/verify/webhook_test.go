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

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/assert"
)

var (
	testSolution = "test/integration/scenarios/01.update/manifestTemplates/oss/solutionversion-container.yaml"
	testSolutionVersion          = "test/integration/scenarios/01.update/manifestTemplates/oss/solutionversion.yaml"
	testTarget            = "test/integration/scenarios/01.update/manifestTemplates/oss/target.yaml"
	testInstance          = "test/integration/scenarios/01.update/manifestTemplates/oss/instance.yaml"
	testCampaign          = "test/integration/scenarios/04.workflow/manifest/campaign.yaml"
	testCampaignContainer = "test/integration/scenarios/04.workflow/manifest/campaign-container.yaml"

	testCampaignWithWrongStage     = "test/integration/scenarios/08.webhook/manifest/campaignWithWrongStages.yaml"
	testCampaignWithLongRunning    = "test/integration/scenarios/08.webhook/manifest/campaignLongRunning.yaml"
	testActivationsWithWrongStage  = "test/integration/scenarios/08.webhook/manifest/activationWithWrongStage.yaml"
	testActivationsWithLongRunning = "test/integration/scenarios/08.webhook/manifest/activationLongRunning.yaml"

	testCatalogVersion          = "test/integration/scenarios/05.catalogversionNconfigmap/manifests/config.yaml"
	testCatalog = "test/integration/scenarios/05.catalogversionNconfigmap/manifests/config-container.yaml"
	testSchema           = "test/integration/scenarios/05.catalogversionNconfigmap/manifests/schema.yaml"
	testSchemaContainer  = "test/integration/scenarios/05.catalogversionNconfigmap/manifests/schema-container.yaml"
	testWrongSchema      = "test/integration/scenarios/05.catalogversionNconfigmap/manifests/wrongconfig.yaml"
	testChildCatalogVersion     = "test/integration/scenarios/08.webhook/manifest/childCatalogVersion.yaml"

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
	historySolutionVersion       = "test/integration/scenarios/08.webhook/manifest/history-solutionversion.yaml"
	historyInstance       = "test/integration/scenarios/08.webhook/manifest/history-instance.yaml"
	historySolutionVersionUpdate = "test/integration/scenarios/08.webhook/manifest/history-solutionversion-update.yaml"
	historyInstanceUpdate = "test/integration/scenarios/08.webhook/manifest/history-instance-update.yaml"
)

var (
	solutionversionContainerFullName = "solutionversion01"
	solutionversionFullName          = "solutionversion01-v-version1"
	targetName                = "target01"
	instanceFullName          = "instance01"
)

// Define a struct to parse the JSON output
type HistoryList struct {
	Items []map[string]interface{} `json:"items"`
}

func TestPrepare(t *testing.T) {
	err := testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), testSolutionVersion), "target01", "solutionversion01", "version1", "instance01", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), testSolution), "target01", "solutionversion01", "version1", "instance01", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), testTarget), "target01", "solutionversion01", "version1", "instance01", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), testInstance), "target01", "solutionversion01", "version1", "instance01", "")
	assert.Nil(t, err)
	if testhelpers.IsTestInAzure() {
		solutionversionContainerFullName = "target01-v-solutionversion01"
		solutionversionFullName = solutionversionContainerFullName + "-v-version1"
		instanceFullName = solutionversionContainerFullName + "-v-instance01"
	}
}

func TestCreateSolutionVersionWithoutContainer(t *testing.T) {
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testSolutionVersion)).CombinedOutput()
	assert.Contains(t, string(output), "rootResource must be a valid container")
	assert.NotNil(t, err, "solutionversion creation without container should fail")
}

func TestInstanceWithoutSolutionVersion(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testTarget))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testInstance)).CombinedOutput()
	assert.Contains(t, string(output), "solutionversion does not exist")
	assert.NotNil(t, err, "instance creation without solutionversion should fail")
	err = shellcmd.Command("kubectl delete targets.fabric.symphony " + targetName).Run()
	assert.Nil(t, err)
}

func TestInstanceWithoutTarget(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolution))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolutionVersion))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testInstance)).CombinedOutput()
	assert.Contains(t, string(output), "target does not exist")
	assert.NotNil(t, err, "instance creation without target should fail")
	output, err = exec.Command("kubectl", "delete", "solutions.solutionversion.symphony", solutionversionContainerFullName).CombinedOutput()
	assert.Contains(t, string(output), "nested resources with root resource '"+solutionversionContainerFullName+"' are not empty")
	assert.NotNil(t, err, "solutionversion container deletion with solutionversion should fail")
	err = shellcmd.Command("kubectl delete solutionversions.solutionversion.symphony " + solutionversionFullName).Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutions.solutionversion.symphony " + solutionversionContainerFullName).Run()
	assert.Nil(t, err)
}

func TestTargetSolutionVersionDeletionWithInstance(t *testing.T) {
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testTarget))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolution))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSolutionVersion))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testInstance))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "delete", "solutionversions.solutionversion.symphony", solutionversionFullName).CombinedOutput()
	assert.Contains(t, string(output), "SolutionVersion has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err)
	output, err = exec.Command("kubectl", "delete", "targets.fabric.symphony", targetName).CombinedOutput()
	assert.Contains(t, string(output), "Target has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err, "target deletion with instance should fail")
	err = shellcmd.Command("kubectl delete instances.solutionversion.symphony " + instanceFullName).Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete targets.fabric.symphony " + targetName).Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutionversions.solutionversion.symphony " + solutionversionFullName).Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutions.solutionversion.symphony " + solutionversionContainerFullName).Run()
	assert.Nil(t, err)
}

func TestCreateActivationWithoutCampaign(t *testing.T) {
	// Already covered in 04.workflow tests
}

func TestCreateActivationWithWrongFirstStage(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
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
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCampaignContainer))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCampaignWithWrongStage)).CombinedOutput()
	assert.Contains(t, string(output), "stageSelector must be one of the stages in the stages list")
	assert.NotNil(t, err, "campaign creation with wrong stages should fail")
	err = shellcmd.Command("kubectl delete campaigncontainers.workflow.symphony 04campaign").Run()
	assert.Nil(t, err)
}

func TestDeleteCampaignWithRunningActivation(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
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

func TestCreateCatalogVersionWithoutContainer(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCatalogVersion)).CombinedOutput()
	assert.Contains(t, string(output), "rootResource must be a valid container")
	assert.NotNil(t, err, "catalogversion creation without container should fail")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogVersion))).Run()
	assert.Nil(t, err)
	output, err = exec.Command("kubectl", "delete", "catalog.federation.symphony", "config").CombinedOutput()
	assert.NotNil(t, err, "catalogversion container deletion with catalogversion should fail")
	assert.Contains(t, string(output), "nested resources with root resource 'config' are not empty")
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogVersionWithoutSchema(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCatalogVersion)).CombinedOutput()
	assert.NotNil(t, err, "catalogversion creation without schema should fail")
	assert.Contains(t, string(output), "could not find the required schema")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogVersion))).Run()
	assert.Nil(t, err)

	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogVersionWithWrongSchema(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testWrongSchema)).CombinedOutput()
	assert.NotNil(t, err, "catalogversion creation without schema should fail")
	assert.Contains(t, string(output), "property does not match pattern: <email>")
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestCreateCatalogVersionWithoutParent(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchemaContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalog))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testSchema))).Run()
	assert.Nil(t, err)
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testChildCatalogVersion)).CombinedOutput()
	assert.Contains(t, string(output), "parent catalogversion not found")
	assert.NotNil(t, err, "catalogversion creation without parent should fail")
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCatalogVersion))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testChildCatalogVersion))).Run()
	assert.Nil(t, err)
	output, err = exec.Command("kubectl", "delete", "catalogversion.federation.symphony", "config-v-version1").CombinedOutput()
	assert.Contains(t, string(output), "CatalogVersion has one or more child catalogversions. Update or Deletion is not allowed")
	assert.NotNil(t, err, "catalogversion deletion with child catalogversion should fail")
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony config-v-version3").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony config-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony schema-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func TestUpdateCatalogVersionWithCircularParentDependency(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularParentContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularParent))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularChildContainer))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testCircularChild))).Run()
	assert.Nil(t, err)

	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testCircularParentUpdate)).CombinedOutput()
	assert.Contains(t, string(output), "parent catalogversion has circular dependency")
	assert.NotNil(t, err, "catalogversion upsert with circular parent dependency should fail")
}

func TestUpdateCatalogVersionRemoveParentLabel(t *testing.T) {
	if testhelpers.IsTestInAzure() {
		return
	}
	err := shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), testNoParentChild))).Run()
	assert.Nil(t, err)

	// Should be able to delete parent catalogversion, because child catalogversion has updated without parent catalogversion
	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony parent-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony parent").Run()
	assert.Nil(t, err)

	err = shellcmd.Command("kubectl delete catalogversions.federation.symphony child-v-version1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony child").Run()
	assert.Nil(t, err)
}

func TestDiagnosticWithoutEdgeLocation(t *testing.T) {
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), diagnostic_01_WithoutEdgeLocation)).CombinedOutput()
	if testhelpers.IsTestInAzure() {
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
	err := testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), historyTarget), "history-target", "history-solutionversion", "version1", "history-instance", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), historySolutionVersion), "history-target", "history-solutionversion", "version1", "history-instance", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), historyInstance), "history-target", "history-solutionversion", "version1", "history-instance", "")
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyTarget))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historySolutionVersion))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyInstance))).Run()
	assert.Nil(t, err)

	instanceName := "history-instance"
	if testhelpers.IsTestInAzure() {
		instanceName = "history-target-v-history-solutionversion-v-history-instance"
	}
	// wait until instance deployed
	for {
		output, err := exec.Command("kubectl", "get", "instances.solutionversion.symphony", instanceName, "-o", "jsonpath={.status.status}").CombinedOutput()
		if err != nil {
			assert.Fail(t, "failed to get instance %s state: %s", instanceName, err.Error())
		}
		err = shellcmd.Command("kubectl get instances.solutionversion.symphony " + instanceName).Run()
		assert.Nil(t, err)
		status := string(output)
		if status == "Succeeded" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), historySolutionVersionUpdate), "history-target", "history-solutionversion", "version2", "history-instance", "")
	assert.Nil(t, err)
	err = testhelpers.ReplacePlaceHolderInManifest(path.Join(getRepoPath(), historyInstanceUpdate), "history-target", "history-solutionversion", "version2", "history-instance", "")
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historySolutionVersionUpdate))).Run()
	assert.Nil(t, err)
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f %s", path.Join(getRepoPath(), historyInstanceUpdate))).Run()
	assert.Nil(t, err)

	// check instance history result
	output, err := exec.Command("kubectl", "get", "instancehistory", "-o", "json").CombinedOutput()
	if err != nil {
		assert.Fail(t, "failed to get instance %s state: %s", instanceName, err.Error())
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
	assert.True(t, strings.HasPrefix(metadata["name"].(string), instanceName+"-v-"))
	assert.Equal(t, instanceName, spec["rootResource"].(string))
	solutionversionRef := "history-solutionversion:version1"
	if testhelpers.IsTestInAzure() {
		solutionversionRef = "/subscriptions/aaaa0a0a-bb1b-cc2c-dd3d-eeeeee4e4e4e/resourcegroups/test-rg/providers/microsoft.edge/targets/history-target/solutionversions/history-solutionversion/versions/version1"
	}
	assert.Equal(t, solutionversionRef, spec["solutionversionId"].(string))
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
