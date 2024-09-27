/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/stretchr/testify/assert"
)

func TestManagerContextInitWithoutPubSub(t *testing.T) {
	v := createVendorContextWithPubSub()
	v.SiteInfo.SiteId = "test"
	m := &ManagerContext{}
	err := m.Init(v, nil)
	assert.Nil(t, err)
	assert.Equal(t, "test", m.SiteInfo.SiteId)
	assert.NotNil(t, m.Logger)
	assert.NotNil(t, m.PubsubProvider)
}

func TestManagerContextInitWithVendorConext(t *testing.T) {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	m := &ManagerContext{}
	err := m.Init(nil, &pubsubProvider)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.NotNil(t, m.PubsubProvider)
}

func TestManagerContextInitWithVendorConextAndPubSub(t *testing.T) {
	m := &ManagerContext{}
	err := m.Init(nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.Nil(t, m.PubsubProvider)
}

func TestManagerContextPublishSubscribe(t *testing.T) {
	v := createVendorContextWithPubSub()
	v.SiteInfo.SiteId = "test"
	m := &ManagerContext{}
	err := m.Init(v, nil)
	assert.Nil(t, err)
	sig := make(chan int)
	msg := ""

	err = m.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			msg = event.Body.(string)
			sig <- 1
			return nil
		},
	})
	assert.Nil(t, err)
	err = m.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)
	<-sig
	assert.Equal(t, "test", msg)
}

func TestManagerContextPublishSubscribeWithoutPubSub(t *testing.T) {
	m := &ManagerContext{}
	err := m.Init(nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.Nil(t, m.PubsubProvider)

	err = m.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, event v1alpha2.Event) error {
			return nil
		},
	})
	assert.Nil(t, err)
	err = m.Publish("test", v1alpha2.Event{
		Body: "test",
	})
	assert.Nil(t, err)
}
