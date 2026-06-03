/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package campaignversions

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

// write test case to create a CampaignVersionSpec using the manager
func TestCreateGetDeleteCampaignVersionSpec(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
	}
	err := manager.UpsertState(context.Background(), "test", model.CampaignVersionState{})
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
func TestCreateCampaignVersionWithMissingContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.CampaignVersionValidator.CampaignLookupFunc = manager.CampaignLookup
	err := manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "test",
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rootResource must be a valid container")
}

func TestCreateCampaignVersionWithContainer(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	manager.CampaignVersionValidator.CampaignLookupFunc = manager.CampaignLookup
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: "test",
			Body: map[string]interface{}{
				"apiVersion": model.WorkflowGroup + "/v1",
				"kind":       "Campaign",
				"metadata": model.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				"spec": model.CampaignSpec{},
			},
			ETag: "1",
		},
		Metadata: map[string]interface{}{
			"namespace": "default",
			"group":     model.WorkflowGroup,
			"version":   "v1",
			"resource":  "campaigns",
			"kind":      "Campaign",
		},
	})

	err := manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
}

func TestCreateCampaignVersionWithRunningActivation(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "test",
		},
	})
	assert.Nil(t, err)
	manager.CampaignVersionValidator.CampaignVersionActivationsLookupFunc = manager.CampaignVersionActivationsLookup
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
						"campaignversion":      "test-v-version1",
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

	err = manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
			RootResource: "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Name: "test",
				},
			},
		},
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CampaignVersion has one or more running activations. Update or Deletion is not allowed")

	err = manager.DeleteState(context.Background(), "test-v-version1", "default")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CampaignVersion has one or more running activations. Update or Deletion is not allowed")
}

func TestCreateCampaignVersionWithWrongStages(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
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

func TestCreateCampaignVersionWithWrongFirstStage(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := CampaignVersionsManager{
		StateProvider: stateProvider,
		needValidate:  true,
	}
	err := manager.UpsertState(context.Background(), "test-v-version1", model.CampaignVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      "test-v-version1",
			Namespace: "default",
		},
		Spec: &model.CampaignVersionSpec{
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
