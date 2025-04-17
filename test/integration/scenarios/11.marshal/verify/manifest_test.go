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
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testSolution_WithEmptyComponentProperties = "test/integration/scenarios/11.marshal/manifest/empty-solution-properties.yaml"
	testTarget_WithEmptyComponentProperties   = "test/integration/scenarios/11.marshal/manifest/empty-target-properties.yaml"
)

func getRepoPath() string {
	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "../../../../../"
	}
	return repoPath
}

func TestSolution_WithEmptyComponentProperties(t *testing.T) {
	// first apply - generation 1
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testSolution_WithEmptyComponentProperties)).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// second apply - generation 1
	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testSolution_WithEmptyComponentProperties)).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// validating metadata.generation
	output, err = exec.Command("kubectl", "get", "solution", "empty-solution-v-version1", "-o=jsonpath='{.metadata.generation}'").CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.Equal(t, "'1'", string(output), fmt.Sprintf("Expected generation '1', got %s", string(output)))
}

func TestTarget_WithEmptyComponentProperties(t *testing.T) {
	// first apply - generation 1
	output, err := exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testTarget_WithEmptyComponentProperties)).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	time.Sleep(5 * time.Second) // wait for reconciliation happens (summary id will be bumpped up)

	// validating metadata.annotations.SummaryJobIdKey
	output, err = exec.Command("kubectl", "get", "target", "empty-target", "-o=jsonpath='{.metadata.annotations.SummaryJobIdKey}'").CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.Equal(t, "'1'", string(output), fmt.Sprintf("Expected job id '1', got %s", string(output)))

	// second apply - generation 1
	output, err = exec.Command("kubectl", "apply", "-f", path.Join(getRepoPath(), testTarget_WithEmptyComponentProperties)).CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))

	// validating metadata.generation
	output, err = exec.Command("kubectl", "get", "target", "empty-target", "-o=jsonpath='{.metadata.generation}'").CombinedOutput()
	assert.Nil(t, err, fmt.Sprintf("No error expected, got %s", string(output)))
	assert.Equal(t, "'1'", string(output), fmt.Sprintf("Expected generation '1', got %s", string(output)))
}
