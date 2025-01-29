/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package instances

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a InstanceSpec using the manager
func TestCreateGetDeleteInstancesState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := InstancesManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.InstanceState{})
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

func TestCreateInstanceWithoutSolutionTargetValidation(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := InstancesManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.InstanceValidator = validation.NewInstanceValidator(manager.instanceUniqueNameLookup, manager.solutionLookup, manager.targetLookup)

	err := manager.UpsertState(context.Background(), "test", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution",
			Target: model.TargetSelector{
				Name: "testtarget",
			},
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "target does not exist")
	assert.Contains(t, err.Error(), "solution does not exist")
}

func TestCreateInstanceWithSolutionTargetValidation(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := InstancesManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.InstanceValidator = validation.NewInstanceValidator(manager.instanceUniqueNameLookup, manager.solutionLookup, manager.targetLookup)

	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "testsolution",
			Body: map[string]interface{}{
				"apiVersion": model.SolutionGroup + "/v1",
				"kind":       "Solution",
				"metadata": model.ObjectMeta{
					Name:      "testsolution",
					Namespace: "default",
				},
				"spec": model.SolutionSpec{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.SolutionGroup,
			"version":   "v1",
			"resource":  "solutions",
			"kind":      "Solution",
		},
	})

	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "testtarget",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": model.ObjectMeta{
					Name:      "testtarget",
					Namespace: "default",
				},
				"spec": model.TargetSpec{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.FabricGroup,
			"version":   "v1",
			"resource":  "targets",
			"kind":      "Target",
		},
	})

	err := manager.UpsertState(context.Background(), "test", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution",
			Target: model.TargetSelector{
				Name: "testtarget",
			},
		},
	})
	assert.Nil(t, err)
	err = manager.UpsertState(context.Background(), "test", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution",
			Target: model.TargetSelector{
				Name: "testtarget2",
			},
		},
	})
	assert.Contains(t, err.Error(), "target does not exist")
	err = manager.UpsertState(context.Background(), "test", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution2",
			Target: model.TargetSelector{
				Name: "testtarget",
			},
		},
	})
	assert.Contains(t, err.Error(), "solution does not exist")
}

func TestCreateInstanceWithSameDisplayNameValidation(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := InstancesManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.InstanceValidator = validation.NewInstanceValidator(manager.instanceUniqueNameLookup, nil, nil)

	err := manager.UpsertState(context.Background(), "test", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution",
			Target: model.TargetSelector{
				Name: "testtarget",
			},
		},
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test2", model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Spec: &model.InstanceSpec{
			DisplayName: "test",
			Solution:    "testsolution",
			Target: model.TargetSelector{
				Name: "testtarget",
			},
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "instance displayName must be unique")
}
