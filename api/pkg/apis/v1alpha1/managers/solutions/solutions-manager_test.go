/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutions

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a SolutionSpec using the manager
func TestCreateGetDeleteSolutionsState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.SolutionState{})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.ObjectMeta.Name)
	specLists, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(specLists))
	assert.Equal(t, "test", specLists[0].ObjectMeta.Name)
	err = manager.DeleteState(context.Background(), "test", "default")
	assert.Nil(t, err)
	spec, err = manager.GetState(context.Background(), "test", "default")
	assert.NotNil(t, err)
}

func TestCreateSolutionWithMissingContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionValidator = validation.NewSolutionValidator(nil, manager.solutionContainerLookup, nil)
	err := manager.UpsertState(context.Background(), "test-v-v1", model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.SolutionSpec{
			DisplayName:  "test-v-v1",
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rootResource must be a valid container")
}

/*
	func TestCreateSolutionWithContainer(t *testing.T) {
		stateProvider := &memorystate.MemoryStateProvider{}
		stateProvider.Init(memorystate.MemoryStateProviderConfig{})
		manager := SolutionsManager{
			StateProvider: stateProvider,
			needValidate:  true,
		}
		manager.SolutionValidator = validation.NewSolutionValidator(nil, manager.solutionContainerLookup, nil)
		stateProvider.Upsert(context.Background(), states.UpsertRequest{
			Value: states.StateEntry{
				ID: "test",
				Body: map[string]interface{}{
					"apiVersion": model.SolutionGroup + "/v1",
					"kind":       "SolutionContainer",
					"metadata": model.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					"spec": model.SolutionContainerSpec{},
				},
				ETag: "1",
			},
			Metadata: map[string]interface{}{
				"namespace": "default",
				"group":     model.SolutionGroup,
				"version":   "v1",
				"resource":  "solutioncontainers",
				"kind":      "SolutionContainer",
			},
		})

		err := manager.UpsertState(context.Background(), "test-v-v1", model.SolutionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "test-v-v1",
				Namespace: "default",
			},
			Spec: &model.SolutionSpec{
				DisplayName:  "test-v-v1",
				RootResource: "test",
			},
		})
		assert.Nil(t, err)
	}
*/
func TestCreateSolutionWithSameDisplayName(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionValidator = validation.NewSolutionValidator(nil, nil, manager.uniqueNameSolutionLookup)
	err := manager.UpsertState(context.Background(), "test-v-v1", model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.SolutionSpec{
			DisplayName:  "test-v-v1",
			RootResource: "test",
		},
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test-v-v2", model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v2",
			Namespace: "default",
		},
		Spec: &model.SolutionSpec{
			DisplayName:  "test-v-v1",
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "solution displayName must be unique")
}

/*
func TestDeleteSolutionWithInstance(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionValidator = validation.NewSolutionValidator(manager.solutionInstanceLookup, nil, nil)
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "test",
			Body: map[string]interface{}{
				"apiVersion": model.SolutionGroup + "/v1",
				"kind":       "Instance",
				"metadata": model.ObjectMeta{
					Name:      "testinstance",
					Namespace: "default",
					Labels: map[string]string{
						"solution": "test-v-v2",
					},
				},
				"spec": model.InstanceSpec{
					DisplayName: "test",
					Solution:    "test-v-v1",
				},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "instances",
			"kind":      "Instance",
		},
	})

	err := manager.UpsertState(context.Background(), "test-v-v2", model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v2",
			Namespace: "default",
		},
		Spec: &model.SolutionSpec{
			DisplayName:  "test-v-v2",
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
	err = manager.DeleteState(context.Background(), "test-v-v2", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Solution has one or more associated instances")
}
*/
