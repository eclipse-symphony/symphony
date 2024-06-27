/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package skills

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
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	manager := SkillsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)
}

func TestInitFail(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	manager := SkillsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "memory-state-fail",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.NotNil(t, err)
}

func TestUpsertAndDeleteSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	manager := SkillsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.SkillState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.SkillSpec{
			Parameters: map[string]string{
				"a": "default-a",
				"c": "default-c",
			},
			Nodes: []model.NodeSpec{
				{
					Id: "1",
					Configurations: map[string]string{
						"v-a": "$param(a)",
						"v-c": "$param(c)",
					},
				},
			},
		},
	})
	assert.Nil(t, err)

	err = manager.DeleteState(context.Background(), "test", "default")
	assert.Nil(t, err)
}

func TestUpsertAndListSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	manager := SkillsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.SkillState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.SkillSpec{
			Parameters: map[string]string{
				"a": "default-a",
				"c": "default-c",
			},
			Nodes: []model.NodeSpec{
				{
					Id: "1",
					Configurations: map[string]string{
						"v-a": "$param(a)",
						"v-c": "$param(c)",
					},
				},
			},
		},
	})
	assert.Nil(t, err)

	skills, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(skills))
	assert.Equal(t, "test", skills[0].ObjectMeta.Name)
	assert.Equal(t, "default-a", skills[0].Spec.Parameters["a"])
	assert.Equal(t, "default-c", skills[0].Spec.Parameters["c"])
}

func TestUpsertAndGetSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)

	manager := SkillsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.persistentstate": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.SkillState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.SkillSpec{
			Parameters: map[string]string{
				"a": "default-a",
				"c": "default-c",
			},
			Nodes: []model.NodeSpec{
				{
					Id: "1",
					Configurations: map[string]string{
						"v-a": "$param(a)",
						"v-c": "$param(c)",
					},
				},
			},
		},
	})
	assert.Nil(t, err)

	skill, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", skill.ObjectMeta.Name)
	assert.Equal(t, "default-a", skill.Spec.Parameters["a"])
	assert.Equal(t, "default-c", skill.Spec.Parameters["c"])
}
