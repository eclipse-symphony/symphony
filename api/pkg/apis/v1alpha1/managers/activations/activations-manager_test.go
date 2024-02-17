/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestCreateGetDeleteActivationSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.ActivationState{Spec: &model.ActivationSpec{}})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	err = manager.DeleteState(context.Background(), "test")
	assert.Nil(t, err)
}

func TestCleanupOldActivationSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})

	manager := ActivationsManager{
		StateProvider: stateProvider,
	}
	cleanupmanager := ActivationsCleanupManager{
		ActivationsManager: manager,
		RetentionInMinutes: 0,
	}
	err := manager.UpsertState(context.Background(), "test", model.ActivationState{Spec: &model.ActivationSpec{}})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	err = manager.ReportStatus(context.Background(), "test", model.ActivationStatus{Status: 9996})
	assert.Nil(t, err)
	errList := cleanupmanager.Poll()
	assert.Empty(t, errList)
	_, err = manager.GetState(context.Background(), "test")
	assert.NotNil(t, err)
}
