/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

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

	testCatalog          = "test/integration/scenarios/05.catalog/catalogs/config.yaml"
	testCatalogContainer = "test/integration/scenarios/05.catalog/catalogs/config-container.yaml"
	testSchema           = "test/integration/scenarios/05.catalog/catalogs/schema.yaml"
	testSchemaContainer  = "test/integration/scenarios/05.catalog/catalogs/schema-container.yaml"
	testWrongSchema      = "test/integration/scenarios/05.catalog/catalogs/wrongconfig.yaml"
	testChildCatalog     = "test/integration/scenarios/08.webhook/manifest/childCatalog.yaml"
)

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
	err = shellcmd.Command("kubectl delete solutions.solution.symphony mysol-v-v1").Run()
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
	output, err := exec.Command("kubectl", "delete", "solutions.solution.symphony", "mysol-v-v1").CombinedOutput()
	assert.Contains(t, string(output), "Solution has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err)
	output, err = exec.Command("kubectl", "delete", "targets.fabric.symphony", "self").CombinedOutput()
	assert.Contains(t, string(output), "Target has one or more associated instances. Deletion is not allowed.")
	assert.NotNil(t, err, "target deletion with instance should fail")
	err = shellcmd.Command("kubectl delete instances.solution.symphony instance").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete targets.fabric.symphony self").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete solutions.solution.symphony mysol-v-v1").Run()
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
	err = shellcmd.Command("kubectl delete campaigns.workflow.symphony 04campaign-v-v1").Run()
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
	output, err := exec.Command("kubectl", "delete", "campaigns.workflow.symphony", "04campaign-v-v3").CombinedOutput()
	assert.Contains(t, string(output), "Campaign has one or more running activations. Update or Deletion is not allowed")
	assert.NotNil(t, err, "campaign deletion with running activation should fail")
	err = shellcmd.Command("kubectl delete activations.workflow.symphony activationlongrunning").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete campaigns.workflow.symphony 04campaign-v-v3").Run()
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
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-v1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-v1").Run()
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

	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-v1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-v1").Run()
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
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-v1").Run()
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
	output, err = exec.Command("kubectl", "delete", "catalog.federation.symphony", "config-v-v1").CombinedOutput()
	assert.Contains(t, string(output), "Catalog has one or more child catalogs. Update or Deletion is not allowed")
	assert.NotNil(t, err, "catalog deletion with child catalog should fail")
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-v3").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony config-v-v1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony config").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogs.federation.symphony schema-v-v1").Run()
	assert.Nil(t, err)
	err = shellcmd.Command("kubectl delete catalogcontainers.federation.symphony schema").Run()
	assert.Nil(t, err)
}

func getRepoPath() string {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../../"
	}
	return repoPath
}
