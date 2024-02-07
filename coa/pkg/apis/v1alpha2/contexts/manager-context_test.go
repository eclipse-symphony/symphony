/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package contexts

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInitWithoutVendorContextAndPubSub(t *testing.T) {
	m := &ManagerContext{}
	err := m.Init(nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.Nil(t, m.PubsubProvider)
}

func TestInitWithPubSub(t *testing.T) {
	m := &ManagerContext{}
	pubsub := &TestPubSubProvider{}
	err := pubsub.Init(nil)
	assert.Nil(t, err)

	err = m.Init(nil, pubsub)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.NotNil(t, m.PubsubProvider)
}

func TestInitWithVendorConext(t *testing.T) {
	m := &ManagerContext{}
	v := &VendorContext{
		SiteInfo: v1alpha2.SiteInfo{
			SiteId: "test",
		},
	}

	pubsub := &TestPubSubProvider{}
	err := pubsub.Init(nil)
	assert.Nil(t, err)

	err = v.Init(pubsub)
	assert.Nil(t, err)

	err = m.Init(v, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.NotNil(t, m.PubsubProvider)
	assert.Equal(t, v.PubsubProvider, m.PubsubProvider)
	assert.Equal(t, v.SiteInfo.SiteId, m.SiteInfo.SiteId)
}

func TestManagerContextPublishSubscribe(t *testing.T) {
	m := ManagerContext{}
	pubSub := &TestPubSubProvider{}
	err := pubSub.Init(nil)
	assert.Nil(t, err)
	err = m.Init(nil, pubSub)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.NotNil(t, m.PubsubProvider)

	called := false
	sig := make(chan bool)
	m.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		called = true
		sig <- true
		return nil
	})

	m.Publish("test", v1alpha2.Event{})
	<-sig
	assert.True(t, called)
}

func TestManagerContextPublishSubscribeWithoutPubSub(t *testing.T) {
	m := ManagerContext{}
	err := m.Init(nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m.Logger)
	assert.Nil(t, m.PubsubProvider)

	called := false
	m.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		called = true
		return nil
	})

	m.Publish("test", v1alpha2.Event{})
	assert.False(t, called)
}
