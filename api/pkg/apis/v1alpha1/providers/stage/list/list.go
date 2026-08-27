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
	"reflect"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/stage"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	utils2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var msLock sync.Mutex
var log = logger.NewLogger("coa.runtime")

type ListStageProviderConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type ListStageProvider struct {
	Config    ListStageProviderConfig
	Context   *contexts.ManagerContext
	ApiClient utils.ApiClient
}

func (s *ListStageProvider) Init(config providers.IProviderConfig) error {
	msLock.Lock()
	defer msLock.Unlock()
	mockConfig, err := toListStageProviderConfig(config)
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
func (s *ListStageProvider) SetContext(ctx *contexts.ManagerContext) {
	s.Context = ctx
}
func toListStageProviderConfig(config providers.IProviderConfig) (ListStageProviderConfig, error) {
	ret := ListStageProviderConfig{}
	data, err := json.Marshal(config)
	if err != nil {
		return ret, err
	}
	err = utils2.UnmarshalJson(data, &ret)
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
func (i *ListStageProvider) Process(ctx context.Context, mgrContext contexts.ManagerContext, inputs map[string]interface{}) (map[string]interface{}, bool, error) {
	ctx, span := observability.StartSpan("[Stage] List Process Provider", ctx, &map[string]string{
		"method": "Process",
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)
	defer observ_utils.EmitUserDiagnosticsLogs(ctx, &err)

	log.InfoCtx(ctx, "  P (List Processor): processing inputs")

	outputs := make(map[string]interface{})

	objectType, ok := inputs["objectType"].(string)
	if !ok {
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("objectType is not a valid string: %v", inputs["objectType"]), v1alpha2.BadRequest)
		return nil, false, err
	}
	namesOnly := false
	if v, ok := inputs["namesOnly"]; ok {
		if vbool, ok := v.(bool); ok {
			namesOnly = vbool
		}
	}
	objectNamespace := stage.GetNamespace(inputs)
	if objectNamespace == "" {
		objectNamespace = "default"
	}

	// items holds the typed slice returned by the API. nameField is the JSON dot-path
	// used to extract an object's name when namesOnly is requested.
	var items interface{}
	nameField := "metadata.name"

	switch objectType {
	case "instance":
		var instances []model.InstanceState
		instances, err = i.ApiClient.GetInstances(ctx, objectNamespace, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to get instances: %v", err)
			return nil, false, err
		}
		items = instances
	case "sites":
		var sites []model.SiteState
		sites, err = i.ApiClient.GetSites(ctx, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to get sites: %v", err)
			return nil, false, err
		}
		filteredSites := make([]model.SiteState, 0)
		for _, site := range sites {
			if site.Spec.Name != mgrContext.SiteInfo.SiteId { //TODO: this should filter to keep just the direct children?
				filteredSites = append(filteredSites, site)
			}
		}
		items = filteredSites
		nameField = "spec.name"
	case "catalogversions":
		var catalogversions []model.CatalogVersionState
		catalogversions, err = i.ApiClient.GetCatalogVersions(ctx, objectNamespace, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to get catalogversions: %v", err)
			return nil, false, err
		}
		items = catalogversions
	case "activations":
		var activations []model.ActivationState
		activations, err = i.ApiClient.GetActivations(ctx, objectNamespace, i.Config.User, i.Config.Password)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to get activations: %v", err)
			return nil, false, err
		}
		items = activations
	default:
		log.ErrorfCtx(ctx, "  P (List Processor): unsupported object type: %s", objectType)
		err = v1alpha2.NewCOAError(nil, fmt.Sprintf("Unsupported object type: %s", objectType), v1alpha2.InternalError)
		return nil, false, err
	}

	// Apply generic field-based filtering (works for any object type).
	var filters []filterSpec
	filters, err = parseFilters(inputs)
	if err != nil {
		log.ErrorfCtx(ctx, "  P (List Processor): failed to parse filters: %v", err)
		return nil, false, err
	}
	if len(filters) > 0 {
		items, err = filterItems(items, filters)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to apply filters: %v", err)
			return nil, false, err
		}
	}

	itemCount := reflect.ValueOf(items).Len()
	if namesOnly {
		var names []string
		names, err = extractNames(items, nameField)
		if err != nil {
			log.ErrorfCtx(ctx, "  P (List Processor): failed to extract names: %v", err)
			return nil, false, err
		}
		outputs["items"] = names
	} else {
		outputs["items"] = items
	}
	outputs["objectType"] = objectType
	outputs["itemCount"] = itemCount
	return outputs, false, nil
}

// filterSpec describes a single generic filter that is matched against a field of
// an object (identified by a JSON dot-path, e.g. "status.status" or
// "metadata.labels.parentActivation").
type filterSpec struct {
	Field    string `json:"field"`
	Value    string `json:"value"`
	Operator string `json:"operator"` // eq (default), ne, contains, exists
}

// parseFilters reads filter definitions from the stage inputs. Two forms are supported:
//   - a single filter via the "filterField"/"filterValue"/"filterOperator" inputs
//   - a list of filters via the "filter" (or "filters") input, each entry being a
//     map with "field", "value" and optional "operator" keys
func parseFilters(inputs map[string]interface{}) ([]filterSpec, error) {
	ret := make([]filterSpec, 0)

	if v, ok := inputs["filterField"]; ok {
		if field, ok := v.(string); ok && field != "" {
			f := filterSpec{Field: field}
			if val, ok := inputs["filterValue"].(string); ok {
				f.Value = val
			}
			if op, ok := inputs["filterOperator"].(string); ok {
				f.Operator = op
			}
			ret = append(ret, f)
		}
	}

	var raw interface{}
	if v, ok := inputs["filter"]; ok {
		raw = v
	} else if v, ok := inputs["filters"]; ok {
		raw = v
	}
	if raw != nil {
		data, err := json.Marshal(raw)
		if err != nil {
			return nil, v1alpha2.NewCOAError(err, "failed to marshal filter input", v1alpha2.BadRequest)
		}
		var list []filterSpec
		if err := json.Unmarshal(data, &list); err != nil {
			// allow a single filter object as well
			var single filterSpec
			if err2 := json.Unmarshal(data, &single); err2 != nil {
				return nil, v1alpha2.NewCOAError(err, "invalid filter input", v1alpha2.BadRequest)
			}
			if single.Field != "" {
				ret = append(ret, single)
			}
		} else {
			for _, f := range list {
				if f.Field != "" {
					ret = append(ret, f)
				}
			}
		}
	}
	return ret, nil
}

// filterItems keeps only the elements of the (typed) slice that match all filters.
// The returned value preserves the original element type.
func filterItems(items interface{}, filters []filterSpec) (interface{}, error) {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return items, nil
	}
	out := reflect.MakeSlice(v.Type(), 0, v.Len())
	for idx := 0; idx < v.Len(); idx++ {
		el := v.Index(idx)
		m, err := toGenericMap(el.Interface())
		if err != nil {
			return nil, err
		}
		if matchFilters(m, filters) {
			out = reflect.Append(out, el)
		}
	}
	return out.Interface(), nil
}

// extractNames returns the value at nameField for each element of the (typed) slice.
func extractNames(items interface{}, nameField string) ([]string, error) {
	v := reflect.ValueOf(items)
	names := make([]string, 0)
	if v.Kind() != reflect.Slice {
		return names, nil
	}
	for idx := 0; idx < v.Len(); idx++ {
		m, err := toGenericMap(v.Index(idx).Interface())
		if err != nil {
			return nil, err
		}
		if val, ok := getByPath(m, nameField); ok {
			names = append(names, utils2.FormatAsString(val))
		} else {
			names = append(names, "")
		}
	}
	return names, nil
}

func toGenericMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func matchFilters(m map[string]interface{}, filters []filterSpec) bool {
	for _, f := range filters {
		val, found := getByPath(m, f.Field)
		operator := strings.ToLower(f.Operator)
		switch operator {
		case "exists":
			if !found {
				return false
			}
		case "ne":
			if found && utils2.FormatAsString(val) == f.Value {
				return false
			}
		case "contains":
			if !found || !strings.Contains(utils2.FormatAsString(val), f.Value) {
				return false
			}
		default: // eq
			if !found || utils2.FormatAsString(val) != f.Value {
				return false
			}
		}
	}
	return true
}

// getByPath resolves a dot-separated path within a generic map produced from JSON.
func getByPath(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = m
	for _, part := range parts {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = asMap[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
