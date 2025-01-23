/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package patch

import (
	"context"
	"encoding/json"
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
	loggerName   = "providers.stage.patch"
	providerName = "P (Patch Stage)"
	patch        = "patch"
)

var (
	msLock                   sync.Mutex
	sLog                     = logger.NewLogger(loggerName)
	once                     sync.Once
	providerOperationMetrics *metrics.Metrics
)

type PatchStageProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type PatchStageProvider struct {
	Config    PatchStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient api_utils.ApiClient
}

func (s *PatchStageProvider) Init(config providers.IProviderConfig) error {
	ctx, span := observability.StartSpan("[Stage] Patch Provider", context.TODO(), &map[string]string{
		"method": "Init",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	msLock.Lock()
	defer msLock.Unlock()
	var mockConfig PatchStageProviderConfig
	mockConfig, err = toPatchStageProviderConfig(config)
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
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): failed to create metrics: %+v", err)
			}
		}
	})
	return err
}
func (s *PatchStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toPatchStageProviderConfig(config providers.IProviderConfig) (PatchStageProviderConfig, error) {
	ret := PatchStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *PatchStageProvider) InitWithMap(properties map[string]string) error {
	config, err := SymphonyStageProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func SymphonyStageProviderConfigFromMap(properties map[string]string) (PatchStageProviderConfig, error) {
	ret := PatchStageProviderConfig{}
	if api_utils.ShouldUseUserCreds() {
		user, err := api_utils.GetString(properties, "user")
		if err != nil {
			return ret, err
		}
		ret.User = user
		if ret.User == "" {
			return ret, v1alpha2.NewCOAError(nil, "user is required", v1alpha2.BadConfig)
		}
		password, err := api_utils.GetString(properties, "password")
		ret.Password = password
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}
func (m *PatchStageProvider) traceValue(ctx context.Context, v interface{}, localContext interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := api_utils.NewParser(val)
		context := m.Context.VencorContext.EvaluationContext.Clone()
		context.Value = localContext
		context.Context = ctx
		v, err := parser.Eval(*context)
		if err != nil {
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return m.traceValue(ctx, v, localContext)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := m.traceValue(ctx, v, localContext)
			if err != nil {
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := m.traceValue(ctx, v, localContext)
			if err != nil {
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
}

func (i *PatchStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] Patch Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	sLog.InfoCtx(ctx, "  P (Patch Stage): start process request")
	processTime := time.Now().UTC()
	functionName := observ_utils.GetFunctionName()
	defer providerOperationMetrics.ProviderOperationLatency(
		processTime,
		patch,
		metrics.ProcessOperation,
		metrics.RunOperationType,
		functionName,
	)
	outputs := make(map[string]interface{})

	objectType := stage.ReadInputString(inputs, "objectType")
	objectName := api_utils.ConvertReferenceToObjectName(stage.ReadInputString(inputs, "objectName"))
	patchSource := stage.ReadInputString(inputs, "patchSource")
	var patchContent interface{}
	if v, ok := inputs["patchContent"]; ok {
		patchContent = v
	}
	componentName := stage.ReadInputString(inputs, "component")
	propertyName := stage.ReadInputString(inputs, "property")
	subKey := stage.ReadInputString(inputs, "subKey")
	dedupKey := stage.ReadInputString(inputs, "dedupKey")
	patchAction := stage.ReadInputString(inputs, "patchAction")
	if patchAction == "" {
		patchAction = "add"
	}
	updated := false
	objectNamespace := stage.GetNamespace(inputs)
	if objectNamespace == "" {
		objectNamespace = "default"
	}

	var catalog model.CatalogState

	switch patchSource {
	case "", "catalog":
		if v, ok := patchContent.(string); ok {
			v := api_utils.ConvertReferenceToObjectName(v)
			catalog, err = i.ApiClient.GetCatalog(ctx, v, objectNamespace, i.Config.User, i.Config.Password)

			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): error getting catalog %s", v)
				providerOperationMetrics.ProviderOperationErrors(
					patch,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.CatalogsGetFailed.String(),
				)
				return nil, false, err
			}
		} else {
			sLog.ErrorfCtx(ctx, "  P (Patch Stage): error getting catalog %s", v)
			err = v1alpha2.NewCOAError(nil, "patchContent is not valid", v1alpha2.BadConfig)
			providerOperationMetrics.ProviderOperationErrors(
				patch,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.BadConfig.String(),
			)
			return nil, false, err
		}
	case "inline":
		if componentName != "" {
			if v, ok := patchContent.(map[string]interface{}); ok {
				catalog = model.CatalogState{
					Spec: &model.CatalogSpec{
						Properties: v,
					},
				}
			} else {
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): error getting catalog %s", v)
				err = v1alpha2.NewCOAError(nil, "patchContent is not valid", v1alpha2.BadConfig)
				providerOperationMetrics.ProviderOperationErrors(
					patch,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.BadConfig.String(),
				)
				return nil, false, err
			}
		} else {
			var componentSpec model.ComponentSpec
			jData, _ := json.Marshal(patchContent)
			if err = json.Unmarshal(jData, &componentSpec); err != nil {
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): error unmarshalling componentSpec")
				providerOperationMetrics.ProviderOperationErrors(
					patch,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.BadConfig.String(),
				)
				return nil, false, err
			}
			catalog = model.CatalogState{
				Spec: &model.CatalogSpec{
					Properties: map[string]interface{}{
						"spec": componentSpec,
					},
				},
			}
		}
	default:
		sLog.ErrorfCtx(ctx, "  P (Patch Stage): unsupported patchSource: %s", patchSource)
		err = v1alpha2.NewCOAError(nil, "patchSource is not valid", v1alpha2.BadConfig)
		providerOperationMetrics.ProviderOperationErrors(
			patch,
			functionName,
			metrics.ProcessOperation,
			metrics.RunOperationType,
			v1alpha2.BadConfig.String(),
		)
		return nil, false, err
	}

	for k, v := range catalog.Spec.Properties {
		var tv interface{}
		tv, err = i.traceValue(ctx, v, inputs["context"])
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Patch Stage): error tracing value %s", k)
			return nil, false, err
		}
		catalog.Spec.Properties[k] = tv
	}

	switch objectType {
	case "solution":
		var solution model.SolutionState
		solution, err = i.ApiClient.GetSolution(ctx, objectName, objectNamespace, i.Config.User, i.Config.Password)
		if err != nil {
			sLog.ErrorfCtx(ctx, "  P (Patch Stage): error getting solution %s", objectName)
			providerOperationMetrics.ProviderOperationErrors(
				patch,
				functionName,
				metrics.ProcessOperation,
				metrics.RunOperationType,
				v1alpha2.SolutionGetFailed.String(),
			)
			return nil, false, err
		}

		if componentName == "" {
			componentSpec, ok := catalog.Spec.Properties["spec"].(model.ComponentSpec)
			if !ok {
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): catalog spec is not valid")
				err = v1alpha2.NewCOAError(nil, "catalog spec is not valid", v1alpha2.BadConfig)
				return nil, false, err
			}
			if solution.Spec.Components == nil {
				solution.Spec.Components = make([]model.ComponentSpec, 0)
			}
			for i, c := range solution.Spec.Components {
				if c.Name == componentSpec.Name {
					if patchAction == "remove" {
						solution.Spec.Components = append(solution.Spec.Components[:i], solution.Spec.Components[i+1:]...)
					} else {
						solution.Spec.Components[i] = componentSpec
					}
					updated = true
					break
				}
			}
			if !updated && patchAction != "remove" {
				solution.Spec.Components = append(solution.Spec.Components, componentSpec)
				updated = true
			}
		} else {
			for i, c := range solution.Spec.Components {
				if c.Name == componentName {
					for k, p := range c.Properties {
						if k == propertyName {
							if subKey != "" {
								if detailedTarget, ok := p.(map[string]interface{}); ok {
									if v, ok := detailedTarget[subKey]; ok {
										if targetMap, ok := v.([]interface{}); ok {
											replaced := false
											if dedupKey != "" {
												for i, v := range targetMap {
													if vmap, ok := v.(map[string]interface{}); ok {
														if vmap[dedupKey] == catalog.Spec.Properties[dedupKey] {
															if patchAction == "remove" {
																targetMap = append(targetMap[:i], targetMap[i+1:]...)
															} else {
																targetMap[i] = catalog.Spec.Properties
															}
															replaced = true
															break
														}
													}
												}
											}
											if !replaced && patchAction != "remove" {
												targetMap = append(targetMap, catalog.Spec.Properties)
											}
											detailedTarget[subKey] = targetMap
											solution.Spec.Components[i].Properties[propertyName] = detailedTarget
											updated = true
										} else {
											sLog.ErrorfCtx(ctx, "  P (Patch Stage): target properties is not valid")
											err = v1alpha2.NewCOAError(nil, "target properties is not valid", v1alpha2.BadConfig)
											providerOperationMetrics.ProviderOperationErrors(
												patch,
												functionName,
												metrics.ProcessOperation,
												metrics.RunOperationType,
												v1alpha2.BadConfig.String(),
											)
											return nil, false, err
										}
									} else {
										sLog.ErrorfCtx(ctx, "  P (Patch Stage): subKey is not valid")
										err = v1alpha2.NewCOAError(nil, "subKey is not valid", v1alpha2.BadConfig)
										providerOperationMetrics.ProviderOperationErrors(
											patch,
											functionName,
											metrics.ProcessOperation,
											metrics.RunOperationType,
											v1alpha2.BadConfig.String(),
										)
										return nil, false, err
									}
								} else {
									sLog.ErrorfCtx(ctx, "  P (Patch Stage): subKey is not valid")
									err = v1alpha2.NewCOAError(nil, "subKey is not valid", v1alpha2.BadConfig)
									providerOperationMetrics.ProviderOperationErrors(
										patch,
										functionName,
										metrics.ProcessOperation,
										metrics.RunOperationType,
										v1alpha2.BadConfig.String(),
									)
									return nil, false, err
								}
							} else {
								if targetMap, ok := p.([]interface{}); ok {
									replaced := false
									if dedupKey != "" {
										for i, v := range targetMap {
											if vmap, ok := v.(map[string]interface{}); ok {
												if vmap[dedupKey] == catalog.Spec.Properties[dedupKey] {
													if patchAction == "remove" {
														targetMap = append(targetMap[:i], targetMap[i+1:]...)
													} else {
														targetMap[i] = catalog.Spec.Properties
													}
													replaced = true
													break
												}
											}
										}
									}
									if !replaced && patchAction != "remove" {
										targetMap = append(targetMap, catalog.Spec.Properties)
									}
									solution.Spec.Components[i].Properties[propertyName] = targetMap
									updated = true
								} else {
									sLog.ErrorfCtx(ctx, "  P (Patch Stage): target properties is not valid")
									err = v1alpha2.NewCOAError(nil, "target properties is not valid", v1alpha2.BadConfig)
									providerOperationMetrics.ProviderOperationErrors(
										patch,
										functionName,
										metrics.ProcessOperation,
										metrics.RunOperationType,
										v1alpha2.BadConfig.String(),
									)
									return nil, false, err
								}
							}
							break
						}
					}
					break
				}
			}
		}
		if updated {
			jData, _ := json.Marshal(solution)
			observ_utils.EmitUserAuditsLogs(ctx, "  P (Patch Stage): updating solution name: %s namespace: %s", objectName, objectNamespace)
			err = i.ApiClient.UpsertSolution(ctx, objectName, jData, objectNamespace, i.Config.User, i.Config.Password)
			if err != nil {
				sLog.ErrorfCtx(ctx, "  P (Patch Stage): error updating solution %s", objectName)
				providerOperationMetrics.ProviderOperationErrors(
					patch,
					functionName,
					metrics.ProcessOperation,
					metrics.RunOperationType,
					v1alpha2.UpdateFailed.String(),
				)
				return nil, false, err
			}
		}

	}
	sLog.InfoCtx(ctx, "  P (Patch Stage): end process request")

	return outputs, false, nil
}
