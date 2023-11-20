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

package delay

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type DelayStageProviderConfig struct {
	ID string `json:"id"`
}
type DelayStageProvider struct {
	Config  DelayStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *DelayStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()

	mockConfig, err := toMockStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	return nil
}
func (s *DelayStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toMockStageProviderConfig(config providers.IProviderConfig) (DelayStageProviderConfig, error) {
	ret := DelayStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *DelayStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MockStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockStageProviderConfigFromMap(properties map[string]string) (DelayStageProviderConfig, error) {
	ret := DelayStageProviderConfig{}
	ret.ID = properties["id"]
	return ret, nil
}
func (i *DelayStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	_, span := observability.StartSpan("[Stage] Delay provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, err)

	outputs := make(map[string]interface{})
	outputs[v1alpha2.StatusOutput] = v1alpha2.OK

	if v, ok := inputs["delay"]; ok {
		switch vs := v.(type) {
		case string:
			duration, err := time.ParseDuration(vs)
			if err != nil {
				if vi, err := strconv.Atoi(vs); err == nil {
					duration = time.Duration(vi) * time.Second
				} else {
					outputs[v1alpha2.StatusOutput] = v1alpha2.InternalError
					outputs[v1alpha2.ErrorOutput] = fmt.Sprintf("Failed to parse delay duration: %s", err.Error())
				}
			}
			time.Sleep(duration)
		case int:
			time.Sleep(time.Duration(vs) * time.Second)
		case int32:
			time.Sleep(time.Duration(vs) * time.Second)
		case int64:
			time.Sleep(time.Duration(vs) * time.Second)
		}
	}

	return outputs, false, nil
}
