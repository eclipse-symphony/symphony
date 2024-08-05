/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	oJsonpath "github.com/oliveagle/jsonpath"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/yaml"
)

const (
	Must   = "must"
	Prefer = "prefer"
	Reject = "reject"
	Any    = "any"
)

func matchString(src string, target string) bool {
	if strings.Contains(src, "*") || strings.Contains(src, "%") {
		p := strings.ReplaceAll(src, "*", ".*")
		p = strings.ReplaceAll(p, "%", ".")
		re := regexp.MustCompile(p)
		return re.MatchString(target)
	} else {
		return src == target
	}
}

func ReadInt32(col map[string]string, key string, defaultVal int32) int32 {
	if v, ok := col[key]; ok {
		i, e := ParseValue(v)
		if e != nil {
			return defaultVal
		}
		if i, iok := i.(int32); iok {
			return i
		}
	}
	return defaultVal
}
func GetString(col map[string]string, key string) (string, error) {
	if v, ok := col[key]; ok {
		i, e := ParseValue(v)
		if e != nil {
			return "", e
		}
		s, sok := i.(string)
		if sok {
			return s, nil
		} else {
			return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("value of %s is not a string", key), v1alpha2.BadConfig)
		}
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("key %s is not found", key), v1alpha2.BadConfig)
}

func ReadStringFromMapCompat(col map[string]interface{}, key string, defaultVal string) string {
	if v, ok := col[key]; ok {
		i, e := ParseValue(fmt.Sprintf("%v", v))
		if e != nil {
			return defaultVal
		}
		s, sok := i.(string)
		if sok {
			return s
		}
	}
	return defaultVal
}

func ReadString(col map[string]string, key string, defaultVal string) string {
	if v, ok := col[key]; ok {
		i, e := ParseValue(v)
		if e != nil {
			return defaultVal
		}
		s, sok := i.(string)
		if sok {
			return s
		}
	}
	return defaultVal
}
func ReadStringWithOverrides(col1 map[string]string, col2 map[string]string, key string, defaultVal string) string {
	val := ReadString(col1, key, defaultVal)
	return ReadString(col2, key, val)
}

func ContainsString(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
func MergeCollection(cols ...map[string]string) map[string]string {
	ret := make(map[string]string)
	for _, col := range cols {
		for k, v := range col {
			ret[k] = v
		}
	}
	return ret
}
func CollectStringMap(col map[string]string, prefix string) map[string]string {
	ret := make(map[string]string)
	for k := range col {
		if strings.HasPrefix(k, prefix) {
			ret[k] = ReadString(col, k, "")
		}
	}
	return ret
}

// TODO: we should get rid of this
func ParseValue(v string) (interface{}, error) { //TODO: make this a generic utiliy
	if v == "$true" {
		return true, nil
	} else if v == "$false" {
		return false, nil
	} else if strings.HasPrefix(v, "#") {
		ri, e := strconv.Atoi(v[1:])
		return int32(ri), e
	} else if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
		var objmap map[string]*json.RawMessage
		e := json.Unmarshal([]byte(v), &objmap)
		return objmap, e
	} else if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		var objmap []map[string]*json.RawMessage
		e := json.Unmarshal([]byte(v), &objmap)
		return objmap, e
	} else if strings.HasPrefix(v, "$") {
		return os.Getenv(v[1:]), nil
	}
	return v, nil
}

// TODO: This should not be used anymore
func ProjectValue(val string, name string) string {
	if strings.Contains(val, "${{$instance()}}") {
		val = strings.ReplaceAll(val, "${{$instance()}}", name)
	}
	return val
}

func FormatObject(obj interface{}, isArray bool, path string, format string) ([]byte, error) {
	jData, _ := json.Marshal(obj)
	if path == "" && format == "" {
		return jData, nil
	}
	var dict interface{}
	if isArray {
		dict = make([]map[string]interface{}, 0)
	} else {
		dict = make(map[string]interface{})
	}
	json.Unmarshal(jData, &dict)
	if path != "" {
		if path == "first_embedded" {
			path = "$.spec.components[0].properties.embedded"
		}
		if isArray {
			if format == "yaml" {
				ret := make([]byte, 0)
				for i, item := range dict.([]interface{}) {
					ob, _ := oJsonpath.JsonPathLookup(item, path)
					if s, ok := ob.(string); ok {
						str, err := strconv.Unquote(strings.TrimSpace(s))
						if err != nil {
							str = strings.TrimSpace(s)
						}
						var o interface{}
						err = yaml.Unmarshal([]byte(str), &o)
						if err != nil {
							jData = []byte(s)
						} else {
							jData, _ = yaml.Marshal(o)
						}
					} else {
						jData, _ = yaml.Marshal(ob)
					}
					if i > 0 {
						ret = append(ret, []byte("---\n")...)
					}
					ret = append(ret, jData...)
				}
				jData = ret
			} else {
				ret := make([]interface{}, 0)
				for _, item := range dict.([]interface{}) {
					ob, _ := oJsonpath.JsonPathLookup(item, path)
					ret = append(ret, ob)
					jData, _ = yaml.Marshal(ob)
				}
				jData, _ = json.Marshal(ret)
			}
		} else {
			ob, _ := oJsonpath.JsonPathLookup(dict, path)
			if format == "yaml" {
				if s, ok := ob.(string); ok {
					str, err := strconv.Unquote(strings.TrimSpace(s))
					if err != nil {
						str = strings.TrimSpace(s)
					}
					var o interface{}
					err = yaml.Unmarshal([]byte(str), &o)
					if err != nil {
						jData = []byte(str)
					} else {
						jData, _ = yaml.Marshal(o)
					}
				} else {
					jData, _ = yaml.Marshal(ob)
				}
			} else {
				jData, _ = json.Marshal(ob)
			}
		}
	}
	return jData, nil
}

func toInterfaceMap(m map[string]string) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range m {
		ret[k] = v
	}
	return ret
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

func ReplaceSeperator(name string) string {
	if strings.Contains(name, ":") {
		name = strings.ReplaceAll(name, ":", constants.ResourceSeperator)
	}
	return name
}

func GetNamespaceFromContext(localContext interface{}) string {
	if localContext != nil {
		if ltx, ok := localContext.(coa_utils.EvaluationContext); ok {
			return ltx.Namespace
		}
	}
	return " "
}

func removeDuplicates(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func AreSlicesEqual(slice1, slice2 []string) bool {
	slice1 = removeDuplicates(slice1)
	slice2 = removeDuplicates(slice2)

	if len(slice1) != len(slice2) {
		return false
	}

	sort.Strings(slice1)
	sort.Strings(slice2)

	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}

	return true
}
