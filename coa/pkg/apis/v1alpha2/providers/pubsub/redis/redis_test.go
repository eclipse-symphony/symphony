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

package redis

import (
	"os"
	"testing"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestWithEmptyConfig(t *testing.T) {
	provider := RedisPubSubProvider{}
	err := provider.Init(RedisPubSubProviderConfig{})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestWithMissingHost(t *testing.T) {
	provider := RedisPubSubProvider{}
	err := provider.Init(RedisPubSubProviderConfig{
		Name:     "test",
		Password: "abc",
	})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestWithZeroWorkerCount(t *testing.T) {
	provider := RedisPubSubProvider{}
	err := provider.Init(RedisPubSubProviderConfig{
		Name:            "test",
		Host:            "abc",
		Password:        "abc",
		NumberOfWorkers: 0,
	})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	// NumberOfWorkers should be set to 1, but initializtion should fail because of invalid host name
	assert.Equal(t, v1alpha2.InternalError, coaErr.State)
	assert.Equal(t, 1, provider.Config.NumberOfWorkers)
}

func TestInit(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	provider := RedisPubSubProvider{}
	err := provider.Init(RedisPubSubProviderConfig{
		Name:            "test",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
	})
	assert.Nil(t, err)
}

func TestBasicPubSub(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	sig := make(chan int)
	msg := ""
	provider := RedisPubSubProvider{}
	err := provider.Init(RedisPubSubProviderConfig{
		Name:            "test",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
	})
	assert.Nil(t, err)
	provider.Subscribe("test", func(topic string, message v1alpha2.Event) error {
		msg = message.Body.(string)
		sig <- 1
		return nil
	})
	provider.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig
	assert.Equal(t, "TEST", msg)
}

func TestBasicPubSubTwoProviders(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	sig := make(chan int)
	msg := ""
	provider1 := RedisPubSubProvider{}
	err := provider1.Init(RedisPubSubProviderConfig{
		Name:            "redis-1",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
	})
	assert.Nil(t, err)
	provider2 := RedisPubSubProvider{}
	err = provider2.Init(RedisPubSubProviderConfig{
		Name:            "redis-2",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
		ConsumerID:      "c",
	})
	assert.Nil(t, err)
	provider2.Subscribe("test", func(topic string, message v1alpha2.Event) error {
		msg = message.Body.(string)
		sig <- 1
		return nil
	})
	provider1.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig
	assert.Equal(t, "TEST", msg)
}
func TestMultipleSubscriber(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	sig1 := make(chan int)
	sig2 := make(chan int)
	msg1 := ""
	msg2 := ""
	provider1 := RedisPubSubProvider{}
	err := provider1.Init(RedisPubSubProviderConfig{
		Name:            "redis-1",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
	})
	assert.Nil(t, err)
	provider2 := RedisPubSubProvider{}
	err = provider2.Init(RedisPubSubProviderConfig{
		Name:            "redis-2",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
		ConsumerID:      "a",
	})
	assert.Nil(t, err)
	provider3 := RedisPubSubProvider{}
	err = provider3.Init(RedisPubSubProviderConfig{
		Name:            "redis-2",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
		ConsumerID:      "b",
	})
	assert.Nil(t, err)
	provider2.Subscribe("test", func(topic string, message v1alpha2.Event) error {
		msg1 = message.Body.(string)
		sig1 <- 1
		return nil
	})
	provider3.Subscribe("test", func(topic string, message v1alpha2.Event) error {
		msg2 = message.Body.(string)
		sig2 <- 1
		return nil
	})
	provider1.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig1
	<-sig2
	assert.Equal(t, "TEST", msg1)
	assert.Equal(t, "TEST", msg2)
}

func TestRedisPubSubProviderConfigFromMap(t *testing.T) {

	configMap := map[string]string{
		"name":              "test",
		"host":              "localhost:6379",
		"password":          "123",
		"requiresTLS":       "true",
		"numberOfWorkers":   "1",
		"queueDepth":        "10",
		"consumerID":        "test-consumer",
		"processingTimeout": "10",
		"redeliverInterval": "10",
	}
	config, err := RedisPubSubProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, "localhost:6379", config.Host)
	assert.Equal(t, "123", config.Password)
	assert.Equal(t, true, config.RequiresTLS)
	assert.Equal(t, 1, config.NumberOfWorkers)
	assert.Equal(t, 10, config.QueueDepth)
	assert.Equal(t, "test-consumer", config.ConsumerID)
	assert.Equal(t, time.Duration(10), config.ProcessingTimeout)
	assert.Equal(t, time.Duration(10), config.RedeliverInterval)
}
