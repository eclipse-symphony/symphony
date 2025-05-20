/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/stretchr/testify/assert"
)

var (
	instance          = "test/integration/scenarios/12.naming/manifest/instance.yaml"
	solutionContainer = "test/integration/scenarios/12.naming/manifest/solution-container.yaml"
	solution          = "test/integration/scenarios/12.naming/manifest/solution.yaml"
	target            = "test/integration/scenarios/12.naming/manifest/target.yaml"
	instanceHistory   = "test/integration/scenarios/12.naming/manifest/instance-history.yaml"

	catalogcontainer = "test/integration/scenarios/12.naming/manifest/catalog-container.yaml"
	catalog          = "test/integration/scenarios/12.naming/manifest/catalog.yaml"

	campaign          = "test/integration/scenarios/12.naming/manifest/campaign.yaml"
	campaigncontainer = "test/integration/scenarios/12.naming/manifest/campaign-container.yaml"
	activation        = "test/integration/scenarios/12.naming/manifest/activation.yaml"

	diagnostic = "test/integration/scenarios/12.naming/manifest/diagnostic.yaml"

	longLength     = 65
	shortLength    = 1
	specialLength  = 10
	diaShortLength = 1
	diaLongLength  = 95
	labelLength    = 35
)

// generateRFC1123Subdomain generates a random string of the specified length
// conforming to the RFC 1123 subdomain validation rule.
func generateRandomName(length int, special bool) string {
	if length < 1 {
		panic("Length must be at least 1 to ensure start and end with alphanumeric characters")
	}

	var alphanumericCharset = "abcdeghijklmnopqrsuvwxyz"
	var middleCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
	var specialCharset = "!@#$%^&*()_=+[]{}|;:',<>?/"

	if special {
		middleCharset = specialCharset
	}
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)

	// Ensure the first character is alphanumeric
	b[0] = alphanumericCharset[seededRand.Intn(len(alphanumericCharset))]

	// Fill the middle characters with the allowed charset, avoiding consecutive dots or hyphens
	for i := 1; i < length-1; i++ {
		char := middleCharset[seededRand.Intn(len(middleCharset))]
		if (char == '.' && b[i-1] == '.') || (char == '-' && b[i-1] == '-') {
			// Avoid consecutive dots or hyphens by replacing with an alphanumeric character
			char = alphanumericCharset[seededRand.Intn(len(alphanumericCharset))]
		}
		b[i] = char
	}

	// Ensure the last character is alphanumeric
	b[length-1] = alphanumericCharset[seededRand.Intn(len(alphanumericCharset))]

	return string(b)
}

func getRepoPath() string {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../../"
	}
	return repoPath
}

func applyManifest(manifest []byte) ([]byte, error) {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("failed to apply manifest: %s, error: %w", string(output), err)
	}

	return output, nil
}

func createNonLinkedResource(file string, nameLength int, special bool) (string, []byte, error) {
	// read the manifest
	manifest, err := os.ReadFile(path.Join(getRepoPath(), file))
	if err != nil {
		return "", nil, fmt.Errorf("Failed to read manifest: %v", err)
	}
	// randomly generate a name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	resourceName := generateRandomName(nameLength, special) // Generate a random name with length characters
	manifest = []byte(strings.ReplaceAll(string(manifest), "${PLACEHOLDER_NAME}", resourceName))

	output, err := applyManifest(manifest)
	return resourceName, output, err
}

func createRootLinkedResource(file string, nameLength int, special bool, rootResource string) (string, []byte, error) {
	// read the manifest
	manifest, err := os.ReadFile(path.Join(getRepoPath(), file))
	if err != nil {
		return "", nil, fmt.Errorf("Failed to read manifest: %v", err)
	}
	// randomly generate a name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	resourceName := generateRandomName(nameLength, special) // Generate a random name with length characters
	manifest = []byte(strings.ReplaceAll(string(manifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", rootResource, resourceName)))
	manifest = []byte(strings.ReplaceAll(string(manifest), "${PLACEHOLDER_ROOT_RESOURCE}", rootResource))

	output, err := applyManifest(manifest)
	return resourceName, output, err
}

func createActivationResource(file string, nameLength int, special bool, campaignResource string) (string, []byte, error) {
	// read the manifest
	manifest, err := os.ReadFile(path.Join(getRepoPath(), file))
	if err != nil {
		return "", nil, fmt.Errorf("Failed to read manifest: %v", err)
	}
	// randomly generate a name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	resourceName := generateRandomName(nameLength, special) // Generate a random name with length characters
	manifest = []byte(strings.ReplaceAll(string(manifest), "${PLACEHOLDER_NAME}", resourceName))
	manifest = []byte(strings.ReplaceAll(string(manifest), "${PLACEHOLDER_CAMPAIGN_NAME}", campaignResource))

	output, err := applyManifest(manifest)
	return resourceName, output, err
}

func TestLongResourceName(t *testing.T) {
	targetName := generateRandomName(longLength, false) // Generate a random name with length characters
	solutionContainerName := generateRandomName(longLength, false)
	solutionName := generateRandomName(longLength, false)
	instanceName := generateRandomName(longLength, false)
	historyName := generateRandomName(longLength, false)
	// create target
	targetManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), target), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err := applyManifest([]byte(targetManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))

	// do the same for the solutioncontainer manifest
	sollutionContainerManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solutionContainer), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(sollutionContainerManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	outputString := strings.ToLower(string(output))
	assert.True(t, strings.Contains(outputString, "name length"))

	// do the same for the solution manifest
	solutionManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solution), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))

	// do the same for the instance manifest
	instanceManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instance), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(instanceManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))

	// do the same for the instance history manifest
	historyManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instanceHistory), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(historyManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))

	// do the same for the diagnostic manifest
	_, output, err = createNonLinkedResource(diagnostic, diaLongLength, false)
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))

	if testhelpers.IsTestInAzure() {
		// skip the campaign and catalog tests in azure
		return
	}
	// do the same for the catalog container manifest
	catalogContainerName, output, err := createNonLinkedResource(catalogcontainer, longLength, false)
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	outputString = strings.ToLower(string(output))
	assert.True(t, strings.Contains(outputString, "name length"))

	// do the same for the catalog manifest
	_, output, err = createRootLinkedResource(catalog, longLength, false, catalogContainerName)
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))

	// do the same for the campaign container manifest
	campaignContainerName, output, err := createNonLinkedResource(campaigncontainer, longLength, false)
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	outputString = strings.ToLower(string(output))
	assert.True(t, strings.Contains(outputString, "name length"))

	// do the same for the campaign manifest
	campaignName, output, err := createRootLinkedResource(campaign, longLength, false, campaignContainerName)
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))

	// do the same for the activation manifest
	_, output, err = createActivationResource(activation, longLength, false, fmt.Sprintf("%s:%s", campaignContainerName, campaignName))
	assert.NotNil(t, err, fmt.Sprintf("Error exepected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "Name length"))
}

func TestLabelLengthResourceName(t *testing.T) {
	targetName := generateRandomName(labelLength, false) // Generate a random name with length characters
	solutionContainerName := generateRandomName(labelLength, false)
	solutionName := generateRandomName(labelLength, false)
	instanceName := generateRandomName(labelLength, false)
	historyName := generateRandomName(labelLength, false)
	// create target
	targetManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), target), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err := applyManifest([]byte(targetManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solutionContainer), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionContainerManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the solution manifest
	solutionManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solution), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the instance manifest
	instanceManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instance), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(instanceManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the instance history manifest
	historyManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instanceHistory), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(historyManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the diagnostic manifest
	diaName, output, err := createNonLinkedResource(diagnostic, labelLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error exepected, got %s", string(output)))

	//delete Diagnostic k8s resource named diaName
	output, err = exec.Command("kubectl", "delete", "Diagnostic", diaName).Output()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	if testhelpers.IsTestInAzure() {
		// skip the campaign and catalog tests in azure
		return
	}
	// do the same for the catalog container manifest
	catalogContainerName, output, err := createNonLinkedResource(catalogcontainer, labelLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the catalog manifest
	_, output, err = createRootLinkedResource(catalog, labelLength, false, catalogContainerName)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign container manifest
	campaignContainerName, output, err := createNonLinkedResource(campaigncontainer, labelLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign manifest
	campaignName, output, err := createRootLinkedResource(campaign, labelLength, false, campaignContainerName)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the activation manifest
	_, output, err = createActivationResource(activation, labelLength, false, fmt.Sprintf("%s:%s", campaignContainerName, campaignName))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
}

func TestForShortResourceName(t *testing.T) {
	targetName := generateRandomName(shortLength, false) // Generate a random name with length characters
	solutionContainerName := generateRandomName(shortLength, false)
	solutionName := generateRandomName(shortLength, false)
	instanceName := generateRandomName(shortLength, false)
	historyName := generateRandomName(shortLength, false)
	// create target
	targetManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), target), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err := applyManifest([]byte(targetManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solutionContainer), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionContainerManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the solution manifest
	solutionManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solution), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the instance manifest
	instanceManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instance), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(instanceManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the instance history manifest
	historyManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instanceHistory), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(historyManifest))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the diagnostic manifest
	_, output, err = createNonLinkedResource(diagnostic, diaShortLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error exepected, got %s", string(output)))

	if testhelpers.IsTestInAzure() {
		// skip the campaign and catalog tests in azure
		return
	}
	// do the same for the catalog container manifest
	catalogContainerName, output, err := createNonLinkedResource(catalogcontainer, shortLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the catalog manifest
	_, output, err = createRootLinkedResource(catalog, shortLength, false, catalogContainerName)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign container manifest
	campaignContainerName, output, err := createNonLinkedResource(campaigncontainer, shortLength, false)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign manifest
	campaignName, output, err := createRootLinkedResource(campaign, shortLength, false, campaignContainerName)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the activation manifest
	_, output, err = createActivationResource(activation, shortLength, false, fmt.Sprintf("%s:%s", campaignContainerName, campaignName))
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
}

func TestForSpecialResourceName(t *testing.T) {
	targetName := generateRandomName(specialLength, true) // Generate a random name with length characters
	solutionContainerName := generateRandomName(specialLength, true)
	solutionName := generateRandomName(specialLength, true)
	instanceName := generateRandomName(specialLength, true)
	historyName := generateRandomName(specialLength, true)
	// create target
	targetManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), target), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err := applyManifest([]byte(targetManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solutionContainer), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionContainerManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the solution manifest
	solutionManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), solution), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(solutionManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the instance manifest
	instanceManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instance), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(instanceManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the instance history manifest
	historyManifest, err := testhelpers.ReplacePlaceHolderInManifestWithString(path.Join(getRepoPath(), instanceHistory), targetName, solutionContainerName, solutionName, instanceName, historyName)
	assert.Nil(t, err, "No error expected")
	output, err = applyManifest([]byte(historyManifest))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the diagnostic manifest
	_, output, err = createNonLinkedResource(diagnostic, specialLength, false)
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	if testhelpers.IsTestInAzure() {
		// skip the campaign and catalog tests in azure
		return
	}
	// do the same for the catalog container manifest
	catalogContainerName, output, err := createNonLinkedResource(catalogcontainer, specialLength, true)
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the catalog manifest

	_, output, err = createRootLinkedResource(catalog, specialLength, true, catalogContainerName)
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the campaign container manifest
	campaignContainerName, output, err := createNonLinkedResource(campaigncontainer, specialLength, true)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the campaign manifest
	campaignName, output, err := createRootLinkedResource(campaign, specialLength, true, campaignContainerName)
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the activation manifest
	_, output, err = createActivationResource(activation, specialLength, true, fmt.Sprintf("%s:%s", campaignContainerName, campaignName))
	assert.NotNil(t, err, fmt.Sprintf("Error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
}
