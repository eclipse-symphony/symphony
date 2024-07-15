/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package counter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.counter"
	providerName = "P (Counter Stage)"
	counter      = "counter"
)

var (
	msLock                   sync.Mutex
	mLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type CounterStageProviderConfig struct {
	ID string `json:"id"`
}
type CounterStageProvider struct {
	Config  CounterStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *CounterStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Counter Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	msLock.Lock()
	defer msLock.Unlock()

	var mockConfig CounterStageProviderConfig
	mockConfig, err = toMockStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				mLog.ErrorfCtx(ctx, "  P (Counter Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
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
	ctx, span := observability.StartSpan("[Stage] Counter provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	mLog.InfofCtx(ctx, "  P (Counter Stage) process started")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()

	outputs := make(map[string]interface{})
	selfState := make(map[string]interface{})
	if state, ok := inputs["__state"]; ok {
		selfState, ok = state.(map[string]interface{})
		if !ok {
			err = v1alpha2.NewCOAError(nil, "input state is not a valid map[string]interface{}", v1alpha2.BadRequest)
			mLog.ErrorfCtx(ctx, "[Stage] Counter provider failed: %+v", err)
			providerOperationMetrics.ProviderOperationErrors(
				counter,
				functionName,
				metrics.ProcessOperation,
				metrics.ValidateRuleOperation,
				v1alpha2.BadConfig.String(),
			)
			return outputs, false, err
		}
	}

	for k, v := range inputs {
		if k != "__state" {
			if !strings.HasSuffix(k, ".init") {
				var iv int64
				if iv, err = getNumber(v); err == nil {
					if s, ok := selfState[k]; ok {
						var sv int64
						if sv, err = getNumber(s); err == nil {
							selfState[k] = sv + iv
							outputs[k] = sv + iv
						}
					} else {
						if vs, ok := inputs[k+".init"]; ok {
							var ivs int64
							if ivs, err = getNumber(vs); err == nil {
								selfState[k] = ivs + iv
								outputs[k] = ivs + iv
							}
						} else {
							selfState[k] = iv
							outputs[k] = iv
						}
					}
				}
			}
		}
	}

	outputs["__state"] = selfState
	mLog.InfofCtx(ctx, "  P (Counter Stage) process completed")
	providerOperationMetrics.ProviderOperationLatency(
		processTime,
		counter,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)
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
	return 0, v1alpha2.NewCOAError(nil, fmt.Sprintf("cannot convert %v to number", val), v1alpha2.BadRequest)
}
