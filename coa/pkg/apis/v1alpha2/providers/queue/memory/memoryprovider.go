/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memoryqueue

import (
	"context"
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

// not implemented
func (s *MemoryQueueProvider) QueryByPaging(ctx context.Context, queueName string, start string, size int) ([][]byte, string, error) {
	// Implement the logic to retrieve items from the queue based on the provided parameters.
	// For now, returning an empty result and a not implemented error.
	return [][]byte{}, "", errors.New("functionality not implemented yet")
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

// not implemented
func (s *MemoryQueueProvider) RemoveFromQueue(ctx context.Context, queue string, messageID string) error {
	return errors.New("functionality not implemented yet")
}

func (s *MemoryQueueProvider) Enqueue(ctx context.Context, queue string, data interface{}) (string, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		s.Data[queue] = make([]interface{}, 0)
	}
	s.Data[queue] = append(s.Data[queue], data)
	return "key", nil
}

func (s *MemoryQueueProvider) Dequeue(ctx context.Context, queue string) (interface{}, error) {
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

func (s *MemoryQueueProvider) Peek(ctx context.Context, queue string) (interface{}, error) {
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

func (s *MemoryQueueProvider) Size(ctx context.Context, queue string) int {
	mLock.Lock()
	defer mLock.Unlock()
	if _, ok := s.Data[queue]; !ok {
		return 0
	}
	return len(s.Data[queue])
}

func (s *MemoryQueueProvider) DeleteQueue(ctx context.Context, queue string) error {
	mLock.Lock()
	defer mLock.Unlock()
	delete(s.Data, queue)
	return nil
}
