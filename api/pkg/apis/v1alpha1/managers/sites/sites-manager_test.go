/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package sites

import (
	"context"
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a SitesSpec using the manager
func TestCreateGetDeleteTargetsSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SitesManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertSpec(context.Background(), "test", model.SiteSpec{})
	assert.Nil(t, err)
	spec, err := manager.GetSpec(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	specLists, err := manager.ListSpec(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(specLists))
	assert.Equal(t, "test", specLists[0].Id)
	err = manager.DeleteSpec(context.Background(), "test")
	assert.Nil(t, err)
	spec, err = manager.GetSpec(context.Background(), "test")
	assert.NotNil(t, err)
}

func TestUpdateTargetStatus(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := SitesManager{
		StateProvider: stateProvider,
	}
	var state model.SiteState
	state.Id = "test"
	state.Spec = &model.SiteSpec{}
	var status model.SiteStatus
	status.IsOnline = true
	state.Id = "test"
	state.Status = &status
	err := manager.ReportState(context.Background(), state)
	assert.Nil(t, err)
	spec, err := manager.GetSpec(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	assert.Equal(t, true, spec.Status.IsOnline)
	assert.NotEqual(t, "", spec.Status.LastReported)
}
