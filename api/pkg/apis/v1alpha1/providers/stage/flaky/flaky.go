/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package flaky

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var fLog = logger.NewLogger("coa.runtime")

// callCounts tracks Process invocations per provider ID across instances,
// since the stage manager creates a new provider instance per trigger.
var (
	callCounts     = make(map[string]int)
	callCountsLock sync.Mutex
)

// GetCallCount returns how many times Process was called for the given ID.
func GetCallCount(id string) int {
	callCountsLock.Lock()
	defer callCountsLock.Unlock()
	return callCounts[id]
}

// ResetCallCount clears the invocation counter for the given ID.
func ResetCallCount(id string) {
	callCountsLock.Lock()
	defer callCountsLock.Unlock()
	delete(callCounts, id)
}

type FlakyStageProviderConfig struct {
	ID string `json:"id"`
	// FailTimes is the number of initial Process calls that return an error
	FailTimes int `json:"failTimes"`
}

type FlakyStageProvider struct {
	Config  FlakyStageProviderConfig
	Context *contexts.ManagerContext
}

func (f *FlakyStageProvider) Init(config providers.IProviderConfig) error {
	flakyConfig, err := toFlakyStageProviderConfig(config)
	if err != nil {
		return err
	}
	f.Config = flakyConfig
	return nil
}

func (f *FlakyStageProvider) SetContext(ctx *contexts.ManagerContext) {
	f.Context = ctx
}

func toFlakyStageProviderConfig(config providers.IProviderConfig) (FlakyStageProviderConfig, error) {
	ret := FlakyStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}

func (f *FlakyStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	callCountsLock.Lock()
	callCounts[f.Config.ID]++
	calls := callCounts[f.Config.ID]
	callCountsLock.Unlock()

	if calls <= f.Config.FailTimes {
		fLog.InfofCtx(ctx, "  P (Flaky Stage): failing attempt %d/%d for %s", calls, f.Config.FailTimes, f.Config.ID)
		return nil, false, fmt.Errorf("flaky stage provider %s failed on attempt %d", f.Config.ID, calls)
	}

	outputs := make(map[string]interface{})
	for k, v := range inputs {
		outputs[k] = v
	}
	return outputs, false, nil
}
