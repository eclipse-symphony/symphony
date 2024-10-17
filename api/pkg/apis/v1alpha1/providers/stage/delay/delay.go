/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package delay

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	loggerName   = "providers.stage.delay"
	providerName = "P (Delay Stage)"
	delay        = "delay"
)

var (
	msLock                   sync.Mutex
	mLog                     = logger.NewLogger(loggerName)
	providerOperationMetrics *metrics.Metrics
	once                     sync.Once
)

type DelayStageProviderConfig struct {
	ID string `json:"id"`
}
type DelayStageProvider struct {
	Config  DelayStageProviderConfig
	Context *contexts.ManagerContext
}

func (m *DelayStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Delay Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	msLock.Lock()
	defer msLock.Unlock()

	var mockConfig DelayStageProviderConfig
	mockConfig, err = toMockStageProviderConfig(config)
	if err != nil {
		return err
	}
	m.Config = mockConfig
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				mLog.ErrorfCtx(ctx, "  P (Delay Stage): failed to create metrics: %+v", err)
			}
		}
	})
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
	ctx, span := observability.StartSpan("[Stage] Delay provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	mLog.InfoCtx(ctx, "  P (Delay Stage) process started")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		delay,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)

	outputs := make(map[string]interface{})
	outputs[v1alpha2.StatusOutput] = v1alpha2.OK

	if v, ok := inputs["delay"]; ok {
		switch vs := v.(type) {
		case string:
			var duration time.Duration
			duration, err = time.ParseDuration(vs)
			if err != nil {
				var vi int
				if vi, err = strconv.Atoi(vs); err == nil {
					duration = time.Duration(vi) * time.Second
				} else {
					outputs[v1alpha2.StatusOutput] = v1alpha2.InternalError
					outputs[v1alpha2.ErrorOutput] = fmt.Sprintf("Failed to parse delay duration: %s", err.Error())
					mLog.ErrorfCtx(ctx, "  P (Delay Stage) process failed: %+v", err)
					providerOperationMetrics.ProviderOperationErrors(
						delay,
						functionName,
						metrics.ProcessOperation,
						metrics.ValidateOperationType,
						v1alpha2.BadConfig.String(),
					)
				}
			}
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Delay Stage): Delaying for %s", duration)
			time.Sleep(duration)
		case int:
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Delay Stage): Delaying for %d seconds", vs)
			time.Sleep(time.Duration(vs) * time.Second)
		case int32:
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Delay Stage): Delaying for %d seconds", vs)
			time.Sleep(time.Duration(vs) * time.Second)
		case int64:
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Delay Stage): Delaying for %d seconds", vs)
			time.Sleep(time.Duration(vs) * time.Second)
		}
	}

	mLog.InfoCtx(ctx, "  P (Delay Stage) process completed")
	return outputs, false, nil
}
