/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package memorystate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	contexts "github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	providers "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	states "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/azure/symphony/coa/pkg/logger"
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
		return errors.New("expected MemoryStateProviderConfig")
	}
	s.Config = stateConfig
	s.Data = make(map[string]interface{}, 0)
	return nil
}

func (s *MemoryStateProvider) Upsert(ctx context.Context, entry states.UpsertRequest) (string, error) {
	tag := "1"
	if entry.Value.ETag != "" {
		if v, err := strconv.ParseInt(entry.Value.ETag, 10, 64); err == nil {
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
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "List",
	})
	sLog.Debug("  P (Memory State): list states")

	var entities []states.StateEntry
	for _, v := range s.Data {
		vE, ok := v.(states.StateEntry)
		if ok {
			if request.Filter != "" {

			}
			entities = append(entities, vE)
		} else {
			err := v1alpha2.NewCOAError(nil, "found invalid state entry", v1alpha2.InternalError)
			observ_utils.CloseSpanWithError(span, err)
			return entities, "", err
		}
	}

	observ_utils.CloseSpanWithError(span, nil)
	return entities, "", nil
}

func (s *MemoryStateProvider) Delete(ctx context.Context, request states.DeleteRequest) error {
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Delete",
	})
	sLog.Debug("  P (Memory State): delete state")

	if _, ok := s.Data[request.ID]; !ok {
		err := v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)
		observ_utils.CloseSpanWithError(span, err)
		return err
	}
	delete(s.Data, request.ID)

	observ_utils.CloseSpanWithError(span, nil)
	return nil
}

func (s *MemoryStateProvider) Get(ctx context.Context, request states.GetRequest) (states.StateEntry, error) {
	_, span := observability.StartSpan("Memory State Provider", ctx, &map[string]string{
		"method": "Get",
	})
	sLog.Debug("  P (Memory State): get state")

	if v, ok := s.Data[request.ID]; ok {
		vE, ok := v.(states.StateEntry)
		if ok {
			observ_utils.CloseSpanWithError(span, nil)
			return vE, nil
		} else {
			err := v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not a valid state entry", request.ID), v1alpha2.InternalError)
			observ_utils.CloseSpanWithError(span, err)
			return states.StateEntry{}, err
		}
	}
	err := v1alpha2.NewCOAError(nil, fmt.Sprintf("entry '%s' is not found", request.ID), v1alpha2.NotFound)

	observ_utils.CloseSpanWithError(span, err)
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
