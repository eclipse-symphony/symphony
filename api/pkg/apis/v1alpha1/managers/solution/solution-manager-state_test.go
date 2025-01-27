/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/stretchr/testify/assert"
)

func TestConstructManagerState(t *testing.T) {
	//		 T1
	// ---------
	//	a	 X
	//	b	 X
	//	c	 X
	state, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 1, len(state.Targets))
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "instance", state.TargetComponent["c::T1"])
}
func TestConstructManagerStateTwoProviders(t *testing.T) {
	//		 T1		T2
	// ----------------
	//	a	 X		 .
	//	b	 .       X
	//	c	 X		 .
	state, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{c}",
			"T2": "{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
			"T2": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 2, len(state.Targets))
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "instance", state.TargetComponent["b::T2"])
	assert.Equal(t, "instance", state.TargetComponent["c::T1"])
}
func Test(t *testing.T) {
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock1",
					},
					{
						Name: "b",
						Type: "mock2",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}",
			"T2": "{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock1",
									Provider: "providers.target.mock",
								},
							},
						},
					},
				},
			},
			"T2": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock2",
									Provider: "providers.target.mock",
								},
							},
						},
					},
				},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Components))
	assert.Equal(t, 2, len(state.Targets))
	assert.Equal(t, 2, len(state.TargetComponent))
	assert.Equal(t, "mock1", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock2", state.TargetComponent["b::T2"])
}
func TestConstructManagerStateThreeProvidersDepedencies(t *testing.T) {
	//		 T1		T2		T3
	// ------------------------
	//	c	 .		 .		X
	//	b	 .       X		.
	//	a	 X		 .		.
	state, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name:         "a",
						Dependencies: []string{"b"},
					},
					{
						Name:         "b",
						Dependencies: []string{"c"},
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}",
			"T2": "{b}",
			"T3": "{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
			"T2": {
				Spec: &model.TargetSpec{},
			},
			"T3": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 3, len(state.Targets))
	assert.Equal(t, "instance", state.TargetComponent["c::T3"])
	assert.Equal(t, "instance", state.TargetComponent["b::T2"])
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
}
func TestMergeStateAddAComponent(t *testing.T) {
	state1, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state2, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state := MergeDeploymentStates(state1, state2)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 1, len(state.Targets))
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "instance", state.TargetComponent["c::T1"])
}
func TestMergeStateRemoveAComponent(t *testing.T) {
	state1, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state2, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state := MergeDeploymentStates(state1, state2)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 1, len(state.Targets))
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "-instance", state.TargetComponent["c::T1"])
}
func TestMergeStateProviderChange(t *testing.T) {
	state1, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state2, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T2": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T2": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state := MergeDeploymentStates(state1, state2)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 2, len(state.Targets))
	assert.Equal(t, "-instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "-instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "-instance", state.TargetComponent["c::T1"])
	assert.Equal(t, "instance", state.TargetComponent["a::T2"])
	assert.Equal(t, "instance", state.TargetComponent["b::T2"])
	assert.Equal(t, "instance", state.TargetComponent["c::T2"])
}
func TestMergeStateUnrelated(t *testing.T) {
	state1, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state2, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "d",
					},
					{
						Name: "e",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T2": "{d}{e}",
		},
		Targets: map[string]model.TargetState{
			"T2": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state := MergeDeploymentStates(state1, state2)
	assert.Equal(t, 5, len(state.Components))
	assert.Equal(t, 2, len(state.Targets))
	assert.Equal(t, "-instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "-instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "-instance", state.TargetComponent["c::T1"])
	assert.Equal(t, "instance", state.TargetComponent["d::T2"])
	assert.Equal(t, "instance", state.TargetComponent["e::T2"])
}
func TestMergeStateAddProvider(t *testing.T) {
	state1, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state2, err := NewDeploymentState(model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
			"T2": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
			"T2": {
				Spec: &model.TargetSpec{},
			},
		},
	})
	assert.Nil(t, err)
	state := MergeDeploymentStates(state1, state2)
	assert.Equal(t, 3, len(state.Components))
	assert.Equal(t, 2, len(state.Targets))
	assert.Equal(t, 5, len(state.TargetComponent))
	assert.Equal(t, "instance", state.TargetComponent["a::T1"])
	assert.Equal(t, "instance", state.TargetComponent["b::T1"])
	assert.Equal(t, "instance", state.TargetComponent["c::T1"])
	assert.Equal(t, "instance", state.TargetComponent["a::T2"])
	assert.Equal(t, "instance", state.TargetComponent["b::T2"])
}
func TestPlanSimple(t *testing.T) {
	//		 T1
	// ---------
	//	a	 X	helm
	//	b	 X	(instance)
	//	c	 X	instance
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
					},
					{
						Name: "c",
						Type: "instance",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1-helm: a
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 1, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	//T1-container: b,c
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "instance", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[1].Components[1].Component.Name)
}
func TestPlanComplex(t *testing.T) {
	//		 T1		T2		T3
	// -------------------------
	//	a	                X
	//	b	 X      X
	//	c	        X       X
	//  d    X              X
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
					},
					{
						Name:         "b",
						Dependencies: []string{"a"},
						Type:         "helm",
					},
					{
						Name:         "c",
						Dependencies: []string{"b"},
						Type:         "helm",
					},
					{
						Name:         "d",
						Dependencies: []string{"b", "c"},
						Type:         "kubectl",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{b}{d}",
			"T2": "{b}{c}",
			"T3": "{a}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
			"T2": {
				Spec: &model.TargetSpec{},
			},
			"T3": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(plan.Steps))
	//T3:a
	assert.Equal(t, "T3", plan.Steps[0].Target)
	assert.Equal(t, "instance", plan.Steps[0].Role)
	assert.Equal(t, 1, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	//T1:b
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "helm", plan.Steps[1].Role)
	assert.Equal(t, 1, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	//T2:b,c
	assert.Equal(t, "T2", plan.Steps[2].Target)
	assert.Equal(t, "helm", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[2].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[2].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[2].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[2].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[2].Components[1].Component.Name)
	//T3:c
	assert.Equal(t, "T3", plan.Steps[3].Target)
	assert.Equal(t, "helm", plan.Steps[3].Role)
	assert.Equal(t, 1, len(plan.Steps[3].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[3].Components[0].Action)
	assert.Equal(t, "c", plan.Steps[3].Components[0].Component.Name)
	//T1:d
	assert.Equal(t, "T1", plan.Steps[4].Target)
	assert.Equal(t, "kubectl", plan.Steps[4].Role)
	assert.Equal(t, 1, len(plan.Steps[4].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[4].Components[0].Action)
	assert.Equal(t, "d", plan.Steps[4].Components[0].Component.Name)
	//T3:d
	assert.Equal(t, "T3", plan.Steps[5].Target)
	assert.Equal(t, 1, len(plan.Steps[5].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[5].Components[0].Action)
	assert.Equal(t, "d", plan.Steps[5].Components[0].Component.Name)
}
func TestProviderStepsMergeNoDepedencies(t *testing.T) {
	//		 T1
	// -------------
	//	a	 X       helm
	//	b	 X       kubectl
	//	c	 X       helm
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
						Type: "kubectl",
					},
					{
						Name: "c",
						Type: "helm",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1:a,c
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 2, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[0].Components[1].Component.Name)
	//T1:b
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "kubectl", plan.Steps[1].Role)
	assert.Equal(t, 1, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
}
func TestProviderStepsMergeWithDepedencies(t *testing.T) {
	//		 T1
	// -------------
	//	a	 X       helm
	//	b	 X       kubectl
	//	c	 X       helm
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
						Type: "kubectl",
					},
					{
						Name:         "c",
						Type:         "helm",
						Dependencies: []string{"b"},
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(plan.Steps))
	//T1:a
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 1, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	//T1:b
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "kubectl", plan.Steps[1].Role)
	assert.Equal(t, 1, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	//T1:c
	assert.Equal(t, "T1", plan.Steps[2].Target)
	assert.Equal(t, "helm", plan.Steps[2].Role)
	assert.Equal(t, 1, len(plan.Steps[2].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[2].Components[0].Action)
	assert.Equal(t, "c", plan.Steps[2].Components[0].Component.Name)
}
func TestDockerHelmMixNoDepedencies(t *testing.T) {
	//	T1
	//------
	//a 	helm
	//b		docker
	//c		helm
	//d		docker
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
						Type: "docker",
					},
					{
						Name: "c",
						Type: "helm",
					},
					{
						Name: "d",
						Type: "docker",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1:a,c
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 2, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[0].Components[1].Component.Name)
	//T1:b,d
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "docker", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[1].Action)
	assert.Equal(t, "d", plan.Steps[1].Components[1].Component.Name)
}
func TestDockerHelmMixPairedDepedencies(t *testing.T) {
	//	T1
	//------
	//a 	helm
	//b		docker -> a
	//c		helm
	//d		docker -> c
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name:         "b",
						Type:         "docker",
						Dependencies: []string{"a"},
					},
					{
						Name: "c",
						Type: "helm",
					},
					{
						Name:         "d",
						Type:         "docker",
						Dependencies: []string{"c"},
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1:a,c
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 2, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[0].Components[1].Component.Name)
	//T1:b,d
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "docker", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[1].Action)
	assert.Equal(t, "d", plan.Steps[1].Components[1].Component.Name)
}
func TestDockerHelmMixCrosseddDepedencies(t *testing.T) {
	//	T1
	//------
	//a 	helm
	//b		docker
	//c		helm -> a
	//d		docker -> b
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
						Type: "docker",
					},
					{
						Name:         "c",
						Type:         "helm",
						Dependencies: []string{"a"},
					},
					{
						Name:         "d",
						Type:         "docker",
						Dependencies: []string{"b"},
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1:a,c
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 2, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[0].Components[1].Component.Name)
	//T1:b,d
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "docker", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[1].Action)
	assert.Equal(t, "d", plan.Steps[1].Components[1].Component.Name)
}
func TestDockerHelmMixCombineddDepedencies(t *testing.T) {
	//	T1
	//------
	//a 	helm
	//b		docker
	//c		helm
	//d		docker -> b,c
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name: "b",
						Type: "docker",
					},
					{
						Name: "c",
						Type: "helm",
					},
					{
						Name:         "d",
						Type:         "docker",
						Dependencies: []string{"b", "c"},
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(plan.Steps))
	//T1:a,c
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 2, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[1].Action)
	assert.Equal(t, "c", plan.Steps[0].Components[1].Component.Name)
	//T1:b,d
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "docker", plan.Steps[1].Role)
	assert.Equal(t, 2, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[1].Action)
	assert.Equal(t, "d", plan.Steps[1].Components[1].Component.Name)
}
func TestDockerHelmMixLinearDepedencies(t *testing.T) {
	//	T1
	//------
	//a 	helm
	//b		docker -> a
	//c		helm -> b
	//d		docker -> c
	deployment := model.DeploymentSpec{
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "helm",
					},
					{
						Name:         "b",
						Type:         "docker",
						Dependencies: []string{"a"},
					},
					{
						Name:         "c",
						Type:         "helm",
						Dependencies: []string{"b"},
					},
					{
						Name:         "d",
						Type:         "docker",
						Dependencies: []string{"c"},
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}{c}{d}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{},
			},
		},
	}
	state, err := NewDeploymentState(deployment)
	assert.Nil(t, err)
	plan, err := PlanForDeployment(deployment, state)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(plan.Steps))
	//T1:a
	assert.Equal(t, "T1", plan.Steps[0].Target)
	assert.Equal(t, "helm", plan.Steps[0].Role)
	assert.Equal(t, 1, len(plan.Steps[0].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[0].Components[0].Action)
	assert.Equal(t, "a", plan.Steps[0].Components[0].Component.Name)
	//T1:b
	assert.Equal(t, "T1", plan.Steps[1].Target)
	assert.Equal(t, "docker", plan.Steps[1].Role)
	assert.Equal(t, 1, len(plan.Steps[1].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[1].Components[0].Action)
	assert.Equal(t, "b", plan.Steps[1].Components[0].Component.Name)
	//T1:c
	assert.Equal(t, "T1", plan.Steps[2].Target)
	assert.Equal(t, "helm", plan.Steps[2].Role)
	assert.Equal(t, 1, len(plan.Steps[2].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[2].Components[0].Action)
	assert.Equal(t, "c", plan.Steps[2].Components[0].Component.Name)
	//T1:d
	assert.Equal(t, "T1", plan.Steps[3].Target)
	assert.Equal(t, "docker", plan.Steps[3].Role)
	assert.Equal(t, 1, len(plan.Steps[3].Components))
	assert.Equal(t, model.ComponentUpdate, plan.Steps[3].Components[0].Action)
	assert.Equal(t, "d", plan.Steps[3].Components[0].Component.Name)
}
