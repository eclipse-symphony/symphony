/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memorystate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	providers "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var sLog = logger.NewLogger("coa.runtime")
var mLock sync.RWMutex

type MemoryStateProviderConfig struct {
	Name string `json:"name"`
}

func MemoryStateProviderConfigFromMap(properties map[string]string) (MemoryStateProviderConfig, error) {
	ret := MemoryStateProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MemoryStateProvider struct {
	Config  MemoryStateProviderConfig
	Data    map[string]interface{}
	Context *contexts.ManagerContext
}

func (s *MemoryStateProvider) ID() string {
	return s.Config.Name
}

func (s *MemoryStateProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *MemoryStateProvider) InitWithMap(properties map[string]string) error {
	config, err := MemoryStateProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func (s *MemoryStateProvider) Init(config providers.IProviderConfig) error {
	// parameter checks
	stateConfig, err := toMemoryStateProviderConfig(config)
	if err != nil {
		sLog.Errorf("  P (Memory State): failed to parse provider config %+v", err)
		return errors.New("expected MemoryStateProviderConfig")
	}
	s.Config = stateConfig
	s.Data = make(map[string]interface{}, 0)
	return nil
}

func (s *MemoryStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	mLock.Lock()
	defer mLock.Unlock()

	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})
	sLog.Debugf("  P (Memory State): upsert states %s, traceId: %s", entry.Value.ID, span.SpanContext().TraceID().String())

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	tag := "1"
	if entry.Value.ETag != "" {
		var v int64
		if v, err = strconv.ParseInt(entry.Value.ETag, 10, 64); err == nil {
			tag = strconv.FormatInt(v+1, 10)
		}
	}
	entry.Value.ETag = tag

	// This hack is to simulate k8s upsert behavior
	if _, ok := entry.Value.Body.(map[string]interface{}); ok {
		mapRef := entry.Value.Body.(map[string]interface{})
		if mapRef["status"] != nil && mapRef["spec"] == nil {

			dataRef := s.Data[entry.Value.ID]
			if dataRef != nil {
				mapRef["spec"] = dataRef.(states.StateEntry).Body.(map[string]interface{})["spec"]
			}
			entry.Value.Body = mapRef
		}
	}

	s.Data[entry.Value.ID] = entry.Value

	return entry.Value.ID, nil
}

func (s *MemoryStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	mLock.RLock()
	defer mLock.RUnlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Debugf("  P (Memory State): list states, traceId: %s", span.SpanContext().TraceID().String())

	var entities []states.StateEntry
	for _, v := range s.Data {
		vE, ok := v.(states.StateEntry)
		if ok {
			if request.Filter != "" {
				//TODO: support filters in the future
			}
			entities = append(entities, vE)
		} else {
			err = v1alpha2.NewCOAError(nil, "found invalid state entry", v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to list states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
			return entities, "", err
		}
	}

	return entities, "", nil
}

func (s *MemoryStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	mLock.Lock()
	defer mLock.Unlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Debug("  P (Memory State): delete state %s, traceId: %s", request.ID, span.SpanContext().TraceID().String())

	if _, ok := s.Data[request.ID]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)
		sLog.Errorf("  P (Memory State): failed to delete %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	delete(s.Data, request.ID)

	return nil
}

func (s *MemoryStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	mLock.RLock()
	defer mLock.RUnlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	sLog.Debug("  P (Memory State): get state %s, traceId: %s", request.ID, span.SpanContext().TraceID().String())

	if v, ok := s.Data[request.ID]; ok {
		vE, ok := v.(states.StateEntry)
		if ok {
			err = nil
			return vE, nil
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", request.ID), v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
			return states.StateEntry{}, err
		}
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)
	sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
	return states.StateEntry{}, err
}

func toMemoryStateProviderConfig(config providers.IProviderConfig) (MemoryStateProviderConfig, error) {
	ret := MemoryStateProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (a *MemoryStateProvider) Clone(config providers.IProviderConfig) (providers.IProvider, error) {
	ret := &MemoryStateProvider{}
	if config == nil {
		err := ret.Init(a.Config)
		if err != nil {
			return nil, err
		}
	} else {
		err := ret.Init(config)
		if err != nil {
			return nil, err
		}
	}
	if a.Context != nil {
		ret.Context = a.Context
	}
	return ret, nil
}
