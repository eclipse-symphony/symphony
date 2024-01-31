/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

func TestCampaignWithSingleMockStageLoop(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Inputs: map[string]interface{}{
			"foo": 0,
		},
		Outputs:  nil,
		Provider: "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider: "providers.stage.mock",
					Inputs: map[string]interface{}{
						"foo": "${{$output(test,foo)}}",
					},
					StageSelector: "${{$if($lt($output(test,foo), 5), test, '')}}",
					Contexts:      "fake",
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(5), status.Outputs["foo"])
}
func TestCampaignWithSingleCounterStageLoop(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Inputs: map[string]interface{}{
			"foo": 1,
		},
		Outputs:  nil,
		Provider: "providers.stage.counter",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.counter",
					StageSelector: "${{$if($lt($output(test,foo), 5), test, '')}}",
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(5), status.Outputs["foo"])
}

func TestCampaignWithSingleMegativeCounterStageLoop(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Inputs: map[string]interface{}{
			"foo": -10,
		},
		Outputs:  nil,
		Provider: "providers.stage.counter",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.counter",
					StageSelector: "${{$if($gt($output(test,foo), -50), test, '')}}",
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(-50), status.Outputs["foo"])
}

func TestCampaignWithTwoCounterStageLoop(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Inputs: map[string]interface{}{
			"foo": 1,
			"bar": 1,
		},
		Outputs:  nil,
		Provider: "providers.stage.counter",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.counter",
					StageSelector: "test2",
				},
				"test2": {
					Provider:      "providers.stage.counter",
					StageSelector: "${{$if($lt($output(test2,bar), 5), test, '')}}",
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(5), status.Outputs["foo"])
	assert.Equal(t, int64(5), status.Outputs["bar"])
}

func TestCampaignWithHTTPCounterStageLoop(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "https://www.bing.com",
					},
				},
				"test2": {
					Provider: "providers.stage.counter",
					Inputs: map[string]interface{}{
						"success": "${{$if($equal($output(test, status), 200), 1, 0)}}",
					},
					StageSelector: "${{$if($lt($output(test2,success), 5), test, '')}}",
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(5), status.Outputs["success"])
}
func TestCampaignWithDelay(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.delay",
	}
	timeStamp := time.Now().UTC()
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider: "providers.stage.delay",
					Inputs: map[string]interface{}{
						"delay": 5,
					},
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, v1alpha2.OK, status.Outputs[v1alpha2.StatusOutput])
	assert.True(t, time.Now().UTC().Sub(timeStamp) > 5*time.Second)
}
func TestErrorHandler(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "bad url",
					},
				},
				"test2": {
					Provider: "providers.stage.counter",
					Inputs: map[string]interface{}{
						"success": "${{$if($equal($output(test, __status), 200), 1, 0)}}",
					},
					StageSelector: "",
					HandleErrors:  true,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, int64(0), status.Outputs["success"])
}
func TestErrorHandlerNotSet(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "bad url",
					},
				},
				"test2": {
					Provider: "providers.stage.counter",
					Inputs: map[string]interface{}{
						"success": "${{$if($equal($output(test, __status), 200), 1, 0)}}",
					},
					StageSelector: "",
					HandleErrors:  false,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.InternalError, status.Status)
}
func TestAccessingPreviousStage(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		_, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "https://www.bing.com",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  false,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, "test", activation.TriggeringStage)
	}
}

func TestAccessingStageStatus(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	var status model.ActivationStatus
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "${{$if($equal($output(test, __status), 200), test2, '')}}",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "https://www.bing.com",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  false,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, "test", activation.TriggeringStage)
		assert.Equal(t, "test2", status.NextStage)
	}
}

func TestIntentionalError(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"__status": 400,
						"__error":  "bad",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  true,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, v1alpha2.BadRequest, status.Outputs["__status"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
}
func TestIntentionalErrorState(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"__status": v1alpha2.DeleteFailed,
						"__error":  "failed",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  true,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, v1alpha2.DeleteFailed, status.Outputs["__status"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
}
func TestIntentionalErrorString(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"__status": "400",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  true,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, v1alpha2.InternalError, status.Outputs["__status"]) // non-successful state is returned without __error, set to InternalError
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
}
func TestIntentionalErrorStringProper(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var status model.ActivationStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"__status": "400",
						"__error":  "this_is_an_error",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  true,
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
		assert.Equal(t, v1alpha2.BadRequest, status.Outputs["__status"]) // non-successful state is returned without __error, set to InternalError
		assert.Equal(t, "this_is_an_error", status.Outputs["__error"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
}
func TestAccessingPreviousStageInExpression(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	var status model.ActivationStatus
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			Name:        "test-campaign",
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"ticket": "bar",
					},
				},
				"test2": {
					Provider:      "providers.stage.mock",
					StageSelector: "",
					HandleErrors:  false,
					Inputs: map[string]interface{}{
						"stcheck": "${{$output($input(__previousStage), __status)}}",
						"stfoo":   "${{$output($input(__previousStage), ticket)}}",
					},
				},
			},
		}, *activation)

		if activation == nil {
			assert.Equal(t, v1alpha2.OK, status.Outputs["stcheck"])
			assert.Equal(t, "bar", status.Outputs["stfoo"])
			break
		}
	}
}
func TestResumeStage(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	campaignName := "campaign1"
	activationName := "activation1"
	activationGen := "1"
	output := map[string]interface{}{
		"__campaign":             campaignName,
		"__activation":           activationName,
		"__activationGeneration": activationGen,
		"__site":                 "fake",
		"__stage":                "test",
	}
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: fmt.Sprintf("%s-%s-%s", campaignName, activationName, activationGen),
			Body: PendingTask{
				Sites:         []string{"fake"},
				OutputContext: map[string]map[string]interface{}{"fake": output},
			},
		},
	})
	activation := model.ActivationStatus{
		Status:  v1alpha2.Done,
		Stage:   "test",
		Outputs: output,
	}
	campaign := model.CampaignSpec{
		Name:        campaignName,
		SelfDriving: true,
		FirstStage:  "test",
		Stages: map[string]model.StageSpec{
			"test": {
				Provider:      "providers.stage.mock",
				StageSelector: "test2",
				Inputs: map[string]interface{}{
					"ticket": "bar",
				},
			},
			"test2": {
				Provider:      "providers.stage.mock",
				StageSelector: "",
				HandleErrors:  false,
				Inputs: map[string]interface{}{
					"stcheck": "${{$output($input(__previousStage), __status)}}",
					"stfoo":   "${{$output($input(__previousStage), ticket)}}",
				},
			},
		},
	}
	nextStage, err := manager.ResumeStage(activation, campaign)
	assert.Nil(t, err)
	assert.Equal(t, "test2", nextStage.Stage)
	assert.Equal(t, campaignName, nextStage.Campaign)
	assert.Equal(t, activationName, nextStage.Activation)
}
func TestHandleDirectTriggerEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	activation := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Inputs: map[string]interface{}{
			"foo":     1,
			"bar":     2,
			"__state": map[string]interface{}{"foo": 1, "bar": 1},
		},
		Outputs:  nil,
		Provider: "providers.stage.counter",
	}
	status := manager.HandleDirectTriggerEvent(context.Background(), activation)
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	assert.Equal(t, "test-activation", status.Outputs["__activation"])
	assert.Equal(t, "1", status.Outputs["__activationGeneration"])
	assert.Equal(t, "fake", status.Outputs["__site"])
	assert.Equal(t, "test", status.Outputs["__stage"])
	assert.Equal(t, int64(2), status.Outputs["foo"])
	assert.Equal(t, int64(3), status.Outputs["bar"])
}
func TestHandleDirectTriggerScheduleEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	activation := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Inputs: map[string]interface{}{
			"foo":     1,
			"bar":     2,
			"__state": map[string]interface{}{"foo": 1, "bar": 1},
		},
		Outputs:  nil,
		Provider: "providers.stage.counter",
		Schedule: &v1alpha2.ScheduleSpec{
			Date: "2020-01-01",
			Time: "12:00:00PM",
			Zone: "PST",
		},
	}
	status := manager.HandleDirectTriggerEvent(context.Background(), activation)
	assert.Equal(t, v1alpha2.Paused, status.Status)

	assert.Equal(t, v1alpha2.Delayed, status.Outputs["__status"])
	assert.Equal(t, false, status.IsActive)
}
func TestHandleActivationEvent(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}

	activationData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.mock",
	}

	campaign := model.CampaignSpec{
		Name:        "test-campaign",
		SelfDriving: true,
		FirstStage:  "test",
		Stages: map[string]model.StageSpec{
			"test": {
				Provider:      "providers.stage.mock",
				StageSelector: "test2",
				Inputs: map[string]interface{}{
					"ticket": "bar",
				},
			},
			"test2": {
				Provider:      "providers.stage.mock",
				StageSelector: "",
				HandleErrors:  false,
				Inputs: map[string]interface{}{
					"stcheck": "${{$output($input(__previousStage), __status)}}",
					"stfoo":   "${{$output($input(__previousStage), ticket)}}",
				},
			},
		},
	}

	activationState := model.ActivationState{
		Spec: &model.ActivationSpec{
			Inputs: map[string]interface{}{
				"foo":     1,
				"bar":     2,
				"__state": map[string]interface{}{"foo": 1, "bar": 1},
			},
		},
	}
	output, err := manager.HandleActivationEvent(context.Background(), activationData, campaign, activationState)
	assert.Nil(t, err)
	assert.Equal(t, "test-campaign", output.Campaign)
	assert.Equal(t, "test-activation", output.Activation)
	assert.Equal(t, "1", output.ActivationGeneration)
	assert.Equal(t, "test", output.Stage)
	assert.Equal(t, int(1), output.Inputs["foo"])
	assert.Equal(t, int(2), output.Inputs["bar"])
	assert.Equal(t, "providers.stage.mock", output.Provider)
}
func TestTriggerEventWithSchedule(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}

	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Inputs: map[string]interface{}{
			"foo": 0,
		},
		Outputs:  nil,
		Provider: "providers.stage.mock",
		Schedule: &v1alpha2.ScheduleSpec{
			Date: "2020-01-01",
			Time: "12:00:00PM",
			Zone: "PST",
		},
	}

	status, _ := manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
		Name:        "test-campaign",
		SelfDriving: true,
		FirstStage:  "test",
		Stages: map[string]model.StageSpec{
			"test": {
				Provider: "providers.stage.mock",
				Inputs: map[string]interface{}{
					"foo": "${{$output(test,foo)}}",
				},
				StageSelector: "${{$if($lt($output(test,foo), 5), test, '')}}",
				Contexts:      "fake",
			},
		},
	}, *activation)
	assert.Equal(t, v1alpha2.Paused, status.Status)
	assert.Equal(t, false, status.IsActive)
}
