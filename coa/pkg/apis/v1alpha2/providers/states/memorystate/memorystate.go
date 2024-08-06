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
	mu      sync.RWMutex
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
	s.mu.Lock()
	defer s.mu.Unlock()

	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	namespace := "default"
	if n, ok := entry.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.Debugf("  P (Memory State): upsert states %s in namespace %s, traceId: %s", entry.Value.ID, namespace, span.SpanContext().TraceID().String())

	if _, ok := s.Data[namespace]; !ok {
		s.Data[namespace] = map[string]interface{}{}
	}

	tag := "1"
	if entry.Value.ETag != "" {
		var v int64
		if v, err = strconv.ParseInt(entry.Value.ETag, 10, 64); err == nil {
			tag = strconv.FormatInt(v+1, 10)
		}
	}
	entry.Value.ETag = tag

	list, ok := s.Data[namespace].(map[string]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to convert entry list to map[string]interface{} for namespace %s", namespace), v1alpha2.InternalError)
		sLog.Errorf("  P (Memory State): failed to upsert %s states: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
		return "", err
	}
	if entry.Options.UpdateStatusOnly {
		existing, ok := list[entry.Value.ID]
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", entry.Value.ID), v1alpha2.NotFound)
			sLog.Errorf("  P (Memory State): failed to upsert %s state: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
			return "", err
		}
		existingEntry, ok := existing.(states.StateEntry)
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", entry.Value.ID), v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to upsert %s state: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
			return "", err
		}

		mapRef, ok := existingEntry.Body.(map[string]interface{})
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' doesn't has a valid body", entry.Value.ID), v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to upsert %s state: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
			return "", err
		}
		var mapType map[string]interface{}
		jBody, _ := json.Marshal(entry.Value.Body)
		json.Unmarshal(jBody, &mapType)

		if mapRef["status"] == nil {
			mapRef["status"] = make(map[string]interface{})
		}
		statusMap, ok := mapType["status"].(map[string]interface{})
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' doesn't has a valid status", entry.Value.ID), v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to upsert %s state: %+v, traceId: %s", entry.Value.ID, err, span.SpanContext().TraceID().String())
			return "", err
		}
		for k, v := range statusMap {
			mapRef["status"].(map[string]interface{})[k] = v
		}

		entry.Value.Body = mapRef
	}

	list[entry.Value.ID] = entry.Value

	return entry.Value.ID, nil
}
func (s *MemoryStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var entities []states.StateEntry
	namespace := ""
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.Debugf("  P (Memory State): list states in namespace %s, traceId: %s", namespace, span.SpanContext().TraceID().String())
	for nKey, nList := range s.Data {
		// If namespace is not specified, get entry for all namespaces
		if namespace == "" || namespace == nKey {
			if list, ok := nList.(map[string]interface{}); ok {
				for _, entry := range list {
					vE, ok := entry.(states.StateEntry)
					if ok {
						if request.FilterType != "" && request.FilterValue != "" {
							var match bool
							match, err = states.MatchFilter(vE, request.FilterType, request.FilterValue)
							if err != nil {
								return entities, "", err
							} else if !match {
								continue
							}
						}
						var copy states.StateEntry
						copy, err = s.ReturnDeepCopy(vE)
						if err != nil {
							err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to create a deep copy of entry '%s'", vE.ID), v1alpha2.InternalError)
							sLog.Errorf("  P (Memory State): failed to list states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
							return entities, "", err
						}
						entities = append(entities, copy)
					} else {
						err = v1alpha2.NewCOAError(nil, "found invalid state entry", v1alpha2.InternalError)
						sLog.Errorf("  P (Memory State): failed to list states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
						return entities, "", err
					}
				}
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to convert entry list to map[string]interface{} for namespace %s", namespace), v1alpha2.InternalError)
				sLog.Errorf("  P (Memory State): failed to list states: %+v, traceId: %s", err, span.SpanContext().TraceID().String())
				return entities, "", err
			}
		}
	}

	return entities, "", nil
}

func (s *MemoryStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	namespace := "default"
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.Debugf("  P (Memory State): delete state %s in namespace %s, traceId: %s", request.ID, namespace, span.SpanContext().TraceID().String())

	if _, ok := s.Data[namespace]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.Errorf("  P (Memory State): failed to delete %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	list, ok := s.Data[namespace].(map[string]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to convert entry list to map[string]interface{} for namespace %s", namespace), v1alpha2.InternalError)
		sLog.Errorf("  P (Memory State): failed to delete %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	if _, ok := list[request.ID]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)
		sLog.Errorf("  P (Memory State): failed to delete %s: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return err
	}
	delete(list, request.ID)

	return nil
}

func (s *MemoryStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	namespace := "default"
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}

	sLog.Debugf("  P (Memory State): get state %s in namespace %s, traceId: %s", request.ID, namespace, span.SpanContext().TraceID().String())

	if _, ok := s.Data[namespace]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	list, ok := s.Data[namespace].(map[string]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to convert entry list to map[string]interface{} for namespace %s", namespace), v1alpha2.InternalError)
		sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	entry, ok := list[request.ID]
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
		return states.StateEntry{}, err
	}
	vE, ok := entry.(states.StateEntry)
	if ok {
		var copy states.StateEntry
		copy, err = s.ReturnDeepCopy(vE)
		if err != nil {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to create a deep copy of entry '%s'", request.ID), v1alpha2.InternalError)
			sLog.Errorf("  P (Memory State): failed to get %s state: %+v, traceId: %s", request.ID, err, span.SpanContext().TraceID().String())
			return states.StateEntry{}, err
		}
		return copy, nil
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", request.ID), v1alpha2.InternalError)
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

func (a *MemoryStateProvider) ReturnDeepCopy(s states.StateEntry) (states.StateEntry, error) {
	var ret states.StateEntry
	jBody, err := json.Marshal(s)
	if err != nil {
		return states.StateEntry{}, err
	}
	err = json.Unmarshal(jBody, &ret)
	if err != nil {
		return states.StateEntry{}, err
	}
	return ret, nil
}
