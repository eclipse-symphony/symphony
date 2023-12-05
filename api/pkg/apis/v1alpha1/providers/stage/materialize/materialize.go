/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package materialize

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/logger"
)

var maLock sync.Mutex
var mLog = logger.NewLogger("coa.runtime")

type MaterializeStageProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type MaterializeStageProvider struct {
	Config  MaterializeStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *MaterializeStageProvider) Init(config providers.IProviderConfig) error {
	maLock.Lock()
	defer maLock.Unlock()
	mockConfig, err := toMaterializeStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func (s *MaterializeStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toMaterializeStageProviderConfig(config providers.IProviderConfig) (MaterializeStageProviderConfig, error) {
	ret := MaterializeStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *MaterializeStageProvider) InitWithMap(properties map[string]string) error {
	config, err := MaterialieStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func MaterializeStageProviderConfigFromVendorMap(properties map[string]string) (MaterializeStageProviderConfig, error) {
	ret := make(map[string]string)
	for k, v := range properties {
		if strings.HasPrefix(k, "wait.") {
			ret[k[5:]] = v
		}
	}
	return MaterialieStageProviderConfigFromMap(ret)
}
func MaterialieStageProviderConfigFromMap(properties map[string]string) (MaterializeStageProviderConfig, error) {
	ret := MaterializeStageProviderConfig{}
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
	ret.Password = password
	return ret, nil
}
func (i *MaterializeStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Materialize Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	outputs := make(map[string]interface{})

	objects := inputs["names"].([]interface{})
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		prefixedNames[i] = fmt.Sprintf("%s-%s", inputs["__origin"], object.(string))
	}

	catalogs, err := utils.GetCatalogs(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password)
	if err != nil {
		return outputs, false, err
	}
	creationCount := 0
	for _, catalog := range catalogs {
		for _, object := range prefixedNames {
			if catalog.Spec.Name == object {
				objectScope := "default"
				if s, ok := inputs["objectScope"]; ok {
					objectScope = s.(string)
				}
				objectData, _ := json.Marshal(catalog.Spec.Properties["spec"]) //TODO: handle errors
				name := strings.TrimPrefix(catalog.Spec.Name, fmt.Sprintf("%s-", inputs["__origin"]))
				switch catalog.Spec.Type {
				case "instance":
					err = utils.CreateInstance(ctx, i.Config.BaseUrl, name, i.Config.User, i.Config.Password, objectData, objectScope) //TODO: is using Spec.Name safe? Needs to support scopes
					if err != nil {
						mLog.Errorf("Failed to create instance %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				case "solution":
					err = utils.UpsertSolution(ctx, i.Config.BaseUrl, name, i.Config.User, i.Config.Password, objectData, objectScope) //TODO: is using Spec.Name safe? Needs to support scopes
					if err != nil {
						mLog.Errorf("Failed to create solution %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				case "target":
					err = utils.UpsertTarget(ctx, i.Config.BaseUrl, name, i.Config.User, i.Config.Password, objectData, objectScope)
					if err != nil {
						mLog.Errorf("Failed to create target %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				default:
					catalog.Spec.Name = name
					catalog.Id = name
					catalog.Spec.SiteId = i.Context.SiteInfo.SiteId
					objectData, _ := json.Marshal(catalog.Spec)
					err = utils.UpsertCatalog(ctx, i.Config.BaseUrl, name, i.Config.User, i.Config.Password, objectData)
					if err != nil {
						mLog.Errorf("Failed to create catalog %s: %s", name, err.Error())
						return outputs, false, err
					}
					creationCount++
				}
			}
		}
	}
	if creationCount < len(objects) {
		err = v1alpha2.NewCOAError(nil, "failed to create all objects", v1alpha2.InternalError)
		return outputs, false, err
	}
	return outputs, true, nil
}
