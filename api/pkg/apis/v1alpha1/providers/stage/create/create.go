/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type CreateStageProviderConfig struct {
	WaitCount    int `json:"wait.count,omitempty"`
	WaitInterval int `json:"wait.interval,omitempty"`
}

type CreateStageProvider struct {
	Config    CreateStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient *utils.APIClient
}

const (
	RemoveAction = "remove"
	CreateAction = "create"
)

func (s *CreateStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toSymphonyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = utils.GetApiClient()
	return nil
}
func (s *CreateStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toSymphonyStageProviderConfig(config providers.IProviderConfig) (CreateStageProviderConfig, error) {
	ret := CreateStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *CreateStageProvider) InitWithMap(properties map[string]string) error {
	config, err := SymphonyStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func SymphonyStageProviderConfigFromMap(properties map[string]string) (CreateStageProviderConfig, error) {
	ret := CreateStageProviderConfig{}
	waitStr, err := utils.GetString(properties, "wait.count")
	if err != nil {
		return ret, err
	}
	waitCount, err := strconv.Atoi(waitStr)
	if err != nil {
		return ret, v1alpha2.NewCOAError(err, "wait.count must be an integer", v1alpha2.BadConfig)
	}
	ret.WaitCount = waitCount
	waitStr, err = utils.GetString(properties, "wait.interval")
	if err != nil {
		return ret, err
	}
	waitInterval, err := strconv.Atoi(waitStr)
	if err != nil {
		return ret, v1alpha2.NewCOAError(err, "wait.interval must be an integer", v1alpha2.BadConfig)
	}
	ret.WaitInterval = waitInterval
	if waitCount <= 0 {
		waitCount = 1
	}
	return ret, nil
}
func (i *CreateStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Create provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	outputs := make(map[string]interface{})

	objectType := stage.ReadInputString(inputs, "objectType")
	objectName := stage.ReadInputString(inputs, "objectName")
	action := stage.ReadInputString(inputs, "action")
	object := inputs["object"]
	var oData []byte
	if object != nil {
		oData, _ = json.Marshal(object)
	}
	lastSummaryMessage := ""
	switch objectType {
	case "instance":
		objectNamespace := stage.GetNamespace(inputs)
		if objectNamespace == "" {
			objectNamespace = "default"
		}

		if strings.EqualFold(action, RemoveAction) {
			err = i.ApiClient.DeleteInstance(ctx, objectName, objectNamespace)
			if err != nil {
				return nil, false, err
			}
		} else if strings.EqualFold(action, CreateAction) {
			err = i.ApiClient.CreateInstance(ctx, objectName, oData, objectNamespace)
			if err != nil {
				return nil, false, err
			}
			for ic := 0; ic < i.Config.WaitCount; ic++ {
				var summary *model.SummaryResult
				summary, err = i.ApiClient.GetSummary(ctx, objectName, objectNamespace)
				lastSummaryMessage = summary.Summary.SummaryMessage
				if err != nil {
					return nil, false, err
				}
				if summary.Summary.SuccessCount == summary.Summary.TargetCount {
					outputs["objectType"] = objectType
					outputs["objectName"] = objectName
					return outputs, false, nil
				}
				time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
			}
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Instance creation failed: %s", lastSummaryMessage), v1alpha2.InternalError)
			return nil, false, err
		} else {
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported action: %s", action), v1alpha2.InternalError)
			return nil, false, err
		}
	default:
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported object type: %s", objectType), v1alpha2.InternalError)
		return nil, false, err
	}
	outputs["objectType"] = objectType
	outputs["objectName"] = objectName

	return outputs, false, nil
}
