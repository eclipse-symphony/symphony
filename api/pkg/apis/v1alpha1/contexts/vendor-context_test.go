/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"sync"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/stretchr/testify/assert"
)

func createVendorContextWithPubSub() *VendorContext {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{
		SubscriberRetryCount:      5,
		SubscriberRetryWaitSecond: 1,
	})
	v := &VendorContext{}
	_ = v.Init(&pubsubProvider)
	return v
}

func createVendorContextWithoutPubSub() *VendorContext {
	v := &VendorContext{}
	_ = v.Init(nil)
	return v
}

func TestVendorContextInitWithoutPubSub(t *testing.T) {
	v := &VendorContext{}
	err := v.Init(nil)
	assert.Nil(t, err)
}

func TestVendorContextInitWithPubSub(t *testing.T) {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	v := &VendorContext{}
	err := v.Init(&pubsubProvider)
	assert.Nil(t, err)
}

func TestVendorContextPublishSubscribe(t *testing.T) {
	v := createVendorContextWithPubSub()
	sig := make(chan int)
	msg := ""

	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		msg = event.Body.(string)
		sig <- 1
		return nil
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)

	<-sig
	assert.Equal(t, "test", msg)
}

func TestVendorContextPublishSubscribeWithoutPubSub(t *testing.T) {
	v := createVendorContextWithoutPubSub()
	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		return nil
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)
}

func TestVendorContextPublishSubscribeBadRequest(t *testing.T) {
	v := createVendorContextWithPubSub()
	var mu sync.Mutex
	count := 0
	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		mu.Lock()
		defer mu.Unlock()
		count += 1
		return v1alpha2.NewCOAError(nil, "insert bad request", v1alpha2.BadRequest)
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)

	time.Sleep(time.Duration(5) * time.Second)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, count)
}

func TestVendorContextPublishSubscribeInternalError(t *testing.T) {
	v := createVendorContextWithPubSub()
	var wg sync.WaitGroup
	wg.Add(5)
	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		wg.Done()
		return v1alpha2.NewCOAError(nil, "insert bad request", v1alpha2.InternalError)
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)

	wg.Wait()
}
