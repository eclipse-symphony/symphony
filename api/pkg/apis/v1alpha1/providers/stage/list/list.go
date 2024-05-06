/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package list

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var log = logger.NewLogger("coa.runtime")

type ListStageProviderConfig struct {
	BaseUrl  string `json:"baseUrl"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ListStageProvider struct {
	Config  ListStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *ListStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toListStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func (s *ListStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toListStageProviderConfig(config providers.IProviderConfig) (ListStageProviderConfig, error) {
	ret := ListStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *ListStageProvider) InitWithMap(properties map[string]string) error {
	config, err := ListStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func ListStageProviderConfigFromMap(properties map[string]string) (ListStageProviderConfig, error) {
	ret := ListStageProviderConfig{}
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
func (i *ListStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] List Process Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info("  P (List Processor): processing inputs")

	outputs := make(map[string]interface{})

	objectType := inputs["objectType"].(string)
	namesOnly := false
	if v, ok := inputs["namesOnly"]; ok {
		if v.(bool) {
			namesOnly = v.(bool)
		}
	}
	objectNamespace := stage.GetNamespace(inputs)
	if objectNamespace == "" {
		objectNamespace = "default"
	}

	switch objectType {
	case "instance":
		var instances []model.InstanceState
		instances, err = utils.GetInstances(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password, objectNamespace)
		if err != nil {
			log.Errorf("  P (List Processor): failed to get instances: %v", err)
			return nil, false, err
		}
		if namesOnly {
			names := make([]string, 0)
			for _, instance := range instances {
				names = append(names, instance.ObjectMeta.Name)
			}
			outputs["items"] = names
		} else {
			outputs["items"] = instances
		}
	case "sites":
		var sites []model.SiteState
		sites, err = utils.GetSites(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password)
		if err != nil {
			log.Errorf("  P (List Processor): failed to get sites: %v", err)
			return nil, false, err
		}
		filteredSites := make([]model.SiteState, 0)
		for _, site := range sites {
			if site.Spec.Name != mgrContext.SiteInfo.SiteId { //TODO: this should filter to keep just the direct children?
				filteredSites = append(filteredSites, site)
			}
		}
		if namesOnly {
			names := make([]string, 0)
			for _, site := range filteredSites {
				names = append(names, site.Spec.Name)
			}
			outputs["items"] = names
		} else {
			outputs["items"] = filteredSites
		}
	case "catalogs":
		var catalogs []model.CatalogState
		catalogs, err = utils.GetCatalogs(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password, objectNamespace)
		if err != nil {
			log.Errorf("  P (List Processor): failed to get catalogs: %v", err)
			return nil, false, err
		}
		if namesOnly {
			names := make([]string, 0)
			for _, catalog := range catalogs {
				names = append(names, catalog.ObjectMeta.Name)
			}
			outputs["items"] = names
		} else {
			outputs["items"] = catalogs
		}
	default:
		log.Errorf("  P (List Processor): unsupported object type: %s", objectType)
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported object type: %s", objectType), v1alpha2.InternalError)
		return nil, false, err
	}
	outputs["objectType"] = objectType
	return outputs, false, nil
}
