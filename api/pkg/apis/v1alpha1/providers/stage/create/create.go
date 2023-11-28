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

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
)

var msLock sync.Mutex

type CreateStageProviderConfig struct {
	BaseUrl      string `json:"baseUrl"`
	User         string `json:"user"`
	Password     string `json:"password"`
	WaitCount    int    `json:"wait.count,omitempty"`
	WaitInterval int    `json:"wait.interval,omitempty"`
}

type CreateStageProvider struct {
	Config  CreateStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *CreateStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toSymphonyStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
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
	baseUrl, err := utils.GetString(properties, "baseUrl")
	if err != nil {
		return ret, err
	}
	ret.BaseUrl = baseUrl
	if ret.BaseUrl == "" {
		return ret, v1alpha2.NewCOAError(nil, "baseUrl is required", v1alpha2.BadConfig)
	}
	user, err := utils.GetString(properties, "user")
	if err != nil {
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
	}
	password, err := utils.GetString(properties, "password")
	if err != nil {
		return ret, err
	}
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
	ret.Password = password
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
		objectScope := stage.ReadInputString(inputs, "objectScope")
		if objectScope == "" {
			objectScope = "default"
		}

		if action == "remove" {
			err = utils.DeleteInstance(ctx, i.Config.BaseUrl, objectName, i.Config.User, i.Config.Password, objectScope)
			if err != nil {
				return nil, false, err
			}
		} else {
			err = utils.CreateInstance(ctx, i.Config.BaseUrl, objectName, i.Config.User, i.Config.Password, oData, objectScope)
			if err != nil {
				return nil, false, err
			}
			for ic := 0; ic < i.Config.WaitCount; ic++ {
				var summary model.SummaryResult
				summary, err = utils.GetSummary(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password, objectName, objectScope)
				lastSummaryMessage = summary.Summary.SummaryMessage
				if err != nil {
					return nil, false, err
				}
				if summary.Summary.SuccessCount == summary.Summary.TargetCount {
					break
				}
				time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
			}
			err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Instance creation failed: %s", lastSummaryMessage), v1alpha2.InternalError)
			return nil, false, err
		}
	}
	outputs["objectType"] = objectType
	outputs["objectName"] = objectName

	return outputs, false, nil
}
