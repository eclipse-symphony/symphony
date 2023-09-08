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

package remote

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var rmtLock sync.Mutex
var log = logger.NewLogger("coa.runtime")

type RemoteStageProviderConfig struct {
}
type RemoteStageProvider struct {
	Config        RemoteStageProviderConfig
	Context       *contexts.ManagerContext
	OutputContext map[string]map[string]interface{}
}

func (m *RemoteStageProvider) Init(config providers.IProviderConfig) error {
	rmtLock.Lock()
	defer rmtLock.Unlock()

	mockConfig, err := toRemoteStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	return nil
}
func toRemoteStageProviderConfig(config providers.IProviderConfig) (RemoteStageProviderConfig, error) {
	ret := RemoteStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *RemoteStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MockStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockStageProviderConfigFromMap(properties map[string]string) (RemoteStageProviderConfig, error) {
	ret := RemoteStageProviderConfig{}
	return ret, nil
}
func (i *RemoteStageProvider) SetOutputsContext(outputs map[string]map[string]interface{}) {
	i.OutputContext = outputs
}
func (i *RemoteStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	_, span := observability.StartSpan("Remote Process Provider", context.Background(), &map[string]string{
		"method": "Process",
	})
	log.Info("  P (Remote Processor): Process")

	outputs := make(map[string]interface{})

	v, ok := inputs["__site"]

	if !ok {
		err := v1alpha2.NewCOAError(nil, "no site found in inputs", v1alpha2.BadRequest)
		log.Errorf("  P (Remote Processor): %v", err)
		observ_utils.CloseSpanWithError(span, err)
		return nil, false, err
	}

	err := mgrContext.Publish("remote", v1alpha2.Event{
		Metadata: map[string]string{
			"site":       v.(string),
			"objectType": "task",
			"origin":     mgrContext.SiteInfo.SiteId,
		},
		Body: v1alpha2.JobData{
			Id:     "",
			Action: "RUN",
			Body: v1alpha2.InputOutputData{
				Inputs:  inputs,
				Outputs: i.OutputContext,
			},
		},
	})
	if err != nil {
		log.Errorf("  P (Remote Processor): publish failed - %v", err)
		observ_utils.CloseSpanWithError(span, err)
		return nil, false, err
	}

	return outputs, true, nil
}
