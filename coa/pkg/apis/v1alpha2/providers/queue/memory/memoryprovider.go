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

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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
		return errors.New("expected MemoryQueueProviderConfig")
	}
	s.Config = stateConfig
	s.Data = make(map[string][]interface{})
	return nil
}

// fake
func (s *MemoryQueueProvider) RemoveFromQueue(queue string, messageID string) error {
	return nil
}
func (s *MemoryQueueProvider) Enqueue(queue string, data interface{}) (string, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		s.Data[queue] = make([]interface{}, 0)
	}
	s.Data[queue] = append(s.Data[queue], data)
	return "key", nil
}
func (s *MemoryQueueProvider) Dequeue(queue string) (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		return nil, errors.New("queue not found")
	}
	if len(s.Data[queue]) == 0 {
		return nil, errors.New("queue is empty")
	}
	// ret := s.Data[queue][len(s.Data[queue])-1]
	// s.Data[queue] = s.Data[queue][:len(s.Data[queue])-1]
	ret := s.Data[queue][0]
	s.Data[queue] = s.Data[queue][1:]
	return ret, nil
}

func (s *MemoryQueueProvider) Peek(queue string) (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		return nil, errors.New("queue not found")
	}
	if len(s.Data[queue]) == 0 {
		return nil, errors.New("queue is empty")
	}
	//return s.Data[queue][len(s.Data[queue])-1], nil
	return s.Data[queue][0], nil
}

func (s *MemoryQueueProvider) Size(queue string) int {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		return 0
	}
	return len(s.Data[queue])
}
