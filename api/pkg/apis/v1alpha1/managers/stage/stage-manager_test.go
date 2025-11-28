/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package stage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Inputs: map[string]interface{}{
			"foo": 0,
		},
		Outputs:   nil,
		Provider:  "providers.stage.mock",
		Namespace: "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(5), status.Outputs["foo"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, "1", status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.counter",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Inputs: map[string]interface{}{
						"foo": 1,
					},
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(5), status.Outputs["foo"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, int64(5), status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
}

func TestCampaignWithSingleNegativeCounterStageLoop(t *testing.T) {
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.counter",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Inputs: map[string]interface{}{
						"foo": -10,
					},
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(-50), status.Outputs["foo"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, int64(5), status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.counter",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
					Inputs: map[string]interface{}{
						"foo": 1,
						"bar": 1,
					},
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(5), status.Outputs["foo"])
	assert.Equal(t, int64(5), status.Outputs["bar"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, int64(5), status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test2", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
}

func TestCampaignWithTriggersCounterStageLoop(t *testing.T) {
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Inputs: map[string]interface{}{
			"foo": 1,
			"bar": 1,
		},
		Outputs:   nil,
		Provider:  "providers.stage.counter",
		Namespace: "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.counter",
					StageSelector: "${{$if($lt($output(test,foo), 5), test, test2)}}",
					Inputs: map[string]interface{}{
						"count": 1,
						"foo":   "${{$trigger(foo, 0)}}",
					},
				},
				"test2": {
					Provider:      "providers.stage.counter",
					StageSelector: "${{$if($lt($output(test2,foo), 10), test2, '')}}",
					Inputs: map[string]interface{}{
						"count.init": "${{$output(test, count)}}",
						"count":      1,
						"foo":        "${{$output(test, foo)}}",
						"bar":        "${{$trigger(bar, -1)}}",
						"bazz":       "${{$trigger(bazz, 0)}}",
					},
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(7), status.Outputs["count"])
	assert.Equal(t, int64(10), status.Outputs["foo"])
	assert.Equal(t, int64(2), status.Outputs["bar"])
	assert.Equal(t, int64(0), status.Outputs["bazz"])
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.http",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(5), status.Outputs["success"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, int64(5), status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test2", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.delay",
		Namespace:            "fakens",
	}
	timeStamp := time.Now().UTC()
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	// assert.Equal(t, v1alpha2.OK, status.Outputs[v1alpha2.StatusOutput])
	assert.True(t, time.Now().UTC().Sub(timeStamp) > 5*time.Second)
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, "1", status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.http",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
						"success": "${{$if($equal($output(test, status), 200), 1, 0)}}",
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
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, int64(0), status.Outputs["success"])
	// assert.Equal(t, "fakens", status.Outputs["__namespace"])
	// assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	// assert.Equal(t, "test-activation", status.Outputs["__activation"])
	// assert.Equal(t, int64(1), status.Outputs["__activationGeneration"])
	// assert.Equal(t, "test2", status.Outputs["__stage"])
	// assert.Equal(t, "fake", status.Outputs["__site"])
}

func TestErrorHandlerWithSelfDrivingDisabled(t *testing.T) {
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		Stage:                "test",
		ActivationGeneration: "1",
		Outputs:              nil,
		Provider:             "providers.stage.http",
		Namespace:            "fakens",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: false,
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
						"success": "${{$if($equal($output(test, status), 200), 1, 0)}}",
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
	assert.Equal(t, v1alpha2.InternalError, status.Status)
	assert.True(t, v1alpha2.InternalError.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
						"success": "${{$if($equal($output(test, status), 200), 1, 0)}}",
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
	assert.True(t, v1alpha2.InternalError.EqualsWithString(status.StatusMessage))
}

func TestErrorAtLastStage(t *testing.T) {
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.http",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider: "providers.stage.http",
					Inputs: map[string]interface{}{
						"method": "GET",
						"url":    "bad url",
					},
				},
			},
		}, *activation)

		if activation == nil {
			break
		}
	}
	assert.Equal(t, v1alpha2.InternalError, status.Status)
	assert.True(t, v1alpha2.InternalError.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.http",
					StageSelector: "${{$if($equal($output(test, status), 200), test2, '')}}",
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"status": 400,
						"error":  "bad",
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
		assert.Equal(t, v1alpha2.BadRequest, status.Outputs["status"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"status": v1alpha2.DeleteFailed,
						"error":  "failed",
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
		assert.Equal(t, v1alpha2.DeleteFailed, status.Outputs["status"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"status": "400",
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
		assert.Equal(t, v1alpha2.InternalError, status.Outputs["status"]) // non-successful state is returned without __error, set to InternalError
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	activation := &v1alpha2.ActivationData{
		Campaign:   "test-campaign",
		Activation: "test-activation",
		Stage:      "test",
		Outputs:    nil,
		Provider:   "providers.stage.mock",
	}
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
			SelfDriving: true,
			FirstStage:  "test",
			Stages: map[string]model.StageSpec{
				"test": {
					Provider:      "providers.stage.mock",
					StageSelector: "test2",
					Inputs: map[string]interface{}{
						"status": "400",
						"error":  "this_is_an_error",
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
		assert.Equal(t, v1alpha2.BadRequest, status.Outputs["status"]) // non-successful state is returned without __error, set to InternalError
		assert.Equal(t, "Bad Request: this_is_an_error", status.Outputs["error"])
	}
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
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
	var status model.StageStatus
	for {
		status, activation = manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
						"stcheck": "${{$output($input(__previousStage), status)}}",
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
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var err error
	manager.apiClient, err = utils.GetApiClient()
	assert.Nil(t, err)

	campaignName := "campaign1"
	activationName := "activation1"
	activationGen := "1"
	output := map[string]interface{}{
		"__campaign":             campaignName,
		"__activation":           activationName,
		"__activationGeneration": activationGen,
		"__site":                 "fake",
		"__stage":                "test",
		"__namespace":            "default",
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
	activation := model.StageStatus{
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
		Stage:         "test",
		Inputs: map[string]interface{}{
			"nextStage": "test2",
		},
		Outputs: output,
	}
	campaign := model.CampaignSpec{
		SelfDriving: true,
		FirstStage:  "test",
		Stages: map[string]model.StageSpec{
			"test": {
				Provider:      "providers.stage.mock",
				StageSelector: "${{$trigger(nextStage,'')}}",
				Inputs: map[string]interface{}{
					"ticket": "bar",
				},
			},
			"test2": {
				Provider:      "providers.stage.mock",
				StageSelector: "",
				HandleErrors:  false,
				Inputs: map[string]interface{}{
					"stcheck": "${{$output($input(__previousStage), status)}}",
					"stfoo":   "${{$output($input(__previousStage), ticket)}}",
				},
			},
		},
	}
	nextStage, err := manager.ResumeStage(context.TODO(), activation, campaign)
	assert.Nil(t, err)
	assert.Equal(t, "test2", nextStage.Stage)
	assert.Equal(t, campaignName, nextStage.Campaign)
	assert.Equal(t, activationName, nextStage.Activation)
}

func TestResumeStageFailed(t *testing.T) {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := StageManager{
		StateProvider: stateProvider,
	}
	ts := InitializeMockSymphonyAPI()
	os.Setenv(constants.SymphonyAPIUrlEnvName, ts.URL+"/")
	os.Setenv(constants.UseServiceAccountTokenEnvName, "false")
	manager.VendorContext = &contexts.VendorContext{
		EvaluationContext: &coa_utils.EvaluationContext{},
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
			CurrentSite: v1alpha2.SiteConnection{
				BaseUrl:  ts.URL + "/",
				Username: "admin",
				Password: "",
			},
		},
	}
	manager.Context = &contexts.ManagerContext{
		VencorContext: manager.VendorContext,
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "fake",
		},
	}
	var err error
	manager.apiClient, err = utils.GetApiClient()
	assert.Nil(t, err)

	campaignName := "campaign1"
	activationName := "activation2"
	activationGen := "1"
	output := map[string]interface{}{}
	stateProvider.Upsert(context.Background(), states.UpsertRequest{
		Value: states.StateEntry{
			ID: fmt.Sprintf("%s-%s-%s", campaignName, activationName, activationGen),
			Body: PendingTask{
				Sites:         []string{"fake"},
				OutputContext: map[string]map[string]interface{}{"fake": output},
			},
		},
	})
	activation := model.StageStatus{
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
		Stage:         "test",
		Outputs:       output,
	}
	campaign := model.CampaignSpec{
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
					"stcheck": "${{$output($input(__previousStage), status)}}",
					"stfoo":   "${{$output($input(__previousStage), ticket)}}",
				},
			},
		},
	}
	_, err = manager.ResumeStage(context.TODO(), activation, campaign)
	assert.NotNil(t, err)
	assert.Equal(t, "Bad Request: ResumeStage: campaign is not valid", err.Error())
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
		Outputs:   nil,
		Provider:  "providers.stage.counter",
		Namespace: "fakens",
	}
	status := manager.HandleDirectTriggerEvent(context.Background(), activation)
	assert.Equal(t, v1alpha2.Done, status.Status)
	assert.True(t, v1alpha2.Done.EqualsWithString(status.StatusMessage))
	assert.Equal(t, "test-campaign", status.Outputs["__campaign"])
	assert.Equal(t, "test-activation", status.Outputs["__activation"])
	assert.Equal(t, "1", status.Outputs["__activationGeneration"])
	assert.Equal(t, "fake", status.Outputs["__site"])
	assert.Equal(t, "test", status.Outputs["__stage"])
	assert.Equal(t, "fakens", status.Outputs["__namespace"])
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
		Schedule: "2020-01-01T12:00:00-08:00",
	}
	status := manager.HandleDirectTriggerEvent(context.Background(), activation)
	assert.Equal(t, v1alpha2.Paused, status.Status)
	assert.True(t, v1alpha2.Paused.EqualsWithString(status.StatusMessage))

	assert.Equal(t, v1alpha2.Paused, status.Outputs["status"])
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
					"stcheck": "${{$output($input(__previousStage), status)}}",
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
		Schedule: "2020-01-01T12:00:00-08:00",
	}

	status, _ := manager.HandleTriggerEvent(context.Background(), model.CampaignSpec{
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
	assert.True(t, v1alpha2.Paused.EqualsWithString(status.StatusMessage))
	assert.Equal(t, false, status.IsActive)
}

func prepareManager() *StageManager {
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
	return &manager
}

func TestParallelTaskProcessing(t *testing.T) {
	// Setup
	manager := prepareManager()

	ctx := context.Background()
	triggerData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Stage:                "test-stage",
		Namespace:            "default",
		Inputs: map[string]interface{}{
			"input1": "stageValue1",
		},
	}

	// Create test tasks
	tasks := []model.TaskSpec{
		{
			Name:     "task1",
			Provider: "providers.stage.mock",
			Config:   map[string]string{},
			Inputs: map[string]interface{}{
				"taskInput1": "value1",
				"input1":     "value1",
				"foo":        1,
			},
		},
		{
			Name:     "task2",
			Provider: "providers.stage.mock",
			Config:   map[string]string{},
			Inputs: map[string]interface{}{
				"taskInput2": "value2",
				"foo":        2,
			},
		},
	}

	// Create processor and handler
	processor := NewGoRoutineTaskProcessor(manager, ctx)
	handler := NewCampaignTaskHandler(manager, triggerData, nil)

	// Test parallel processing
	results, err := processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_StopOnAnyFailure,
	}, 2, "test-site")

	// Verify results
	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))

	// examine: taskInput
	assert.Equal(t, int64(2), results["task1"].(map[string]interface{})["foo"])
	assert.Equal(t, "value1", results["task1"].(map[string]interface{})["taskInput1"])
	assert.Equal(t, int64(3), results["task2"].(map[string]interface{})["foo"])
	assert.Equal(t, "value2", results["task2"].(map[string]interface{})["taskInput2"])

	// examine: stageInput
	assert.Equal(t, "value1", results["task1"].(map[string]interface{})["input1"]) // will be overridden by taskInput
	assert.Equal(t, "stageValue1", results["task2"].(map[string]interface{})["input1"])
}

func TestTaskErrorHandling(t *testing.T) {
	// Setup
	manager := prepareManager()
	ctx := context.Background()
	triggerData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Stage:                "test-stage",
		Namespace:            "default",
	}

	// Create test tasks with one failing task
	tasks := []model.TaskSpec{
		{
			Name:     "task1",
			Provider: "providers.stage.invalid",
			Config:   map[string]string{},
			Target:   "test-target1",
		},
		{
			Name:     "task2",
			Provider: "providers.stage.mock",
			Config:   map[string]string{},
			Target:   "test-target2",
		},
	}

	// Create processor and handler
	processor := NewGoRoutineTaskProcessor(manager, ctx)
	handler := NewCampaignTaskHandler(manager, triggerData, nil)

	// Test error handling with StopOnAnyFailure
	results, err := processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_StopOnAnyFailure,
	}, 2, "test-site")

	// Verify error handling
	assert.NotNil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	assert.Equal(t, "test-target2", results["task2"].(map[string]interface{})["__target"])

	// Test error handling with ErrorActionMode_StopOnNFailures, maxToleratedFailures = 0
	results, err = processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode:                 model.ErrorActionMode_StopOnNFailures,
		MaxToleratedFailures: 0,
	}, 2, "test-site")

	// Verify error handling
	assert.NotNil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	assert.Equal(t, "test-target2", results["task2"].(map[string]interface{})["__target"])

	// Test error handling with ErrorActionMode_StopOnNFailures, maxToleratedFailures = 1
	results, err = processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode:                 model.ErrorActionMode_StopOnNFailures,
		MaxToleratedFailures: 1,
	}, 2, "test-site")

	// Verify error handling
	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	assert.Equal(t, "test-target2", results["task2"].(map[string]interface{})["__target"])

	// Test error handling with ErrorActionMode_SilentlyContinue
	results, err = processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_SilentlyContinue,
	}, 2, "test-site")

	// Verify error handling
	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	assert.Equal(t, "test-target2", results["task2"].(map[string]interface{})["__target"])
}

func TestTaskErrorHandling_TaskNumberExceedsConcurrency(t *testing.T) {
	// Setup
	manager := prepareManager()
	ctx := context.Background()
	triggerData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Stage:                "test-stage",
		Namespace:            "default",
	}

	// Create test tasks with one failing task
	tasks := []model.TaskSpec{
		{
			Name:     "task1",
			Provider: "providers.stage.invalid",
			Config:   map[string]string{},
		},
		{
			Name:     "task2",
			Provider: "providers.stage.invalid",
			Config:   map[string]string{},
		},
		{
			Name:     "task3",
			Provider: "providers.stage.mock",
			Config:   map[string]string{},
		},
	}

	// Create processor and handler
	processor := NewGoRoutineTaskProcessor(manager, ctx)
	handler := NewCampaignTaskHandler(manager, triggerData, nil)

	// Test error handling with StopOnAnyFailure
	results, err := processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_StopOnAnyFailure,
	}, 1, "test-site")

	// Verify error handling
	assert.NotNil(t, err)
	assert.LessOrEqual(t, len(results), 2)
	if results["task1"] != nil {
		assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
		assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	}
	if results["task2"] != nil {
		assert.Equal(t, v1alpha2.BadConfig, results["task2"].(map[string]interface{})["status"])
		assert.NotNil(t, results["task2"].(map[string]interface{})["error"])
	}
	if results["task3"] != nil {
		assert.Nil(t, results["task3"].(map[string]interface{})["error"])
	}

	// Test error handling with StopOnNFailures
	results, err = processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode:                 model.ErrorActionMode_StopOnNFailures,
		MaxToleratedFailures: 1,
	}, 1, "test-site")

	// Verify error handling
	assert.NotNil(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	if results["task1"] != nil {
		assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
		assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	}
	if results["task2"] != nil {
		assert.Equal(t, v1alpha2.BadConfig, results["task2"].(map[string]interface{})["status"])
		assert.NotNil(t, results["task2"].(map[string]interface{})["error"])
	}
	if results["task3"] != nil {
		assert.Nil(t, results["task3"].(map[string]interface{})["error"])
	}

	// Test error handling with SilentlyContinue
	results, err = processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_SilentlyContinue,
	}, 1, "test-site")

	// Verify error handling
	assert.Nil(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, v1alpha2.BadConfig, results["task1"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task1"].(map[string]interface{})["error"])
	assert.Equal(t, v1alpha2.BadConfig, results["task2"].(map[string]interface{})["status"])
	assert.NotNil(t, results["task2"].(map[string]interface{})["error"])
	assert.Nil(t, results["task3"].(map[string]interface{})["error"])
}

func TestTaskConcurrencyControl(t *testing.T) {
	// Setup
	manager := prepareManager()
	ctx := context.Background()
	triggerData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Stage:                "test-stage",
		Namespace:            "default",
	}

	// Create multiple test tasks
	taskNumber := 5
	tasks := make([]model.TaskSpec, taskNumber)
	for i := 0; i < taskNumber; i++ {
		tasks[i] = model.TaskSpec{
			Name:     fmt.Sprintf("task%d", i+1),
			Provider: "providers.stage.mock",
			Config:   map[string]string{},
		}
	}

	// Create processor and handler
	processor := NewGoRoutineTaskProcessor(manager, ctx)
	handler := NewCampaignTaskHandler(manager, triggerData, nil)

	// Test with concurrency limit of 2
	results, err := processor.Process(ctx, tasks, triggerData.Inputs, handler, model.ErrorAction{
		Mode: model.ErrorActionMode_StopOnAnyFailure,
	}, 2, "test-site")

	assert.Nil(t, err)
	assert.Equal(t, taskNumber, len(results))
}

func TestTaskHandlerInterface(t *testing.T) {
	// Setup
	manager := prepareManager()
	ctx := context.Background()
	triggerData := v1alpha2.ActivationData{
		Campaign:             "test-campaign",
		Activation:           "test-activation",
		ActivationGeneration: "1",
		Stage:                "test-stage",
		Namespace:            "default",
	}

	// Create test task
	task := model.TaskSpec{
		Name:     "test-task",
		Provider: "providers.stage.mock",
		Config:   map[string]string{},
		Inputs: map[string]interface{}{
			"testInput": "testValue",
		},
	}

	// Create handler
	handler := NewCampaignTaskHandler(manager, triggerData, nil)

	// Test task handling
	outputs, err := handler.HandleTask(ctx, task, triggerData.Inputs, "test-site")

	// Verify results
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
}

func InitializeMockSymphonyAPI() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		log.Info("Mock Symphony API called", "path", r.URL.Path)
		switch r.URL.Path {
		case "/activations/registry/activation1":
			response = model.ActivationState{
				ObjectMeta: model.ObjectMeta{
					Name: "activation1",
				},
				Spec: &model.ActivationSpec{
					Campaign: "campaign1",
					Stage:    "test",
					Inputs: map[string]interface{}{
						"nextStage": "test2",
					},
				},
				Status: &model.ActivationStatus{
					Status:        v1alpha2.Done,
					StatusMessage: v1alpha2.Done.String(),
				},
			}
		case "/activations/registry/activation2":
			response = model.ActivationState{
				ObjectMeta: model.ObjectMeta{
					Name: "activation2",
				},
				Spec: &model.ActivationSpec{
					Campaign: "campaign1",
					Stage:    "test",
					Inputs:   map[string]interface{}{},
				},
				Status: &model.ActivationStatus{
					Status:        v1alpha2.Done,
					StatusMessage: v1alpha2.Done.String(),
				},
			}
		default:
			response = utils.AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	return ts
}
