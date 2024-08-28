/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package users

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := UsersManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.volatilestate": "StateProvider",
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
	manager := UsersManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.volatilestate": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
	err = manager.UpsertUser(context.Background(), "test", "password", []string{"testrole"})
	assert.Nil(t, err)
	err = manager.DeleteUser(context.Background(), "test")
	assert.Nil(t, err)
}

func TestUpsertAndCheck(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := UsersManager{
		StateProvider: stateProvider,
	}
	config := managers.ManagerConfig{
		Properties: map[string]string{
			"providers.volatilestate": "StateProvider",
		},
	}
	providers := make(map[string]providers.IProvider)
	providers["StateProvider"] = stateProvider
	err := manager.Init(nil, config, providers)
	assert.Nil(t, err)
	roles := []string{"testrole"}
	err = manager.UpsertUser(context.Background(), "test", "password", roles)
	assert.Nil(t, err)
	rolescheck, res := manager.CheckUser(context.Background(), "test", "wrongpassword")
	assert.False(t, res)
	assert.Nil(t, rolescheck)
	rolescheck, res = manager.CheckUser(context.Background(), "test", "password")
	assert.Equal(t, roles, rolescheck)
	assert.True(t, res)
	err = manager.DeleteUser(context.Background(), "test")
	assert.Nil(t, err)
}
