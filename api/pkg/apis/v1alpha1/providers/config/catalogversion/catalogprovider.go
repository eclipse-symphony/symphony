/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalogversion

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var clog = logger.NewLogger("coa.runtime")

type CatalogVersionConfigProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type CatalogVersionConfigProvider struct {
	Config    CatalogVersionConfigProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func (s *CatalogVersionConfigProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toCatalogVersionConfigProviderConfig(config)
	if err != nil {
		return err
	}
	s.Config = mockConfig
	s.ApiClient, err = utils.GetApiClient()
	if err != nil {
		return err
	}
	return nil
}
func (s *CatalogVersionConfigProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func toCatalogVersionConfigProviderConfig(config providers.IProviderConfig) (CatalogVersionConfigProviderConfig, error) {
	ret := CatalogVersionConfigProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *CatalogVersionConfigProvider) InitWithMap(properties map[string]string) error {
	config, err := CatalogVersionConfigProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func CatalogVersionConfigProviderConfigFromMap(properties map[string]string) (CatalogVersionConfigProviderConfig, error) {
	ret := CatalogVersionConfigProviderConfig{}
	if utils.ShouldUseUserCreds() {
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
	}
	return ret, nil
}

func (m *CatalogVersionConfigProvider) unwindOverrides(ctx context.Context, override string, field string, namespace string, localcontext interface{}, dependencyList map[string]map[string]bool) (interface{}, error) {
	override = utils.ConvertReferenceToObjectName(override)
	catalogversion, err := m.ApiClient.GetCatalogVersion(ctx, override, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): Unwind overrides error:", err)
		return "", err
	}
	if v, ok := utils.JsonParseProperty(catalogversion.Spec.Properties, field); ok {
		return m.traceValue(ctx, v, localcontext, dependencyList)
	}
	if catalogversion.Spec.ParentName != "" {
		return m.unwindOverrides(ctx, catalogversion.Spec.ParentName, field, namespace, localcontext, dependencyList)
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, override), v1alpha2.NotFound)
	clog.ErrorCtx(ctx, "  P (CatalogVersion): Unwind overrides error:", err)
	return "", err
}

func (m *CatalogVersionConfigProvider) Read(ctx context.Context, object string, field string, localcontext interface{}) (interface{}, error) {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "Read",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (CatalogVersion): Read, object: %s, field: %s", object, field)
	namespace := utils.GetNamespaceFromContext(localcontext)
	object = utils.ConvertReferenceToObjectName(object)
	catalogversion, err := m.ApiClient.GetCatalogVersion(ctx, object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): Read error:", err)
		return "", err
	}

	// check circular dependency
	var dependencyList map[string]map[string]bool = nil
	if localcontext != nil {
		if evalContext, ok := localcontext.(coa_utils.EvaluationContext); ok {
			if coa_utils.HasCircularDependency(object, field, evalContext) {
				clog.ErrorfCtx(ctx, "  P (CatalogVersion): Read detect circular dependency. Object: %s, field: %s, ", object, field)
				return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Detect circular dependency, object: %s, field: %s", object, field), v1alpha2.BadConfig)
			}
			dependencyList = coa_utils.DeepCopyDependencyList(evalContext.ParentConfigs)
			dependencyList = coa_utils.UpdateDependencyList(object, field, dependencyList)
		}
	}

	if v, ok := utils.JsonParseProperty(catalogversion.Spec.Properties, field); ok {
		return m.traceValue(ctx, v, localcontext, dependencyList)
	}

	if catalogversion.Spec.ParentName != "" {
		overrid, err := m.unwindOverrides(ctx, catalogversion.Spec.ParentName, field, namespace, localcontext, dependencyList)
		if err != nil {
			return "", err
		} else {
			return overrid, nil
		}
	}

	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
	clog.ErrorCtx(ctx, "  P (CatalogVersion): Read error:", err)
	return "", err
}

func (m *CatalogVersionConfigProvider) ReadObject(ctx context.Context, object string, localcontext interface{}) (map[string]interface{}, error) {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "ReadObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (CatalogVersion): ReadObject, object: %s", object)
	namespace := utils.GetNamespaceFromContext(localcontext)
	object = utils.ConvertReferenceToObjectName(object)

	catalogversion, err := m.ApiClient.GetCatalogVersion(ctx, object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): ReadObject error:", err)
		return nil, err
	}
	errList := make([]error, 0)
	allProperties, err := m.getCatalogVersionPropertiesAll(ctx, catalogversion, namespace)
	ret := map[string]interface{}{}
	for k, v := range allProperties {
		tv, err := m.traceValue(ctx, v, localcontext, nil)
		if err != nil {
			// Wrap the error using fmt.Errorf("%w", err)
			wrappedErr := fmt.Errorf("%w", err)
			errList = append(errList, wrappedErr)
			tv = err.Error()
		}

		ret[k] = tv
	}

	if len(errList) > 0 {
		// concatenate all errors into a single error
		msg := ""
		for _, err := range errList {
			msg += err.Error() + "\n"
		}
		return ret, v1alpha2.NewCOAError(nil, msg, v1alpha2.BadRequest)
	} else {
		return ret, nil
	}
}

func (m *CatalogVersionConfigProvider) getCatalogVersionPropertiesAll(ctx context.Context, catalogversion model.CatalogVersionState, namespace string) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	if catalogversion.Spec.ParentName != "" {
		metaName := utils.ConvertReferenceToObjectName(catalogversion.Spec.ParentName)
		parent, err := m.ApiClient.GetCatalogVersion(ctx, metaName, namespace, m.Config.User, m.Config.Password)
		if err != nil {
			return nil, err
		}
		parentProperties, err := m.getCatalogVersionPropertiesAll(ctx, parent, namespace)
		if err != nil {
			return nil, err
		}
		for k, v := range parentProperties {
			ret[k] = v
		}
	}
	for k, v := range catalogversion.Spec.Properties {
		// we should deep extend the properties
		// if the property is a map, we should deep extend the map
		// if the property is a list, we should deep extend the list
		// if the property is a string, we should just set the string
		// if the property is a number, we should just set the number
		// if the property is a boolean, we should just set the boolean
		// if the property is a null, we should just set the null
		ret[k] = deepExtend(ret[k], v)
	}
	return ret, nil
}

func deepExtend(dst, src interface{}) interface{} {
	switch src := src.(type) {
	case map[string]interface{}:
		if dstMap, ok := dst.(map[string]interface{}); ok {
			for k, v := range src {
				// if the key is not in the dstMap, just set the key
				if _, ok := dstMap[k]; !ok {
					dstMap[k] = v
				} else {
					dstMap[k] = deepExtend(dstMap[k], v)
				}
			}
			return dstMap
		}
		return src
	case []interface{}:
		return src
	default:
		return src
	}
}

func (m *CatalogVersionConfigProvider) traceValue(ctx context.Context, v interface{}, localcontext interface{}, dependencyList map[string]map[string]bool) (interface{}, error) {
	switch val := v.(type) {
	case string:
		parser := utils.NewParser(val)
		context := m.Context.VencorContext.EvaluationContext.Clone()
		context.DeploymentSpec = m.Context.VencorContext.EvaluationContext.DeploymentSpec
		context.Context = ctx
		if localcontext != nil {
			if ltx, ok := localcontext.(coa_utils.EvaluationContext); ok {
				context.Inputs = ltx.Inputs
				context.Outputs = ltx.Outputs
				context.Value = ltx.Value
				context.Properties = ltx.Properties
				context.Component = ltx.Component
				context.Namespace = ltx.Namespace
				if ltx.DeploymentSpec != nil {
					context.DeploymentSpec = ltx.DeploymentSpec
				}
				if dependencyList != nil {
					context.ParentConfigs = coa_utils.DeepCopyDependencyList(dependencyList)
				}
			}
		}
		v, err := parser.Eval(*context)
		if err != nil {
			clog.ErrorCtx(ctx, "  P (CatalogVersion): trace value error:", err)
			return "", err
		}
		switch vt := v.(type) {
		case string:
			return vt, nil
		default:
			return m.traceValue(ctx, v, localcontext, dependencyList)
		}
	case []interface{}:
		ret := []interface{}{}
		for _, v := range val {
			tv, err := m.traceValue(ctx, v, localcontext, dependencyList)
			if err != nil {
				clog.ErrorCtx(ctx, "  P (CatalogVersion): trace value error:", err)
				return "", err
			}
			ret = append(ret, tv)
		}
		return ret, nil
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range val {
			tv, err := m.traceValue(ctx, v, localcontext, dependencyList)
			if err != nil {
				clog.ErrorCtx(ctx, "  P (CatalogVersion): trace value error:", err)
				return "", err
			}
			ret[k] = tv
		}
		return ret, nil
	default:
		return val, nil
	}
}

// TODO: IConfigProvider interface methods should be enhanced to accept namespace as a parameter
// so we can get rid of getCatalogVersionInDefaultNamespace.
func (m *CatalogVersionConfigProvider) Set(ctx context.Context, object string, field string, value interface{}) error {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "Set",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (CatalogVersion): Set, object: %s, field: %s", object, field)
	catalogversion, err := m.getCatalogVersionInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): Set error:", err)
		return err
	}
	catalogversion.Spec.Properties[field] = value
	data, _ := json.Marshal(catalogversion)
	return m.ApiClient.UpsertCatalogVersion(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogVersionConfigProvider) SetObject(ctx context.Context, object string, value map[string]interface{}) error {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "SetObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (CatalogVersion): SetObject, object: %s", object)
	catalogversion, err := m.getCatalogVersionInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): SetObject error:", err)
		return err
	}
	catalogversion.Spec.Properties = map[string]interface{}{}
	for k, v := range value {
		catalogversion.Spec.Properties[k] = v
	}
	data, _ := json.Marshal(catalogversion)
	return m.ApiClient.UpsertCatalogVersion(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogVersionConfigProvider) Remove(ctx context.Context, object string, field string) error {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (CatalogVersion): Remove, object: %s, field: %s", object, field)
	catlog, err := m.getCatalogVersionInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): Remove error:", err)
		return err
	}
	if _, ok := catlog.Spec.Properties[field]; !ok {
		clog.ErrorCtx(ctx, "  P (CatalogVersion): Remove: field not found.")
		return v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	delete(catlog.Spec.Properties, field)
	data, _ := json.Marshal(catlog)
	return m.ApiClient.UpsertCatalogVersion(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogVersionConfigProvider) RemoveObject(ctx context.Context, object string) error {
	ctx, span := observability.StartSpan("CatalogVersion Provider", ctx, &map[string]string{
		"method": "RemoveObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (CatalogVersion): RemoveObject, object: %s", object)
	object = utils.ConvertReferenceToObjectName(object)
	return m.ApiClient.DeleteCatalogVersion(ctx, object, m.Config.User, m.Config.Password)
}

func (m *CatalogVersionConfigProvider) getCatalogVersionInDefaultNamespace(ctx context.Context, catalogversion string) (model.CatalogVersionState, error) {
	catalogversion = utils.ConvertReferenceToObjectName(catalogversion)
	return m.ApiClient.GetCatalogVersion(ctx, catalogversion, "", m.Config.User, m.Config.Password)
}
