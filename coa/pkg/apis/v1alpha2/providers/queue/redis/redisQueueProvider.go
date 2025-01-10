/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package redisqueue

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
var mLock sync.Mutex

type RedisQueueProviderConfig struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Password    string `json:"password,omitempty"`
	RequiresTLS bool   `json:"requiresTLS,omitempty"`
	queueName   string
	Context     *contexts.ManagerContext
}

type RedisQueueProvider struct {
	client     *redis.Client
	Context    *contexts.ManagerContext
	Ctx        context.Context
	Cancel     context.CancelFunc
	MaxRetries int
}

func NewRedisQueue(client *redis.Client, queueName string) *RedisQueueProvider {
	return &RedisQueueProvider{
		client: client,
	}
}
func RedisQueueProviderConfigFromMap(properties map[string]string) (RedisQueueProviderConfig, error) {
	ret := RedisQueueProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

func (s *RedisQueueProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *RedisQueueProvider) InitWithMap(properties map[string]string) error {
	config, err := RedisQueueProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func toRedisQueueProviderConfig(config providers.IProviderConfig) (RedisQueueProviderConfig, error) {
	ret := RedisQueueProviderConfig{}
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
func RedisPubSubProviderConfigFromMap(properties map[string]string) (RedisQueueProviderConfig, error) {
	ret := RedisQueueProviderConfig{}
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
	//TODO: Finish this
	return ret, nil
}
func (rq *RedisQueueProvider) Size(queue string) int {
	xMessages, err := rq.client.XRangeN(rq.Ctx, queue, "0", "+", 1000).Result()
	if err != nil {
		return 0
	}
	return len(xMessages)
}
func (rq *RedisQueueProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toRedisQueueProviderConfig(config)
	if err != nil {
		mLog.Errorf("  P (Redis PubSub): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid redis pub-sub provider config", v1alpha2.BadConfig)
	}
	if vConfig.Host == "" {
		return v1alpha2.NewCOAError(nil, "Redis host is not supplied", v1alpha2.MissingConfig)
	}
	rq.MaxRetries = 3
	rq.Ctx, rq.Cancel = context.WithCancel(context.Background())

	options := &redis.Options{
		Addr:            vConfig.Host,
		Password:        vConfig.Password,
		DB:              0,
		MaxRetries:      3,
		MaxRetryBackoff: time.Second * 2,
	}
	if vConfig.RequiresTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: !vConfig.RequiresTLS,
		}
	}
	client := redis.NewClient(options)
	if _, err := client.Ping(rq.Ctx).Result(); err != nil {
		mLog.Errorf("  P (Redis Queue): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", vConfig.Host), v1alpha2.InternalError)
	}
	rq.client = client

	return nil
}

func (rq *RedisQueueProvider) Enqueue(queue string, element interface{}) (string, error) {
	data, err := json.Marshal(element)
	if err != nil {
		return "", err
	}
	return rq.client.XAdd(rq.Ctx, &redis.XAddArgs{
		Stream: queue,
		Values: map[string]interface{}{"data": data},
	}).Result()
}
func (rq *RedisQueueProvider) PeekFromBegining(queue string, fromBegining bool) (interface{}, error) {
	// Get the last ID processed by this consumer

	// Read message
	xMessages, err := rq.client.XRangeN(rq.Ctx, queue, "0", "+", 1).Result()
	if err != nil {
		return nil, err
	}
	if len(xMessages) == 0 {
		return nil, nil
	}
	xMsg := xMessages[0]
	jsonData := xMsg.Values["data"].(string)
	var result interface{}
	err = json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	// update last read ID
	lastReadKey := fmt.Sprintf("%s:lastID", queue)
	err = rq.client.Set(rq.Ctx, lastReadKey, xMsg.ID, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to update last read ID: %w", err)
	}
	return result, nil
}

func (rq *RedisQueueProvider) Peek(queue string) (interface{}, error) {
	var start string
	// Get the last ID processed by this consumer
	var err error
	lastIDkey := fmt.Sprintf("%s:lastID", queue)
	start, err = rq.client.Get(rq.Ctx, lastIDkey).Result()
	mLog.Errorf("  P redis queue: start is  %s", start)
	if err == redis.Nil {
		start = "0"
	} else if err != nil {
		return nil, err
	}
	// Read message
	xMessages, err := rq.client.XRangeN(rq.Ctx, queue, start, "+", 1).Result()
	if err != nil {
		return nil, err
	}
	if len(xMessages) == 0 {
		return nil, nil
	}
	xMsg := xMessages[0]
	mLog.Errorf("  P redis queue:xmsg id %s", xMsg.ID)
	jsonData := xMsg.Values["data"].(string)
	var result interface{}
	err = json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	// update last read ID
	lastReadKey := fmt.Sprintf("%s:lastID", queue)
	err = rq.client.Set(rq.Ctx, lastReadKey, "("+xMsg.ID, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to update last read ID: %w", err)
	}
	return result, nil
}

func (rq *RedisQueueProvider) RemoveFromQueue(queue string, messageID string) error {
	return rq.client.XDel(rq.Ctx, queue, messageID).Err()
}

func (rq *RedisQueueProvider) Dequeue(queue string) (interface{}, error) {
	// Get the last ID processed by this consumer
	lastIDkey := fmt.Sprintf("%s:lastID", queue)
	start, err := rq.client.Get(context.TODO(), lastIDkey).Result()
	if err == redis.Nil {
		start = "0"
	} else if err != nil {
		return nil, err
	}

	// Read message
	xMessages, err := rq.client.XRangeN(context.TODO(), queue, start, "+", 1).Result()
	if err != nil {
		return nil, err
	}
	if len(xMessages) == 0 {
		return nil, nil
	}
	xMsg := xMessages[0]
	jsonData := xMsg.Values["data"].(string)
	var result interface{}
	err = json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	// Delete message
	err = rq.client.XDel(context.TODO(), queue, xMsg.ID).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	// Update last read ID
	err = rq.client.Set(context.TODO(), lastIDkey, "("+xMsg.ID, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to update last read ID: %w", err)
	}

	return result, nil
}
