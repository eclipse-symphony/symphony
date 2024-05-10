/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import (
	"fmt"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestByTargetNameLen(t *testing.T) {
	p := ByTargetName{
		{
			Name: "a",
		},
		{
			Name: "b",
		},
	}
	assert.Equal(t, p.Len(), 2)
}

func TestByTargetNameLess(t *testing.T) {
	p := ByTargetName{
		{
			Name: "a",
		},
		{
			Name: "b",
		},
	}
	assert.True(t, p.Less(0, 1))
}

func TestByTargetNameSwap(t *testing.T) {
	p := ByTargetName{
		{
			Name: "a",
		},
		{
			Name: "b",
		},
	}
	p.Swap(0, 1)
	assert.Equal(t, p[0].Name, "b")
	assert.Equal(t, p[1].Name, "a")
}

func createSampleDeploymentPlan() DeploymentPlan {
	p := DeploymentPlan{
		Steps: []DeploymentStep{
			{
				Target:  "sample-grpc-target",
				Role:    "instance",
				IsFirst: true,
				Components: []ComponentStep{
					{
						Action: ComponentUpdate,
						Component: ComponentSpec{
							Name: "sample-grpc-solution",
							Type: "instance",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
				},
			},
		},
	}
	return p
}

func createSampleDeploymentPlanForDelete() DeploymentPlan {
	p := DeploymentPlan{
		Steps: []DeploymentStep{
			{
				Target:  "sample-grpc-target-d1",
				Role:    "instance-d1",
				IsFirst: true,
				Components: []ComponentStep{
					{
						Action: ComponentDelete,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-d1",
							Type: "instance-d1",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
				},
			},
			{
				Target:  "sample-grpc-target-d2",
				Role:    "instance-d2",
				IsFirst: true,
				Components: []ComponentStep{
					{
						Action: ComponentDelete,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-d2",
							Type: "instance-d2",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
				},
			},
		},
	}
	return p
}

func createSampleDeploymentPlanForUpdateAndDelete() DeploymentPlan {
	p := DeploymentPlan{
		Steps: []DeploymentStep{
			{
				Target:  "sample-grpc-target",
				Role:    "instance",
				IsFirst: true,
				Components: []ComponentStep{
					{
						Action: ComponentDelete,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-d1",
							Type: "instance",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
					{
						Action: ComponentUpdate,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-u1",
							Type: "instance",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
					{
						Action: ComponentDelete,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-d2",
							Type: "instance",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
					{
						Action: ComponentUpdate,
						Component: ComponentSpec{
							Name: "sample-grpc-solution-u2",
							Type: "instance",
							Properties: map[string]interface{}{
								"file.content": "hello world",
							},
						},
					},
				},
			},
		},
	}
	return p
}

func createSampleDeploymentStepWithUpdateComponent() DeploymentStep {
	s := DeploymentStep{
		Target: "sample-grpc-target",
		Components: []ComponentStep{
			{
				Action: ComponentUpdate,
				Component: ComponentSpec{
					Name: "sample-grpc-solution",
					Type: "instance",
					Properties: map[string]interface{}{
						"file.content": "hello world",
					},
				},
			},
		},
		Role:    "instance",
		IsFirst: true,
	}
	return s
}

func createSampleDeploymentStepWithDeleteComponent() DeploymentStep {
	s := DeploymentStep{
		Target: "sample-grpc-target",
		Components: []ComponentStep{
			{
				Action: ComponentDelete,
				Component: ComponentSpec{
					Name: "sample-grpc-solution",
					Type: "instance",
					Properties: map[string]interface{}{
						"file.content": "hello world",
					},
				},
			},
		},
		Role:    "instance",
		IsFirst: true,
	}
	return s
}

func TestPrepareResultMap(t *testing.T) {
	s := createSampleDeploymentStepWithUpdateComponent()
	resultMap := s.PrepareResultMap()
	assert.Equal(t, v1alpha2.Untouched, resultMap["sample-grpc-solution"].Status)
	assert.Equal(t, fmt.Sprintf("No error. %s is untouched", "sample-grpc-solution"), resultMap["sample-grpc-solution"].Message)
}

func TestGetComponents(t *testing.T) {
	s := createSampleDeploymentStepWithUpdateComponent()
	components := s.GetComponents()
	assert.Equal(t, len(components), 1)
	assert.Equal(t, components[0].Name, "sample-grpc-solution")
	assert.Equal(t, components[0].Type, "instance")
	assert.Equal(t, components[0].Properties["file.content"], "hello world")
}

func TestGetUpdatedComponents(t *testing.T) {
	s := createSampleDeploymentStepWithUpdateComponent()
	components := s.GetUpdatedComponents()
	assert.Equal(t, len(components), 1)
	assert.Equal(t, components[0].Name, "sample-grpc-solution")
	assert.Equal(t, components[0].Type, "instance")
	assert.Equal(t, components[0].Properties["file.content"], "hello world")
}

func TestGetDeletedComponents(t *testing.T) {
	s := createSampleDeploymentStepWithDeleteComponent()
	components := s.GetDeletedComponents()
	assert.Equal(t, len(components), 1)
	assert.Equal(t, components[0].Name, "sample-grpc-solution")
	assert.Equal(t, components[0].Type, "instance")
	assert.Equal(t, components[0].Properties["file.content"], "hello world")
}

func TestGetUpdatedComponentSteps(t *testing.T) {
	s := createSampleDeploymentStepWithUpdateComponent()
	components := s.GetUpdatedComponentSteps()
	assert.Equal(t, len(components), 1)
	assert.Equal(t, components[0].Action, ComponentUpdate)
	assert.Equal(t, components[0].Component.Name, "sample-grpc-solution")
	assert.Equal(t, components[0].Component.Type, "instance")
	assert.Equal(t, components[0].Component.Properties["file.content"], "hello world")
}

func TestMarkRemoveAll(t *testing.T) {
	s := DeploymentState{
		TargetComponent: map[string]string{
			"a::T1": "mock1",
		},
	}
	s.MarkRemoveAll()
	assert.Equal(t, s.TargetComponent["a::T1"], "-mock1")
}

func TestClearAllRemoved(t *testing.T) {
	s := DeploymentState{
		TargetComponent: map[string]string{
			"a::T1": "mock1",
		},
	}
	s.MarkRemoveAll()
	assert.Equal(t, s.TargetComponent["a::T1"], "-mock1")

	s.ClearAllRemoved()
	assert.Equal(t, len(s.TargetComponent), 0)
}

func TestFindLastTargetRole(t *testing.T) {
	p := createSampleDeploymentPlan()
	index := p.FindLastTargetRole("sample-grpc-target", "instance")
	assert.Equal(t, index, 0)

	index = p.FindLastTargetRole("sample-grpc-target", "instance1")
	assert.Equal(t, index, -1)
}

func TestCanAppendToStep(t *testing.T) {
	p := createSampleDeploymentPlan()

	// no dependencies, can add
	canAppend := p.CanAppendToStep(0, ComponentSpec{
		Name: "sample-grpc-solution2",
		Type: "instance",
		Properties: map[string]interface{}{
			"file.content": "hello world",
		},
	})
	assert.Equal(t, true, canAppend)

	// has dependencies, and dependencies include plan component, can add
	canAppend = p.CanAppendToStep(0, ComponentSpec{
		Name: "sample-grpc-solution2",
		Type: "instance",
		Properties: map[string]interface{}{
			"file.content": "hello world",
		},
		Dependencies: []string{"sample-grpc-solution"},
	})
	assert.Equal(t, true, canAppend)

	// has dependencies, but dependencies not include plan component, can not add
	canAppend = p.CanAppendToStep(0, ComponentSpec{
		Name: "sample-grpc-solution2",
		Type: "instance",
		Properties: map[string]interface{}{
			"file.content": "hello world",
		},
		Dependencies: []string{"sample-grpc-solution3"},
	})
	assert.Equal(t, false, canAppend)

}

func TestRevisedForDeletion(t *testing.T) {
	// no delete component, no change
	p := createSampleDeploymentPlan()
	p = p.RevisedForDeletion()
	assert.Equal(t, len(p.Steps), 1)
	assert.Equal(t, p.Steps[0].Components[0].Action, ComponentUpdate)
	assert.Equal(t, p.Steps[0].Components[0].Component.Name, "sample-grpc-solution")
	assert.Equal(t, p.Steps[0].Components[0].Component.Type, "instance")
	assert.Equal(t, p.Steps[0].Components[0].Component.Properties["file.content"], "hello world")

	// only delete component, no change, last-to-first order
	p = createSampleDeploymentPlanForDelete()
	p = p.RevisedForDeletion()
	assert.Equal(t, len(p.Steps), 2)
	assert.Equal(t, p.Steps[0].Components[0].Action, ComponentDelete)
	assert.Equal(t, p.Steps[0].Components[0].Component.Name, "sample-grpc-solution-d2")
	assert.Equal(t, p.Steps[0].Components[0].Component.Type, "instance-d2")
	assert.Equal(t, p.Steps[0].Components[0].Component.Properties["file.content"], "hello world")

	assert.Equal(t, p.Steps[1].Components[0].Action, ComponentDelete)
	assert.Equal(t, p.Steps[1].Components[0].Component.Name, "sample-grpc-solution-d1")
	assert.Equal(t, p.Steps[1].Components[0].Component.Type, "instance-d1")
	assert.Equal(t, p.Steps[1].Components[0].Component.Properties["file.content"], "hello world")

	// update and delete component,
	// step1: update at first, orignial order
	// step2: then delete, last-to-first order
	p = createSampleDeploymentPlanForUpdateAndDelete()
	p = p.RevisedForDeletion()
	assert.Equal(t, len(p.Steps), 2)
	assert.Equal(t, p.Steps[0].Components[0].Action, ComponentUpdate)
	assert.Equal(t, p.Steps[0].Components[0].Component.Name, "sample-grpc-solution-u1")
	assert.Equal(t, p.Steps[0].Components[0].Component.Type, "instance")
	assert.Equal(t, p.Steps[0].Components[0].Component.Properties["file.content"], "hello world")

	assert.Equal(t, p.Steps[0].Components[1].Action, ComponentUpdate)
	assert.Equal(t, p.Steps[0].Components[1].Component.Name, "sample-grpc-solution-u2")
	assert.Equal(t, p.Steps[0].Components[1].Component.Type, "instance")
	assert.Equal(t, p.Steps[0].Components[1].Component.Properties["file.content"], "hello world")

	assert.Equal(t, p.Steps[1].Components[0].Action, ComponentDelete)
	assert.Equal(t, p.Steps[1].Components[0].Component.Name, "sample-grpc-solution-d2")
	assert.Equal(t, p.Steps[1].Components[0].Component.Type, "instance")
	assert.Equal(t, p.Steps[1].Components[0].Component.Properties["file.content"], "hello world")

	assert.Equal(t, p.Steps[1].Components[1].Action, ComponentDelete)
	assert.Equal(t, p.Steps[1].Components[1].Component.Name, "sample-grpc-solution-d1")
	assert.Equal(t, p.Steps[1].Components[1].Component.Type, "instance")
	assert.Equal(t, p.Steps[1].Components[1].Component.Properties["file.content"], "hello world")
}
