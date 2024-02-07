/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/stretchr/testify/assert"
)

type TestPubSubProvider struct {
	Subscribers map[string][]v1alpha2.EventHandler `json:"subscribers"`
}

func (t *TestPubSubProvider) Init(config providers.IProviderConfig) error {
	t.Subscribers = make(map[string][]v1alpha2.EventHandler)
	return nil
}

func (t *TestPubSubProvider) Publish(topic string, event v1alpha2.Event) error {
	arr, ok := t.Subscribers[topic]
	if ok && arr != nil {
		for _, s := range arr {
			go func(handler v1alpha2.EventHandler, topic string, event v1alpha2.Event) {
				handler(topic, event)
			}(s, topic, event)
		}
	}
	return nil
}

func (t *TestPubSubProvider) Subscribe(topic string, handler v1alpha2.EventHandler) error {
	arr, ok := t.Subscribers[topic]
	if !ok || arr == nil {
		t.Subscribers[topic] = make([]v1alpha2.EventHandler, 0)
	}
	t.Subscribers[topic] = append(t.Subscribers[topic], handler)

	return nil
}

func TestVendorContextInit(t *testing.T) {
	v := VendorContext{}
	err := v.Init(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v.Logger)
	assert.Nil(t, v.PubsubProvider)
}

func TestVendorContextWithPubSub(t *testing.T) {
	v := VendorContext{}
	pubSub := &TestPubSubProvider{}
	err := pubSub.Init(nil)
	assert.Nil(t, err)
	err = v.Init(pubSub)
	assert.Nil(t, err)
	assert.NotNil(t, v.Logger)
	assert.NotNil(t, v.PubsubProvider)
}

func TestVendorContextPublishSubscribe(t *testing.T) {
	v := VendorContext{}
	pubSub := &TestPubSubProvider{}
	err := pubSub.Init(nil)
	assert.Nil(t, err)
	err = v.Init(pubSub)
	assert.Nil(t, err)
	assert.NotNil(t, v.Logger)
	assert.NotNil(t, v.PubsubProvider)

	called := false
	sig := make(chan bool)
	v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		called = true
		sig <- true
		return nil
	})

	v.Publish("test", v1alpha2.Event{})
	<-sig
	assert.True(t, called)
}

func TestVendorContextPublishSubscribeWithoutPubSub(t *testing.T) {
	v := VendorContext{}
	err := v.Init(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v.Logger)
	assert.Nil(t, v.PubsubProvider)

	called := false
	v.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		called = true
		return nil
	})

	v.Publish("test", v1alpha2.Event{})
	assert.False(t, called)
}
