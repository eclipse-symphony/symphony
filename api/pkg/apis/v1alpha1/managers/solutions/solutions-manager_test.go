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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a SolutionSpec using the manager
func TestCreateGetDeleteSolutionsSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SolutionsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertSpec(context.Background(), "test", model.SolutionSpec{}, "default")
	assert.Nil(t, err)
	spec, err := manager.GetSpec(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	specLists, err := manager.ListSpec(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(specLists))
	assert.Equal(t, "test", specLists[0].Id)
	err = manager.DeleteSpec(context.Background(), "test", "default")
	assert.Nil(t, err)
	spec, err = manager.GetSpec(context.Background(), "test", "default")
	assert.NotNil(t, err)
}
