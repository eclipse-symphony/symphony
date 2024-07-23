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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
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
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.ObjectMeta.Name)
	err = manager.DeleteState(context.Background(), "test", "default")
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
		RetentionDuration:  0,
	}
	err := manager.UpsertState(context.Background(), "test", model.ActivationState{Spec: &model.ActivationSpec{}})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.ObjectMeta.Name)
	err = manager.ReportStatus(context.Background(), "test", "default", model.ActivationStatus{
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
	})
	assert.Nil(t, err)
	errList := cleanupmanager.Poll()
	assert.Empty(t, errList)
	_, err = manager.GetState(context.Background(), "test", "default")
	assert.NotNil(t, err)
	assert.True(t, v1alpha2.IsNotFound(err))
}
