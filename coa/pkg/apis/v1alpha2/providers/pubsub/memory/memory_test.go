/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/

package memory

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestBasicPubSub(t *testing.T) {
	sig := make(chan int)
	msg := ""
	provider := InMemoryPubSubProvider{}
	provider.Init(InMemoryPubSubConfig{Name: "test"})
	provider.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		msg = event.Body.(string)
		sig <- 1
		return nil
	})
	provider.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig
	assert.Equal(t, "TEST", msg)
}

func TestMultipleSubscriber(t *testing.T) {
	sig1 := make(chan int)
	sig2 := make(chan int)
	msg1 := ""
	msg2 := ""
	provider := InMemoryPubSubProvider{}
	provider.Init(InMemoryPubSubConfig{Name: "test"})
	provider.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		msg1 = event.Body.(string)
		sig1 <- 1
		return nil
	})
	provider.Subscribe("test", func(topic string, event v1alpha2.Event) error {
		msg2 = event.Body.(string)
		sig2 <- 1
		return nil
	})
	provider.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig1
	<-sig2
	assert.Equal(t, "TEST", msg1)
	assert.Equal(t, "TEST", msg2)
}

func TestMultipleTopics(t *testing.T) {
	sig1 := make(chan int)
	sig2 := make(chan int)
	msg1 := ""
	msg2 := ""
	provider := InMemoryPubSubProvider{}
	provider.Init(InMemoryPubSubConfig{Name: "test"})
	provider.Subscribe("test1", func(topic string, event v1alpha2.Event) error {
		msg1 = event.Body.(string)
		sig1 <- 1
		return nil
	})
	provider.Subscribe("test2", func(topic string, event v1alpha2.Event) error {
		msg2 = event.Body.(string)
		sig2 <- 1
		return nil
	})
	provider.Publish("test1", v1alpha2.Event{Body: "TEST1"})
	provider.Publish("test2", v1alpha2.Event{Body: "TEST2"})
	<-sig1
	<-sig2
	assert.Equal(t, "TEST1", msg1)
	assert.Equal(t, "TEST2", msg2)
}
func TestMemoryPubsubProviderConfigFromMapNil(t *testing.T) {
	_, err := InMemoryPubSubConfigFromMap(nil)
	assert.Nil(t, err)
}

func TestMemoryPubsubProviderConfigFromMapEmpty(t *testing.T) {
	_, err := InMemoryPubSubConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestMemoryPubsubProviderConfigFromMap(t *testing.T) {
	config, err := InMemoryPubSubConfigFromMap(map[string]string{
		"name": "my-name",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-name", config.Name)
}
