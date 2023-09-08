/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package solution

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/mock"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFindAgentEmpty(t *testing.T) {
	agent := findAgent(model.TargetSpec{})
	assert.Equal(t, "", agent)
}
func TestFindAgentMatch(t *testing.T) {
	agent := findAgent(model.TargetSpec{
		Components: []model.ComponentSpec{
			{
				Name: "symphony-agent",
				Properties: map[string]interface{}{
					model.ContainerImage: "possprod.azurecr.io/symphony-agent:0.38.0",
				},
			},
		},
	})
	assert.Equal(t, "symphony-agent", agent)
}
func TestFindAgentNotMatch(t *testing.T) {
	agent := findAgent(model.TargetSpec{
		Components: []model.ComponentSpec{
			{
				Name: "symphony-agent",
				Properties: map[string]interface{}{
					model.ContainerImage: "possprod.azurecr.io/symphony-api:0.38.0",
				},
			},
		},
	})
	assert.Equal(t, "", agent)
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
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
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
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetSpec{
			"T1": {
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
	}
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		StateProvider: stateProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	_, err = manager.Reconcile(context.Background(), deployment, false)
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Components))
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 2, len(state.TargetComponent))
	assert.Equal(t, "mock", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock", state.TargetComponent["b::T1"])
}
func TestMockGetTwoTargets(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
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
		Assignments: map[string]string{
			"T1": "{a}{b}",
			"T2": "{a}{b}",
		},
		Targets: map[string]model.TargetSpec{
			"T1": {
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
			"T2": {
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
	}
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		StateProvider: stateProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	_, err = manager.Reconcile(context.Background(), deployment, false)
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment)
	assert.Nil(t, err)
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 4, len(state.TargetComponent))
	assert.Equal(t, "mock", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock", state.TargetComponent["b::T1"])
	assert.Equal(t, "mock", state.TargetComponent["a::T2"])
	assert.Equal(t, "mock", state.TargetComponent["b::T2"])
}
func TestMockGetTwoTargetsTwoProviders(t *testing.T) {
	id := uuid.New().String()
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
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
		Assignments: map[string]string{
			"T1": "{a}",
			"T2": "{b}",
		},
		Targets: map[string]model.TargetSpec{
			"T1": {
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
			"T2": {
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
	}
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		StateProvider: stateProvider,
	}
	state, components, err := manager.Get(context.Background(), deployment)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(components))
	assert.Equal(t, 0, len(state.TargetComponent))

	_, err = manager.Reconcile(context.Background(), deployment, false)
	assert.Nil(t, err)

	state, _, err = manager.Get(context.Background(), deployment)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Components))
	assert.Equal(t, "a", state.Components[0].Name)
	assert.Equal(t, "b", state.Components[1].Name)
	assert.Equal(t, 4, len(state.TargetComponent))
	assert.Equal(t, "mock1", state.TargetComponent["a::T1"])
	assert.Equal(t, "mock2", state.TargetComponent["b::T2"])
}
func TestMockApply(t *testing.T) {
	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
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
		Assignments: map[string]string{
			"T1": "{a}{b}",
		},
		Targets: map[string]model.TargetSpec{
			"T1": {
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
	}
	targetProvider := &mock.MockTargetProvider{}
	targetProvider.Init(mock.MockTargetProviderConfig{})
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionManager{
		TargetProviders: map[string]target.ITargetProvider{
			"mock": targetProvider,
		},
		StateProvider: stateProvider,
	}
	summary, err := manager.Reconcile(context.Background(), deployment, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, summary.SuccessCount)
}
