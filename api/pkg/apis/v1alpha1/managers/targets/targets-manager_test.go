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
	err := manager.UpsertState(context.Background(), "test", "default", model.TargetState{
		Spec: &model.TargetSpec{},
	})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	specLists, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(specLists))
	assert.Equal(t, "test", specLists[0].Id)
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
	err := manager.UpsertState(context.Background(), "test", "default", model.TargetState{
		Id:   "test",
		Spec: &model.TargetSpec{},
	})
	assert.Nil(t, err)

	var state model.TargetState

	state.Id = "test"
	state.Namespace = "default"
	state.Status = map[string]string{"label": "test"}
	_, err = manager.ReportState(context.Background(), state)
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.Id)
	assert.Equal(t, "test", state.Status["label"])
}
