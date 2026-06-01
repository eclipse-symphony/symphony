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

func TestUpdateStageStatus(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.ActivationState{Spec: &model.ActivationSpec{}})
	assert.Nil(t, err)
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test1",
		Status:        v1alpha2.Running,
		StatusMessage: v1alpha2.Running.String(),
	})
	assert.Nil(t, err)
	state, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 1, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Running, state.Status.StageHistory[0].Status)
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test1",
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
	})
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 1, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test2",
		Status:        v1alpha2.Running,
		StatusMessage: v1alpha2.Running.String(),
	})
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 2, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	assert.Equal(t, "test2", state.Status.StageHistory[1].Stage)
	assert.Equal(t, v1alpha2.Running, state.Status.StageHistory[1].Status)
	err = manager.DeleteState(context.Background(), "test", "default")
	assert.Nil(t, err)
}

func TestUpdateStageStatusRemote(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.ActivationState{Spec: &model.ActivationSpec{}})
	assert.Nil(t, err)
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test1",
		Status:        v1alpha2.Running,
		StatusMessage: v1alpha2.Running.String(),
		Outputs: map[string]interface{}{
			"child1.status": v1alpha2.Untouched.String(),
			"child2.status": v1alpha2.Untouched.String(),
		},
	})
	assert.Nil(t, err)
	state, err := manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 1, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Running, state.Status.StageHistory[0].Status)
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test1",
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
		Outputs: map[string]interface{}{
			"__site":  "child1",
			"__stage": "test1",
		},
	})
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 1, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Paused, state.Status.StageHistory[0].Status)
	assert.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[0].Outputs["child1.status"])
	assert.Equal(t, v1alpha2.Untouched.String(), state.Status.StageHistory[0].Outputs["child2.status"])
	err = manager.ReportStageStatus(context.Background(), "test", "default", model.StageStatus{
		Stage:         "test1",
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
		Outputs: map[string]interface{}{
			"__site":  "child2",
			"__stage": "test1",
		},
	})
	assert.Nil(t, err)
	state, err = manager.GetState(context.Background(), "test", "default")
	assert.Nil(t, err)
	assert.Equal(t, "test", state.ObjectMeta.Name)
	assert.Equal(t, 1, len(state.Status.StageHistory))
	assert.Equal(t, "test1", state.Status.StageHistory[0].Stage)
	assert.Equal(t, v1alpha2.Done, state.Status.StageHistory[0].Status)
	assert.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[0].Outputs["child1.status"])
	assert.Equal(t, v1alpha2.Done.String(), state.Status.StageHistory[0].Outputs["child2.status"])
	assert.Nil(t, err)
}

/*
func TestCreateActivationWithMissingCampaignVersion(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.Validator.CampaignVersionLookupFunc = manager.CampaignVersionLookup

	err := manager.UpsertState(context.Background(), "testactivation", model.ActivationState{
		ObjectMeta: model.ObjectMeta{
			Name:      "testactivation",
			Namespace: "default",
		},
		Spec: &model.ActivationSpec{
			CampaignVersion: "testcampaignversion",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "campaignversion reference must be a valid CampaignVersion object in the same namespace")
}

func TestCreateActivationWithCampaignVersion(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.Validator.CampaignVersionLookupFunc = manager.CampaignVersionLookup
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "testcampaignversion",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "CampaignVersion",
				"metadata": model.ObjectMeta{
					Name:      "testcampaignversion",
					Namespace: "default",
				},
				"spec": model.CampaignVersionSpec{
					Stages: map[string]model.StageSpec{},
				},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaignversions",
			"kind":      "CampaignVersion",
		},
	})

	err := manager.UpsertState(context.Background(), "testactivation", model.ActivationState{
		ObjectMeta: model.ObjectMeta{
			Name:      "testactivation",
			Namespace: "default",
		},
		Spec: &model.ActivationSpec{
			CampaignVersion: "testcampaignversion",
		},
	})
	assert.Nil(t, err)
}

func TestUpdateActivationWithRunningStatus(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := ActivationsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "testactivation", model.ActivationState{
		ObjectMeta: model.ObjectMeta{
			Name:      "testactivation",
			Namespace: "default",
			Labels: map[string]string{
				"statusMessage": "Running",
			},
		},
		Spec: &model.ActivationSpec{
			CampaignVersion: "testcampaignversion",
		},
	})
	assert.Nil(t, err)

	err = manager.UpsertState(context.Background(), "testactivation", model.ActivationState{
		ObjectMeta: model.ObjectMeta{
			Name:      "testactivation",
			Namespace: "default",
			Labels: map[string]string{
				"statusMessage": "Running",
			},
		},
		Spec: &model.ActivationSpec{
			CampaignVersion: "testcampaignversion",
			Stage:    "test",
		},
	})
	assert.Contains(t, err.Error(), "spec is immutable: stage doesn't match")
}
*/
