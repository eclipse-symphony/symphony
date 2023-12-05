/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memoryqueue

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/azure/symphony/coa/pkg/logger"
)

var mLog = logger.NewLogger("coa.runtime")
var mLock sync.Mutex

type MemoryQueueProviderConfig struct {
	Name string `json:"name"`
}

func MemoryQueueProviderConfigFromMap(properties map[string]string) (MemoryQueueProviderConfig, error) {
	ret := MemoryQueueProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MemoryQueueProvider struct {
	Config  MemoryQueueProviderConfig
	Data    map[string][]interface{}
	Context *contexts.ManagerContext
}

func (s *MemoryQueueProvider) ID() string {
	return s.Config.Name
}

func (s *MemoryQueueProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func (i *MemoryQueueProvider) InitWithMap(properties map[string]string) error {
	config, err := MemoryQueueProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func toMemoryQueueProviderConfig(config providers.IProviderConfig) (MemoryQueueProviderConfig, error) {
	ret := MemoryQueueProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (s *MemoryQueueProvider) Init(config providers.IProviderConfig) error {
	// parameter checks
	stateConfig, err := toMemoryQueueProviderConfig(config)
	if err != nil {
		return errors.New("expected MemoryStackProviderConfig")
	}
	s.Config = stateConfig
	s.Data = make(map[string][]interface{})
	return nil
}

func (s *MemoryQueueProvider) Enqueue(stack string, data interface{}) error {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[stack]; !ok {
		s.Data[stack] = make([]interface{}, 0)
	}
	s.Data[stack] = append(s.Data[stack], data)
	return nil
}
func (s *MemoryQueueProvider) Dequeue(stack string) (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[stack]; !ok {
		return nil, errors.New("stack not found")
	}
	if len(s.Data[stack]) == 0 {
		return nil, errors.New("stack is empty")
	}
	// ret := s.Data[stack][len(s.Data[stack])-1]
	// s.Data[stack] = s.Data[stack][:len(s.Data[stack])-1]
	ret := s.Data[stack][0]
	s.Data[stack] = s.Data[stack][1:]
	return ret, nil
}

func (s *MemoryQueueProvider) Peek(stack string) (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[stack]; !ok {
		return nil, errors.New("stack not found")
	}
	if len(s.Data[stack]) == 0 {
		return nil, errors.New("stack is empty")
	}
	//return s.Data[stack][len(s.Data[stack])-1], nil
	return s.Data[stack][0], nil
}

func (s *MemoryQueueProvider) Size(stack string) int {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[stack]; !ok {
		return 0
	}
	return len(s.Data[stack])
}
