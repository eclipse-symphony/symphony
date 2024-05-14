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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

func createInstancesVendor() InstancesVendor {
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	vendor := InstancesVendor{}
	vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "instances-manager",
				Type: "managers.symphony.instances",
				Properties: map[string]string{
					"providers.state": "mem-state",
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
		"instances-manager": {
			"mem-state": &stateProvider,
		},
	}, nil)
	return vendor
}
func TestInstancesEndpoints(t *testing.T) {
	vendor := createInstancesVendor()
	vendor.Route = "instances"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 2, len(endpoints))
}
func TestInstancesInfo(t *testing.T) {
	vendor := createInstancesVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
func TestInstancesOnInstances(t *testing.T) {
	vendor := createInstancesVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)
	vendor.Config.Properties = map[string]string{
		"useJobManager": "true",
	}
	succeededCount := 0
	sig := make(chan bool)
	vendor.Context.Subscribe("job", func(topic string, event v1alpha2.Event) error {
		var job v1alpha2.JobData
		jData, _ := json.Marshal(event.Body)
		err := json.Unmarshal(jData, &job)
		assert.Nil(t, err)
		assert.Equal(t, "instance", event.Metadata["objectType"])
		assert.Equal(t, "instance1-v1", job.Id)
		assert.Equal(t, true, job.Action == v1alpha2.JobUpdate || job.Action == v1alpha2.JobDelete)
		succeededCount += 1
		sig <- true
		return nil
	})
	instanceSpec := model.InstanceSpec{}
	data, _ := json.Marshal(instanceSpec)
	resp := vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":    "instance1",
			"__version": "v1",
			"target":    "target1",
			"solution":  "solution1",
		},
		Context: context.Background(),
	})
	<-sig
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "instance1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	var instance model.InstanceState
	err := json.Unmarshal(resp.Body, &instance)
	assert.Nil(t, err)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, "instance1-v1", instance.ObjectMeta.Name)
	assert.Equal(t, "target1", instance.Spec.Target.Name)

	resp = vendor.onInstancesList(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	var instances []model.InstanceState
	err = json.Unmarshal(resp.Body, &instances)
	assert.Nil(t, err)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "instance1-v1", instances[0].ObjectMeta.Name)
	assert.Equal(t, "target1", instances[0].Spec.Target.Name)

	resp = vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "instance1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	<-sig
	assert.Equal(t, v1alpha2.OK, resp.State)
	time.Sleep(time.Second)
	assert.Equal(t, 2, succeededCount)

	vendor.Config.Properties = map[string]string{
		"useJobManager": "false",
	}
	resp = vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodDelete,
		Parameters: map[string]string{
			"__name":    "instance1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onInstancesList(v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	})
	err = json.Unmarshal(resp.Body, &instances)
	assert.Nil(t, err)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, 0, len(instances))
}

func TestInstancesTargetSelector(t *testing.T) {
	vendor := createInstancesVendor()
	vendor.Context = &contexts.VendorContext{}
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor.Context.Init(&pubSubProvider)

	instanceSpec := model.InstanceSpec{}
	data, _ := json.Marshal(instanceSpec)
	resp := vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodPost,
		Body:   data,
		Parameters: map[string]string{
			"__name":          "instance1",
			"__version":       "v1",
			"target-selector": "property1=value1",
			"solution":        "solution1",
		},
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.OK, resp.State)

	resp = vendor.onInstances(v1alpha2.COARequest{
		Method: fasthttp.MethodGet,
		Parameters: map[string]string{
			"__name":    "instance1",
			"__version": "v1",
		},
		Context: context.Background(),
	})
	var instance model.InstanceState
	err := json.Unmarshal(resp.Body, &instance)
	assert.Nil(t, err)
	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, "instance1-v1", instance.ObjectMeta.Name)
	assert.Equal(t, "value1", instance.Spec.Target.Selector["property1"])
}

func TestInstancesWrongMethod(t *testing.T) {
	vendor := createInstancesVendor()
	resp := vendor.onInstances(v1alpha2.COARequest{
		Method:  fasthttp.MethodPut,
		Context: context.Background(),
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}
