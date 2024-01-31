/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createEchoVendor(route string) *EchoVendor {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	vendor := EchoVendor{}
	_ = vendor.Init(vendors.VendorConfig{
		Route: route,
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{}, &pubsubProvider)

	return &vendor
}

func TestEchoVendorInit(t *testing.T) {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	vendor := EchoVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{}, &pubsubProvider)
	assert.Nil(t, err)
}

func TestEchoVendorInitWithTraceMessage(t *testing.T) {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	vendor := EchoVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{}, &pubsubProvider)
	assert.Nil(t, err)

	messageBufferCnt := 20
	// runtime.Gosched() is not very reliable. We cannot control the order of execution of goroutines.
	// Basic idea is to yield once and let subscribe execute at first, then validate the received messages.
	// However, there is no guarantee that subscribe will execute before next publish, add sleep to protect
	for i := 0; i < (messageBufferCnt + 1); i++ {
		vendor.Context.Publish("trace", v1alpha2.Event{
			Body: fmt.Sprintf("hello world %d", i),
		})
		runtime.Gosched() // yield to other goroutines: subscribe and process
		time.Sleep(1 * time.Millisecond)
		if i < messageBufferCnt {
			assert.Equal(t, i+1, len(vendor.GetMessages()))
			assert.Equal(t, fmt.Sprintf("hello world %d", i), vendor.GetMessages()[i])
		} else {
			assert.Equal(t, messageBufferCnt, len(vendor.GetMessages()))
			assert.Equal(t, fmt.Sprintf("hello world %d", i-messageBufferCnt+1), vendor.GetMessages()[0])
			assert.Equal(t, fmt.Sprintf("hello world %d", i), vendor.GetMessages()[messageBufferCnt-1])
		}
	}
}

func TestEchoVendorGetInfo(t *testing.T) {
	vendor := createEchoVendor("")
	info := vendor.GetInfo()
	assert.Equal(t, "Echo", info.Name)
	assert.Equal(t, "Microsoft", info.Producer)
}

func TestEchoVendorEndpoints(t *testing.T) {
	vendor := createEchoVendor("")
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))

	vendor = createEchoVendor("greetings")
	endpoints = vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, "greetings", endpoints[0].Route)
}

func TestEchoVendorOnHello_Get(t *testing.T) {
	vendor := createEchoVendor("")

	additionalMsg := "hello world"
	vendor.Context.Publish("trace", v1alpha2.Event{
		Body: additionalMsg,
	})
	runtime.Gosched() // yield to other goroutines: subscribe and process
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, len(vendor.GetMessages()))
	assert.Equal(t, "hello world", vendor.GetMessages()[0])

	req := v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	response := vendor.onHello(req)
	assert.Equal(t, v1alpha2.OK, response.State)
	assert.Equal(t, "Hello from Symphony K8s control plane (S8C)"+"\r\n"+additionalMsg, string(response.Body))
}

func TestEchoVendorOnHello_Post(t *testing.T) {
	vendor := createEchoVendor("")

	additionalMsg := "post body"
	req := v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Context: context.Background(),
		Body:    []byte(additionalMsg),
	}
	response := vendor.onHello(req)
	assert.Equal(t, v1alpha2.OK, response.State)

	runtime.Gosched() // yield to other goroutines: subscribe and process
	time.Sleep(1 * time.Second)

	req = v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	response = vendor.onHello(req)
	assert.Equal(t, v1alpha2.OK, response.State)
	assert.Equal(t, "Hello from Symphony K8s control plane (S8C)"+"\r\n"+additionalMsg, string(response.Body))
}

func TestEchoVendorOnHello_InvalidMethod(t *testing.T) {
	vendor := createEchoVendor("")

	req := v1alpha2.COARequest{
		Method:  fasthttp.MethodHead,
		Context: context.Background(),
	}
	response := vendor.onHello(req)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
	assert.Equal(t, "{\"result\":\"405 - method not allowed\"}", string(response.Body))
}
