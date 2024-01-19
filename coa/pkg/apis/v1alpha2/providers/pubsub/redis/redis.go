/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/go-redis/redis/v7"
)

var mLog = logger.NewLogger("coa.runtime")

type RedisPubSubProvider struct {
	Config      RedisPubSubProviderConfig          `json:"config"`
	Subscribers map[string][]v1alpha2.EventHandler `json:"subscribers"`
	Client      *redis.Client
	Queue       chan RedisMessageWrapper
	Ctx         context.Context
	Cancel      context.CancelFunc
	Context     *contexts.ManagerContext
}

type RedisMessageWrapper struct {
	MessageID string
	Topic     string
	Message   interface{}
	Handler   v1alpha2.EventHandler
}

type RedisPubSubProviderConfig struct {
	Name              string        `json:"name"`
	Host              string        `json:"host"`
	Password          string        `json:"password,omitempty"`
	RequiresTLS       bool          `json:"requiresTLS,omitempty"`
	NumberOfWorkers   int           `json:"numberOfWorkers,omitempty"`
	QueueDepth        int           `json:"queueDepth,omitempty"`
	ConsumerID        string        `json:"consumerID"`
	ProcessingTimeout time.Duration `json:"processingTimeout,omitempty"`
	RedeliverInterval time.Duration `json:"redeliverInterval,omitempty"`
}

func RedisPubSubProviderConfigFromMap(properties map[string]string) (RedisPubSubProviderConfig, error) {
	ret := RedisPubSubProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = v // providers.LoadEnv(v)
	}
	if v, ok := properties["host"]; ok {
		ret.Host = v //providers.LoadEnv(v)
	} else {
		return ret, v1alpha2.NewCOAError(nil, "Redis pub-sub provider host name is not set", v1alpha2.BadConfig)
	}
	if v, ok := properties["password"]; ok {
		ret.Password = v //providers.LoadEnv(v)
	}
	if v, ok := properties["requiresTLS"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'requiresTLS' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.RequiresTLS = bVal
		}
	}
	if v, ok := properties["numberOfWorkers"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'numberOfWorkers' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.NumberOfWorkers = n
		} else {
			ret.NumberOfWorkers = 1
		}
	}
	if v, ok := properties["queueDepth"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'queueDepth' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.QueueDepth = n
		}
	}
	if v, ok := properties["consumerID"]; ok {
		ret.ConsumerID = v // providers.LoadEnv(v)
	}

	if v, ok := properties["processingTimeout"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			n, err := utils.UnmarshalDuration(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'processingTimeout' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.ProcessingTimeout = n
		}
	}

	if v, ok := properties["redeliverInterval"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			n, err := utils.UnmarshalDuration(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'redeliverInterval' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.RedeliverInterval = n
		}
	}

	if ret.NumberOfWorkers <= 0 {
		ret.NumberOfWorkers = 1
	}
	//TODO: Finish this
	return ret, nil
}

func (v *RedisPubSubProvider) ID() string {
	return v.Config.Name
}

func (s *RedisPubSubProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *RedisPubSubProvider) InitWithMap(properties map[string]string) error {
	config, err := RedisPubSubProviderConfigFromMap(properties)
	if err != nil {
		mLog.Debugf("  P (Redis PubSub) : failed to initialize provider %v", err)
		return err
	}
	return i.Init(config)
}

func (i *RedisPubSubProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toRedisPubSubProviderConfig(config)
	if err != nil {
		mLog.Debugf("  P (Redis PubSub): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid redis pub-sub provider config", v1alpha2.BadConfig)
	}
	i.Config = vConfig
	if i.Config.Host == "" {
		return v1alpha2.NewCOAError(nil, "Redis host is not supplied", v1alpha2.MissingConfig)
	}

	i.Subscribers = make(map[string][]v1alpha2.EventHandler)
	options := &redis.Options{
		Addr:            i.Config.Host,
		Password:        i.Config.Password,
		DB:              0,
		MaxRetries:      3,
		MaxRetryBackoff: time.Second * 2,
	}
	if i.Config.RequiresTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: !i.Config.RequiresTLS,
		}
	}
	client := redis.NewClient(options)
	if _, err := client.Ping().Result(); err != nil {
		mLog.Debugf("  P (Redis PubSub): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", i.Config.Host), v1alpha2.InternalError)
	}
	i.Client = client
	i.Ctx, i.Cancel = context.WithCancel(context.Background())
	i.Queue = make(chan RedisMessageWrapper, int(i.Config.QueueDepth))
	for k := uint(0); k < uint(i.Config.NumberOfWorkers); k++ {
		go i.worker()
	}
	return nil
}

func (i *RedisPubSubProvider) worker() {
	for {
		select {
		case <-i.Ctx.Done():
			return
		case msg := <-i.Queue:
			i.processMessage(msg)
		}
	}
}
func (i *RedisPubSubProvider) processMessage(msg RedisMessageWrapper) error {
	var evt v1alpha2.Event
	err := json.Unmarshal([]byte(msg.Message.(string)), &evt)
	if err != nil {
		return v1alpha2.NewCOAError(err, "failed to unmarshal event", v1alpha2.InternalError)
	}
	if err := msg.Handler(msg.Topic, evt); err != nil {
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to handle message %s", msg.MessageID), v1alpha2.InternalError)
	}
	if err := i.Client.XAck(msg.Topic, i.Config.ConsumerID, msg.MessageID).Err(); err != nil {
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to acknowledge message %s", msg.MessageID), v1alpha2.InternalError)
	}
	return nil
}

func (i *RedisPubSubProvider) Publish(topic string, event v1alpha2.Event) error {
	_, err := i.Client.XAdd(&redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{"data": event},
	}).Result()
	if err != nil {
		mLog.Debugf("  P (Redis PubSub) : failed to publish message %v", err)
		return v1alpha2.NewCOAError(err, "failed to publish message", v1alpha2.InternalError)
	}
	return nil
}
func (i *RedisPubSubProvider) Subscribe(topic string, handler v1alpha2.EventHandler) error {
	err := i.Client.XGroupCreateMkStream(topic, i.Config.ConsumerID, "0").Err()
	//Ignore BUSYGROUP errors
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		mLog.Debugf("  P (Redis PubSub) : failed to subscribe %v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to subsceribe to topic %s", topic), v1alpha2.InternalError)
	}
	go i.pollNewMessagesLoop(topic, handler)
	go i.reclaimPendingMessagesLoop(topic, handler)
	return nil
}

func (i *RedisPubSubProvider) pollNewMessagesLoop(topic string, handler v1alpha2.EventHandler) {
	for {
		if i.Ctx.Err() != nil {
			return
		}
		streams, err := i.Client.XReadGroup(&redis.XReadGroupArgs{
			Group:    i.Config.ConsumerID,
			Consumer: i.Config.ConsumerID,
			Streams:  []string{topic, ">"},
			Count:    int64(i.Config.QueueDepth),
			Block:    0,
		}).Result()
		if err != nil {
			mLog.Debugf("  P (Redis PubSub) : failed to poll message %v", err)
			time.Sleep(30 * time.Second)
			continue
		}
		for _, s := range streams {
			i.enqueueMessages(s.Stream, handler, s.Messages)
		}
	}
}

func (i *RedisPubSubProvider) enqueueMessages(topic string, handler v1alpha2.EventHandler, msgs []redis.XMessage) {
	for _, msg := range msgs {
		rmsg := createRedisMessageWrapper(topic, handler, msg)
		select {
		case i.Queue <- rmsg:
		case <-i.Ctx.Done():
			return
		}
	}
}

func createRedisMessageWrapper(topic string, handler v1alpha2.EventHandler, msg redis.XMessage) RedisMessageWrapper {
	var data interface{}
	if dataValue, exists := msg.Values["data"]; exists && dataValue != nil {
		data = dataValue
	}
	return RedisMessageWrapper{
		Topic:     topic,
		Message:   data,
		MessageID: msg.ID,
		Handler:   handler,
	}
}

func (i *RedisPubSubProvider) reclaimPendingMessagesLoop(topic string, handler v1alpha2.EventHandler) {
	if i.Config.ProcessingTimeout == 0 || i.Config.RedeliverInterval == 0 {
		return
	}
	i.reclaimPendingMessages(topic, handler)
	reclaimTicker := time.NewTicker(i.Config.RedeliverInterval)
	for {
		select {
		case <-i.Ctx.Done():
			return
		case <-reclaimTicker.C:
			i.reclaimPendingMessages(topic, handler)
		}
	}
}

func (i *RedisPubSubProvider) reclaimPendingMessages(topic string, handler v1alpha2.EventHandler) {
	for {
		pendingResult, err := i.Client.XPendingExt(&redis.XPendingExtArgs{
			Stream: topic,
			Group:  i.Config.ConsumerID,
			Start:  "-",
			End:    "+",
			Count:  int64(i.Config.QueueDepth),
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			mLog.Debugf("  P (Redis PubSub) : failed to get pending message %v", err)
			break
		}
		msgIDs := make([]string, 0, len(pendingResult))
		for _, msg := range pendingResult {
			if msg.Idle >= i.Config.ProcessingTimeout {
				msgIDs = append(msgIDs, msg.ID)
			}
		}
		if len(msgIDs) == 0 {
			break
		}
		claimResult, err := i.Client.XClaim(&redis.XClaimArgs{
			Stream:   topic,
			Group:    i.Config.ConsumerID,
			Consumer: i.Config.ConsumerID,
			MinIdle:  i.Config.ProcessingTimeout,
			Messages: msgIDs,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			mLog.Debugf("  P (Redis PubSub) : failed to reclaim pending message %v", err)
			break
		}
		i.enqueueMessages(topic, handler, claimResult)
		// If the Redis nil error is returned, it means some messages in the pending
		// state no longer exist. We need to acknowledge these mesages to
		// remove them from the pending list
		if errors.Is(err, redis.Nil) {
			// Build a set of message IDs that were not returned
			// that potentitally no longer exist
			expectedMsgIDs := make(map[string]struct{}, len(msgIDs))
			for _, id := range msgIDs {
				expectedMsgIDs[id] = struct{}{}
			}
			for _, claimed := range claimResult {
				delete(expectedMsgIDs, claimed.ID)
			}
			i.removeMessagesThatNoLongerExistFromPending(topic, expectedMsgIDs, handler)
		}
	}
}

func (i *RedisPubSubProvider) removeMessagesThatNoLongerExistFromPending(topic string, messageIDs map[string]struct{}, handler v1alpha2.EventHandler) {
	for pendingID := range messageIDs {
		claimResultSingleMsg, err := i.Client.XClaim(&redis.XClaimArgs{
			Stream:   topic,
			Group:    i.Config.ConsumerID,
			Consumer: i.Config.ConsumerID,
			MinIdle:  i.Config.ProcessingTimeout,
			Messages: []string{pendingID},
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			mLog.Debugf("  P (Redis PubSub) : failed to remove pending message %v", err)
			continue
		}
		if errors.Is(err, redis.Nil) {
			if err = i.Client.XAck(topic, i.Config.ConsumerID, pendingID).Err(); err != nil {
				mLog.Debugf("  P (Redis PubSub) : error acknowledging Redis message %s after failed claim for %s - %v", i.Config.ConsumerID, pendingID, err)
			} else {
				i.enqueueMessages(topic, handler, claimResultSingleMsg)
			}
		}
	}
}

func toRedisPubSubProviderConfig(config providers.IProviderConfig) (RedisPubSubProviderConfig, error) {
	ret := RedisPubSubProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	//ret.Host = providers.LoadEnv(ret.Host)
	//ret.Password = providers.LoadEnv(ret.Password)
	if ret.NumberOfWorkers <= 0 {
		ret.NumberOfWorkers = 1
	}
	return ret, err
}
