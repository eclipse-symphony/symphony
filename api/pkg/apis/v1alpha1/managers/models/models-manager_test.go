/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
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
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state-fail",
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
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.ModelSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
			Constraints: "constraints",
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
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.ModelSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
			Constraints: "constraints",
		},
	})
	assert.Nil(t, err)
	list, err := manager.ListState(context.Background(), "default")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "test", list[0].ObjectMeta.Name)
	assert.Equal(t, "device", list[0].Spec.DisplayName)
	assert.Equal(t, "a", list[0].Spec.Properties["a"])
}

func TestUpsertAndGetSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state",
		},
	}, map[string]providers.IProvider{
		"memory-state": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "test", model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "test",
		},
		Spec: &model.ModelSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
			Constraints: "constraints",
		},
	})
	assert.Nil(t, err)
	spec, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", spec.ObjectMeta.Name)
	assert.Equal(t, "device", spec.Spec.DisplayName)
	assert.Equal(t, "a", spec.Spec.Properties["a"])
}

func TestUpsertSpecFail(t *testing.T) {
	stateProvider := &MemoryStateProviderFail{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state-fail",
		},
	}, map[string]providers.IProvider{
		"memory-state-fail": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "mockError", model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "mockError",
		},
		Spec: &model.ModelSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
			Constraints: "constraints",
		},
	})
	assert.NotNil(t, err)
}

func TestListSpecFail(t *testing.T) {
	stateProvider := &MemoryStateProviderFail{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state-fail",
		},
	}, map[string]providers.IProvider{
		"memory-state-fail": stateProvider,
	})
	assert.Nil(t, err)

	_, err = manager.ListState(context.Background(), "default")
	assert.NotNil(t, err)
}

func TestGetSpecFail(t *testing.T) {
	stateProvider := &MemoryStateProviderFail{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state-fail",
		},
	}, map[string]providers.IProvider{
		"memory-state-fail": stateProvider,
	})
	assert.Nil(t, err)

	_, err = manager.GetState(context.Background(), "mockError", "default")
	assert.NotNil(t, err)

	_, err = manager.GetState(context.Background(), "mockJsonError", "default")
	assert.NotNil(t, err)
}

func TestUpsertAndListSpecFail(t *testing.T) {
	stateProvider := &MemoryStateProviderFail{}
	err := stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	assert.Nil(t, err)
	manager := ModelsManager{}
	err = manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "memory-state-fail",
		},
	}, map[string]providers.IProvider{
		"memory-state-fail": stateProvider,
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "mockJsonError", model.ModelState{
		ObjectMeta: model.ObjectMeta{
			Name: "mockJsonError",
		},
		Spec: &model.ModelSpec{
			DisplayName: "device",
			Properties: map[string]string{
				"a": "a",
			},
			Constraints: "constraints",
		},
	})
	assert.Nil(t, err)

	_, err = manager.ListState(context.Background(), "default")
	assert.NotNil(t, err)
}

type MemoryStateProviderFail struct {
	Data map[string]interface{}
}

func (m *MemoryStateProviderFail) Init(config providers.IProviderConfig) error {
	m.Data = make(map[string]interface{})
	return nil
}

func (m *MemoryStateProviderFail) Upsert(ctx context.Context, request states.UpsertRequest) (string, error) {
	if request.Value.ID == "mockError" {
		return "", assert.AnError
	} else {
		if request.Value.ID == "mockJsonError" {
			request.Value.Body = map[string]interface{}{
				"spec": []byte("invalid json"),
			}
		}
		m.Data[request.Value.ID] = request.Value
		return request.Value.ID, nil
	}
}

func (m *MemoryStateProviderFail) Delete(context.Context, states.DeleteRequest) error {
	return assert.AnError
}

func (m *MemoryStateProviderFail) Get(ctx context.Context, getRequest states.GetRequest) (states.StateEntry, error) {
	if getRequest.ID == "mockError" {
		return states.StateEntry{}, assert.AnError
	}

	if getRequest.ID == "mockJsonError" {
		return states.StateEntry{
			ID: "mockJsonError",
			Body: map[string]interface{}{
				"spec": []byte("invalid json"),
			},
		}, nil
	}

	var err error
	if v, ok := m.Data[getRequest.ID]; ok {
		vE, ok := v.(states.StateEntry)
		if ok {
			return vE, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", getRequest.ID), v1alpha2.InternalError)
			return states.StateEntry{}, err
		}
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", getRequest.ID), v1alpha2.NotFound)
	return states.StateEntry{}, err
}

func (m *MemoryStateProviderFail) GetLatest(ctx context.Context, getRequest states.GetRequest) (states.StateEntry, error) {
	return states.StateEntry{}, nil
}

func (m *MemoryStateProviderFail) List(context.Context, states.ListRequest) ([]states.StateEntry, string, error) {
	if (m.Data == nil) || (len(m.Data) == 0) {
		return nil, "", assert.AnError
	}
	var entities []states.StateEntry
	for _, v := range m.Data {
		vE, ok := v.(states.StateEntry)
		if ok {
			entities = append(entities, vE)
		} else {
			parseErr := v1alpha2.NewCOAError(nil, "found invalid state entry", v1alpha2.InternalError)
			return entities, "", parseErr
		}
	}
	return entities, "", nil
}

func (m *MemoryStateProviderFail) SetContext(context *contexts.ManagerContext) {
}
