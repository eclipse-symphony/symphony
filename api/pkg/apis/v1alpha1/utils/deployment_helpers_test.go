/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/require"
)

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
	res, err := CreateSymphonyDeploymentFromTarget(ctx, model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name: "someTargetName",
			Annotations: map[string]string{
				"Guid": "someGuid",
			},
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
		SolutionVersionName: "target-runtime-someTargetName",
		SolutionVersion: model.SolutionVersionState{
			ObjectMeta: model.ObjectMeta{
				Name: "target-runtime-someTargetName",
			},
			Spec: &model.SolutionVersionSpec{
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
				Annotations: map[string]string{
					"Guid": "someGuid",
				},
			},
			Spec: &model.InstanceSpec{
				Scope:           "targetScope",
				DisplayName:     "target-runtime-someTargetName",
				SolutionVersion: "target-runtime-someTargetName",
				Target: model.TargetSelector{
					Name: "someTargetName",
				},
			},
		},
		Targets: map[string]model.TargetState{
			"someTargetName": {
				ObjectMeta: model.ObjectMeta{
					Name: "someTargetName",
					Annotations: map[string]string{
						"Guid": "someGuid",
					},
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
	res, err := CreateSymphonyDeployment(ctx, model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "someOtherId",
			Namespace: "instanceScope",
			Annotations: map[string]string{
				"Guid": "someGuid",
			},
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
	}, model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "someOtherId",
			Namespace: "solutionversionsScope",
		},
		Spec: &model.SolutionVersionSpec{
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
		SolutionVersionName: "someOtherId",
		SolutionVersion: model.SolutionVersionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "someOtherId",
				Namespace: "solutionversionsScope",
			},
			Spec: &model.SolutionVersionSpec{
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
				Annotations: map[string]string{
					"Guid": "someGuid",
				},
			},
			Spec: &model.InstanceSpec{
				SolutionVersion: "",
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
	res, err := AssignComponentsToTargets(ctx, []model.ComponentSpec{
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
