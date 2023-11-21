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
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

type MockTargetProviderConfig struct {
	ID string `json:"id"`
}
type MockTargetProvider struct {
	Config  MockTargetProviderConfig
	Context *contexts.ManagerContext
}

var cache map[string][]model.ComponentSpec
var mLock sync.Mutex

func (m *MockTargetProvider) Init(config providers.IProviderConfig) error {
	_, span := observability.StartSpan(
		"Mock Target Provider",
		context.Background(),
		&map[string]string{
			"method": "Init",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	mLock.Lock()
	defer mLock.Unlock()

	mockConfig, err := toMockTargetProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	if cache == nil {
		cache = make(map[string][]model.ComponentSpec)
	}
	return nil
}
func (s *MockTargetProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func toMockTargetProviderConfig(config providers.IProviderConfig) (MockTargetProviderConfig, error) {
	ret := MockTargetProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MockTargetProvider) InitWithMap(properties map[string]string) error {
	config, err := MockTargetProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MockTargetProviderConfigFromMap(properties map[string]string) (MockTargetProviderConfig, error) {
	ret := MockTargetProviderConfig{}
	ret.ID = properties["id"]
	return ret, nil
}
func (m *MockTargetProvider) Get(ctx context.Context, deployment model.DeploymentSpec, references []model.ComponentStep) ([]model.ComponentSpec, error) {
	mLock.Lock()
	defer mLock.Unlock()

	_, span := observability.StartSpan(
		"Mock Target Provider",
		ctx,
		&map[string]string{
			"method": "Get",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	return cache[m.Config.ID], nil
}
func (m *MockTargetProvider) Apply(ctx context.Context, deployment model.DeploymentSpec, step model.DeploymentStep, isDryRun bool) (map[string]model.ComponentResultSpec, error) {
	_, span := observability.StartSpan(
		"Mock Target Provider",
		ctx,
		&map[string]string{
			"method": "Apply",
		},
	)
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	mLock.Lock()
	defer mLock.Unlock()
	if cache[m.Config.ID] == nil {
		cache[m.Config.ID] = make([]model.ComponentSpec, 0)
	}
	for _, c := range step.Components {
		found := false
		for i, _ := range cache[m.Config.ID] {
			if cache[m.Config.ID][i].Name == c.Component.Name {
				found = true
				if c.Action == "delete" {
					cache[m.Config.ID] = append(cache[m.Config.ID][:i], cache[m.Config.ID][i+1:]...)
				}
				break
			}
		}
		if !found {
			cache[m.Config.ID] = append(cache[m.Config.ID], c.Component)
		}
	}
	ret := make(map[string]model.ComponentResultSpec)
	for _, c := range cache[m.Config.ID] {
		ret[c.Name] = model.ComponentResultSpec{
			Status:  v1alpha2.OK,
			Message: "",
		}
	}
	return ret, nil
}
func (m *MockTargetProvider) GetValidationRule(ctx context.Context) model.ValidationRule {
	return model.ValidationRule{}
}
