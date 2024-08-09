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
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var mLog = logger.NewLogger("coa.runtime")

type RedisPubSubProvider struct {
	Config          RedisPubSubProviderConfig          `json:"config"`
	Subscribers     map[string][]v1alpha2.EventHandler `json:"subscribers"`
	Client          *redis.Client
	Queue           chan RedisMessageWrapper
	Ctx             context.Context
	Cancel          context.CancelFunc
	Context         *contexts.ManagerContext
	ClaimedMessages map[string]bool
	TopicLock       map[string]*sync.Mutex
	MapLock         *sync.Mutex
}

type RedisMessageWrapper struct {
	MessageID string
	Topic     string
	Message   interface{}
	Handler   v1alpha2.EventHandler
}

type RedisPubSubProviderConfig struct {
	Name            string `json:"name"`
	Host            string `json:"host"`
	Password        string `json:"password,omitempty"`
	RequiresTLS     bool   `json:"requiresTLS,omitempty"`
	NumberOfWorkers int    `json:"numberOfWorkers,omitempty"`
	QueueDepth      int    `json:"queueDepth,omitempty"`
	ConsumerID      string `json:"consumerID"`
	MultiInstance   bool   `json:"multiInstance,omitempty"`
}

const (
	RedisGroup = "symphony"
	// defines the interval in which the provider should check for pending messages
	PendingMessagesScanInterval = 5 * time.Second
	// defines after how much idle time the provider should check for pending messages that previously claimed
	// by itself and reset the idle time of them to prevent them from being claimed by other clients
	ExtendMessageOwnershipWithIdleTime = 30 * time.Second
	// defines the interval in which the provider should check for pending messages from other clients
	PendingMessagesScanIntervalOtherClient = 60 * time.Second
	// defines after how much idle time the provider should check for pending messages that previously claimed
	// by other clients
	ClaimMessageFromOtherClientWithIdleTime = 60 * time.Second
)

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
	if v, ok := properties["multiInstance"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			bVal, err := strconv.ParseBool(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid bool value in the 'requiresTLS' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.MultiInstance = bVal
		}
	} else {
		ret.MultiInstance = false
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
	if ret.QueueDepth <= 0 {
		ret.QueueDepth = 10
	}
	if v, ok := properties["consumerID"]; ok {
		ret.ConsumerID = v // providers.LoadEnv(v)
	} else {
		ret.ConsumerID = ""
	}
	ret.ConsumerID = ret.ConsumerID + generateConsumerIDSuffix()

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
		mLog.Errorf("  P (Redis PubSub) : failed to initialize provider %v", err)
		return err
	}
	return i.Init(config)
}

func (i *RedisPubSubProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toRedisPubSubProviderConfig(config)
	if err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid redis pub-sub provider config", v1alpha2.BadConfig)
	}
	i.Config = vConfig
	if i.Config.Host == "" {
		return v1alpha2.NewCOAError(nil, "Redis host is not supplied", v1alpha2.MissingConfig)
	}

	i.Ctx, i.Cancel = context.WithCancel(context.Background())
	i.ClaimedMessages = make(map[string]bool)
	i.MapLock = &sync.Mutex{}
	i.TopicLock = make(map[string]*sync.Mutex)

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
	if _, err := client.Ping(i.Ctx).Result(); err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", i.Config.Host), v1alpha2.InternalError)
	}
	i.Client = client
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
	i.ClaimedMessages[msg.MessageID] = true
	var evt v1alpha2.Event
	err := json.Unmarshal([]byte(utils.FormatAsString(msg.Message)), &evt)
	if err != nil {
		return v1alpha2.NewCOAError(err, "failed to unmarshal event", v1alpha2.InternalError)
	}
	shouldRetry := v1alpha2.EventShouldRetryWrapper(msg.Handler, msg.Topic, evt)
	//err = msg.Handler(msg.Topic, evt)
	lock := i.getTopicLock(msg.Topic)
	lock.Lock()
	defer lock.Unlock()
	if shouldRetry {
		delete(i.ClaimedMessages, msg.MessageID)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to handle message %s", msg.MessageID), v1alpha2.InternalError)
	}
	i.Client.XAck(i.Ctx, msg.Topic, RedisGroup, msg.MessageID)
	delete(i.ClaimedMessages, msg.MessageID)

	return nil
}

func (i *RedisPubSubProvider) Publish(topic string, event v1alpha2.Event) error {
	_, err := i.Client.XAdd(i.Ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{"data": event},
	}).Result()
	if err != nil {
		mLog.Errorf("  P (Redis PubSub) : failed to publish message %v", err)
		return v1alpha2.NewCOAError(err, "failed to publish message", v1alpha2.InternalError)
	}
	return nil
}
func (i *RedisPubSubProvider) Subscribe(topic string, handler v1alpha2.EventHandler) error {
	err := i.Client.XGroupCreateMkStream(i.Ctx, topic, RedisGroup, "0").Err()
	//Ignore BUSYGROUP errors
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		mLog.Errorf("  P (Redis PubSub) : failed to subscribe %v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to subsceribe to topic %s", topic), v1alpha2.InternalError)
	}
	go i.pollNewMessagesLoop(topic, handler)
	go i.ClaimMessageLoop(topic, i.Config.ConsumerID, handler, PendingMessagesScanInterval, ExtendMessageOwnershipWithIdleTime)
	if i.Config.MultiInstance {
		go i.ClaimMessageLoop(topic, "", handler, PendingMessagesScanIntervalOtherClient, ClaimMessageFromOtherClientWithIdleTime)
	}
	return nil
}

func (i *RedisPubSubProvider) pollNewMessagesLoop(topic string, handler v1alpha2.EventHandler) {
	for {
		if i.Ctx.Err() != nil {
			return
		}
		streams, err := i.Client.XReadGroup(i.Ctx, &redis.XReadGroupArgs{
			Group:    RedisGroup,
			Consumer: i.Config.ConsumerID,
			Streams:  []string{topic, ">"},
			Count:    int64(i.Config.QueueDepth),
			Block:    0,
		}).Result()
		if err != nil {
			mLog.Errorf("  P (Redis PubSub) : failed to poll message %v", err)
			continue
		}
		for _, s := range streams {
			i.enqueueMessages(s.Stream, handler, s.Messages)
		}
		time.Sleep(PendingMessagesScanInterval)
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

func (i *RedisPubSubProvider) ClaimMessageLoop(topic string, consumerId string, handler v1alpha2.EventHandler, scanInterval time.Duration, messageIdleTime time.Duration) {
	i.reclaimPendingMessages(topic, messageIdleTime, consumerId, handler)
	reclaimTicker := time.NewTicker(scanInterval)
	defer reclaimTicker.Stop()
	for {
		select {
		case <-i.Ctx.Done():
			return
		case <-reclaimTicker.C:
			i.reclaimPendingMessages(topic, messageIdleTime, consumerId, handler)
		}
	}
}

func (i *RedisPubSubProvider) reclaimPendingMessages(topic string, idleTime time.Duration, consumer string, handler v1alpha2.EventHandler) {
	start := "-"
	for {
		pendingResult, err := i.Client.XPendingExt(i.Ctx, &redis.XPendingExtArgs{
			Stream:   topic,
			Group:    RedisGroup,
			Start:    start,
			End:      "+",
			Count:    int64(i.Config.QueueDepth),
			Idle:     idleTime,
			Consumer: consumer,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			mLog.Errorf("  P (Redis PubSub) : failed to get pending message %v", err)
			break
		}
		if len(pendingResult) == 0 {
			break
		}
		start = pendingResult[len(pendingResult)-1].ID
		msgIDs := make([]string, 0, len(pendingResult))
		for _, msg := range pendingResult {
			msgIDs = append(msgIDs, msg.ID)
		}
		i.XClaimWrapper(topic, idleTime, consumer, msgIDs, handler)
	}
}
func (i *RedisPubSubProvider) XClaimWrapper(topic string, minIdle time.Duration, consumer string, msgIDs []string, handler v1alpha2.EventHandler) {
	lock := i.getTopicLock(topic)
	lock.Lock()
	defer lock.Unlock()
	claimResult, err := i.Client.XClaim(i.Ctx, &redis.XClaimArgs{
		Stream:   topic,
		Group:    RedisGroup,
		Consumer: consumer,
		MinIdle:  minIdle,
		Messages: msgIDs,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		mLog.Error("  P (Redis PubSub) : failed to reclaim pending message %v", err)
		return
	}
	filteredClaimResult := make([]redis.XMessage, 0, len(claimResult))
	for _, msg := range claimResult {
		if _, ok := i.ClaimedMessages[msg.ID]; !ok {
			filteredClaimResult = append(filteredClaimResult, msg)
		}
	}
	i.enqueueMessages(topic, handler, filteredClaimResult)
}

func toRedisPubSubProviderConfig(config providers.IProviderConfig) (RedisPubSubProviderConfig, error) {
	ret := RedisPubSubProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	var configs map[string]interface{}
	err = json.Unmarshal(data, &configs)
	if err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to parse to map[string]interface{} %+v", err)
		return ret, err
	}
	configStrings := map[string]string{}
	for k, v := range configs {
		configStrings[k] = utils.FormatAsString(v)
	}

	ret, err = RedisPubSubProviderConfigFromMap(configStrings)
	if err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to parse to RedisPubSubProviderConfig %+v", err)
		return ret, err
	}
	return ret, err
}

func generateConsumerIDSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (i *RedisPubSubProvider) getTopicLock(topic string) *sync.Mutex {
	i.MapLock.Lock()
	defer i.MapLock.Unlock()
	if _, ok := i.TopicLock[topic]; !ok {
		i.TopicLock[topic] = &sync.Mutex{}
	}
	return i.TopicLock[topic]
}
