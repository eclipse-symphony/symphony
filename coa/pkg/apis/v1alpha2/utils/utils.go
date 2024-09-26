/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"k8s.io/client-go/util/jsonpath"
)

func UnmarshalDuration(duration string) (time.Duration, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(duration), &v); err != nil {
		return 1 * time.Millisecond, err
	}
	switch value := v.(type) {
	case float64:
		return time.Duration(value), nil
	case string:
		ret, err := time.ParseDuration(value)
		if err != nil {
			return 1 * time.Microsecond, err
		}
		return ret, nil
	default:
		return 1 * time.Microsecond, errors.New("invalid duration format")
	}
}

func ParseProperty(val string) string {
	if strings.HasPrefix(val, "$env:") {
		return os.Getenv(val[5:])
	}
	return val
}

type EvaluationContext struct {
	ConfigProvider config.IExtConfigProvider
	SecretProvider secret.IExtSecretProvider
	DeploymentSpec interface{}
	Properties     map[string]string
	Inputs         map[string]interface{}
	Outputs        map[string]map[string]interface{}
	Component      string
	Value          interface{}
	Namespace      string
	ParentConfigs  map[string]map[string]bool
	Context        context.Context
}

func (e *EvaluationContext) Clone() *EvaluationContext {
	// The Clone() method shares references to the same ConfigProvider and SecretProvider
	// Other fields are not shared and need to be filled in by the caller
	if e == nil {
		return nil
	}
	return &EvaluationContext{
		ConfigProvider: e.ConfigProvider,
		SecretProvider: e.SecretProvider,
	}
}

func HasCircularDependency(object string, field string, context EvaluationContext) bool {
	if context.ParentConfigs == nil {
		return false
	}
	if catalogFields, exist := context.ParentConfigs[object]; exist {
		if catalogFields[field] {
			return true
		}
	}

	return false
}

func UpdateDependencyList(object string, field string, dependencyList map[string]map[string]bool) map[string]map[string]bool {
	if dependencyList == nil {
		dependencyList = make(map[string]map[string]bool)
	}
	if _, ok := dependencyList[object]; !ok {
		dependencyList[object] = make(map[string]bool)
	}
	dependencyList[object][field] = true
	return dependencyList
}

func DeepCopyDependencyList(dependencyList map[string]map[string]bool) map[string]map[string]bool {
	if dependencyList == nil {
		return nil
	}

	newMapConfigs := make(map[string]map[string]bool)
	for key, innerMap := range dependencyList {
		newInnerMap := make(map[string]bool)
		for innerKey, value := range innerMap {
			newInnerMap[innerKey] = value
		}
		newMapConfigs[key] = newInnerMap
	}
	return newMapConfigs
}

func JsonPathQuery(obj interface{}, jsonPath string) (interface{}, error) {
	jPath := jsonPath
	if !strings.HasPrefix(jPath, "{") {
		jPath = "{" + jsonPath + "}" // k8s.io/client-go/util/jsonpath requires JsonPath expression to be wrapped in {}
	}

	result, err := jsonPathQuery(obj, jPath)
	if err == nil {
		return result, nil
	}

	// This is a workaround for filtering by root-level attributes. In this case, we need to
	// wrap the object into an array and then query the array.
	var arr []interface{}
	switch obj.(type) {
	case []interface{}:
		// the object is already an array, so the query didn't work
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("no matches found by JsonPath query '%s'", jsonPath), v1alpha2.InternalError)
	default:
		arr = append(arr, obj)
	}
	return jsonPathQuery(arr, jPath)
}
func jsonPathQuery(obj interface{}, jsonPath string) (interface{}, error) {
	jpLookup := jsonpath.New("lookup")
	jpLookup.AllowMissingKeys(true)
	jpLookup.EnableJSONOutput(true)

	err := jpLookup.Parse(jsonPath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = jpLookup.Execute(&buf, obj)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	err = json.Unmarshal(buf.Bytes(), &result)

	if err != nil {
		return nil, err
	} else if len(result) == 0 {
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("no matches found by JsonPath query '%s'", jsonPath), v1alpha2.InternalError)
	} else if len(result) == 1 {
		return result[0], nil
	} else {
		return result, nil
	}
}
func FormatAsString(val interface{}) string {
	switch tv := val.(type) {
	case string:
		return tv
	case int:
		return strconv.Itoa(tv)
	case int32:
		return strconv.Itoa(int(tv))
	case int64:
		return strconv.Itoa(int(tv))
	case float32:
		return strconv.FormatFloat(float64(tv), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(tv, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(tv)
	case map[string]interface{}:
		ret, _ := json.Marshal(tv)
		return string(ret)
	case []interface{}:
		ret, _ := json.Marshal(tv)
		return string(ret)
	default:
		return fmt.Sprintf("%v", tv)
	}
}

func ConvertStringToValidLabel(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
