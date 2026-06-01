/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutionversions

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a SolutionVersionSpec using the manager
func TestCreateGetDeleteSolutionVersionsState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionVersionsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.SolutionVersionState{})
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

func TestCreateSolutionVersionWithMissingContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionVersionValidator = validation.NewSolutionVersionValidator(nil, manager.solutionversionContainerLookup, nil)
	err := manager.UpsertState(context.Background(), "test-v-version1", model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.SolutionVersionSpec{
			DisplayName:  "test-v-version1",
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rootResource must be a valid container")
}

/*
	func TestCreateSolutionVersionWithContainer(t *testing.T) {
		stateProvider := &memorystate.MemoryStateProvider{}
		stateProvider.Init(memorystate.MemoryStateProviderConfig{})
		manager := SolutionVersionsManager{
			StateProvider: stateProvider,
			needValidate:  true,
		}
		manager.SolutionVersionValidator = validation.NewSolutionVersionValidator(nil, manager.solutionversionContainerLookup, nil)
		stateProvider.Upsert(context.Background(), states.UpsertRequest{
			Value: states.StateEntry{
				ID: "test",
				Body: map[string]interface{}{
					"apiVersion": model.SolutionVersionGroup + "/v1",
					"kind":       "Solution",
					"metadata": model.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					"spec": model.SolutionSpec{},
				},
				ETag: "1",
			},
			Metadata: map[string]interface{}{
				"namespace": "default",
				"group":     model.SolutionVersionGroup,
				"version":   "v1",
				"resource":  "solutions",
				"kind":      "Solution",
			},
		})

		err := manager.UpsertState(context.Background(), "test-v-version1", model.SolutionVersionState{
			ObjectMeta: model.ObjectMeta{
				Name:      "test-v-version1",
				Namespace: "default",
			},
			Spec: &model.SolutionVersionSpec{
				DisplayName:  "test-v-version1",
				RootResource: "test",
			},
		})
		assert.Nil(t, err)
	}
*/
func TestCreateSolutionVersionWithSameDisplayName(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionVersionValidator = validation.NewSolutionVersionValidator(nil, nil, manager.uniqueNameSolutionVersionLookup)
	err := manager.UpsertState(context.Background(), "test-v-version1", model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.SolutionVersionSpec{
			DisplayName:  "test-v-version1",
			RootResource: "test",
		},
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test-v-version2", model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version2",
			Namespace: "default",
		},
		Spec: &model.SolutionVersionSpec{
			DisplayName:  "test-v-version1",
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "solutionversion displayName must be unique")
}

/*
func TestDeleteSolutionVersionWithInstance(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.SolutionVersionValidator = validation.NewSolutionVersionValidator(manager.solutionversionInstanceLookup, nil, nil)
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "test",
			Body: map[string]interface{}{
				"apiVersion": model.SolutionVersionGroup + "/v1",
				"kind":       "Instance",
				"metadata": model.ObjectMeta{
					Name:      "testinstance",
					Namespace: "default",
					Labels: map[string]string{
						"solutionversion": "test-v-version2",
					},
				},
				"spec": model.InstanceSpec{
					DisplayName: "test",
					SolutionVersion:    "test-v-version1",
				},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.SolutionVersionGroup,
			"version":   "v1",
			"resource":  "instances",
			"kind":      "Instance",
		},
	})

	err := manager.UpsertState(context.Background(), "test-v-version2", model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version2",
			Namespace: "default",
		},
		Spec: &model.SolutionVersionSpec{
			DisplayName:  "test-v-version2",
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
	err = manager.DeleteState(context.Background(), "test-v-version2", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SolutionVersion has one or more associated instances")
}
*/
