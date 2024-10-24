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
	"github.com/redis/go-redis/v9"
)

var rLog = logger.NewLogger("coa.runtime")

const (
	entryCountPerList = 100
	separator         = "*"
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
	Ctx     context.Context
	Cancel  context.CancelFunc
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
	r.Ctx, r.Cancel = context.WithCancel(context.Background())
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
	if _, err := client.Ping(r.Ctx).Result(); err != nil {
		rLog.Debugf("  P (Redis State): failed to connect to redis %+v", err)
		return v1alpha2.NewCOAError(err, fmt.Sprintf("redis stream: error connecting to redis at %s", r.Config.Host), v1alpha2.InternalError)
	}
	r.Client = client
	rLog.Debug("  P (Redis State): Successfully launch redis state provider")
	return nil
}

func (r *RedisStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	ctx, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var keyPrefix string
	keyPrefix, err = getKeyNamePrefix(entry.Metadata)
	if err != nil {
		rLog.ErrorfCtx(ctx, "  P (Redis State): upsert states %s failed to get key prefix with error %s", entry.Value.ID, err.Error())
		return entry.Value.ID, err
	}

	rLog.DebugfCtx(ctx, "  P (Redis State): upsert states %s with keyPrefix %s", entry.Value.ID, keyPrefix)

	key := fmt.Sprintf("%s%s%s", keyPrefix, separator, entry.Value.ID)
	var body []byte
	body, err = json.Marshal(entry.Value.Body)
	if err != nil {
		return entry.Value.ID, err
	}
	if entry.Options.UpdateStatusOnly {
		var existing string
		existing, err = r.Client.HGet(r.Ctx, key, "values").Result()
		if err != nil {
			return entry.Value.ID, v1alpha2.NewCOAError(nil, fmt.Sprintf("redis state %s not found. Cannot update state only", entry.Value.ID), v1alpha2.BadRequest)
		}
		var oldEntryDict map[string]interface{}
		var oldStatusDict map[string]interface{}
		oldEntryDict, oldStatusDict, err = getStatusDictFromMarshalStateEntryBody([]byte(existing))
		if err != nil {
			return entry.Value.ID, v1alpha2.NewCOAError(nil, fmt.Sprintf("old redis state %s status cannot be parsed", entry.Value.ID), v1alpha2.InternalError)
		}
		var newStatusDict map[string]interface{}
		_, newStatusDict, err = getStatusDictFromMarshalStateEntryBody(body)
		if err != nil {
			return entry.Value.ID, v1alpha2.NewCOAError(nil, fmt.Sprintf("new redis state %s cannot be parsed", entry.Value.ID), v1alpha2.InternalError)
		}
		for k, v := range newStatusDict {
			oldStatusDict[k] = v
		}
		oldEntryDict["status"] = oldStatusDict
		body, _ = json.Marshal(oldEntryDict)
		_, err = r.Client.HSet(r.Ctx, key, "values", string(body)).Result()
		return entry.Value.ID, err
	}

	properties := map[string]interface{}{
		"values": string(body),
		"etag":   entry.Value.ETag,
	}
	_, err = r.Client.HSet(r.Ctx, key, properties).Result()
	return entry.Value.ID, err
}

func (r *RedisStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	ctx, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var entities []states.StateEntry
	var keyPrefix string
	keyPrefix, err = getObjectTypePrefixForList(request.Metadata)
	if err != nil {
		rLog.ErrorfCtx(ctx, "  P (Redis State): list states failed to get key prefix with error %s", err.Error())
		return entities, "", err
	}
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			keyPrefix = keyPrefix + separator + nstring
		}
	}
	// Scheduled events will call List periodically. Comment this log line to reduce the log.
	// rLog.DebugfCtx(ctx, "  P (Redis State): list states with keyPrefix %s", keyPrefix)

	filter := fmt.Sprintf("%s%s*", keyPrefix, separator)
	var cursor uint64 = 0
	var keys []string

	for {
		var err error
		keys, cursor, err = r.Client.Scan(r.Ctx, cursor, filter, entryCountPerList).Result()
		if err != nil {
			rLog.Errorf("  P (Redis State): failed to get all the keys matching pattern %s: %+v", keyPrefix, err)
		}

		for _, key := range keys {
			result, err := r.Client.HGetAll(r.Ctx, key).Result()
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
			if request.FilterType != "" && request.FilterValue != "" {
				var match bool
				match, err = states.MatchFilter(entry, request.FilterType, request.FilterValue)
				if err != nil {
					return entities, "", err
				} else if !match {
					continue
				}
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
	ctx, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var keyPrefix string
	keyPrefix, err = getKeyNamePrefix(request.Metadata)
	if err != nil {
		rLog.ErrorfCtx(ctx, "  P (Redis State): delete state %s failed to get key prefix with error %s", request.ID, err.Error())
		return err
	}
	rLog.DebugfCtx(ctx, "  P (Redis State): delete state %s with keyPrefix %s", request.ID, keyPrefix)

	HKey := fmt.Sprintf("%s%s%s", keyPrefix, separator, request.ID)
	_, err = r.Client.Del(r.Ctx, HKey).Result()
	return nil
}

func (r *RedisStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	ctx, span := observability.StartSpan("Redis State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var keyPrefix string
	keyPrefix, err = getKeyNamePrefix(request.Metadata)
	if err != nil {
		rLog.ErrorfCtx(ctx, "  P (Redis State): get state %s failed to get key prefix with error %s", request.ID, err.Error())
		return states.StateEntry{}, err
	}

	rLog.DebugfCtx(ctx, "  P (Redis State): get state %s with keyPrefix %s", request.ID, keyPrefix)
	HKey := fmt.Sprintf("%s%s%s", keyPrefix, separator, request.ID)

	var data map[string]string
	data, err = r.Client.HGetAll(r.Ctx, HKey).Result()
	if err != nil {
		rLog.ErrorfCtx(ctx, "  P (Redis State): failed to get state %s with keyPrefix %s", request.ID, keyPrefix)
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

func getKeyNamePrefix(metadata map[string]interface{}) (string, error) {
	namespace := "default"
	if n, ok := metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	// Construct object type
	objectType, err := getObjectTypePrefixForList(metadata)
	if err != nil {
		return "", err
	}
	return objectType + separator + namespace, nil
}

func getObjectTypePrefixForList(metadata map[string]interface{}) (string, error) {
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
	if objectType == "" {
		return "", v1alpha2.NewCOAError(nil, "Redis state provider object type is not specified", v1alpha2.BadConfig)
	}
	return objectType, nil
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

func getStatusDictFromMarshalStateEntryBody(body []byte) (map[string]interface{}, map[string]interface{}, error) {
	var EntryDict map[string]interface{}
	err := json.Unmarshal([]byte(body), &EntryDict)
	if err != nil {
		return nil, nil, err
	}
	if EntryDict == nil {
		EntryDict = make(map[string]interface{})
	}
	var statusDict map[string]interface{}
	var j []byte
	j, _ = json.Marshal(EntryDict["status"])
	err = json.Unmarshal(j, &statusDict)
	if err != nil {
		return nil, nil, err
	}
	if statusDict == nil {
		statusDict = make(map[string]interface{})
	}
	return EntryDict, statusDict, nil
}
