/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/require"
)

const (
	baseUrl  = "http://localhost:8090/v1alpha2/"
	user     = "admin"
	password = ""
)

var (
	testApiClient ApiClient = &apiClient{
		baseUrl: baseUrl,
	}
)

func TestGetInstancesWhenSomeInstances(t *testing.T) {
	testSymphonyApi := os.Getenv("TEST_SYMPHONY_API")
	if testSymphonyApi != "yes" {
		t.Skip("Skipping becasue TEST_SYMPHONY_API is missing or not set to 'yes'")
	}

	solutionName := "solution1"
	solution1JsonObj := map[string]interface{}{
		"name": "e4i-assets",
		"type": "helm.v3",
		"properties": map[string]interface{}{
			"chart": map[string]interface{}{
				"repo":    "e4ipreview.azurecr.io/helm/az-e4i-demo-assets",
				"version": "0.4.0",
			},
		},
	}

	solution1, err := json.Marshal(solution1JsonObj)
	if err != nil {
		panic(err)
	}

	err = testApiClient.UpsertSolution(context.Background(), solutionName, solution1, "default", user, password)
	require.NoError(t, err)

	targetName := "target1"
	target1JsonObj := map[string]interface{}{
		"id": "self",
		"spec": map[string]interface{}{
			"displayName": "int-virtual-02",
			"scope":       "alice-springs",
			"components": []interface{}{
				map[string]interface{}{
					"name": "observability",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "symphonycr.azurecr.io/sample-dashboard",
							"version": "0.4.0-dev",
						},
						"values": map[string]interface{}{
							"obsConfig": map[string]interface{}{
								"bluefin": true,
								"e4i":     true,
								"e4k":     true,
							},
						},
					},
				},
				map[string]interface{}{
					"name": "e4k",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "e4kpreview.azurecr.io/helm/az-e4k",
							"version": "0.3.0",
						},
					},
				},
				map[string]interface{}{
					"name": "e4i",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "e4ipreview.azurecr.io/helm/az-e4i",
							"version": "0.4.0",
						},
						"values": map[string]interface{}{
							"mqttBroker": map[string]interface{}{
								"authenticationMethod": "serviceAccountToken",
								"name":                 "azedge-dmqtt-frontend",
								"namespace":            "alice-springs",
							},
							"opcPlcSimulation": map[string]interface{}{
								"deploy": "true",
							},
							"openTelemetry": map[string]interface{}{
								"enabled":  "true",
								"endpoint": "http://otel-collector.alice-springs.svc.cluster.local:4317",
								"protocol": "grpc",
							},
						},
					},
				},
				map[string]interface{}{
					"name": "bluefin",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
							"version": "0.1.2",
						},
					},
				},
			},
			"topologies": []interface{}{
				map[string]interface{}{
					"bindings": []interface{}{
						map[string]interface{}{
							"role":     "instance",
							"provider": "providers.target.k8s",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
						map[string]interface{}{
							"role":     "helm.v3",
							"provider": "providers.target.helm",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
						map[string]interface{}{
							"role":     "yaml.k8s",
							"provider": "providers.target.kubectl",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
					},
				},
			},
		},
	}
	target1, err := json.Marshal(target1JsonObj)
	require.NoError(t, err)

	err = testApiClient.CreateTarget(context.Background(), targetName, target1, "default", user, password)
	require.NoError(t, err)

	instanceName := "instance1"
	instance1JsonObj := map[string]interface{}{
		"scope":    "default",
		"solution": solutionName,
		"target": map[string]interface{}{
			"name": targetName,
		},
	}

	instance1, err := json.Marshal(instance1JsonObj)
	if err != nil {
		panic(err)
	}

	err = testApiClient.CreateInstance(context.Background(), instanceName, instance1, "default", user, password)
	require.NoError(t, err)

	// ensure instance gets created properly
	time.Sleep(time.Second)

	instancesRes, err := testApiClient.GetInstances(context.Background(), "default", user, password)
	require.NoError(t, err)

	require.Equal(t, 1, len(instancesRes))
	require.Equal(t, instanceName, instancesRes[0].Spec.DisplayName)
	require.Equal(t, solutionName, instancesRes[0].Spec.Solution)
	require.Equal(t, targetName, instancesRes[0].Spec.Target.Name)
	require.Equal(t, "1", instancesRes[0].Status.Properties["targets"])
	require.Equal(t, "OK", instancesRes[0].Status.Properties["status"])

	instanceRes, err := testApiClient.GetInstance(context.Background(), instanceName, "default", user, password)
	require.NoError(t, err)

	require.Equal(t, instanceName, instanceRes.Spec.DisplayName)
	require.Equal(t, solutionName, instanceRes.Spec.Solution)
	require.Equal(t, targetName, instanceRes.Spec.Target.Name)
	require.Equal(t, "1", instanceRes.Status.Properties["targets"])
	require.Equal(t, "OK", instanceRes.Status.Properties["status"])

	err = testApiClient.DeleteTarget(context.Background(), targetName, "default", user, password)
	require.NoError(t, err)

	err = testApiClient.DeleteSolution(context.Background(), solutionName, "default", user, password)
	require.NoError(t, err)

	err = testApiClient.DeleteInstance(context.Background(), instanceName, "default", user, password)
	require.NoError(t, err)
}

func TestGetSolutionsWhenSomeSolution(t *testing.T) {
	testSymphonyApi := os.Getenv("TEST_SYMPHONY_API")
	if testSymphonyApi != "yes" {
		t.Skip("Skipping becasue TEST_SYMPHONY_API is missing or not set to 'yes'")
	}

	solutionName := "solution1"
	solution1JsonObj := map[string]interface{}{
		"name": "e4i-assets",
		"type": "helm.v3",
		"properties": map[string]interface{}{
			"chart": map[string]interface{}{
				"repo":    "e4ipreview.azurecr.io/helm/az-e4i-demo-assets",
				"version": "0.4.0",
			},
		},
	}

	solution1, err := json.Marshal(solution1JsonObj)
	if err != nil {
		panic(err)
	}

	err = testApiClient.UpsertSolution(context.Background(), solutionName, solution1, "default", user, password)
	require.NoError(t, err)

	solutionsRes, err := testApiClient.GetSolutions(context.Background(), "default", user, password)
	require.NoError(t, err)

	require.Equal(t, 1, len(solutionsRes))
	require.Equal(t, solutionName, solutionsRes[0].Spec.DisplayName)

	solutionRes, err := testApiClient.GetSolution(context.Background(), solutionName, "default", user, password)
	require.NoError(t, err)

	require.Equal(t, solutionName, solutionRes.Spec.DisplayName)

	err = testApiClient.DeleteSolution(context.Background(), solutionName, "default", user, password)
	require.NoError(t, err)
}

func TestGetTargetsWithSomeTargets(t *testing.T) {
	testSymphonyApi := os.Getenv("TEST_SYMPHONY_API")
	if testSymphonyApi != "yes" {
		t.Skip("Skipping becasue TEST_SYMPHONY_API is missing or not set to 'yes'")
	}

	targetName := "target1"
	target1JsonObj := map[string]interface{}{
		"id": "self",
		"spec": map[string]interface{}{
			"displayName": "int-virtual-02",
			"scope":       "alice-springs",
			"components": []interface{}{
				map[string]interface{}{
					"name": "observability",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "symphonycr.azurecr.io/sample-dashboard",
							"version": "0.4.0-dev",
						},
						"values": map[string]interface{}{
							"obsConfig": map[string]interface{}{
								"bluefin": true,
								"e4i":     true,
								"e4k":     true,
							},
						},
					},
				},
				map[string]interface{}{
					"name": "e4k",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "e4kpreview.azurecr.io/helm/az-e4k",
							"version": "0.3.0",
						},
					},
				},
				map[string]interface{}{
					"name": "e4i",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "e4ipreview.azurecr.io/helm/az-e4i",
							"version": "0.4.0",
						},
						"values": map[string]interface{}{
							"mqttBroker": map[string]interface{}{
								"authenticationMethod": "serviceAccountToken",
								"name":                 "azedge-dmqtt-frontend",
								"namespace":            "alice-springs",
							},
							"opcPlcSimulation": map[string]interface{}{
								"deploy": "true",
							},
							"openTelemetry": map[string]interface{}{
								"enabled":  "true",
								"endpoint": "http://otel-collector.alice-springs.svc.cluster.local:4317",
								"protocol": "grpc",
							},
						},
					},
				},
				map[string]interface{}{
					"name": "bluefin",
					"type": "helm.v3",
					"properties": map[string]interface{}{
						"chart": map[string]interface{}{
							"repo":    "azbluefin.azurecr.io/helmcharts/bluefin-arc-extension/bluefin-arc-extension",
							"version": "0.1.2",
						},
					},
				},
			},
			"topologies": []interface{}{
				map[string]interface{}{
					"bindings": []interface{}{
						map[string]interface{}{
							"role":     "instance",
							"provider": "providers.target.k8s",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
						map[string]interface{}{
							"role":     "helm.v3",
							"provider": "providers.target.helm",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
						map[string]interface{}{
							"role":     "yaml.k8s",
							"provider": "providers.target.kubectl",
							"config": map[string]interface{}{
								"inCluster": "true",
							},
						},
					},
				},
			},
		},
	}
	target1, err := json.Marshal(target1JsonObj)
	require.NoError(t, err)

	err = testApiClient.CreateTarget(context.Background(), targetName, target1, "default", user, password)
	require.NoError(t, err)

	// Ensure target gets created properly
	time.Sleep(time.Second)

	targetsRes, err := testApiClient.GetTargets(context.Background(), "default", user, password)
	require.NoError(t, err)

	require.Equal(t, 1, len(targetsRes))
	require.Equal(t, targetName, targetsRes[0].Spec.DisplayName)
	require.Equal(t, "default", targetsRes[0].ObjectMeta.Namespace)
	require.Equal(t, "1", targetsRes[0].Status.Properties["targets"])
	require.Equal(t, "OK", targetsRes[0].Status.Properties["status"])

	targetRes, err := testApiClient.GetTarget(context.Background(), targetName, "default", user, password)
	require.NoError(t, err)

	require.Equal(t, targetName, targetRes.Spec.DisplayName)
	require.Equal(t, "default", targetRes.ObjectMeta.Namespace)
	require.Equal(t, "1", targetRes.Status.Properties["targets"])
	require.Equal(t, "OK", targetRes.Status.Properties["status"])

	err = testApiClient.DeleteTarget(context.Background(), targetName, "default", user, password)
	require.NoError(t, err)
}

func TestMatchTargetsWithTargetName(t *testing.T) {
	res := MatchTargets(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name: "someId",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
			},
		},
		Status: model.InstanceStatus{},
	}, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someTargetName",
		},
		Spec: &model.TargetSpec{
			Metadata: map[string]string{
				"key": "value",
			},
		},
	}})

	require.Equal(t, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someTargetName",
		},
		Spec: &model.TargetSpec{
			Metadata: map[string]string{
				"key": "value",
			},
		},
	}}, res)
}

func TestMatchTargetsWithUnmatchedName(t *testing.T) {
	res := MatchTargets(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name: "someId",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
			},
		},
		Status: model.InstanceStatus{},
	}, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someDifferentTargetName",
		},
		Spec: &model.TargetSpec{},
	}})

	require.Equal(t, []model.TargetState{}, res)
}

func TestMatchTargetsWithSelectors(t *testing.T) {
	res := MatchTargets(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name: "someId",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
		Status: model.InstanceStatus{},
	}, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someDifferentTargetName",
		},
		Spec: &model.TargetSpec{
			DisplayName: "someDisplayName",
			Properties: map[string]string{
				"OS": "windows",
			},
		},
	}})

	require.Equal(t, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someDifferentTargetName",
		},
		Spec: &model.TargetSpec{
			DisplayName: "someDisplayName",
			Properties: map[string]string{
				"OS": "windows",
			},
		},
	}}, res)
}

func TestMatchTargetsWithUnmatchedSelectors(t *testing.T) {
	res := MatchTargets(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name: "someId",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
		Status: model.InstanceStatus{},
	}, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someDifferentTargetName",
		},
		Spec: &model.TargetSpec{
			Properties: map[string]string{
				"OS": "linux",
			},
		},
	}})

	require.Equal(t, []model.TargetState{}, res)

	res = MatchTargets(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name: "someId",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
		Status: model.InstanceStatus{},
	}, []model.TargetState{{
		ObjectMeta: model.ObjectMeta{
			Name: "someDifferentTargetName",
		},
		Spec: &model.TargetSpec{
			Properties: map[string]string{
				"company": "linux",
			},
		},
	}})

	require.Equal(t, []model.TargetState{}, res)
}

func TestCreateSymphonyDeploymentFromTarget(t *testing.T) {
	res, err := CreateSymphonyDeploymentFromTarget(model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name: "someTargetName",
		},
		Spec: &model.TargetSpec{
			DisplayName: "someDisplayName",
			Scope:       "targetScope",
			Components: []model.ComponentSpec{
				{
					Name: "componentName1",
					Type: "componentType1",
					Metadata: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				{
					Name: "componentName2",
					Type: "componentType2",
				},
			},
			Properties: map[string]string{
				"OS": "windows",
			},
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}, "default")
	require.NoError(t, err)

	ret, err := res.DeepEquals(model.DeploymentSpec{
		SolutionName: "target-runtime-someTargetName",
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name: "target-runtime-someTargetName",
			},
			Spec: &model.SolutionSpec{
				DisplayName: "target-runtime-someTargetName",
				Components: []model.ComponentSpec{
					{
						Name: "componentName1",
						Type: "componentType1",
						Metadata: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					{
						Name: "componentName2",
						Type: "componentType2",
					},
				},
			},
		},
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "target-runtime-someTargetName",
			},
			Spec: &model.InstanceSpec{
				Scope:       "targetScope",
				DisplayName: "target-runtime-someTargetName",
				Solution:    "target-runtime-someTargetName",
				Target: model.TargetSelector{
					Name: "someTargetName",
				},
			},
		},
		Targets: map[string]model.TargetState{
			"someTargetName": {
				ObjectMeta: model.ObjectMeta{
					Name: "someTargetName",
				},
				Spec: &model.TargetSpec{
					DisplayName: "someDisplayName",
					Scope:       "targetScope",
					Properties: map[string]string{
						"OS": "windows",
					},
					Components: []model.ComponentSpec{
						{
							Name: "componentName1",
							Type: "componentType1",
							Metadata: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
						{
							Name: "componentName2",
							Type: "componentType2",
						},
					},
					ForceRedeploy: false,
					Metadata: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
		Assignments: map[string]string{
			"someTargetName": "{componentName1}{componentName2}",
		},
	})
	require.NoError(t, err)
	require.True(t, ret)
}

func TestCreateSymphonyDeployment(t *testing.T) {
	res, err := CreateSymphonyDeployment(model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "someOtherId",
			Namespace: "instanceScope",
		},
		Spec: &model.InstanceSpec{
			Target: model.TargetSelector{
				Name: "someTargetName",
				Selector: map[string]string{
					"OS": "windows",
				},
			},
		},
		Status: model.InstanceStatus{},
	}, model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "someOtherId",
			Namespace: "solutionsScope",
		},
		Spec: &model.SolutionSpec{
			DisplayName: "someDisplayName",
			Components: []model.ComponentSpec{
				{
					Name: "componentName1",
					Type: "componentType1",
				},
				{
					Name: "componentName2",
					Type: "componentType2",
					Metadata: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}, []model.TargetState{
		{
			ObjectMeta: model.ObjectMeta{
				Name:      "someTargetName1",
				Namespace: "targetScope",
			},
			Spec: &model.TargetSpec{
				Properties: map[string]string{
					"company": "microsoft",
				},
				Metadata: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
	}, []model.DeviceState{
		{
			ObjectMeta: model.ObjectMeta{
				Name: "someTargetName2",
			},
			Spec: &model.DeviceSpec{
				DisplayName: "someDeviceDisplayName",
				Properties: map[string]string{
					"company": "microsoft",
				},
			},
		},
	}, "default")
	require.NoError(t, err)

	jData, _ := json.Marshal(res)
	t.Log(string(jData))

	ret, err := res.DeepEquals(model.DeploymentSpec{ //require.Equal( doesn't seem to compare pointer fields correctly
		SolutionName: "someOtherId",
		Solution: model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "someOtherId",
				Namespace: "solutionsScope",
			},
			Spec: &model.SolutionSpec{
				DisplayName: "someDisplayName",
				Components: []model.ComponentSpec{
					{
						Name: "componentName1",
						Type: "componentType1",
					},
					{
						Name: "componentName2",
						Type: "componentType2",
						Metadata: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
				Metadata: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name:      "someOtherId",
				Namespace: "instanceScope",
			},
			Spec: &model.InstanceSpec{
				Solution: "",
				Scope:    "default", // CreateSymphonyDeployment will give default if instance.Spec.Scope is empty
				Target: model.TargetSelector{
					Name: "someTargetName",
					Selector: map[string]string{
						"OS": "windows",
					},
				},
			},
			Status: model.InstanceStatus{},
		},
		Targets: map[string]model.TargetState{
			"someTargetName1": {
				ObjectMeta: model.ObjectMeta{
					Name:      "someTargetName1",
					Namespace: "targetScope",
				},
				Spec: &model.TargetSpec{
					Properties: map[string]string{
						"company": "microsoft",
					},
					ForceRedeploy: false,
					Metadata: map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					},
				},
			},
		},
		Assignments: map[string]string{
			"someTargetName1": "{componentName1}{componentName2}",
		},
	})
	require.NoError(t, err)
	require.True(t, ret)
}

func TestAssignComponentsToTargetsWithMixedConstraints(t *testing.T) {
	res, err := AssignComponentsToTargets([]model.ComponentSpec{
		{
			Name:        "componentName1",
			Constraints: "${{$equal($property(OS),windows)}}",
		},
		{
			Name:        "componentName2",
			Constraints: "${{$equal($property(OS),linux)}}",
		},
		{
			Name:        "componentName3",
			Constraints: "${{$equal($property(OS),unix)}}",
		},
	}, map[string]model.TargetState{
		"target1": {
			Spec: &model.TargetSpec{
				Properties: map[string]string{
					"OS": "windows",
				},
			},
		},
		"target2": {
			Spec: &model.TargetSpec{
				Properties: map[string]string{
					"OS": "linux",
				},
			},
		},
		"target3": {
			Spec: &model.TargetSpec{
				Properties: map[string]string{
					"OS": "unix",
				},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"target1": "{componentName1}",
		"target2": "{componentName2}",
		"target3": "{componentName3}",
	}, res)
}
