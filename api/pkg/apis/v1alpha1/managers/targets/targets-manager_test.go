/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package targets

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/validation"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a TargetSpec using the manager
func TestCreateGetDeleteTargetsState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := TargetsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.TargetState{
		Spec: &model.TargetSpec{},
	})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.ObjectMeta.Name)
	specLists, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(specLists))
	assert.Equal(t, "test", specLists[0].ObjectMeta.Name)
	err = manager.DeleteSpec(context.Background(), "test", "default")
	assert.Nil(t, err)
	spec, err = manager.GetState(context.Background(), "test", "default")
	assert.NotNil(t, err)
}

func TestUpdateTargetStatus(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := TargetsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.TargetSpec{},
	})
	assert.Nil(t, err)

	var state model.TargetState

	state.ObjectMeta.Name = "test"
	state.ObjectMeta.Namespace = "default"
	state.Status = model.TargetStatus{
		Properties: map[string]string{
			"label": "test",
		},
	}
	_, err = manager.ReportState(context.Background(), state)
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, "test", state.Status.Properties["label"])
}

func TestCreateTargetWithSameDisplayName(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := TargetsManager{
		StateProvider:   stateProvider,
		needValidate:    true,
		TargetValidator: validation.TargetValidator{},
	}
	manager.TargetValidator = validation.NewTargetValidator(nil, manager.targetUniqueNameLookup)
	err := manager.UpsertState(context.Background(), "test-v-v1", model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.TargetSpec{
			DisplayName: "test-v-v1",
		},
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test-v-v2", model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v2",
			Namespace: "default",
		},
		Spec: &model.TargetSpec{
			DisplayName: "test-v-v1",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "target displayName must be unique")
}

/*
func TestDeleteTargetWithInstance(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := TargetsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.TargetValidator = validation.NewTargetValidator(manager.targetInstanceLookup, nil)
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "testinstance",
			Body: map[string]interface{}{
				"apiVersion": model.FabricGroup + "/v1",
				"kind":       "Target",
				"metadata": model.ObjectMeta{
					Name:      "testinstance",
					Namespace: "default",
					Labels: map[string]string{
						"target": "testtarget",
					},
				},
				"spec": model.InstanceSpec{
					DisplayName: "test",
					Target: model.TargetSelector{
						Name: "testtarget",
					},
				},
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

	err := manager.UpsertState(context.Background(), "testtarget", model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:      "testtarget",
			Namespace: "default",
		},
		Spec: &model.TargetSpec{
			DisplayName: "testtarget",
		},
	})
	assert.Nil(t, err)
	err = manager.DeleteSpec(context.Background(), "testtarget", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Target has one or more associated instances")
}
*/
