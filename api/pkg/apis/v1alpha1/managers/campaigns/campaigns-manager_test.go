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
	err := manager.UpsertState(context.Background(), "test", model.CampaignState{})
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
}

/*
func TestCreateCampaignWithMissingContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.CampaignValidator.CampaignContainerLookupFunc = manager.CampaignContainerLookup
	err := manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rootResource must be a valid container")
}

func TestCreateCampaignWithContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.CampaignValidator.CampaignContainerLookupFunc = manager.CampaignContainerLookup
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "test",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "CampaignContainer",
				"metadata": model.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				"spec": model.CampaignContainerSpec{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigncontainers",
			"kind":      "CampaignContainer",
		},
	})

	err := manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
}

func TestCreateCampaignWithRunningActivation(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
	manager.CampaignValidator.CampaignActivationsLookupFunc = manager.CampaignActivationsLookup
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "testactivation",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Activation",
				"metadata": model.ObjectMeta{
					Name:      "testactivation",
					Namespace: "default",
					Labels: map[string]string{
						"campaign":      "test-v-v1",
						"statusMessage": "Running",
					},
				},
				"spec": model.ActivationSpec{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "activations",
			"kind":      "Activation",
		},
	})

	err = manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Name: "test",
				},
			},
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Campaign has one or more running activations. Update or Deletion is not allowed")

	err = manager.DeleteState(context.Background(), "test-v-v1", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Campaign has one or more running activations. Update or Deletion is not allowed")
}

func TestCreateCampaignWithWrongStages(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Name:          "test",
					StageSelector: "wrongstage",
				},
			},
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "stageSelector must be one of the stages in the stages list")
}

func TestCreateCampaignWithWrongFirstStage(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-v1", model.CampaignState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-v1",
			Namespace: "default",
		},
		Spec: &model.CampaignSpec{
			RootResource: "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Name: "test",
				},
			},
			FirstStage: "wrongstage",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "firstStage must be one of the stages in the stages list")
}
*/
