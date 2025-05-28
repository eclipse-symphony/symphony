/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solutioncontainers

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a SolutionContainerSpec using the manager
func TestCreateGetDeleteSolutionContainersState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionContainersManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.SolutionContainerState{})
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
