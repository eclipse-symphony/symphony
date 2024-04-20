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
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/jobs"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createJobVendor() JobVendor {
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := jobs.JobsManager{
		StateProvider: stateProvider,
	}
	vendor := JobVendor{
		JobsManager: &manager,
	}
	return vendor
}
func TestJobsEndpoints(t *testing.T) {
	vendor := createJobVendor()
	vendor.Route = "instances"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}
func TestJobsInfo(t *testing.T) {
	vendor := createJobVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestJobsInit(t *testing.T) {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := JobVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "jobs-manager",
				Type: "managers.symphony.jobs",
				Properties: map[string]string{
					"providers.state": "mem-state",
					"baseUrl":         "http://localhost:8082/v1alpha2/",
					"user":            "admin",
					"password":        "",
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
		"jobs-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	assert.Nil(t, err)
}
func TestJobsonHello(t *testing.T) {
	vendor := createJobVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	succeededCount := 0
	sig := make(chan bool)
	vendor.Context.Subscribe("activation", func(topic string, event v1alpha2.Event) error {
		var activation v1alpha2.ActivationData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &activation)
		assert.Nil(t, err)
		assert.Equal(t, "activation1", activation.Activation)
		assert.Equal(t, "campaign1", activation.Campaign)
		succeededCount += 1
		sig <- true
		return nil
	})
	activation := v1alpha2.ActivationData{
		Activation: "activation1",
		Campaign:   "campaign1",
	}
	data, _ := json.Marshal(activation)
	resp := vendor.onHello(v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	})
	<-sig
	assert.Equal(t, v1alpha2.OK, resp.State)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, succeededCount)
}
func TestJobWrongMethod(t *testing.T) {
	vendor := createJobVendor()
	resp := vendor.onHello(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
