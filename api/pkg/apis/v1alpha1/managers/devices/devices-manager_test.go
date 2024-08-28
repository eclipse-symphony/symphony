/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package devices

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := DevicesManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
}

func TestUpsertAndDelete(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := DevicesManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
	deviceState := model.DeviceState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.DeviceSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
		},
	}
	err = manager.UpsertState(context.Background(), "test", deviceState)
	assert.Nil(t, err)
	err = manager.DeleteState(context.Background(), "test", "default")
	assert.Nil(t, err)
}

func TestUpsertAndGet(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := DevicesManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
	deviceState := model.DeviceState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: &model.DeviceSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
		},
	}
	err = manager.UpsertState(context.Background(), "test", deviceState)
	assert.Nil(t, err)
	state, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, deviceState, state)
	states, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(states))
	assert.Equal(t, deviceState, states[len(states)-1])
}
