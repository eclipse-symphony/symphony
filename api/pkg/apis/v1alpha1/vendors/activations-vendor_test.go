/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/activations"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createActivationsVendor() ActivationsVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := activations.ActivationsManager{
		StateProvider: stateProvider,
	}
	vendor := ActivationsVendor{
		ActivationsManager: &manager,
	}
	return vendor
}
func TestActivationsEndpoints(t *testing.T) {
	vendor := createActivationsVendor()
	vendor.Route = "activations"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 2, len(endpoints))
}
func TestActivationsInfo(t *testing.T) {
	vendor := createActivationsVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestActivationsOnStatus(t *testing.T) {
	vendor := createActivationsVendor()
	status := model.ActivationStatus{
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
	}
	data, _ := json.Marshal(status)
	resp := vendor.onStatus(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "Not Found: entry 'activation1' is not found in namespace default", string(resp.Body))

	resp = vendor.onStatus(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   []byte{},
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	assert.Equal(t, "unexpected end of JSON input", string(resp.Body))
}
func TestActivationsOnActivations(t *testing.T) {
	vendor := createActivationsVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	activationName := "activation1"
	campaignRefName := "campaign1:v1"
	succeededCount := 0
	sigs := make(chan bool)
	vendor.Context.Subscribe("activation", func(topic string, event v1alpha2.Event) error {
		var activation v1alpha2.ActivationData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &activation)
		assert.Nil(t, err)
		assert.Equal(t, campaignRefName, activation.Campaign)
		assert.Equal(t, activationName, activation.Activation)
		succeededCount += 1
		sigs <- true
		return nil
	})
	resp := vendor.onActivations(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.InternalError, resp.State)
	activationState := model.ActivationState{
		Spec: &model.ActivationSpec{
			Campaign: campaignRefName,
		},
		ObjectMeta: model.ObjectMeta{
			Name: activationName,
		},
	}
	data, _ := json.Marshal(activationState)
	resp = vendor.onActivations(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	<-sigs
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, 1, succeededCount)

	resp = vendor.onActivations(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	var activation model.ActivationState
	assert.Equal(t, v1alpha2.OK, resp.State)
	err := json.Unmarshal(resp.Body, &activation)
	assert.Nil(t, err)
	assert.Equal(t, activationName, activation.ObjectMeta.Name)
	assert.Equal(t, campaignRefName, activation.Spec.Campaign)

	resp = vendor.onActivations(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	var activations []model.ActivationState
	err = json.Unmarshal(resp.Body, &activations)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(activations))
	assert.Equal(t, activationName, activations[0].ObjectMeta.Name)
	assert.Equal(t, campaignRefName, activations[0].Spec.Campaign)

	status := model.ActivationStatus{
		Status:        v1alpha2.Done,
		StatusMessage: v1alpha2.Done.String(),
	}
	data, _ = json.Marshal(status)
	resp = vendor.onStatus(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onActivations(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}
func TestActivationsWrongMethod(t *testing.T) {
	vendor := createActivationsVendor()
	resp := vendor.onActivations(v1alpha2.COARequest{
		Method: fasthttp.MethodPut,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)

	resp = vendor.onStatus(v1alpha2.COARequest{
		Method: fasthttp.MethodPut,
		Parameters: map[string]string{
			"__name": "activation1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
