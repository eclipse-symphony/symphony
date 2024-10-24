/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package testhelpers

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// BuildManifestFile modifies the target/solution manifest files
func BuildManifestFile(inputFolderPath string, outputFolderPath string, targetType string, components []string) error {
	inputFilePath := fmt.Sprintf("%s/%s.yaml", inputFolderPath, targetType)
	outputFilePath := fmt.Sprintf("%s/%s.yaml", outputFolderPath, targetType)

	// Read the YAML file
	data, err := os.ReadFile(inputFilePath)
	if err != nil {
		return err
	}

	var newData []byte
	switch targetType {
	case "solution":
		fmt.Println("Building manifest file - Solution!")
		solution, err := addComponentsToSolution(data, components)
		if err != nil {
			return err
		}
		newData, err = yaml.Marshal(&solution)
		if err != nil {
			return err
		}

	case "target":
		fmt.Println("Building manifest file - Target")
		target, err := addComponentsToTarget(data, components)
		if err != nil {
			return err
		}
		newData, err = yaml.Marshal(&target)
		if err != nil {
			return err
		}

	default:
		fmt.Println("target type not implemented yet")
	}

	// Create the directory
	err = os.MkdirAll(outputFolderPath, 0755)
	if err != nil {
		return err
	}

	// Write the data back to the file
	err = os.WriteFile(outputFilePath, newData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func addComponentsToSolution(data []byte, components []string) (Solution, error) {
	var solution Solution
	err := yaml.Unmarshal(data, &solution)
	if err != nil {
		return Solution{}, err
	}
	yamlComponents := make([]ComponentSpec, 0)
	for _, name := range components {
		if val, ok := ComponetsMap[name]; ok {
			yamlComponents = append(yamlComponents, val)
		}
	}
	solution.Spec.Components = yamlComponents

	return solution, nil
}

func addComponentsToTarget(data []byte, components []string) (Target, error) {
	var target Target
	err := yaml.Unmarshal(data, &target)
	if err != nil {
		return Target{}, err
	}
	yamlComponents := make([]ComponentSpec, 0)
	for _, name := range components {
		if val, ok := ComponetsMap[name]; ok {
			yamlComponents = append(yamlComponents, val)
		}
	}
	target.Spec.Components = yamlComponents

	return target, nil
}

type (
	InstanceOptions struct {
		NamePostfix string
		Scope       string
		Namespace   string
		Parameters  map[string]interface{}
		PostProcess func(*Instance)
		Solution    string
	}

	SolutionOptions struct {
		NamePostfix    string
		ComponentNames []string
		Namespace      string
		PostProcess    func(*Solution)
		SolutionName   string
	}

	TargetOptions = struct {
		NamePostfix    string
		Scope          string
		Namespace      string
		ComponentNames []string
		Properties     map[string]string
		PostProcess    func(*Target)
	}

	ContainerOptions = struct {
		Namespace string
	}
)

const (
	AzureOperationIdKey = "management.azure.com/operationId"
)

var leadingDash = regexp.MustCompile(`^-`)

func PatchSolution(data []byte, opts SolutionOptions) ([]byte, error) {
	var solution Solution
	err := yaml.Unmarshal(data, &solution)
	if err != nil {
		return nil, err
	}
	yamlComponents := make([]ComponentSpec, 0)
	for _, name := range opts.ComponentNames {
		if val, ok := ComponetsMap[name]; ok {
			yamlComponents = append(yamlComponents, val)
		} else {
			return nil, fmt.Errorf("component %s not found", name)
		}
	}

	if solution.Metadata.Annotations == nil {
		solution.Metadata.Annotations = make(map[string]string)
	}

	if opts.NamePostfix != "" {
		solution.Metadata.Name = fmt.Sprintf("%s-%s", solution.Metadata.Name, opts.NamePostfix)
		solution.Metadata.Name = leadingDash.ReplaceAllString(solution.Metadata.Name, "")
	}

	if opts.Namespace != "" {
		solution.Metadata.Namespace = opts.Namespace
	}

	if opts.SolutionName != "" {
		solution.Metadata.Name = opts.SolutionName
	}

	solution.Metadata.Annotations[AzureOperationIdKey] = string(uuid.NewUUID())
	solution.Spec.Components = yamlComponents
	if opts.PostProcess != nil {
		opts.PostProcess(&solution)
	}
	return yaml.Marshal(solution)
}

func PatchTarget(data []byte, opts TargetOptions) ([]byte, error) {
	var target Target
	err := yaml.Unmarshal(data, &target)
	if err != nil {
		return nil, err
	}

	for _, name := range opts.ComponentNames {
		if val, ok := ComponetsMap[name]; ok {
			target.Spec.Components = append(target.Spec.Components, val)
		} else {
			return nil, fmt.Errorf("component %s not found", name)
		}
	}
	if opts.NamePostfix != "" {
		target.Metadata.Name = fmt.Sprintf("%s-%s", target.Metadata.Name, opts.NamePostfix)
		target.Metadata.Name = leadingDash.ReplaceAllString(target.Metadata.Name, "")
	}

	if opts.Namespace != "" {
		target.Metadata.Namespace = opts.Namespace
	}

	if opts.Scope != "" {
		target.Spec.Scope = opts.Scope
	}

	if target.Metadata.Annotations == nil {
		target.Metadata.Annotations = make(map[string]string)
	}

	if opts.Properties != nil {
		target.Spec.Properties = opts.Properties
	}

	target.Metadata.Annotations[AzureOperationIdKey] = string(uuid.NewUUID())
	if opts.PostProcess != nil {
		opts.PostProcess(&target)
	}

	return yaml.Marshal(target)
}

func PatchInstance(data []byte, opts InstanceOptions) ([]byte, error) {
	var instance Instance
	err := yaml.Unmarshal(data, &instance)
	if err != nil {
		return nil, err
	}

	if opts.NamePostfix != "" {
		instance.Metadata.Name = fmt.Sprintf("%s-%s", instance.Metadata.Name, opts.NamePostfix)
		instance.Metadata.Name = leadingDash.ReplaceAllString(instance.Metadata.Name, "")
	}

	if opts.Namespace != "" {
		instance.Metadata.Namespace = opts.Namespace
	}

	if opts.Scope != "" {
		instance.Spec.Scope = opts.Scope
	}

	if opts.Solution != "" {
		instance.Spec.Solution = opts.Solution
	}

	if opts.Parameters != nil {
		instance.Spec.Parameters = opts.Parameters
	}

	if instance.Metadata.Annotations == nil {
		instance.Metadata.Annotations = make(map[string]string)
	}

	instance.Metadata.Annotations[AzureOperationIdKey] = string(uuid.NewUUID())
	if opts.PostProcess != nil {
		opts.PostProcess(&instance)
	}
	return yaml.Marshal(instance)
}

func PatchSolutionContainer(data []byte, opts ContainerOptions) ([]byte, error) {
	var solutionContainer SolutionContainer
	err := yaml.Unmarshal(data, &solutionContainer)
	if err != nil {
		return nil, err
	}

	if opts.Namespace != "" {
		solutionContainer.Metadata.Namespace = opts.Namespace
	}

	if solutionContainer.Metadata.Annotations == nil {
		solutionContainer.Metadata.Annotations = make(map[string]string)
	}

	return yaml.Marshal(solutionContainer)
}
