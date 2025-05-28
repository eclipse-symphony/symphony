/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package redis

import (
	"context"
	"encoding/json"
	"errors"

	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/host"
	"github.com/redis/go-redis/v9"
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
	assert.Equal(t, 20, provider.Config.NumberOfWorkers)
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

func TestInitWithMap(t *testing.T) {
	provider := RedisPubSubProvider{}
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis != "" {
		err := provider.InitWithMap(
			map[string]string{
				"name": "test",
				"host": "localhost:6379",
			},
		)
		assert.Nil(t, err) // Provider initialization succeeds if redis is running
	}

	err := provider.InitWithMap(
		map[string]string{
			"name": "test",
		},
	)
	assert.NotNil(t, err)
	err = provider.InitWithMap(
		map[string]string{
			"name":        "test",
			"host":        "localhost:6379",
			"requiresTLS": "abcd",
		},
	)
	assert.NotNil(t, err)
	err = provider.InitWithMap(
		map[string]string{
			"name":            "test",
			"host":            "localhost:6379",
			"numberOfWorkers": "abcd",
		},
	)
	assert.NotNil(t, err)
	err = provider.InitWithMap(
		map[string]string{
			"name": "test",
			"host": "localhost:6379",
		},
	)
	assert.NotNil(t, err)
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
	provider.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, message v1alpha2.Event) error {
			msg = message.Body.(string)
			sig <- 1
			return nil
		},
	})
	host.SetHostReadyFlag(true)
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
	provider2.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, message v1alpha2.Event) error {
			msg = message.Body.(string)
			sig <- 1
			return nil
		},
	})
	host.SetHostReadyFlag(true)
	provider1.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig
	assert.Equal(t, "TEST", msg)
}
func TestBasicPubSubTwoProvidersComplexEvent(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	sig := make(chan int)
	var msg v1alpha2.JobData
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
	provider2.Subscribe("testjob", v1alpha2.EventHandler{
		Handler: func(topic string, message v1alpha2.Event) error {
			jData, _ := json.Marshal(message.Body)
			json.Unmarshal(jData, &msg)
			sig <- 1
			return nil
		},
	})
	host.SetHostReadyFlag(true)
	provider1.Publish("testjob", v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": "mock",
		},
		Body: v1alpha2.JobData{
			Id:     "123",
			Action: "do-it",
		},
	})
	<-sig
	assert.Equal(t, v1alpha2.JobAction("do-it"), msg.Action)
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
	provider2.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, message v1alpha2.Event) error {
			msg1 = message.Body.(string)
			sig1 <- 1
			return nil
		},
		Group: "test0",
	})
	provider3.Subscribe("test", v1alpha2.EventHandler{
		Handler: func(topic string, message v1alpha2.Event) error {
			msg2 = message.Body.(string)
			sig2 <- 1
			return nil
		},
		Group: "test1",
	})
	provider1.Publish("test", v1alpha2.Event{Body: "TEST"})
	<-sig1
	<-sig2
	assert.Equal(t, "TEST", msg1)
	assert.Equal(t, "TEST", msg2)
}

func TestSubscribePublish(t *testing.T) {
	provider := RedisPubSubProvider{}
	provider.Init(RedisPubSubProviderConfig{
		Name:            "test",
		Host:            "localhost:6379",
		Password:        "",
		NumberOfWorkers: 1,
	})
	// assert.Nil(t, err) // Provider initialization succeeds if redis is running

	// var msg string
	// sig := make(chan int)
	// provider.Subscribe("test", func(topic string, message v1alpha2.Event) error {
	// 	msg = message.Body.(string)
	// 	sig <- 1
	// 	return nil
	// })
	// provider.Publish("test", v1alpha2.Event{Body: "TEST"})
	// <-sig
	// assert.Equal(t, "TEST", msg)
}

func TestRedisPubSubProviderConfigFromMap(t *testing.T) {
	configMap := map[string]string{
		"name":            "test",
		"host":            "localhost:6379",
		"password":        "123",
		"requiresTLS":     "true",
		"numberOfWorkers": "1",
		"consumerID":      "test-consumer",
	}
	config, err := RedisPubSubProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, "localhost:6379", config.Host)
	assert.Equal(t, "123", config.Password)
	assert.Equal(t, true, config.RequiresTLS)
	assert.Equal(t, 1, config.NumberOfWorkers)
	assert.Contains(t, config.ConsumerID, "test-consumer")
}

// This test mostly test the behavior of redis API rather than pubsub
func TestRedisStreamBasic(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS enviornment variable is not set")
	}
	Ctx, _ := context.WithCancel(context.Background())
	options := &redis.Options{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		MaxRetryBackoff: time.Second * 2,
	}
	client := redis.NewClient(options)
	consumerId := "testconsumer"
	topic := "test"
	err := client.XGroupCreateMkStream(Ctx, topic, consumerId, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		mLog.Debugf("  P (Redis PubSub) : failed to subscribe %v", err)
		assert.Nil(t, err)
	}
	_, err = client.XAdd(Ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{"data": v1alpha2.Event{Body: "TEST"}},
	}).Result()
	assert.Nil(t, err)

	streams, err := client.XReadGroup(Ctx, &redis.XReadGroupArgs{
		Group:    consumerId,
		Consumer: consumerId,
		Streams:  []string{topic, ">"},
		Count:    int64(10),
		Block:    0,
	}).Result()
	assert.NotNil(t, streams)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)
	pendingResult, err := client.XPendingExt(Ctx, &redis.XPendingExtArgs{
		Stream:   topic,
		Group:    consumerId,
		Start:    "-",
		End:      "+",
		Count:    int64(10),
		Consumer: "",
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		mLog.Debugf("  P (Redis PubSub) : failed to get pending message %v", err)
		assert.Nil(t, err)
	}
	msgIDs := make([]string, 0, len(pendingResult))
	for _, msg := range pendingResult {
		msgIDs = append(msgIDs, msg.ID)
	}
	assert.NotNil(t, msgIDs)
	claimResult, err := client.XClaim(Ctx, &redis.XClaimArgs{
		Stream:   topic,
		Group:    consumerId,
		Consumer: consumerId,
		MinIdle:  time.Duration(1 * time.Second),
		Messages: msgIDs,
	}).Result()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(claimResult))

	pendingResult, err = client.XPendingExt(Ctx, &redis.XPendingExtArgs{
		Stream: topic,
		Group:  consumerId,
		Start:  "-",
		End:    "+",
		Count:  int64(10),
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		mLog.Debugf("  P (Redis PubSub) : failed to get pending message %v", err)
		assert.Nil(t, err)
	}
	msgIDs = make([]string, 0, len(pendingResult))
	for _, msg := range pendingResult {
		msgIDs = append(msgIDs, msg.ID)
	}
	assert.NotNil(t, msgIDs)
	claimResult, err = client.XClaim(Ctx, &redis.XClaimArgs{
		Stream:   topic,
		Group:    consumerId,
		Consumer: consumerId,
		MinIdle:  time.Duration(100 * time.Second),
		Messages: msgIDs,
	}).Result()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(claimResult))
}
