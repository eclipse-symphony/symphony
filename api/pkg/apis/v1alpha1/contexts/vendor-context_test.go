/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/stretchr/testify/assert"
)

func createVendorContextWithPubSub() *VendorContext {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{
		SubscriberRetryCount:      4,
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
	ch := make(chan struct{})
	count := 0

	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		count += 1
		ch <- struct{}{}
		return v1alpha2.NewCOAError(nil, "insert bad request", v1alpha2.BadRequest)
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)

	signal := 0
	for signal < 1 {
		select {
		case <-ch:
			close(ch)
		case <-time.After(5 * time.Second):
			// Timeout, function was not called
			t.Fatal("Function was not called within the timeout period")
		}
		signal += 1
	}
	time.Sleep(5 * time.Second) // Wait to ensure no further calls are made
	assert.Equal(t, 1, count)
}

func TestVendorContextPublishSubscribeInternalError(t *testing.T) {
	v := createVendorContextWithPubSub()
	ch := make(chan struct{})
	count := 0

	err := v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		count += 1
		ch <- struct{}{}
		return v1alpha2.NewCOAError(nil, "insert internal error", v1alpha2.InternalError)
	})
	assert.Nil(t, err)
	err = v.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)

	signal := 0
	for signal < 5 {
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			// Timeout, function was not called
			t.Fatal("Function was not called within the timeout period")
		}
		signal += 1
	}
	close(ch)
	time.Sleep(2 * time.Second) // Wait to ensure no further calls are made
	assert.Equal(t, 5, count)
}
