/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package redisstate

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/go-redis/redis/v7"
)

var rLog = logger.NewLogger("coa.runtime")

const (
	entryCountPerList = 100
	separator         = ":"
)

type RedisStateProviderConfig struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Password    string `json:"password,omitempty"`
	RequiresTLS bool   `json:"requiresTLS,omitempty"`
}

func RedisStateProviderConfigFromMap(properties map[string]string) (RedisStateProviderConfig, error) {
	ret := RedisStateProviderConfig{}
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

type RedisStateProvider struct {
	Config  RedisStateProviderConfig
	Context *contexts.ManagerContext
	Client  *redis.Client
}

func (r *RedisStateProvider) ID() string {
	return r.Config.Name
}

func (r *RedisStateProvider) SetContext(ctx *contexts.ManagerContext) {
	r.Context = ctx
}

func (i *RedisStateProvider) InitWithMap(properties map[string]string) error {
	config, err := RedisStateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (r *RedisStateProvider) Init(config providers.IProviderConfig) error {
	vConfig, err := toRedisStateProviderConfig(config)
	if err != nil {
		rLog.Debugf("  P (Redis State): failed to parse provider config %+v", err)
		return v1alpha2.NewCOAError(nil, "provided config is not a valid redis pub-sub provider config", v1alpha2.BadConfig)
	}
	r.Config = vConfig
	if r.Config.Host == "" {
		return v1alpha2.NewCOAError(nil, "Redis host is not supplied", v1alpha2.MissingConfig)
	}

	options := &redis.Options{
		Addr:            r.Config.Host,
		Password:        r.Config.Password,
		DB:              0,
		MaxRetries:      3,
		MaxRetryBackoff: time.Second * 2,
	}
	if r.Config.RequiresTLS {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: !r.Config.RequiresTLS,
		}
	}
	client := redis.NewClient(options)
	if _, err := client.Ping().Result(); err != nil {
		rLog.Debugf("  P (Redis State): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", r.Config.Host), v1alpha2.InternalError)
	}
	r.Client = client
	rLog.Debug("  P (Redis State): Successfully launch redis state provider")
	return nil
}

func (r *RedisStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	_, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	keyPrefix := getKeyNamePrefix(entry.Metadata)
	rLog.Debugf("  P (Redis State): upsert states %s with keyPrefix %s, traceId: %s", entry.Value.ID, keyPrefix, span.SpanContext().TraceID().String())

	key := fmt.Sprintf("%s%s%s", keyPrefix, separator, entry.Value.ID)
	var body []byte
	body, err = json.Marshal(entry.Value.Body)
	if err != nil {
		return entry.Value.ID, err
	}
	properties := map[string]interface{}{
		"values": string(body),
		"etag":   entry.Value.ETag,
	}
	_, err = r.Client.HSet(key, properties).Result()
	return entry.Value.ID, err
}

func (r *RedisStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	_, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var entities []states.StateEntry
	keyPrefix := getKeyNamePrefix(request.Metadata)
	rLog.Debugf("  P (Redis State): list states with keyPrefix %s, traceId: %s", keyPrefix, span.SpanContext().TraceID().String())

	filter := fmt.Sprintf("%s%s*", keyPrefix, separator)
	var cursor uint64 = 0
	var keys []string

	for {
		var err error
		keys, cursor, err = r.Client.Scan(cursor, filter, entryCountPerList).Result()
		if err != nil {
			rLog.Errorf("  P (Redis State): failed to get all the keys matching pattern %s: %+v", keyPrefix, err)
		}

		for _, key := range keys {
			result, err := r.Client.HGetAll(key).Result()
			if err != nil || len(result) == 0 {
				rLog.Errorf("  P (Redis State): failed to get entry for key %s: %+v", key, err)
				continue
			}
			parts := strings.Split(key, separator)
			if len(parts) != 3 {
				rLog.Errorf("  P (Redis State): key is not valid %s: %+v", key, err)
				continue
			}
			entry, err := CastRedisPropertiesToStateEntry(parts[2], result)
			if err != nil {
				rLog.Errorf("  P (Redis State): failed to cast entry for key %s: %+v", key, err)
				continue
			}
			entities = append(entities, entry)
		}

		if cursor == 0 {
			break
		}
	}

	return entities, "", nil
}

func (r *RedisStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	_, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	keyPrefix := getKeyNamePrefix(request.Metadata)
	rLog.Debugf("  P (Redis State): delete state %s with keyPrefix %s, traceId: %s", request.ID, keyPrefix, span.SpanContext().TraceID().String())

	HKey := fmt.Sprintf("%s%s%s", keyPrefix, separator, request.ID)
	_, err = r.Client.Del(HKey).Result()
	return nil
}

func (r *RedisStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	_, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	keyPrefix := getKeyNamePrefix(request.Metadata)

	rLog.Debugf("  P (Redis State): get state %s with keyPrefix %s, traceId: %s", request.ID, keyPrefix, span.SpanContext().TraceID().String())
	HKey := fmt.Sprintf("%s%s%s", keyPrefix, separator, request.ID)

	var data map[string]string
	data, err = r.Client.HGetAll(HKey).Result()
	if err != nil {
		rLog.Errorf("  P (Redis State): failed to get state %s with keyPrefix %s, traceId: %s", request.ID, keyPrefix, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	if len(data) == 0 {
		return states.StateEntry{}, v1alpha2.NewCOAError(nil, fmt.Sprintf("state %s not found", request.ID), v1alpha2.NotFound)
	}
	return CastRedisPropertiesToStateEntry(request.ID, data)
}

func toRedisStateProviderConfig(config providers.IProviderConfig) (RedisStateProviderConfig, error) {
	ret := RedisStateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func getKeyNamePrefix(metadata map[string]interface{}) string {
	namespace := "default"
	if n, ok := metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	// Construct object type
	objectType := ""
	if resource, ok := metadata["resource"]; ok {
		if rstring, ok := resource.(string); ok && rstring != "" {
			objectType = rstring
		}
	}
	if group, ok := metadata["group"]; ok {
		if gstring, ok := group.(string); ok && gstring != "" {
			objectType = objectType + "." + gstring
		}
	}
	return namespace + separator + objectType
}

func CastRedisPropertiesToStateEntry(id string, properties map[string]string) (states.StateEntry, error) {
	entry := states.StateEntry{}
	entry.ETag = properties["etag"]
	// Body should be a map[string]interface{} to be align with other state providers
	var BodyDict map[string]interface{}
	err := json.Unmarshal([]byte(properties["values"]), &BodyDict)
	if err != nil {
		return entry, err
	}
	entry.Body = BodyDict
	entry.ID = id
	return entry, nil
}
