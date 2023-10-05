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

package counter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type CounterStageProviderConfig struct {
	ID string `json:"id"`
}
type CounterStageProvider struct {
	Config  CounterStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *CounterStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()

	mockConfig, err := toMockStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	return nil
}
func (s *CounterStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toMockStageProviderConfig(config providers.IProviderConfig) (CounterStageProviderConfig, error) {
	ret := CounterStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *CounterStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MockStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockStageProviderConfigFromMap(properties map[string]string) (CounterStageProviderConfig, error) {
	ret := CounterStageProviderConfig{}
	ret.ID = properties["id"]
	return ret, nil
}
func (i *CounterStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {

	outputs := make(map[string]interface{})
	selfState := make(map[string]interface{})
	if state, ok := inputs["__state"]; ok {
		selfState = state.(map[string]interface{})
	}

	for k, v := range inputs {
		if k != "__state" {
			if v, err := getNumber(v); err == nil {
				if s, ok := selfState[k]; ok {
					if sv, err := getNumber(s); err == nil {
						selfState[k] = sv + v
						outputs[k] = sv + v
					}
				} else {
					selfState[k] = v
					outputs[k] = v
				}
			}
		}
	}

	outputs["__state"] = selfState
	return outputs, false, nil
}

func getNumber(val interface{}) (int64, error) {
	if v, ok := val.(int64); ok {
		return v, nil
	}
	if v, ok := val.(int); ok {
		return int64(v), nil
	}
	if v, ok := val.(float64); ok {
		return int64(v), nil
	}
	if v, ok := val.(float32); ok {
		return int64(v), nil
	}
	if v, ok := val.(string); ok {
		if v, err := strconv.ParseInt(v, 10, 64); err == nil {
			return v, nil
		}
	}
	return 0, fmt.Errorf("cannot convert %v to number", val)
}
