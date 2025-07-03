/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/test/integration/lib/testhelpers"
	"github.com/princjef/mageutil/shellcmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var (
	// Global Kubernetes client configuration and dynamic client
	cfg *rest.Config
	dyn dynamic.Interface

	// Sample workflow manifests to deploy
	testCampaign           = "../manifest/campaign.yaml"
	campaigncontainer      = "../manifest/campaign-container.yaml"
	testActivation         = "../manifest/activation.yaml"
	helmTarget             = "../manifest/target.yaml"
	solutionContainer      = "../manifest/solution-container.yaml"
	solutionSuccess        = "../manifest/successSolution.yaml"
	solutionReconcileError = "../manifest/reconcileErrorSolution.yaml"

	namespace = "default"
)

// YamlDocument represents a single YAML document structure
type YamlDocument struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Annotations map[string]string `yaml:"annotations,omitempty"`
	} `yaml:"metadata"`
	Spec map[string]interface{} `yaml:"spec"`
}

// TaskConfig represents a task configuration in a stage
type TaskConfig struct {
	Name     string                 `yaml:"name"`
	Provider string                 `yaml:"provider"`
	Target   string                 `yaml:"target"`
	Config   map[string]interface{} `yaml:"config"`
	Inputs   TaskInputs             `yaml:"inputs"`
}

// TaskInputs represents the inputs for a task
type TaskInputs struct {
	Action     string     `yaml:"action"`
	Object     TaskObject `yaml:"object"`
	ObjectName string     `yaml:"objectName"`
	ObjectType string     `yaml:"objectType"`
}

// TaskObject represents the object being created/managed by a task
type TaskObject struct {
	Metadata TaskMetadata `yaml:"metadata"`
	Spec     TaskSpec     `yaml:"spec"`
}

// TaskMetadata represents the metadata of a task object
type TaskMetadata struct {
	Name string `yaml:"name"`
}

// TaskSpec represents the spec of a task object
type TaskSpec struct {
	Solution string `yaml:"solution"`
	Scope    string `yaml:"scope,omitempty"` // Optional field for solution scope
}

// StageConfig represents a stage configuration
type StageConfig struct {
	Name          string                 `yaml:"name"`
	TaskOption    TaskOption             `yaml:"taskOption,omitempty"`
	Tasks         []TaskConfig           `yaml:"tasks,omitempty"`
	Target        string                 `yaml:"target,omitempty"` // Optional field for stage target
	Provider      string                 `yaml:"provider,omitempty"`
	Config        map[string]interface{} `yaml:"config,omitempty"`
	Inputs        TaskInputs             `yaml:"inputs,omitempty"`
	StageSelector string                 `yaml:"stageSelector,omitempty"` // Optional field for stage selector
}

// TaskOption represents task execution options
type TaskOption struct {
	Concurrency int         `yaml:"concurrency"`
	ErrorAction ErrorAction `yaml:"errorAction"`
}

// ErrorAction represents error handling configuration
type ErrorAction struct {
	Mode                 string `yaml:"mode"`
	MaxToleratedFailures int    `yaml:"maxToleratedFailures,omitempty"`
}

// readYamlFile reads a YAML file
func readYamlFile(filePath string) (YamlDocument, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return YamlDocument{}, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	var yamlDoc YamlDocument
	err = yaml.Unmarshal(data, &yamlDoc)
	if err != nil {
		return YamlDocument{}, fmt.Errorf("failed to parse YAML document: %v", err)
	}

	return yamlDoc, nil
}

func readTarget(filePath string, name string) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = name

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	if doc.Spec == nil {
		doc.Spec = make(map[string]interface{})
	}
	// Update solution scope in spec if it exists
	doc.Spec["solutionScope"] = name

	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

func readSolutionContainer(filePath string, name string, targetName string) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = fmt.Sprintf("%s-v-%s", targetName, name)

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s", targetName, name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

func readSolution(filePath string, name string, targetName string, containerName string) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = fmt.Sprintf("%s-v-%s-v-%s", targetName, containerName, name)

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName, containerName, name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	if doc.Spec == nil {
		doc.Spec = make(map[string]interface{})
	}
	// Update root resource in spec if it exists
	doc.Spec["rootResource"] = fmt.Sprintf("%s-v-%s", targetName, containerName)

	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

func readCampaignContainer(filePath string, name string) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = fmt.Sprintf("%s-v-%s", "context1", name)

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/contexts/context1/workflows/%s", name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

func readActivation(filePath, name, campaignContainerName, campaignName string) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = fmt.Sprintf("context1-v-%s-v-%s-v-%s", campaignContainerName, campaignName, name)

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/contexts/context1/workflows/%s/versions/%s/executions/%s", campaignContainerName, campaignName, name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	if doc.Spec == nil {
		doc.Spec = make(map[string]interface{})
	}
	if testhelpers.IsTestInAzure() {
		doc.Spec["campaign"] = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/contexts/context1/workflows/%s/versions/%s", campaignContainerName, campaignName)
	} else {
		doc.Spec["campaign"] = fmt.Sprintf("context1-v-%s:%s", campaignContainerName, campaignName)
	}
	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

func readCampaign(filePath, name, campaignContainerName string, spec map[string]interface{}) (string, error) {
	doc, err := readYamlFile(filePath)
	if err != nil {
		return "", err
	}

	// Modify the fields with the new name
	doc.Metadata.Name = fmt.Sprintf("context1-v-%s-v-%s", campaignContainerName, name)

	// Initialize annotations if nil
	if doc.Metadata.Annotations == nil {
		doc.Metadata.Annotations = make(map[string]string)
	}

	// Update the resource ID annotation
	resourceId := fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/contexts/context1/workflows/%s/versions/%s", campaignContainerName, name)
	doc.Metadata.Annotations["management.azure.com/resourceId"] = resourceId

	doc.Spec = spec
	// Marshal the modified document back to YAML
	modifiedDoc, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal modified YAML document: %v", err)
	}

	// Join all documents back together with document separators
	return string(modifiedDoc), nil
}

// DeployManifests deploys the specified manifests to the given namespace
func TestParallelErrorContinue(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	solutionName1 := "parallelerrorcontinue1"
	solutionName2 := "parallelerrorcontinue2"
	targetName1 := "target1"
	targetName2 := "target2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelerrorcontinue"
	aName := "pecactivation"
	instanceName1 := "pecinstance1"
	instanceName2 := "pecinstance2"
	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	// Intentionally set to an invalid provider to test
	task1.Provider = "providers.stage.invalid"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1

	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2

	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, stName1, solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, stName2, solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, stName1, solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, stName2, solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 2,
				ErrorAction: ErrorAction{
					Mode: "silentlyContinue",
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./test.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid stage provider and error mode as "silentlyContinue"
	// As a result, we would see overall activation status as "Done", stage status as "Done"
	// The output status would be 200, task1 status would be 1000 with error message and task2 status would be 200
	// Task1 would not have failedDeploymentCount as it is not executed
	// Task2 would have failedDeploymentCount as 0 as it is executed successfully
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "Bad Config: task provider providers.stage.invalid is not found")
	require.Equal(t, float64(1000), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}

func TestParallelErrorStop(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelerrorstop1"
	solutionName2 := "parallelerrorstop2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelerrorstop"
	aName := "pesactivation"
	instanceName1 := "pesinstance1"
	instanceName2 := "pesinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	// Intentionally set to an invalid provider to test
	task1.Provider = "providers.stage.invalid"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1

	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2

	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode: "stopOnAnyFailure",
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./test.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid stage provider and error mode as "stopOnAnyFailure", concurrency as 1, maxToleratedFailures as 0
	// As a result, we would see overall activation status as "InternalError", stage status as "InternalError"
	// The output status would be 500, task1 status would be 1000 with error message and task2 status would be nil
	// Task1 would not have failedDeploymentCount as it is not executed
	require.Equal(t, v1alpha2.InternalError, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.InternalError, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(500), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "Bad Config: task provider providers.stage.invalid is not found")
	require.Equal(t, float64(1000), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Nil(t, state.Status.StageHistory[0].Outputs["task2"])
}

func TestParallelReconcileContinue(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelreconcilecontinue1"
	solutionName2 := "parallelreconcilecontinue2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelreconcilecontinue"
	aName := "prcactivation"
	instanceName1 := "prcinstance1"
	instanceName2 := "prcinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	// solution 1 is with invalid chart repo
	absSolution := filepath.Join(repoPath, solutionReconcileError)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode:                 "silentlyContinue",
					MaxToleratedFailures: 1,
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid chart repo and error mode as "silentlyContinue", concurrency as 1, maxToleratedFailures as 1
	// As a result, we would see overall activation status as "Done", stage status as "Done"
	// The output status would be 200, task1 status would be 400
	// with error message and task2 status would be 200
	// Task1 would have failedDeploymentCount as 1 as it is executed and failed
	// Task2 would have failedDeploymentCount as 0 as it is executed successfully
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "failed to pull char")
	require.Equal(t, float64(400), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(1), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["failedDeploymentCount"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}

func TestParallelReconcileStop(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelreconcilestop1"
	solutionName2 := "parallelreconcilestop2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelreconcilestop"
	aName := "prsactivation"
	instanceName1 := "prsinstance1"
	instanceName2 := "prsinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	// solution 1 is with invalid chart repo
	absSolution := filepath.Join(repoPath, solutionReconcileError)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode:                 "stopOnNFailures",
					MaxToleratedFailures: 0,
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid chart repo and error mode as "stopOnNFailures", concurrency as 1, maxToleratedFailures as 0
	// As a result, we should see overall activation status as "InternalServerError", stage status as "InternalServerError"
	// The output status would be 500, task1 status would be 400
	// However, the overall status is Done and stage status is also Done, output status is 200 and task2 is also executed
	// I think this should be a bug
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "failed to pull char")
	require.Equal(t, float64(400), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(1), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["failedDeploymentCount"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}

func TestParallelStopN(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelstopn1"
	solutionName2 := "parallelstopn2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelstopn"
	aName := "psnactivation"
	instanceName1 := "psninstance1"
	instanceName2 := "psninstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionReconcileError)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	// Set the provider to an invalid one to simulate a failure
	task1.Provider = "providers.stage.invalid"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode:                 "stopOnNFailures",
					MaxToleratedFailures: 0,
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid stage provider and error mode as "stopOnNFailures", concurrency as 1, maxToleratedFailures as 0
	// As a result, we would see overall activation status as "InternalServerError", stage status as "InternalServerError"
	// The output status would be 500, task1 status would be 400
	// with error message and task2 status would be nil
	require.Equal(t, v1alpha2.InternalError, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.InternalError, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(500), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "providers.stage.invalid is not found")
	require.Equal(t, float64(1000), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	assert.Nil(t, state.Status.StageHistory[0].Outputs["task2"])
}

func TestParallelStopN1(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelstop1n1"
	solutionName2 := "parallelstop1n2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelstop1n"
	aName := "psn1activation"
	instanceName1 := "psn1instance1"
	instanceName2 := "psn1instance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	// solution 1 is with invalid chart repo
	absSolution := filepath.Join(repoPath, solutionReconcileError)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	// Set the provider to an invalid one to simulate a failure
	task1.Provider = "providers.stage.invalid"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode:                 "stopOnNFailures",
					MaxToleratedFailures: 1,
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with task1 failed with invalid stage provider and error mode as "stopOnNFailures", concurrency as 1, maxToleratedFailures as 1
	// As a result, we would see overall activation status as "Done", stage status as "Done"
	// The output status would be 200, task1 status would be 1000 with error message and task2 status would be 200
	// Task2 should be executed as maxToleratedFailures is 1 and faileDeploymentCount is 0
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["error"], "providers.stage.invalid is not found")
	require.Equal(t, float64(1000), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}

func TestParallelSuccess(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallelsuccess"
	solutionName2 := "parallelsuccess2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallelsuccess"
	aName := "psactivation"
	instanceName1 := "psinstance1"
	instanceName2 := "psinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 1,
				ErrorAction: ErrorAction{
					Mode: "silentlyContinue",
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with both task1 and task2 executed successfully
	// As a result, we would see overall activation status as "Done", stage status as "Done"
	// The output status would be 200, task1 status would be 200
	// and task2 status would be 200
	// failedDeploymentCount would be 0 for both tasks
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["failedDeploymentCount"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}

func TestParallel2Success(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "parallel2success"
	solutionName2 := "parallel2success2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "parallel2success"
	aName := "p2sactivation"
	instanceName1 := "p2sinstance1"
	instanceName2 := "p2sinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	task1 := generateTaskConfig()
	task1.Name = "task1"
	task1.Inputs.Object.Metadata.Name = instanceName1
	task1.Inputs.Object.Spec.Scope = instanceName1
	task1.Inputs.ObjectName = instanceName1
	task2 := generateTaskConfig()
	task2.Name = "task2"
	task2.Inputs.Object.Metadata.Name = instanceName2
	task2.Inputs.Object.Spec.Scope = instanceName2
	task2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each task
	if testhelpers.IsTestInAzure() {
		task1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		task2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		task1.Target = targetName1
		task1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		task2.Target = targetName2
		task2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}
	stages := map[string]StageConfig{
		"stage1": {
			Name: "stage1",
			TaskOption: TaskOption{
				Concurrency: 2,
				ErrorAction: ErrorAction{
					Mode: "silentlyContinue",
				},
			},
			Tasks: []TaskConfig{task1, task2},
		},
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an parallel execution with both task1 and task2 executed successfully with concurrency as 2
	// As a result, we would see overall activation status as "Done", stage status as "Done"
	// The output status would be 200, task1 status would be 200
	// and task2 status would be 200
	// failedDeploymentCount would be 0 for both tasks
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task1"].(map[string]interface{})["failedDeploymentCount"])
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["task2"].(map[string]interface{})["failedDeploymentCount"])
}
func TestStageReconcileTimeout(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "stagereconciletimeout1"
	solutionName2 := "stagereconciletimeout2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "stagereconciletimeout"
	aName := "srtactivation"
	instanceName1 := "srtinstance1"
	instanceName2 := "srtinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	// This solution will have a reconcile error
	absSolution := filepath.Join(repoPath, solutionReconcileError)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	absSolution = filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	stage1 := generateStageConfig()
	stage1.Name = "stage1"
	stage1.Inputs.Object.Metadata.Name = instanceName1
	stage1.Inputs.Object.Spec.Scope = instanceName1
	stage1.Inputs.ObjectName = instanceName1
	stage1.StageSelector = "stage2"
	stage2 := generateStageConfig()
	stage2.Name = "stage2"
	stage2.Inputs.Object.Metadata.Name = instanceName2
	stage2.Inputs.Object.Spec.Scope = instanceName2
	stage2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each stage
	if testhelpers.IsTestInAzure() {
		stage1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		stage2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		stage1.Target = targetName1
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		stage2.Target = targetName2
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}

	stages := map[string]StageConfig{
		"stage1": stage1,
		"stage2": stage2,
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an sequential execution with stage1 and stage2 executed failed
	// As a result, we would see overall activation status as "InternalError", stage1 status as "InternalError"
	// The output status would be 500, stage1 status would be 400
	// stage length should be 1 as stage 2 will not be executed
	require.Equal(t, v1alpha2.InternalError, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.InternalError, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(400), state.Status.StageHistory[0].Outputs["status"])
	require.Equal(t, float64(1), state.Status.StageHistory[0].Outputs["failedDeploymentCount"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["error"], "failed to pull chart")
}

func TestStageError(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "stageerror1"
	solutionName2 := "stageerror2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "stageerror"
	aName := "seactivation"
	instanceName1 := "seinstance1"
	instanceName2 := "seinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	stage1 := generateStageConfig()
	stage1.Name = "stage1"
	stage1.Inputs.Object.Metadata.Name = instanceName1
	stage1.Inputs.Object.Spec.Scope = instanceName1
	stage1.Inputs.ObjectName = instanceName1
	stage1.StageSelector = "stage2"
	stage2 := generateStageConfig()
	stage2.Name = "stage2"
	stage2.Inputs.Object.Metadata.Name = instanceName2
	stage2.Inputs.Object.Spec.Scope = instanceName2
	stage2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each stage
	// stage1 will fail with solution not found error as we are using an invalid solution version
	if testhelpers.IsTestInAzure() {
		stage1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", "invalid")
		stage2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		stage1.Target = targetName1
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", "invalid")
		stage2.Target = targetName2
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}

	stages := map[string]StageConfig{
		"stage1": stage1,
		"stage2": stage2,
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an sequential execution with stage1 and stage2 executed failed
	// As a result, we would see overall activation status as "InternalError", stage1 status as "InternalError"
	// The output status would be 500, stage1 status would be 400
	// stage length should be 1 as stage 2 will not be executed
	require.Equal(t, v1alpha2.InternalError, state.Status.Status)
	require.Equal(t, 1, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.InternalError, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(400), state.Status.StageHistory[0].Outputs["status"])
	require.Contains(t, state.Status.StageHistory[0].Outputs["error"], "solution does not exist")
}

func TestStageSuccess(t *testing.T) {
	var err error
	cfg, err = testhelpers.RestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to get REST config: %v", err))
	}

	dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create dynamic client: %v", err))
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		// Get current working directory and navigate to repository root
		wd, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get working directory: %v", err))
		}
		// Navigate up from verify directory to the repository root
		repoPath = wd
	}
	targetName1 := "target1"
	targetName2 := "target2"
	solutionName1 := "stagesuccess1"
	solutionName2 := "stagesuccess2"
	stName1 := "scontainer"
	stName2 := "scontainer2"
	ccName := "ccontainer"
	cName := "stagesuccess"
	aName := "ssactivation"
	instanceName1 := "ssinstance1"
	instanceName2 := "ssinstance2"

	// create target
	absTarget := filepath.Join(repoPath, helmTarget)
	output, err := readTarget(absTarget, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	output, err = readTarget(absTarget, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read target manifest %s: %s", absTarget, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absTarget, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absTarget, output))
	os.Remove("./test.yaml")

	// create solution container
	absSolutionContainer := filepath.Join(repoPath, solutionContainer)
	output, err = readSolutionContainer(absSolutionContainer, stName1, targetName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	output, err = readSolutionContainer(absSolutionContainer, stName2, targetName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read st manifest %s: %s", absSolutionContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolutionContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolutionContainer, output))
	os.Remove("./test.yaml")

	// create solution
	absSolution := filepath.Join(repoPath, solutionSuccess)
	output, err = readSolution(absSolution, solutionName1, targetName1, stName1)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write solution manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy solution manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	output, err = readSolution(absSolution, solutionName2, targetName2, stName2)
	assert.Nil(t, err, fmt.Sprintf("Failed to read solution manifest %s: %s", absSolution, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absSolution, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absSolution, output))
	os.Remove("./test.yaml")

	// create campaign container
	absCampaignContainer := filepath.Join(repoPath, campaigncontainer)
	output, err = readCampaignContainer(absCampaignContainer, ccName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read cc manifest %s: %s", absCampaignContainer, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write target manifest %s: %s", absCampaignContainer, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy target manifest %s: %s", absCampaignContainer, output))
	os.Remove("./test.yaml")

	absCampaign := filepath.Join(repoPath, testCampaign)
	// Create task configurations using structs
	stage1 := generateStageConfig()
	stage1.Name = "stage1"
	stage1.Inputs.Object.Metadata.Name = instanceName1
	stage1.Inputs.Object.Spec.Scope = instanceName1
	stage1.Inputs.ObjectName = instanceName1
	stage1.StageSelector = "stage2"
	stage2 := generateStageConfig()
	stage2.Name = "stage2"
	stage2.Inputs.Object.Metadata.Name = instanceName2
	stage2.Inputs.Object.Spec.Scope = instanceName2
	stage2.Inputs.ObjectName = instanceName2
	// Set the target and solution for each stage
	if testhelpers.IsTestInAzure() {
		stage1.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName1)
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName1, "scontainer", solutionName1)
		stage2.Target = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s", targetName2)
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/%s/solutions/%s/versions/%s", targetName2, "scontainer2", solutionName2)
	} else {
		stage1.Target = targetName1
		stage1.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName1, "scontainer", solutionName1)
		stage2.Target = targetName2
		stage2.Inputs.Object.Spec.Solution = fmt.Sprintf("%s-v-%s:%s", targetName2, "scontainer2", solutionName2)
	}

	stages := map[string]StageConfig{
		"stage1": stage1,
		"stage2": stage2,
	}
	spec := map[string]interface{}{
		"rootResource": fmt.Sprintf("context1-v-%s", ccName),
		"firstStage":   "stage1",
		"selfDriving":  true,
		"stages":       stages,
	}

	// Create campaign
	output, err = readCampaign(absCampaign, cName, ccName, spec)
	assert.Nil(t, err, fmt.Sprintf("Failed to read campaign manifest %s: %s", absCampaign, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./campaigntest.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write campaign manifest %s: %s", absCampaign, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./campaigntest.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy campaign manifest %s: %s", absCampaign, output))
	os.Remove("./campaigntest.yaml")

	// Create activation
	absActivation := filepath.Join(repoPath, testActivation)
	output, err = readActivation(absActivation, aName, ccName, cName)
	assert.Nil(t, err, fmt.Sprintf("Failed to read activation manifest %s: %s", absActivation, output))
	err = testhelpers.WriteYamlStringsToFile(output, "./test.yaml")
	assert.Nil(t, err, fmt.Sprintf("Failed to write activation manifest %s: %s", absActivation, output))
	err = shellcmd.Command(fmt.Sprintf("kubectl apply -f ./test.yaml -n %s", namespace)).Run()
	assert.Nil(t, err, fmt.Sprintf("Failed to deploy activation manifest %s: %s", absActivation, output))
	os.Remove("./test.yaml")

	// Wait for the activation to complete
	state, err := waitForActivationToComplete(namespace, fmt.Sprintf("%s-v-%s-v-%s-v-%s", "context1", ccName, cName, aName), 60*5)
	assert.Nil(t, err, fmt.Sprintf("Failed to wait for activation to complete: %s", err))

	// This is an sequential execution with stage1 and stage2 executed successfully
	// As a result, we would see overall activation status as "Done", stage1 status as "Done"
	// The output status would be 200, stage1 status would be 200
	// stage length should be 2 as both stages are executed successfully
	// failedDeploymentCount should be 0 as both stages are executed successfully
	require.Equal(t, v1alpha2.Done, state.Status.Status)
	require.Equal(t, 2, len(state.Status.StageHistory))
	require.Equal(t, "stage1", state.Status.StageHistory[0].Stage)
	require.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	require.Equal(t, float64(200), state.Status.StageHistory[0].Outputs["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[0].Outputs["failedDeploymentCount"])
	require.Equal(t, float64(200), state.Status.StageHistory[1].Outputs["status"])
	require.Equal(t, float64(0), state.Status.StageHistory[1].Outputs["failedDeploymentCount"])
}

func generateTaskConfig() TaskConfig {
	return TaskConfig{
		Name:     "task1",
		Provider: "providers.stage.create",
		Target:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/target111",
		Config: map[string]interface{}{
			"wait.count":    12,
			"wait.interval": 5,
		},
		Inputs: TaskInputs{
			Action: "create",
			Object: TaskObject{
				Metadata: TaskMetadata{
					Name: "instance",
				},
				Spec: TaskSpec{
					Solution: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/target111/solutions/sol1/versions/version1",
				},
			},
			ObjectName: "instance",
			ObjectType: "instance",
		},
	}
}

func generateStageConfig() StageConfig {
	return StageConfig{
		Name:     "stage1",
		Provider: "providers.stage.create",
		Target:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/target111",
		Config: map[string]interface{}{
			"wait.count":    12,
			"wait.interval": 5,
		},
		Inputs: TaskInputs{
			Action: "create",
			Object: TaskObject{
				Metadata: TaskMetadata{
					Name: "instance",
				},
				Spec: TaskSpec{
					Solution: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Edge/targets/target111/solutions/sol1/versions/version1",
				},
			},
			ObjectName: "instance",
			ObjectType: "instance",
		},
	}
}

func waitForActivationToComplete(namespace string, activationName string, timeout int) (model.ActivationState, error) {
	timeoutDuration := time.Duration(timeout) * time.Second
	deadline := time.Now().Add(timeoutDuration)
	sleepDuration := 30 * time.Second
	var state model.ActivationState
	for time.Now().Before(deadline) {
		resource, err := dyn.Resource(schema.GroupVersionResource{
			Group:    "workflow.symphony",
			Version:  "v1",
			Resource: "activations",
		}).Namespace(namespace).Get(context.Background(), activationName, metav1.GetOptions{})
		if err != nil {
			return model.ActivationState{}, fmt.Errorf("failed to get activation: %v", err)
		}

		if resource == nil {
			return model.ActivationState{}, fmt.Errorf("activation named '%s' not found", activationName)
		}

		bytes, _ := json.Marshal(resource.Object)
		err = json.Unmarshal(bytes, &state)
		if err != nil {
			return model.ActivationState{}, fmt.Errorf("failed to unmarshal activation status: %v", err)
		}
		if state.Status != nil {
			status := state.Status.Status
			fmt.Printf("Current activation status: %s\n", status)
			// Check if status is not the zero value and is one of the terminal states
			if status == v1alpha2.Done || status == v1alpha2.BadRequest || status == v1alpha2.InternalError {
				return state, nil
			}
		}

		// Check if we have enough time left for another sleep cycle
		if time.Now().Add(sleepDuration).After(deadline) {
			break
		}

		time.Sleep(sleepDuration)
	}

	return model.ActivationState{}, fmt.Errorf("timeout after %d seconds waiting for activation '%s' to complete", timeout, activationName)
}
