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

package memorystack

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

type MemoryStackProviderConfig struct {
	Name string `json:"name"`
}

func MemoryStackProviderConfigFromMap(properties map[string]string) (MemoryStackProviderConfig, error) {
	ret := MemoryStackProviderConfig{}
	if v, ok := properties["name"]; ok {
		ret.Name = utils.ParseProperty(v)
	}
	return ret, nil
}

type MemoryStackProvider struct {
	Config  MemoryStackProviderConfig
	Data    []interface{}
	Context *contexts.ManagerContext
}

func (s *MemoryStackProvider) ID() string {
	return s.Config.Name
}

func (s *MemoryStackProvider) SetContext(ctx *contexts.ManagerContext) error {
	s.Context = ctx
	return nil
}

func (i *MemoryStackProvider) InitWithMap(properties map[string]string) error {
	config, err := MemoryStackProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}

func toMemoryStackProviderConfig(config providers.IProviderConfig) (MemoryStackProviderConfig, error) {
	ret := MemoryStackProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	//ret.Name = providers.LoadEnv(ret.Name)
	return ret, err
}

func (s *MemoryStackProvider) Init(config providers.IProviderConfig) error {
	// parameter checks
	stateConfig, err := toMemoryStackProviderConfig(config)
	if err != nil {
		return errors.New("expected MemoryStackProviderConfig")
	}
	s.Config = stateConfig
	s.Data = make([]interface{}, 0)
	return nil
}

func (s *MemoryStackProvider) Push(data interface{}) error {
	mLock.Lock()
	defer mLock.Unlock()
	s.Data = append(s.Data, data)
	return nil
}
func (s *MemoryStackProvider) Pop() (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if len(s.Data) == 0 {
		return nil, errors.New("stack is empty")
	}
	ret := s.Data[len(s.Data)-1]
	s.Data = s.Data[:len(s.Data)-1]
	return ret, nil
}

func (s *MemoryStackProvider) Peek() (interface{}, error) {
	mLock.Lock()
	defer mLock.Unlock()
	if len(s.Data) == 0 {
		return nil, errors.New("stack is empty")
	}
	return s.Data[len(s.Data)-1], nil
}

func (s *MemoryStackProvider) Size() int {
	mLock.Lock()
	defer mLock.Unlock()
	return len(s.Data)
}
