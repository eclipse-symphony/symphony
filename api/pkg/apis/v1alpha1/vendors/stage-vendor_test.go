/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
)

func TestStageEndpoints(t *testing.T) {
	vendor := createStageVendor()
	vendor.Route = "stage"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 0, len(endpoints))
}

func TestStageInfo(t *testing.T) {
	vendor := createStageVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func createStageVendor() StageVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := StageVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "stage-manager",
				Type: "managers.symphony.stage",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
			{
				Name: "campaigns-manager",
				Type: "managers.symphony.campaigns",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
			{
				Name: "activations-manager",
				Type: "managers.symphony.activations",
				Properties: map[string]string{
					"providers.persistentstate": "mem-state",
				},
				Providers: map[string]managers.ProviderConfig{
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"stage-manager": {
			"mem-state": &stateProvider,
		},
		"campaigns-manager": {
			"mem-state": &stateProvider,
		},
		"activations-manager": {
			"mem-state": &stateProvider,
		},
	}, &pubSubProvider)
	return vendor
}

// Comment out this test temporarily due to data racing issue in memory state provider: https://github.com/eclipse-symphony/symphony/issues/84
// func TestStageActivateCampaign(t *testing.T) {
// 	vendor := createStageVendor()
// 	vendor.Context.EvaluationContext = &utils.EvaluationContext{}
// 	err := vendor.CampaignsManager.UpsertSpec(context.Background(), "test-campaign", model.CampaignSpec{
// 		Name:        "test-campaign",
// 		SelfDriving: true,
// 		FirstStage:  "test",
// 		Stages: map[string]model.StageSpec{
// 			"test": {
// 				Provider: "providers.stage.mock",
// 				Inputs: map[string]interface{}{
// 					"foo": "${{$output(test,foo)}}",
// 				},
// 				StageSelector: "${{$if($lt($output(test,foo), 5), test, '')}}",
// 			},
// 		},
// 	})
// 	assert.Nil(t, err)
// 	err = vendor.ActivationsManager.UpsertSpec(context.Background(), "test-activation", model.ActivationSpec{
// 		Campaign: "test-campaign",
// 		Name:     "test-activation",
// 		Stage:    "test",
// 		Inputs: map[string]interface{}{
// 			"foo": 0,
// 		},
// 		Generation: "1",
// 	})
// 	assert.Nil(t, err)
// 	vendor.Context.Publish("activation", v1alpha2.Event{
// 		Body: v1alpha2.ActivationData{
// 			Campaign:   "test-campaign",
// 			Activation: "test-activation",
// 			Stage:      "test",
// 			Inputs: map[string]interface{}{
// 				"foo": 0,
// 			},
// 		},
// 	})
// 	vendor.Context.Publish("job-report", v1alpha2.Event{
// 		Body: model.ActivationStatus{
// 			Status: v1alpha2.Done,
// 			Outputs: map[string]interface{}{
// 				"__campaign":             "test-campaign",
// 				"__activation":           "test-activation",
// 				"__stage":                "test",
// 				"__activationGeneration": "1",
// 				"__site":                 "fake",
// 			},
// 		},
// 	})
// 	vendor.Context.Publish("remote-job", v1alpha2.Event{
// 		Body: v1alpha2.JobData{
// 			Body: v1alpha2.InputOutputData{
// 				Inputs: map[string]interface{}{
// 					"__activation":           "test-activation",
// 					"__activationGeneration": "1",
// 					"__campaign":             "test-campaign",
// 					"__stage":                "test",
// 					"operation":              "wait",
// 				},
// 			},
// 		},
// 	})
// 	vendor.Context.Publish("remote-job", v1alpha2.Event{
// 		Body: v1alpha2.JobData{
// 			Body: v1alpha2.InputOutputData{
// 				Inputs: map[string]interface{}{
// 					"__activation":           "test-activation",
// 					"__activationGeneration": "1",
// 					"__campaign":             "test-campaign",
// 					"__stage":                "test",
// 					"operation":              "materialize",
// 				},
// 			},
// 		},
// 	})
// 	vendor.Context.Publish("remote-job", v1alpha2.Event{
// 		Body: v1alpha2.JobData{
// 			Body: v1alpha2.InputOutputData{
// 				Inputs: map[string]interface{}{
// 					"__activation":           "test-activation",
// 					"__activationGeneration": "1",
// 					"__campaign":             "test-campaign",
// 					"__stage":                "test",
// 					"operation":              "mock",
// 				},
// 			},
// 		},
// 	})
// 	time.Sleep(2 * time.Second)

// 	activation, err := vendor.ActivationsManager.GetSpec(context.Background(), "test-activation")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, activation)
// 	assert.Equal(t, "test-activation", activation.Id)
// 	assert.NotNil(t, activation.Status.UpdateTime)
//  assert.True(t, v1alpha2.Done.EqualsWithString(activation.Status.Status))
// }
