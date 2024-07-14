/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var mLog = logger.NewLogger("coa.runtime")

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
	ctx, span := observability.StartSpan("[Stage] Mock Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	mLog.InfofCtx(ctx, "\n\n====================================================\n")
	mLog.InfofCtx(ctx, "MOCK STAGE PROVIDER IS PROCESSING INPUTS:\n")
	for k, v := range inputs {
		mLog.InfofCtx(ctx, "%v: \t%v\n", k, v)
	}
	mLog.InfofCtx(ctx, "----------------------------------------\n")
	mLog.InfofCtx(ctx, "TIME (UTC)  : %s\n", time.Now().UTC().Format(time.RFC3339))
	mLog.InfofCtx(ctx, "TIME (Local): %s\n", time.Now().Local().Format(time.RFC3339))
	mLog.InfofCtx(ctx, "----------------------------------------\n")
	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	if v, ok := inputs["foo"]; ok {
		if v == nil || v == "" {
			outputs["foo"] = 1
		} else {
			var val int64
			val, err = strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
			if err == nil {
				outputs["foo"] = val + 1
			}
		}
	}
	mLog.InfofCtx(ctx, "MOCK STAGE PROVIDER IS DONE PROCESSING WITH OUTPUTS:\n")
	for k, v := range outputs {
		mLog.InfofCtx(ctx, "%v: \t%v\n", k, v)
	}
	mLog.InfofCtx(ctx, "====================================================\n\n\n")
	return outputs, false, nil
}
