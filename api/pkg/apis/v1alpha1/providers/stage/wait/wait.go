/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package wait

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var mwLock sync.Mutex
var log = logger.NewLogger("coa.runtime")

type WaitStageProviderConfig struct {
	BaseUrl      string `json:"baseUrl"`
	User         string `json:"user"`
	Password     string `json:"password"`
	WaitInterval int    `json:"wait.interval,omitempty"`
	WaitCount    int    `json:"wait.count,omitempty"`
}

type WaitStageProvider struct {
	Config  WaitStageProviderConfig
	Context *contexts.ManagerContext
}

func (s *WaitStageProvider) Init(config providers.IProviderConfig) error {
	mwLock.Lock()
	defer mwLock.Unlock()
	mockConfig, err := toWaitStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	return nil
}
func (s *WaitStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toWaitStageProviderConfig(config providers.IProviderConfig) (WaitStageProviderConfig, error) {
	ret := WaitStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *WaitStageProvider) InitWithMap(properties map[string]string) error {
	config, err := WaitStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func WaitStageProviderConfigFromVendorMap(properties map[string]string) (WaitStageProviderConfig, error) {
	ret := make(map[string]string)
	for k, v := range properties {
		if strings.HasPrefix(k, "wait.") {
			ret[k[5:]] = v
		}
	}
	return WaitStageProviderConfigFromMap(ret)
}
func WaitStageProviderConfigFromMap(properties map[string]string) (WaitStageProviderConfig, error) {
	_, span := observability.StartSpan("Wait Process Provider", context.TODO(), &map[string]string{
		"method": "WaitStageProviderConfigFromMap",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info("  P (Wait Processor): getting configuration from properties")
	ret := WaitStageProviderConfig{}
	baseUrl, err := utils.GetString(properties, "baseUrl")
	if err != nil {
		log.Errorf("  P (Wait Processor): failed to get baseUrl: %v", err)
		return ret, err
	}
	ret.BaseUrl = baseUrl
	if ret.BaseUrl == "" {
		log.Errorf("  P (Wait Processor): baseUrl is required")
		err = v1alpha2.NewCOAError(nil, "baseUrl is required", v1alpha2.BadConfig)
		return ret, err
	}
	user, err := utils.GetString(properties, "user")
	if err != nil {
		log.Errorf("  P (Wait Processor): failed to get user: %v", err)
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		log.Errorf("  P (Wait Processor): user is required")
		err = v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		return ret, err
	}
	password, err := utils.GetString(properties, "password")
	if err != nil {
		log.Errorf("  P (Wait Processor): failed to get password: %v", err)
		return ret, err
	}
	ret.Password = password
	if v, ok := properties["wait.interval"]; ok {
		var interval int
		interval, err = strconv.Atoi(v)
		if err != nil {
			cErr := v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait interval %v", v), v1alpha2.BadConfig)
			log.Errorf("  P (Wait Processor): failed to parse wait interval %v", cErr)
			return ret, cErr
		}
		ret.WaitInterval = interval
	}
	if v, ok := properties["wait.count"]; ok {
		var count int
		count, err = strconv.Atoi(v)
		if err != nil {
			cErr := v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait count %v", v), v1alpha2.BadConfig)
			log.Errorf("  P (Wait Processor): failed to parse wait count %v", cErr)
			return ret, cErr
		}
		ret.WaitCount = count
	}
	err = nil
	return ret, nil
}
func (i *WaitStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Wait Process Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Info("  P (Wait Processor): processing inputs")
	outputs := make(map[string]interface{})

	objectType := inputs["objectType"].(string)
	objects := inputs["names"].([]interface{})
	prefixedNames := make([]string, len(objects))
	for i, object := range objects {
		prefixedNames[i] = fmt.Sprintf("%v-%v", inputs["__origin"], object)
	}
	log.Debugf("  P (Wait Processor): waiting for %v %v", objectType, prefixedNames)
	counter := 0
	scope := "default"
	if s, ok := inputs["scope"]; ok {
		scope = s.(string)
	}
	for counter < i.Config.WaitCount || i.Config.WaitCount == 0 {
		foundCount := 0
		switch objectType {
		case "instance":
			var instances []model.InstanceState
			instances, err = utils.GetInstances(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password, scope)
			if err != nil {
				log.Errorf("  P (Wait Processor): failed to get instances: %v", err)
				return nil, false, err
			}
			for _, instance := range instances {
				for _, object := range prefixedNames {
					if instance.Spec.Name == object {
						foundCount++
					}
				}
			}
		case "sites":
			var sites []model.SiteState
			sites, err = utils.GetSites(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password)
			if err != nil {
				log.Errorf("  P (Wait Processor): failed to get sites: %v", err)
				return nil, false, err
			}
			for _, site := range sites {
				for _, object := range prefixedNames {
					if site.Spec.Name == object {
						foundCount++
					}
				}
			}
		case "catalogs":
			var catalogs []model.CatalogState
			catalogs, err = utils.GetCatalogs(ctx, i.Config.BaseUrl, i.Config.User, i.Config.Password)
			if err != nil {
				log.Errorf("  P (Wait Processor): failed to get catalogs: %v", err)
				return nil, false, err
			}
			for _, catalog := range catalogs {
				for _, object := range prefixedNames {
					if catalog.Spec.Name == object {
						foundCount++
					}
				}
			}
		}
		if foundCount == len(objects) {
			outputs["objectType"] = objectType
			outputs["status"] = "OK"
			log.Infof("  P (Wait Processor): found %v %v", objectType, objects)
			return outputs, false, nil
		}
		counter++
		if i.Config.WaitInterval > 0 {
			time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
		}
	}

	outputs["objectType"] = objectType
	log.Errorf("  P (Wait Processor): failed to wait for %v %v", objectType, objects)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to wait for %v %v", objectType, objects), v1alpha2.NotFound)
	return outputs, false, err
}
