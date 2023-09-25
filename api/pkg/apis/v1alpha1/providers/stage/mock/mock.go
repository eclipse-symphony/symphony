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

package mock

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

type MockStageProviderConfig struct {
	ID string `json:"id"`
}
type MockStageProvider struct {
	Config  MockStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *MockStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()

	mockConfig, err := toMockStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	return nil
}
func (s *MockStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toMockStageProviderConfig(config providers.IProviderConfig) (MockStageProviderConfig, error) {
	ret := MockStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MockStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MockStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockStageProviderConfigFromMap(properties map[string]string) (MockStageProviderConfig, error) {
	ret := MockStageProviderConfig{}
	ret.ID = properties["id"]
	return ret, nil
}
func (i *MockStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	fmt.Printf("MOCK STAGE PROVIDER IS BUSY PROCESSING INPUTS: %v\n", inputs)
	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	if v, ok := inputs["foo"]; ok {
		if v == nil || v == "" {
			outputs["foo"] = 1
		} else {
			val, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			if err == nil {
				outputs["foo"] = val + 1
			}
		}
	}
	fmt.Printf("MOCK STAGE PROVIDER IS DONE PROCESSING WITH OUTPUTS: %v\n", outputs)
	return outputs, false, nil
}
