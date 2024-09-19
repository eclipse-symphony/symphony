/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package catalog

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

type CatalogConfigProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type CatalogConfigProvider struct {
	Config    CatalogConfigProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func (s *CatalogConfigProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toCatalogConfigProviderConfig(config)
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
func (s *CatalogConfigProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}

func toCatalogConfigProviderConfig(config providers.IProviderConfig) (CatalogConfigProviderConfig, error) {
	ret := CatalogConfigProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, err
}
func (i *CatalogConfigProvider) InitWithMap(properties map[string]string) error {
	config, err := CatalogConfigProviderConfigFromMap(properties)
	if err != nil {
		return err
	}
	return i.Init(config)
}
func CatalogConfigProviderConfigFromMap(properties map[string]string) (CatalogConfigProviderConfig, error) {
	ret := CatalogConfigProviderConfig{}
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

func (m *CatalogConfigProvider) unwindOverrides(ctx context.Context, override string, field string, namespace string, localcontext interface{}, dependencyList map[string]map[string]bool) (interface{}, error) {
	override = utils.ConvertReferenceToObjectName(override)
	catalog, err := m.ApiClient.GetCatalog(ctx, override, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): Unwind overrides error:", err)
		return "", err
	}
	if v, ok := utils.JsonParseProperty(catalog.Spec.Properties, field); ok {
		return m.traceValue(ctx, v, localcontext, dependencyList)
	}
	if catalog.Spec.ParentName != "" {
		return m.unwindOverrides(ctx, catalog.Spec.ParentName, field, namespace, localcontext, dependencyList)
	}
	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, override), v1alpha2.NotFound)
	clog.ErrorCtx(ctx, "  P (Catalog): Unwind overrides error:", err)
	return "", err
}

func (m *CatalogConfigProvider) Read(ctx context.Context, object string, field string, localcontext interface{}) (interface{}, error) {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "Read",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (Catalog): Read, object: %s, field: %s", object, field)
	namespace := utils.GetNamespaceFromContext(localcontext)
	object = utils.ConvertReferenceToObjectName(object)
	catalog, err := m.ApiClient.GetCatalog(ctx, object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): Read error:", err)
		return "", err
	}

	// check circular dependency
	var dependencyList map[string]map[string]bool = nil
	if localcontext != nil {
		if evalContext, ok := localcontext.(coa_utils.EvaluationContext); ok {
			if coa_utils.HasCircularDependency(object, field, evalContext) {
				clog.ErrorfCtx(ctx, "  P (Catalog): Read detect circular dependency. Object: %s, field: %s, ", object, field)
				return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("Detect circular dependency, object: %s, field: %s", object, field), v1alpha2.BadConfig)
			}
			dependencyList = coa_utils.DeepCopyDependencyList(evalContext.ParentConfigs)
			dependencyList = coa_utils.UpdateDependencyList(object, field, dependencyList)
		}
	}

	if v, ok := utils.JsonParseProperty(catalog.Spec.Properties, field); ok {
		return m.traceValue(ctx, v, localcontext, dependencyList)
	}

	if catalog.Spec.ParentName != "" {
		overrid, err := m.unwindOverrides(ctx, catalog.Spec.ParentName, field, namespace, localcontext, dependencyList)
		if err != nil {
			return "", err
		} else {
			return overrid, nil
		}
	}

	err = v1alpha2.NewCOAError(nil, fmt.Sprintf("field '%s' is not found in configuration '%s'", field, object), v1alpha2.NotFound)
	clog.ErrorCtx(ctx, "  P (Catalog): Read error:", err)
	return "", err
}

func (m *CatalogConfigProvider) ReadObject(ctx context.Context, object string, localcontext interface{}) (map[string]interface{}, error) {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "ReadObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (Catalog): ReadObject, object: %s", object)
	namespace := utils.GetNamespaceFromContext(localcontext)
	object = utils.ConvertReferenceToObjectName(object)

	catalog, err := m.ApiClient.GetCatalog(ctx, object, namespace, m.Config.User, m.Config.Password)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): ReadObject error:", err)
		return nil, err
	}
	errList := make([]error, 0)
	ret := map[string]interface{}{}
	for k, v := range catalog.Spec.Properties {
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

func (m *CatalogConfigProvider) traceValue(ctx context.Context, v interface{}, localcontext interface{}, dependencyList map[string]map[string]bool) (interface{}, error) {
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
			clog.ErrorCtx(ctx, "  P (Catalog): trace value error:", err)
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
				clog.ErrorCtx(ctx, "  P (Catalog): trace value error:", err)
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
				clog.ErrorCtx(ctx, "  P (Catalog): trace value error:", err)
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
// so we can get rid of getCatalogInDefaultNamespace.
func (m *CatalogConfigProvider) Set(ctx context.Context, object string, field string, value interface{}) error {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "Set",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)
	clog.DebugfCtx(ctx, "  P (Catalog): Set, object: %s, field: %s", object, field)
	catalog, err := m.getCatalogInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): Set error:", err)
		return err
	}
	catalog.Spec.Properties[field] = value
	data, _ := json.Marshal(catalog)
	return m.ApiClient.UpsertCatalog(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogConfigProvider) SetObject(ctx context.Context, object string, value map[string]interface{}) error {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "SetObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (Catalog): SetObject, object: %s", object)
	catalog, err := m.getCatalogInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): SetObject error:", err)
		return err
	}
	catalog.Spec.Properties = map[string]interface{}{}
	for k, v := range value {
		catalog.Spec.Properties[k] = v
	}
	data, _ := json.Marshal(catalog)
	return m.ApiClient.UpsertCatalog(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogConfigProvider) Remove(ctx context.Context, object string, field string) error {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "Remove",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (Catalog): Remove, object: %s, field: %s", object, field)
	catlog, err := m.getCatalogInDefaultNamespace(ctx, object)
	if err != nil {
		clog.ErrorCtx(ctx, "  P (Catalog): Remove error:", err)
		return err
	}
	if _, ok := catlog.Spec.Properties[field]; !ok {
		clog.ErrorCtx(ctx, "  P (Catalog): Remove: field not found.")
		return v1alpha2.NewCOAError(nil, "field not found", v1alpha2.NotFound)
	}
	delete(catlog.Spec.Properties, field)
	data, _ := json.Marshal(catlog)
	return m.ApiClient.UpsertCatalog(ctx, object, data, m.Config.User, m.Config.Password)
}

func (m *CatalogConfigProvider) RemoveObject(ctx context.Context, object string) error {
	ctx, span := observability.StartSpan("Catalog Provider", ctx, &map[string]string{
		"method": "RemoveObject",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	clog.DebugfCtx(ctx, "  P (Catalog): RemoveObject, object: %s", object)
	object = utils.ConvertReferenceToObjectName(object)
	return m.ApiClient.DeleteCatalog(ctx, object, m.Config.User, m.Config.Password)
}

func (m *CatalogConfigProvider) getCatalogInDefaultNamespace(ctx context.Context, catalog string) (model.CatalogState, error) {
	catalog = utils.ConvertReferenceToObjectName(catalog)
	return m.ApiClient.GetCatalog(ctx, catalog, "", m.Config.User, m.Config.Password)
}
