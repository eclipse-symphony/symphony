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

	longLength    = 65
	shortLength   = 3
	specialLength = 10
)

// generateRFC1123Subdomain generates a random string of the specified length
// conforming to the RFC 1123 subdomain validation rule.
func generateRandomName(length int, special bool) string {
	if length < 2 {
		panic("Length must be at least 2 to ensure start and end with alphanumeric characters")
	}

	var alphanumericCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
	var middleCharset = alphanumericCharset
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

func TestLongResourceName(t *testing.T) {
	// first apply - generation 1
	// read the solution manifest
	targetManifest, err := os.ReadFile(path.Join(getRepoPath(), target))
	if err != nil {
		t.Fatalf("Failed to read target manifest: %v", err)
	}
	// randomly generate a target name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	targetName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	targetManifest = []byte(strings.ReplaceAll(string(targetManifest), "${PLACEHOLDER_NAME}", targetName))

	output, err := applyManifest(targetManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), solutionContainer))
	if err != nil {
		t.Fatalf("Failed to read solution container manifest: %v", err)
	}
	// randomly generate a solution container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionContainerName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	solutionContainerManifest = []byte(strings.ReplaceAll(string(solutionContainerManifest), "${PLACEHOLDER_NAME}", solutionContainerName))
	output, err = applyManifest(solutionContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the solution manifest
	solutionManifest, err := os.ReadFile(path.Join(getRepoPath(), solution))
	if err != nil {
		t.Fatalf("Failed to read solution manifest: %v", err)
	}
	// randomly generate a solution name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", solutionContainerName, solutionName)))
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_ROOT_RESOURCE}", solutionContainerName))
	output, err = applyManifest(solutionManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the instance manifest
	instanceManifest, err := os.ReadFile(path.Join(getRepoPath(), instance))
	if err != nil {
		t.Fatalf("Failed to read instance manifest: %v", err)
	}
	// randomly generate a instance name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_NAME}", instanceName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_TARGET}", targetName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_SOLUTION}", fmt.Sprintf("%s:%s", solutionContainerName, solutionName)))
	output, err = applyManifest(instanceManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the instance history manifest
	instanceHistoryManifest, err := os.ReadFile(path.Join(getRepoPath(), instanceHistory))
	if err != nil {
		t.Fatalf("Failed to read instance history manifest: %v", err)
	}
	// randomly generate a instance history name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceHistoryName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", instanceName, instanceHistoryName)))
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_ROOT_RESOURCE}", instanceName))
	output, err = applyManifest(instanceHistoryManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the catalog container manifest
	catalogContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), catalogcontainer))
	if err != nil {
		t.Fatalf("Failed to read catalog container manifest: %v", err)
	}
	// randomly generate a catalog container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogContainerName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	catalogContainerManifest = []byte(strings.ReplaceAll(string(catalogContainerManifest), "${PLACEHOLDER_NAME}", catalogContainerName))
	output, err = applyManifest(catalogContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the catalog manifest
	catalogManifest, err := os.ReadFile(path.Join(getRepoPath(), catalog))
	if err != nil {
		t.Fatalf("Failed to read catalog manifest: %v", err)
	}
	// randomly generate a catalog name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", catalogContainerName, catalogName)))
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_ROOT_RESOURCE}", catalogContainerName))
	output, err = applyManifest(catalogManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign container manifest
	campaignContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), campaigncontainer))
	if err != nil {
		t.Fatalf("Failed to read campaign container manifest: %v", err)
	}
	// randomly generate a campaign container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignContainerName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	campaignContainerManifest = []byte(strings.ReplaceAll(string(campaignContainerManifest), "${PLACEHOLDER_NAME}", campaignContainerName))
	output, err = applyManifest(campaignContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign manifest
	campaignManifest, err := os.ReadFile(path.Join(getRepoPath(), campaign))
	if err != nil {
		t.Fatalf("Failed to read campaign manifest: %v", err)
	}
	// randomly generate a campaign name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", campaignContainerName, campaignName)))
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_ROOT_RESOURCE}", campaignContainerName))
	output, err = applyManifest(campaignManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the activation manifest
	activationManifest, err := os.ReadFile(path.Join(getRepoPath(), activation))
	if err != nil {
		t.Fatalf("Failed to read activation manifest: %v", err)
	}
	// randomly generate a activation name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	activationName := generateRandomName(longLength, false) // Generate a random name with longLength characters
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_NAME}", activationName))
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_CAMPAIGN_NAME}", fmt.Sprintf("%s:%s", campaignContainerName, campaignName)))
	output, err = applyManifest(activationManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
}

func TestForShortResourceName(t *testing.T) {
	// first apply - generation 1
	// read the solution manifest
	targetManifest, err := os.ReadFile(path.Join(getRepoPath(), target))
	if err != nil {
		t.Fatalf("Failed to read target manifest: %v", err)
	}
	// randomly generate a target name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	targetName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	targetManifest = []byte(strings.ReplaceAll(string(targetManifest), "${PLACEHOLDER_NAME}", targetName))

	output, err := applyManifest(targetManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), solutionContainer))
	if err != nil {
		t.Fatalf("Failed to read solution container manifest: %v", err)
	}
	// randomly generate a solution container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionContainerName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	solutionContainerManifest = []byte(strings.ReplaceAll(string(solutionContainerManifest), "${PLACEHOLDER_NAME}", solutionContainerName))
	output, err = applyManifest(solutionContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the solution manifest
	solutionManifest, err := os.ReadFile(path.Join(getRepoPath(), solution))
	if err != nil {
		t.Fatalf("Failed to read solution manifest: %v", err)
	}
	// randomly generate a solution name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", solutionContainerName, solutionName)))
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_ROOT_RESOURCE}", solutionContainerName))
	output, err = applyManifest(solutionManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the instance manifest
	instanceManifest, err := os.ReadFile(path.Join(getRepoPath(), instance))
	if err != nil {
		t.Fatalf("Failed to read instance manifest: %v", err)
	}
	// randomly generate a instance name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_NAME}", instanceName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_TARGET}", targetName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_SOLUTION}", fmt.Sprintf("%s:%s", solutionContainerName, solutionName)))
	output, err = applyManifest(instanceManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the instance history manifest
	instanceHistoryManifest, err := os.ReadFile(path.Join(getRepoPath(), instanceHistory))
	if err != nil {
		t.Fatalf("Failed to read instance history manifest: %v", err)
	}
	// randomly generate a instance history name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceHistoryName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", instanceName, instanceHistoryName)))
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_ROOT_RESOURCE}", instanceName))
	output, err = applyManifest(instanceHistoryManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// do the same for the catalog container manifest
	catalogContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), catalogcontainer))
	if err != nil {
		t.Fatalf("Failed to read catalog container manifest: %v", err)
	}
	// randomly generate a catalog container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogContainerName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	catalogContainerManifest = []byte(strings.ReplaceAll(string(catalogContainerManifest), "${PLACEHOLDER_NAME}", catalogContainerName))
	output, err = applyManifest(catalogContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the catalog manifest
	catalogManifest, err := os.ReadFile(path.Join(getRepoPath(), catalog))
	if err != nil {
		t.Fatalf("Failed to read catalog manifest: %v", err)
	}
	// randomly generate a catalog name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", catalogContainerName, catalogName)))
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_ROOT_RESOURCE}", catalogContainerName))
	output, err = applyManifest(catalogManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign container manifest
	campaignContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), campaigncontainer))
	if err != nil {
		t.Fatalf("Failed to read campaign container manifest: %v", err)
	}
	// randomly generate a campaign container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignContainerName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	campaignContainerManifest = []byte(strings.ReplaceAll(string(campaignContainerManifest), "${PLACEHOLDER_NAME}", campaignContainerName))
	output, err = applyManifest(campaignContainerManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the campaign manifest
	campaignManifest, err := os.ReadFile(path.Join(getRepoPath(), campaign))
	if err != nil {
		t.Fatalf("Failed to read campaign manifest: %v", err)
	}
	// randomly generate a campaign name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", campaignContainerName, campaignName)))
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_ROOT_RESOURCE}", campaignContainerName))
	output, err = applyManifest(campaignManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	// do the same for the activation manifest
	activationManifest, err := os.ReadFile(path.Join(getRepoPath(), activation))
	if err != nil {
		t.Fatalf("Failed to read activation manifest: %v", err)
	}
	// randomly generate a activation name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	activationName := generateRandomName(shortLength, false) // Generate a random name with shortLength characters
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_NAME}", activationName))
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_CAMPAIGN_NAME}", fmt.Sprintf("%s:%s", campaignContainerName, campaignName)))
	output, err = applyManifest(activationManifest)
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
}

func TestForSpecialResourceName(t *testing.T) {
	// first apply - generation 1
	// read the solution manifest
	targetManifest, err := os.ReadFile(path.Join(getRepoPath(), target))
	if err != nil {
		t.Fatalf("Failed to read target manifest: %v", err)
	}
	// randomly generate a target name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	targetName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	targetManifest = []byte(strings.ReplaceAll(string(targetManifest), "${PLACEHOLDER_NAME}", targetName))

	output, err := applyManifest(targetManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the solutioncontainer manifest
	solutionContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), solutionContainer))
	if err != nil {
		t.Fatalf("Failed to read solution container manifest: %v", err)
	}
	// randomly generate a solution container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionContainerName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	solutionContainerManifest = []byte(strings.ReplaceAll(string(solutionContainerManifest), "${PLACEHOLDER_NAME}", solutionContainerName))
	output, err = applyManifest(solutionContainerManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the solution manifest
	solutionManifest, err := os.ReadFile(path.Join(getRepoPath(), solution))
	if err != nil {
		t.Fatalf("Failed to read solution manifest: %v", err)
	}
	// randomly generate a solution name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	solutionName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", solutionContainerName, solutionName)))
	solutionManifest = []byte(strings.ReplaceAll(string(solutionManifest), "${PLACEHOLDER_ROOT_RESOURCE}", solutionContainerName))
	output, err = applyManifest(solutionManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the instance manifest
	instanceManifest, err := os.ReadFile(path.Join(getRepoPath(), instance))
	if err != nil {
		t.Fatalf("Failed to read instance manifest: %v", err)
	}
	// randomly generate a instance name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_NAME}", instanceName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_TARGET}", targetName))
	instanceManifest = []byte(strings.ReplaceAll(string(instanceManifest), "${PLACEHOLDER_SOLUTION}", fmt.Sprintf("%s:%s", solutionContainerName, solutionName)))
	output, err = applyManifest(instanceManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the instance history manifest
	instanceHistoryManifest, err := os.ReadFile(path.Join(getRepoPath(), instanceHistory))
	if err != nil {
		t.Fatalf("Failed to read instance history manifest: %v", err)
	}
	// randomly generate a instance history name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	instanceHistoryName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", instanceName, instanceHistoryName)))
	instanceHistoryManifest = []byte(strings.ReplaceAll(string(instanceHistoryManifest), "${PLACEHOLDER_ROOT_RESOURCE}", instanceName))
	output, err = applyManifest(instanceHistoryManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))

	// do the same for the catalog container manifest
	catalogContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), catalogcontainer))
	if err != nil {
		t.Fatalf("Failed to read catalog container manifest: %v", err)
	}
	// randomly generate a catalog container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogContainerName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	catalogContainerManifest = []byte(strings.ReplaceAll(string(catalogContainerManifest), "${PLACEHOLDER_NAME}", catalogContainerName))
	output, err = applyManifest(catalogContainerManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the catalog manifest
	catalogManifest, err := os.ReadFile(path.Join(getRepoPath(), catalog))
	if err != nil {
		t.Fatalf("Failed to read catalog manifest: %v", err)
	}
	// randomly generate a catalog name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	catalogName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", catalogContainerName, catalogName)))
	catalogManifest = []byte(strings.ReplaceAll(string(catalogManifest), "${PLACEHOLDER_ROOT_RESOURCE}", catalogContainerName))
	output, err = applyManifest(catalogManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the campaign container manifest
	campaignContainerManifest, err := os.ReadFile(path.Join(getRepoPath(), campaigncontainer))
	if err != nil {
		t.Fatalf("Failed to read campaign container manifest: %v", err)
	}
	// randomly generate a campaign container name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignContainerName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	campaignContainerManifest = []byte(strings.ReplaceAll(string(campaignContainerManifest), "${PLACEHOLDER_NAME}", campaignContainerName))
	output, err = applyManifest(campaignContainerManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the campaign manifest
	campaignManifest, err := os.ReadFile(path.Join(getRepoPath(), campaign))
	if err != nil {
		t.Fatalf("Failed to read campaign manifest: %v", err)
	}
	// randomly generate a campaign name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	campaignName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_NAME}", fmt.Sprintf("%s-v-%s", campaignContainerName, campaignName)))
	campaignManifest = []byte(strings.ReplaceAll(string(campaignManifest), "${PLACEHOLDER_ROOT_RESOURCE}", campaignContainerName))
	output, err = applyManifest(campaignManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
	// do the same for the activation manifest
	activationManifest, err := os.ReadFile(path.Join(getRepoPath(), activation))
	if err != nil {
		t.Fatalf("Failed to read activation manifest: %v", err)
	}
	// randomly generate a activation name with length as a param and replace ${PLACEHOLDER_NAME} with the actual name
	activationName := generateRandomName(specialLength, true) // Generate a random name with specialLength characters
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_NAME}", activationName))
	activationManifest = []byte(strings.ReplaceAll(string(activationManifest), "${PLACEHOLDER_CAMPAIGN_NAME}", fmt.Sprintf("%s:%s", campaignContainerName, campaignName)))
	output, err = applyManifest(activationManifest)
	assert.NotNil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.True(t, strings.Contains(string(output), "invalid"))
}
