/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package testhelpers

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type (
	Metadata struct {
		Annotations map[string]string `yaml:"annotations,omitempty"`
		Name        string            `yaml:"name,omitempty"`
	}

	// Solution describes the structure of symphony solution yaml file
	Solution struct {
		ApiVersion string       `yaml:"apiVersion"`
		Kind       string       `yaml:"kind"`
		Metadata   Metadata     `yaml:"metadata"`
		Spec       SolutionSpec `yaml:"spec"`
	}

	SolutionSpec struct {
		DisplayName string            `yaml:"displayName,omitempty"`
		Metadata    map[string]string `yaml:"metadata,omitempty"`
		Components  []ComponentSpec   `yaml:"components,omitempty"`
	}

	// Target describes the structure of symphony target yaml file
	Target struct {
		ApiVersion string     `yaml:"apiVersion"`
		Kind       string     `yaml:"kind"`
		Metadata   Metadata   `yaml:"metadata"`
		Spec       TargetSpec `yaml:"spec"`
	}

	TargetSpec struct {
		DisplayName string          `yaml:"displayName"`
		Scope       string          `yaml:"scope,omitempty"`
		Components  []ComponentSpec `yaml:"components,omitempty"`
		Topologies  []Topology      `yaml:"topologies"`
	}

	Topology struct {
		Bindings []Binding `yaml:"bindings"`
	}

	Binding struct {
		Config   Config `yaml:"config"`
		Provider string `yaml:"provider"`
		Role     string `yaml:"role"`
	}

	Config struct {
		InCluster string `yaml:"inCluster"`
	}

	ComponentSpec struct {
		Name       string                 `yaml:"name"`
		Properties map[string]interface{} `yaml:"properties"`
		Type       string                 `yaml:"type"`
	}
)

var (
	ComponetsMap = map[string]ComponentSpec{
		"e4k": {
			Name: "e4k",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "e4kpreview.azurecr.io/helm/az-e4k",
					"version": "0.3.0",
				},
			},
			Type: "helm.v3",
		},
		"e4k-broker": {
			Name: "e4k-high-availability-broker",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "symphonycr.azurecr.io/az-e4k-broker",
					"version": "0.1.0",
				},
			},
			Type: "helm.v3",
		},
		"bluefin-extension": {
			Name: "bluefin",
			Properties: map[string]interface{}{
				"chart": map[string]interface{}{
					"repo":    "azbluefin.azurecr.io/helm/bluefin-arc-extension",
					"version": "0.2.0-20230706.3-develop",
				},
			},
			Type: "helm.v3",
		},
		"bluefin-instance": {
			Name: "bluefin-instance",
			Properties: map[string]interface{}{
				"resource": map[string]interface{}{
					"apiVersion": "bluefin.az-bluefin.com/v1",
					"kind":       "Instance",
					"metadata": map[string]interface{}{
						"name":      "bf-instance",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"displayName":          "Test Instance",
						"otelCollectorAddress": "otel-collector.alice-springs.svc.cluster.local:4317",
					},
				},
			},
			Type: "yaml.k8s",
		},

		"bluefin-pipeline": {
			Name: "test-pipeline",
			Properties: map[string]interface{}{
				"resource": map[string]interface{}{
					"apiVersion": "bluefin.az-bluefin.com/v1",
					"kind":       "Pipeline",
					"metadata": map[string]interface{}{
						"name":      "bf-pipeline",
						"namespace": "default",
					},
					"spec": map[string]interface{}{
						"displayName": "bf-pipeline",
						"enabled":     true,
						"input": map[string]interface{}{
							"description": "Read from topic Thermostat 3",
							"displayName": "E4K",
							"format":      map[string]interface{}{"type": "json"},
							"mqttConnectionInfo": map[string]interface{}{
								"broker":   "tcp://azedge-dmqtt-frontend:1883",
								"password": "password",
								"username": "client1",
							},
							"next": []interface{}{"node-22f2"},
							"topics": []interface{}{
								map[string]interface{}{
									"name": "alice-springs/data/opc-ua-connector/opc-ua-connector/thermostat-sample-3",
								},
							},
							"type": "input/mqtt@v1",
							"viewOptions": map[string]interface{}{
								"position": map[string]interface{}{
									"x": 0,
									"y": 80,
								},
							},
						},
						"partitionCount": 6,
						"stages": map[string]interface{}{
							"node-22f2": map[string]interface{}{
								"displayName": "No-op",
								"next":        []interface{}{"output"},
								"query":       ".",
								"type":        "processor/transform@v1",
								"viewOptions": map[string]interface{}{
									"position": map[string]interface{}{
										"x": 0,
										"y": 208,
									},
								},
							},
							"output": map[string]interface{}{
								"broker":      "tcp://azedge-dmqtt-frontend:1883",
								"description": "Publish to topic demo-output-topic",
								"displayName": "E4K",
								"format":      map[string]interface{}{"type": "json"},
								"password":    "password",
								"timeout":     "45ms",
								"topic":       "alice-springs/data/demo-output",
								"type":        "output/mqtt@v1",
								"username":    "client1",
								"viewOptions": map[string]interface{}{
									"position": map[string]interface{}{
										"x": 0,
										"y": 336,
									},
								},
							},
						},
					},
				},
			},
			Type: "yaml.k8s",
		},
	}
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
