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

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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
	Data    map[string]map[string]interface{}
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
	s.Data = make(map[string]map[string]interface{}, 0)
	return nil
}

// MemoryStateProvider Upsert will store object in the memory map
// It bumps generation automatically when "spec" is changed on objects with metadata field
func (s *MemoryStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Upsert",
	})

	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := "default"
	if n, ok := entry.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.DebugfCtx(ctx, "  P (Memory State): upsert states %s in namespace %s", entry.Value.ID, namespace)

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

	// Get existing entry
	var oldEntryMap map[string]interface{} = nil
	if s.Data[namespace] != nil && s.Data[namespace][entry.Value.ID] != nil {
		existingEntry, ok := s.Data[namespace][entry.Value.ID].(states.StateEntry)
		if ok {
			oldEntryMap, ok = existingEntry.Body.(map[string]interface{})
			if !ok {
				sLog.ErrorfCtx(ctx, "  P (Memory State): failed to convert old state to map[string]interface{} for %s", entry.Value.ID)
				oldEntryMap = nil
			}
		} else {
			sLog.ErrorfCtx(ctx, "  P (Memory State): failed to convert old state to states.StateEntry for %s: %+v", entry.Value.ID, err)
		}
	} else {
		sLog.InfofCtx(ctx, "  P (Memory State): failed to get old state for %s: %+v", entry.Value.ID, err)
	}

	// Get new entry
	var newEntryMap map[string]interface{}
	jBody, _ := json.Marshal(entry.Value.Body)
	json.Unmarshal(jBody, &newEntryMap)

	if entry.Options.UpdateStatusOnly {
		if oldEntryMap == nil {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found or is invalid", entry.Value.ID), v1alpha2.NotFound)
			sLog.ErrorfCtx(ctx, "  P (Memory State): failed to upsert %s state: %+v", entry.Value.ID, err)
			return "", err
		}
		if oldEntryMap["status"] == nil {
			oldEntryMap["status"] = make(map[string]interface{})
		}
		statusMap, ok := newEntryMap["status"].(map[string]interface{})
		if !ok {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' doesn't has a valid status", entry.Value.ID), v1alpha2.InternalError)
			sLog.ErrorfCtx(ctx, "  P (Memory State): failed to upsert %s state: %+v", entry.Value.ID, err)
			return "", err
		}
		for k, v := range statusMap {
			oldEntryMap["status"].(map[string]interface{})[k] = v
		}
		entry.Value.Body = oldEntryMap
	} else {
		if newEntryMap != nil && newEntryMap["metadata"] != nil {
			s.BumpGeneration(ctx, newEntryMap, oldEntryMap)
			entry.Value.Body = newEntryMap
		}
	}

	s.Data[namespace][entry.Value.ID] = entry.Value

	return entry.Value.ID, nil
}
func (s *MemoryStateProvider) List(ctx context.Context, request states.ListRequest) ([]states.StateEntry, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ctx, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "List",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	var entities []states.StateEntry
	namespace := ""
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.DebugfCtx(ctx, "  P (Memory State): list states in namespace %s", namespace)
	for nKey, nList := range s.Data {
		// If namespace is not specified, get entry for all namespaces
		if namespace == "" || namespace == nKey {
			if nList != nil {
				for _, entry := range nList {
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
							sLog.ErrorfCtx(ctx, "  P (Memory State): failed to list states: %+v", err)
							return entities, "", err
						}
						entities = append(entities, copy)
					} else {
						err = v1alpha2.NewCOAError(nil, "found invalid state entry", v1alpha2.InternalError)
						sLog.ErrorfCtx(ctx, "  P (Memory State): failed to list states: %+v", err)
						return entities, "", err
					}
				}
			} else {
				err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to convert entry list to map[string]interface{} for namespace %s", namespace), v1alpha2.InternalError)
				sLog.ErrorfCtx(ctx, "  P (Memory State): failed to list states: %+v", err)
				return entities, "", err
			}
		}
	}

	return entities, "", nil
}

func (s *MemoryStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := "default"
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}
	sLog.DebugfCtx(ctx, "  P (Memory State): delete state %s in namespace %s", request.ID, namespace)

	if _, ok := s.Data[namespace]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.ErrorfCtx(ctx, "  P (Memory State): failed to delete %s: %+v", request.ID, err)
		return err
	}
	if s.Data[namespace] == nil || s.Data[namespace][request.ID] == nil {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)
		sLog.ErrorfCtx(ctx, "  P (Memory State): failed to delete %s: %+v", request.ID, err)
		return err
	}
	delete(s.Data[namespace], request.ID)

	return nil
}

func (s *MemoryStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ctx, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	namespace := "default"
	if n, ok := request.Metadata["namespace"]; ok {
		if nstring, ok := n.(string); ok && nstring != "" {
			namespace = nstring
		}
	}

	sLog.DebugfCtx(ctx, "  P (Memory State): get state %s in namespace %s", request.ID, namespace)

	if _, ok := s.Data[namespace]; !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.ErrorfCtx(ctx, "  P (Memory State): failed to get %s state: %+v", request.ID, err)
		return states.StateEntry{}, err
	}

	if s.Data[namespace] == nil || s.Data[namespace][request.ID] == nil {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found in namespace %s", request.ID, namespace), v1alpha2.NotFound)
		sLog.ErrorfCtx(ctx, "  P (Memory State): failed to get %s state: %+v", request.ID, err)
		return states.StateEntry{}, err
	}
	vE, ok := s.Data[namespace][request.ID].(states.StateEntry)
	if ok {
		var copy states.StateEntry
		copy, err = s.ReturnDeepCopy(vE)
		if err != nil {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to create a deep copy of entry '%s'", request.ID), v1alpha2.InternalError)
			sLog.ErrorfCtx(ctx, "  P (Memory State): failed to get %s state: %+v", request.ID, err)
			return states.StateEntry{}, err
		}
		return copy, nil
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", request.ID), v1alpha2.InternalError)
	sLog.ErrorfCtx(ctx, "  P (Memory State): failed to get %s state: %+v", request.ID, err)
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

func (a *MemoryStateProvider) BumpGeneration(ctx context.Context, new map[string]interface{}, old map[string]interface{}) {
	var newGeneration int64 = 0
	if old != nil {
		// If old state exists, bump generation if spec is changed
		var oldObjectMeta model.ObjectMeta
		if old["metadata"] == nil {
			oldObjectMeta = model.ObjectMeta{}
		} else {
			oldObjectMeta = old["metadata"].(model.ObjectMeta)
		}
		oldSpecPtr := old["spec"]
		newSpecPtr := new["spec"]
		if oldSpecPtr != nil && newSpecPtr != nil {
			oldSpecString, _ := json.Marshal(oldSpecPtr)
			newSpecString, _ := json.Marshal(newSpecPtr)
			if string(oldSpecString) != string(newSpecString) {
				newGeneration = oldObjectMeta.Generation + 1
			} else {
				newGeneration = oldObjectMeta.Generation
			}
		} else if oldSpecPtr == nil && newSpecPtr == nil {
			newGeneration = oldObjectMeta.Generation
		} else {
			newGeneration = oldObjectMeta.Generation + 1
		}
	}

	// Store metadata field as ObjectMeta in the memory state
	var newObjectMeta model.ObjectMeta
	jData, _ := json.Marshal(new["metadata"])
	json.Unmarshal(jData, &newObjectMeta)
	newObjectMeta.Generation = newGeneration
	sLog.DebugfCtx(ctx, "  P (Memory State): new generation %d for object %s", newGeneration, newObjectMeta.Name)
	new["metadata"] = newObjectMeta
}
