/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	memorykeylock "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/keylock/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFindAgentEmpty(t *testing.T) {
	deploymentState, _ := NewDeploymentState(model.DeploymentSpec{
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
	agent := findAgentFromDeploymentState(deploymentState, "T1")
	assert.Equal(t, "", agent)
}
func TestFindAgentMatch(t *testing.T) {
	deploymentState, _ := NewDeploymentState(model.DeploymentSpec{
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
				Spec: &model.TargetSpec{
					Components: []model.ComponentSpec{
						{
							Name: "symphony-agent",
							Properties: map[string]interface{}{
								model.ContainerImage: "ghcr.io/eclipse-symphony/symphony-agent:0.38.0",
							},
						},
					},
				},
			},
		},
	})
	agent := findAgentFromDeploymentState(deploymentState, "T1")
	assert.Equal(t, "symphony-agent", agent)
}

func TestFindAgentNotMatch(t *testing.T) {
	deploymentState, _ := NewDeploymentState(model.DeploymentSpec{
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
				Spec: &model.TargetSpec{
					Components: []model.ComponentSpec{
						{
							Name: "symphony-agent",
							Properties: map[string]interface{}{
								model.ContainerImage: "ghcr.io/eclipse-symphony/symphony-api:0.38.0",
							},
						},
					},
				},
			},
		},
	})
	agent := findAgentFromDeploymentState(deploymentState, "T1")
	assert.Equal(t, "", agent)
}

func TestFindAgentMatchMultiTargets(t *testing.T) {
	deploymentState, _ := NewDeploymentState(model.DeploymentSpec{
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
				Spec: &model.TargetSpec{
					Components: []model.ComponentSpec{
						{
							Name: "symphony-agent1",
							Properties: map[string]interface{}{
								model.ContainerImage: "ghcr.io/eclipse-symphony/symphony-agent:0.38.0",
							},
						},
					},
				},
			},
			"T2": {
				Spec: &model.TargetSpec{
					Components: []model.ComponentSpec{
						{
							Name: "symphony-agent2",
							Properties: map[string]interface{}{
								model.ContainerImage: "ghcr.io/eclipse-symphony/symphony-agent:0.38.0",
							},
						},
					},
				},
			},
		},
	})
	agent := findAgentFromDeploymentState(deploymentState, "T1")
	assert.Equal(t, "symphony-agent1", agent)
}

func TestSortByDepedenciesSingleChain(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com1", ret[0].Name)
	assert.Equal(t, "com2", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesSingleCircle(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name:         "com1",
			Dependencies: []string{"com3"},
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesSelfCircle(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com2"},
		},
		{
			Name:         "com2",
			Dependencies: []string{"com1"},
		},
		{
			Name:         "com1",
			Dependencies: []string{"com1"}, // note: generally self-depedencies should not be allowed
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesNoDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name: "com3",
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com3", ret[0].Name)
	assert.Equal(t, "com2", ret[1].Name)
	assert.Equal(t, "com1", ret[2].Name)
}
func TestSortByDepedenciesParitalDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com1"},
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com2", ret[0].Name)
	assert.Equal(t, "com1", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesMultiDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com1", "com2"},
		},
		{
			Name: "com2",
		},
		{
			Name:         "com1",
			Dependencies: []string{"com2"},
		},
	}
	ret, err := sortByDepedencies(components)
	assert.Nil(t, err)
	assert.Equal(t, "com2", ret[0].Name)
	assert.Equal(t, "com1", ret[1].Name)
	assert.Equal(t, "com3", ret[2].Name)
}
func TestSortByDepedenciesForeignDependencies(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com4"},
		},
		{
			Name: "com2",
		},
		{
			Name: "com1",
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestSortByDepedenciesAllSelfReferences(t *testing.T) {
	components := []model.ComponentSpec{
		{
			Name:         "com3",
			Dependencies: []string{"com3"}, //note: unlike TestSortByDepedenciesSelfCircle, this self-depedency is not resolved
		},
		{
			Name: "com2",
		},
		{
			Name:         "com1",
			Dependencies: []string{"com2"},
		},
	}
	_, err := sortByDepedencies(components)
	assert.NotNil(t, err)
}
func TestMockGet(t *testing.T) {
	id := uuid.New().String()
	name := "testInstance"
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: name,
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
									Config: map[string]string{
										"id": id,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	guid := uuid.New().String()
	deployment.Instance.ObjectMeta.SetGuid(guid)
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	summaryKey := deployment.Instance.ObjectMeta.GetSummaryId()
	_, err = manager.GetSummary(context.Background(), summaryKey, name, "default")
	assert.NotNil(t, err)

	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment, "")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Components))
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 2, len(state.TargetComponent))
	assert.Equal(t, "mock", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock", state.TargetComponent["b::T1"])

	_, err = manager.GetSummary(context.Background(), summaryKey, name, "default")
	assert.Nil(t, err)

	// Test reconcile idempotency
	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)

	// Test summary deletion
	err = manager.DeleteSummary(context.Background(), summaryKey, "default")
	assert.Nil(t, err)
	result, err := manager.GetSummary(context.Background(), summaryKey, name, "default")
	assert.Nil(t, err)
	assert.True(t, result.Summary.Removed, "Summary should have set the removed flag")
}
func TestMockGetTwoTargets(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			ObjectMeta: model.ObjectMeta{
				Name: "instance",
			},
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
			"T2": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
									Config: map[string]string{
										"id": id,
									},
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
									Role:     "mock",
									Provider: "providers.target.mock",
									Config: map[string]string{
										"id": id,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment, "")
	assert.Nil(t, err)
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 4, len(state.TargetComponent))
	assert.Equal(t, "mock", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock", state.TargetComponent["b::T1"])
	assert.Equal(t, "mock", state.TargetComponent["a::T2"])
	assert.Equal(t, "mock", state.TargetComponent["b::T2"])

	// Test reconcile idempotency
	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
}
func TestMockGetTwoTargetsTwoProviders(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
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
									Config: map[string]string{
										"id": id,
									},
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
									Config: map[string]string{
										"id": id,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock1": targetProvider,
			"mock2": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment, "")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Components))
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 2, len(state.TargetComponent))
	assert.Equal(t, "mock1", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock2", state.TargetComponent["b::T2"])

	// Test reconcile idempotency
	_, err = manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
}
func TestMockApply(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	summary, err := manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
}
func TestMockApply_InActive(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{
				ActiveState: model.ActiveState_Inactive,
			},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
								},
							},
						},
					},
				},
			},
		},
		IsInActive: true,
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	vendorContext := &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendorContext.Init(&pubSubProvider)

	manager2 := SolutionManager{}
	manager2.Init(vendorContext,
		managers.ManagerConfig{
			Name: "solution-manager",
			Type: "managers.symphony.solution",
			Properties: map[string]string{
				"providers.persistentstate": "mem-state",
				"providers.keylock":         "mem-keylock",
			},
		},
		map[string]providers.IProvider{
			"mem-state":   stateProvider,
			"mem-keylock": keyLockProvider,
			"targets":     targetProvider,
		})
	summary, err := manager2.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
	assert.Equal(t, 0, summary.SuccessCount)
}
func TestMockApplyMultiRoles(t *testing.T) {
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock2",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
								},
								{
									Role:     "mock2",
									Provider: "providers.target.mock2",
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider2 := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id1})
	targetProvider2.Init(mock.MockTargetProviderConfig{ID: id2})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock":  targetProvider,
			"mock2": targetProvider2,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	summary, err := manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
	assert.Equal(t, 2, len(summary.TargetResults["T1"].ComponentResults))
}
func TestMockApplyWithUpdateAndRemove(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock",
					},
					{
						Name: "b",
						Type: "mock",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
				Spec: &model.TargetSpec{
					Topologies: []model.TopologySpec{
						{
							Bindings: []model.BindingSpec{
								{
									Role:     "mock",
									Provider: "providers.target.mock",
								},
							},
						},
					},
				},
			},
		},
	}
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	summary, err := manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
}
func TestMockApplyWithError(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Instance: model.InstanceState{
			Spec: &model.InstanceSpec{},
		},
		Solution: model.SolutionState{
			Spec: &model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "a",
						Type: "mock1",
					},
				},
			},
		},
		Assignments: map[string]string{
			"T1": "{a}",
		},
		Targets: map[string]model.TargetState{
			"T1": {
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
	deployment.Instance.ObjectMeta.SetGuid(uuid.New().String())
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{ID: id})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	keyLockProvider := &memorykeylock.MemoryKeyLockProvider{}
	keyLockProvider.Init(memorykeylock.MemoryKeyLockProviderConfig{Mode: memorykeylock.Dedicated})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		SummaryManager: SummaryManager{
			StateProvider: stateProvider,
		},
		KeyLockProvider: keyLockProvider,
	}
	summary, err := manager.Reconcile(context.Background(), deployment, false, "default", "")
	assert.NotNil(t, err)
	assert.Equal(t, 0, summary.SuccessCount)
}
