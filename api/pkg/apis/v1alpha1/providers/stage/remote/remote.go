/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
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
func (s *RemoteStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
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
	ctx, span := observability.StartSpan("[Stage] Remote Process Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, "  P (Remote Processor): Process")

	outputs := make(map[string]interface{})

	v, ok := inputs["__site"]

	if !ok {
		err = v1alpha2.NewCOAError(nil, "no site found in inputs", v1alpha2.BadRequest)
		log.ErrorfCtx(ctx, "  P (Remote Processor): %v", err)
		return nil, false, err
	}

	siteString, ok := v.(string)
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("site name is not a valid string: %v", v), v1alpha2.BadRequest)
		log.ErrorfCtx(ctx, "  P (Remote Processor): %v", err)
		return nil, false, err
	}

	err = mgrContext.Publish("remote", v1alpha2.Event{
		Metadata: map[string]string{
			"site":       siteString,
			"objectType": "task",
			"origin":     mgrContext.SiteInfo.SiteId,
		},
		Body: v1alpha2.JobData{
			Id:     "",
			Action: v1alpha2.JobRun,
			Body: v1alpha2.InputOutputData{
				Inputs:  inputs,
				Outputs: i.OutputContext,
			},
		},
		Context: ctx,
	})
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Remote Processor): publish failed - %v", err)
		return nil, false, err
	}

	return outputs, true, nil
}
