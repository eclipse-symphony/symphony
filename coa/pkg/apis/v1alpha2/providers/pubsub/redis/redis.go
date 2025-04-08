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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/host"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var mLog = logger.NewLogger("coa.runtime")

type RedisPubSubProvider struct {
	Config      RedisPubSubProviderConfig          `json:"config"`
	Subscribers map[string][]v1alpha2.EventHandler `json:"subscribers"`
	Client      *redis.Client
	Ctx         context.Context
	Cancel      context.CancelFunc
	Context     *contexts.ManagerContext
	WorkerLock  *sync.Mutex
	IdleWorkers int
	rwLock      sync.RWMutex
	readyFlag   bool
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
	ConsumerID      string `json:"consumerID"`
}

const (
	// ResetIdleTimeInterval
	ResetIdleTimeInterval = 5 * time.Second
	// ClaimPendingMessage
	ClaimMessageInterval = 10 * time.Second
	// ClaimPendingMessageIdleTime
	ClaimMessageIdleTime = 30 * time.Second

	DefaultNumberOfWorkers = 20
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
	if v, ok := properties["numberOfWorkers"]; ok {
		val := v //providers.LoadEnv(v)
		if val != "" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return ret, v1alpha2.NewCOAError(err, "invalid int value in the 'numberOfWorkers' setting of Redis pub-sub provider", v1alpha2.BadConfig)
			}
			ret.NumberOfWorkers = n
		} else {
			ret.NumberOfWorkers = DefaultNumberOfWorkers
		}
	}
	if v, ok := properties["consumerID"]; ok {
		ret.ConsumerID = v // providers.LoadEnv(v)
	} else {
		ret.ConsumerID = ""
	}
	ret.ConsumerID = ret.ConsumerID + generateConsumerIDSuffix()

	if ret.NumberOfWorkers <= 0 {
		ret.NumberOfWorkers = DefaultNumberOfWorkers
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
	i.IdleWorkers = i.Config.NumberOfWorkers
	i.WorkerLock = &sync.Mutex{}
	client := redis.NewClient(options)
	if _, err := client.Ping(i.Ctx).Result(); err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", i.Config.Host), v1alpha2.InternalError)
	}
	i.Client = client

	return nil
}

func (i *RedisPubSubProvider) Publish(topic string, event v1alpha2.Event) error {
	messageId, err := i.Client.XAdd(i.Ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{"data": event},
	}).Result()
	if err != nil {
		mLog.Errorf("  P (Redis PubSub) : failed to publish message %v", err)
		return v1alpha2.NewCOAError(err, "failed to publish message", v1alpha2.InternalError)
	}
	mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : published message %s for topic %s", messageId, topic)
	return nil
}
func (i *RedisPubSubProvider) Subscribe(topic string, handler v1alpha2.EventHandler) error {
	mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : subscribing to topic %s with Group %s", topic, handler.Group)
	err := i.Client.XGroupCreateMkStream(i.Ctx, topic, handler.Group, "0").Err()
	//Ignore BUSYGROUP errors
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		mLog.Errorf("  P (Redis PubSub) : failed to subscribe %v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to subscribe to topic %s and group %s", topic, handler.Group), v1alpha2.InternalError)
	}

	go func() {
		mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : check host initialization, status topic %s with Group %s", topic, handler.Group)
		for {
			if host.IsHostReady() {
				mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : start poll message, topic %s with Group %s", topic, handler.Group)
				go i.pollNewMessagesLoop(topic, handler)
				go i.ClaimMessageLoop(topic, handler)
				return
			}
			mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : host status not ready, topic %s with Group %s", topic, handler.Group)
			time.Sleep(1 * time.Second)
		}
	}()
	return nil
}

func (i *RedisPubSubProvider) pollNewMessagesLoop(topic string, handler v1alpha2.EventHandler) {
	i.pollNewMessages(topic, handler)
	reclaimTicker := time.NewTicker(ClaimMessageInterval)
	defer reclaimTicker.Stop()
	for {
		select {
		case <-i.Ctx.Done():
			return
		case <-reclaimTicker.C:
			i.pollNewMessages(topic, handler)
		}
	}
}

func (i *RedisPubSubProvider) pollNewMessages(topic string, handler v1alpha2.EventHandler) {
	// If worker is claimed but not started, release it in defer function
	claimWorker := false
	workerStarted := false
	defer func() {
		if claimWorker && !workerStarted {
			i.ReleaseWorker(topic)
		}
	}()

	for {
		// DO NOT REMOVE THIS COMMENT
		// gofail: var PollNewMessagesLoop string
		if i.Ctx.Err() != nil {
			return
		}
		claimWorker = false
		workerStarted = false

		streams, err := i.Client.XReadGroup(i.Ctx, &redis.XReadGroupArgs{
			Group:    handler.Group,
			Consumer: i.Config.ConsumerID,
			Streams:  []string{topic, ">"},
			Count:    1,
			Block:    1 * time.Second,
		}).Result()
		if err != nil {
			break
		}
		if len(streams) == 1 && len(streams[0].Messages) == 1 {
			if claimWorker = i.WaitForIdleWorkers(streams[0].Messages[0].ID, time.Second); !claimWorker {
				mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : no idle workers, abort current pollNewMessages for topic %s and group %s", topic, handler.Group)
				return
			}
			workerStarted = false
			mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : new message for topic %s, group %s, messages %s", topic, handler.Group, streams[0].Messages[0].ID)
			go i.processMessage(topic, handler, &streams[0].Messages[0])
			workerStarted = true
		} else {
			break
		}
		mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : processed pollnewmessages for topic %s", topic)
	}
}

func (i *RedisPubSubProvider) ClaimMessageLoop(topic string, handler v1alpha2.EventHandler) {
	i.reclaimPendingMessages(topic, handler)
	reclaimTicker := time.NewTicker(ClaimMessageInterval)
	defer reclaimTicker.Stop()
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
	// If worker is claimed but not started, release it in defer function
	claimWorker := false
	workerStarted := false
	defer func() {
		if claimWorker && !workerStarted {
			i.ReleaseWorker(topic)
		}
	}()
	for {
		claimWorker = false
		workerStarted = false
		pendingResult, err := i.Client.XPendingExt(i.Ctx, &redis.XPendingExtArgs{
			Stream:   topic,
			Group:    handler.Group,
			Start:    "-",
			End:      "+",
			Count:    1,
			Idle:     ClaimMessageIdleTime,
			Consumer: "",
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			mLog.Errorf("  P (Redis PubSub) : failed to get pending message %v", err)
			return
		}
		if len(pendingResult) != 1 {
			return
		}
		if claimWorker = i.WaitForIdleWorkers(pendingResult[0].ID, time.Second); !claimWorker {
			mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : unable to claim idle workers in %s for topic %s, group %s, message %s", time.Second, topic, handler.Group, pendingResult[0].ID)
			return
		}
		msg, succeeded := i.ClaimMessage(topic, handler.Group, ClaimMessageIdleTime, pendingResult[0].ID)
		if !succeeded {
			mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : failed to claim message %s for topic %s, group %s", msg.ID, topic, handler.Group)
			continue
		}
		go i.processMessage(topic, handler, msg)
		workerStarted = true
	}
}

func (i *RedisPubSubProvider) processMessage(topic string, handler v1alpha2.EventHandler, msg *redis.XMessage) error {
	defer i.ReleaseWorker(topic)
	mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : processing message %s for topic %s, group %s", msg.ID, topic, handler.Group)

	// Reset the idle time for the message until process finishes so other processes won't pick it up
	stopCh := make(chan struct{})
	defer close(stopCh)
	go i.ResetIdleTimeLoop(topic, handler.Group, msg.ID, stopCh)

	var data interface{}
	if dataValue, exists := msg.Values["data"]; exists && dataValue != nil {
		data = dataValue
	}
	var evt v1alpha2.Event
	err := json.Unmarshal([]byte(utils.FormatAsString(data)), &evt)
	if err != nil {
		mLog.ErrorfCtx(i.Ctx, "  P (Redis PubSub) : failed to unmarshal event for message %s and topic %s, group %s: %v", msg.ID, topic, handler.Group, err.Error())
		return v1alpha2.NewCOAError(err, "failed to unmarshal event", v1alpha2.InternalError)
	}
	shouldRetry := v1alpha2.EventShouldRetryWrapper(handler, topic, evt)
	if shouldRetry {
		mLog.ErrorfCtx(i.Ctx, "  P (Redis PubSub) : processing failed with retriable error for message %s for topic %s, group %s", msg.ID, topic, handler.Group)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("failed to handle message %s", msg.ID), v1alpha2.InternalError)
	}
	_, err = i.Client.XAck(i.Ctx, topic, handler.Group, msg.ID).Result()
	if err != nil {
		mLog.ErrorfCtx(i.Ctx, "  P (Redis PubSub) : failed to acknowledge message %s for topic %s, group %s: %v", msg.ID, topic, handler.Group, err)
	}
	mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : processing succeeded for message %s for topic %s, group %s", msg.ID, topic, handler.Group)
	// TODO: This only works when we have only one consumer group for each topic
	_, err = i.Client.XDel(i.Ctx, topic, msg.ID).Result()
	if err != nil {
		mLog.ErrorfCtx(i.Ctx, "  P (Redis PubSub) : failed to delete message %s for topic %s, group %s: %v", msg.ID, topic, handler.Group, err)
	}
	return nil
}

func (i *RedisPubSubProvider) ClaimMessage(topic string, group string, minIdle time.Duration, msgID string) (*redis.XMessage, bool) {
	claimResult, err := i.Client.XClaim(i.Ctx, &redis.XClaimArgs{
		Stream:   topic,
		Group:    group,
		Consumer: i.Config.ConsumerID,
		MinIdle:  minIdle,
		Messages: []string{msgID},
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		mLog.Error("  P (Redis PubSub) : failed to reclaim pending message %s, topic %s, group %s: %v", msgID, topic, group, err)
		return nil, false
	}
	if len(claimResult) == 1 {
		return &claimResult[0], true
	}
	return nil, false
}

func (i *RedisPubSubProvider) ResetIdleTimeLoop(topic string, group string, msgID string, stopCh chan struct{}) {
	ticker := time.NewTicker(ResetIdleTimeInterval)
	claimIdleTime := ResetIdleTimeInterval - 1*time.Second
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mLog.InfofCtx(i.Ctx, "  P (Redis PubSub) : resetting idle time for message %s for topic %s, group %s", msgID, topic, group)
			_, succeeded := i.ClaimMessage(topic, group, claimIdleTime, msgID)
			if !succeeded {
				mLog.ErrorfCtx(i.Ctx, "  P (Redis PubSub) : failed to reset idle time for message %s for topic %s, group %s", msgID, topic, group)
			}
		case <-stopCh:
			return // Exit the goroutine when the stop signal is received
		}
	}
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

func (i *RedisPubSubProvider) WaitForIdleWorkers(msgID string, timeout time.Duration) bool {
	timeoutChan := time.After(timeout)
	claimed := false
	for {
		select {
		case <-timeoutChan:
			return claimed
		default:
			if claimed = i.ClaimWorker(msgID); claimed {
				return true
			}
		}
		time.Sleep(timeout / 10)
	}
}

func (i *RedisPubSubProvider) ClaimWorker(msgID string) bool {
	i.WorkerLock.Lock()
	defer i.WorkerLock.Unlock()
	if i.IdleWorkers == 0 {
		return false
	}
	mLog.DebugfCtx(i.Ctx, "  P (Redis PubSub) : claimWorker for message %s, remaining %d", msgID, i.IdleWorkers)
	i.IdleWorkers--
	return true
}

func (i *RedisPubSubProvider) ReleaseWorker(msgID string) {
	i.WorkerLock.Lock()
	defer i.WorkerLock.Unlock()
	i.IdleWorkers++
	mLog.DebugfCtx(i.Ctx, "  P (Redis PubSub) : releaseWorker for message %s, remaining %d", msgID, i.IdleWorkers)
}
