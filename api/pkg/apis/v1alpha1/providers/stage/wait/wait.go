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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/metrics"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	loggerName   = "providers.stage.wait"
	providerName = "P (Wait Stage)"
	wait         = "wait"
)

var (
	log                      = logger.NewLogger(loggerName)
	mwLock                   sync.Mutex
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type WaitStageProviderConfig struct {
	User         string `json:"user"`
	Password     string `json:"password"`
	WaitInterval int    `json:"wait.interval,omitempty"`
	WaitCount    int    `json:"wait.count,omitempty"`
}

type WaitStageProvider struct {
	Config    WaitStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient api_utils.ApiClient
}

func (s *WaitStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Wait Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	mwLock.Lock()
	defer mwLock.Unlock()
	var mockConfig WaitStageProviderConfig
	mockConfig, err = toWaitStageProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = api_utils.GetApiClient()
	if err != nil {
		return err
	}
	once.Do(func() {
		if providerOperationMetrics == nil {
			providerOperationMetrics, err = metrics.New()
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Wait Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
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
	ctx, span := observability.StartSpan("Wait Process Provider", context.TODO(), &map[string]string{
		"method": "WaitStageProviderConfigFromMap",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, "  P (Wait Processor): getting configuration from properties")
	ret := WaitStageProviderConfig{}

	user, err := api_utils.GetString(properties, "user")
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Wait Processor): failed to get user: %v", err)
		return ret, err
	}
	ret.User = user
	if ret.User == "" {
		log.ErrorfCtx(ctx, "  P (Wait Processor): user is required")
		err = v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		return ret, err
	}
	password, err := api_utils.GetString(properties, "password")
	if err != nil {
		log.ErrorfCtx(ctx, "  P (Wait Processor): failed to get password: %v", err)
		return ret, err
	}
	ret.Password = password

	if v, ok := properties["wait.interval"]; ok {
		var interval int
		interval, err = strconv.Atoi(v)
		if err != nil {
			cErr := v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait interval %v", v), v1alpha2.BadConfig)
			log.ErrorfCtx(ctx, "  P (Wait Processor): failed to parse wait interval %v", cErr)
			return ret, cErr
		}
		ret.WaitInterval = interval
	}
	if v, ok := properties["wait.count"]; ok {
		var count int
		count, err = strconv.Atoi(v)
		if err != nil {
			cErr := v1alpha2.NewCOAError(err, fmt.Sprintf("failed to parse wait count %v", v), v1alpha2.BadConfig)
			log.ErrorfCtx(ctx, "  P (Wait Processor): failed to parse wait count %v", cErr)
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
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, "  P (Wait Processor): processing inputs")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	outputs := make(map[string]interface{})

	objectType, ok := inputs["objectType"].(string)
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("objectType is not a valid string: %v", inputs["objectType"]), v1alpha2.BadRequest)
		providerOperationMetrics.ProviderOperationErrors(
			wait,
			functionName,
			metrics.ProcessOperation,
			metrics.ValidateOperationType,
			v1alpha2.BadConfig.String(),
		)
		return nil, false, err
	}
	objects, ok := inputs["names"].([]interface{})
	if !ok {
		err = v1alpha2.NewCOAError(nil, "input names is not a valid list", v1alpha2.BadRequest)
		providerOperationMetrics.ProviderOperationErrors(
			wait,
			functionName,
			metrics.ProcessOperation,
			metrics.ValidateOperationType,
			v1alpha2.BadConfig.String(),
		)
		return outputs, false, err
	}
	prefixedNames := make([]string, len(objects))
	if inputs["__origin"] == nil || inputs["__origin"] == "" {
		for i, object := range objects {
			prefixedNames[i] = fmt.Sprintf("%v", object)
		}
	} else {
		for i, object := range objects {
			prefixedNames[i] = fmt.Sprintf("%v-%v", inputs["__origin"], object)
		}
	}
	namespace := stage.GetNamespace(inputs)
	if namespace == "" {
		namespace = "default"
	}

	log.DebugfCtx(ctx, "  P (Wait Processor): waiting for %v %v in namespace %s", objectType, prefixedNames, namespace)
	observ_utils.EmitUserAuditsLogs(ctx, "  P (Wait Processor): Start to wait for %v %v in namespace %s", objectType, prefixedNames, namespace)
	counter := 0
	for counter < i.Config.WaitCount || i.Config.WaitCount == 0 {
		foundCount := 0
		switch objectType {
		case "instance":
			var instances []model.InstanceState
			instances, err = i.ApiClient.GetInstances(ctx, namespace, i.Config.User, i.Config.Password)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Wait Processor): failed to get instances: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					wait,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.WaitToGetInstancesFailed.String(),
				)
				return nil, false, err
			}
			for _, instance := range instances {
				for _, object := range prefixedNames {
					object = api_utils.ConvertReferenceToObjectName(object)
					if instance.ObjectMeta.Name == object {
						foundCount++
					}
				}
			}
		case "sites":
			var sites []model.SiteState
			sites, err = i.ApiClient.GetSites(ctx, i.Config.User, i.Config.Password)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Wait Processor): failed to get sites: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					wait,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.WaitToGetSitesFailed.String(),
				)
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
			catalogs, err = i.ApiClient.GetCatalogs(ctx, namespace, i.Config.User, i.Config.Password)
			if err != nil {
				log.ErrorfCtx(ctx, "  P (Wait Processor): failed to get catalogs: %v", err)
				providerOperationMetrics.ProviderOperationErrors(
					wait,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.WaitToGetCatalogsFailed.String(),
				)
				return nil, false, err
			}
			for _, catalog := range catalogs {
				for _, object := range prefixedNames {
					object = api_utils.ConvertReferenceToObjectName(object)
					if catalog.ObjectMeta.Name == object {
						foundCount++
					}
				}
			}
		}
		if foundCount == len(objects) {
			outputs["objectType"] = objectType
			outputs["status"] = "OK"
			log.InfofCtx(ctx, "  P (Wait Processor): found %v %v", objectType, objects)
			providerOperationMetrics.ProviderOperationLatency(
				processTime,
				wait,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				functionName,
			)
			return outputs, false, nil
		}
		counter++
		if counter%10 == 0 {
			// Avoid to check the activation status too frequently
			activationName, ok := inputs["__activation"].(string)
			if !ok || activationName == "" {
				log.InfoCtx(ctx, "  P (Wait Processor): __activation is not provided, exiting endless wait.")
				return nil, false, v1alpha2.NewCOAError(nil, "related activation name is not provided", v1alpha2.BadConfig)
			}
			_, err := i.ApiClient.GetActivation(ctx, activationName, namespace, i.Config.User, i.Config.Password)
			if err != nil && strings.Contains(err.Error(), v1alpha2.NotFound.String()) {
				// Since we use ApiClient to get the activation, not found error will become v1alpha2.InternalError
				// format: utils.SummarySpecError{Code:\"Symphony API: [500]\", Message:\"Not Found: ..."}
				// We need to check if it contains v1alpha2.NotFound
				log.InfoCtx(ctx, "  P (Wait Processor): detected activation got deleted, exiting endless wait.")
				return nil, false, v1alpha2.NewCOAError(err, "related activation got deleted", v1alpha2.NotFound)
			}
			log.InfoCtx(ctx, "  P (Wait Processor): waiting for objects to be ready...")
		}
		if i.Config.WaitInterval > 0 {
			time.Sleep(time.Duration(i.Config.WaitInterval) * time.Second)
		}
	}

	outputs["objectType"] = objectType
	log.ErrorfCtx(ctx, "  P (Wait Processor): failed to wait for %v %v", objectType, objects)
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("failed to wait for %v %v", objectType, objects), v1alpha2.NotFound)
	providerOperationMetrics.ProviderOperationErrors(
		wait,
		functionName,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		v1alpha2.InvalidWaitObjectType.String(),
	)
	return outputs, false, err
}
