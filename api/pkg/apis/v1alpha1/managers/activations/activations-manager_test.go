/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package activations

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestCreateGetDeleteActivationSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertSpec(context.Background(), "test", model.ActivationSpec{})
	assert.Nil(t, err)
	spec, err := manager.GetSpec(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	err = manager.DeleteSpec(context.Background(), "test")
	assert.Nil(t, err)
}
