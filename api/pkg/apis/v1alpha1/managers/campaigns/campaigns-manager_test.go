/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package campaigns

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a CampaignSpec using the manager
func TestCreateGetDeleteCampaignSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertSpec(context.Background(), "test", model.CampaignSpec{})
	assert.Nil(t, err)
	spec, err := manager.GetSpec(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.Id)
	err = manager.DeleteSpec(context.Background(), "test")
	assert.Nil(t, err)
}
