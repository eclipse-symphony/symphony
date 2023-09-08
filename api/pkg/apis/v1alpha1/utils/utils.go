/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/oliveagle/jsonpath"
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
		return i.(int32)
	}
	return defaultVal
}
func GetString(col map[string]string, key string) (string, error) {
	if v, ok := col[key]; ok {
		i, e := ParseValue(v)
		if e != nil {
			return "", e
		}
		return i.(string), nil
	}
	return "", fmt.Errorf("key %s is not found", key)
}

func ReadStringFromMapCompat(col map[string]interface{}, key string, defaultVal string) string {
	if v, ok := col[key]; ok {
		i, e := ParseValue(fmt.Sprintf("%v", v))
		if e != nil {
			return defaultVal
		}
		return i.(string)
	}
	return defaultVal
}

func ReadString(col map[string]string, key string, defaultVal string) string {
	if v, ok := col[key]; ok {
		i, e := ParseValue(v)
		if e != nil {
			return defaultVal
		}
		return i.(string)
	}
	return defaultVal
}
func ReadStringWithOverrides(col1 map[string]string, col2 map[string]string, key string, defaultVal string) string {
	val := ReadString(col1, key, defaultVal)
	return ReadString(col2, key, val)
}
func MergeCollection(col1 map[string]string, col2 map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range col1 {
		ret[k] = v
	}
	for k, v := range col2 {
		ret[k] = v
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
	if strings.Contains(val, "$instance()") {
		val = strings.ReplaceAll(val, "$instance()", name)
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
					ob, _ := jsonpath.JsonPathLookup(item, path)
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
					ob, _ := jsonpath.JsonPathLookup(item, path)
					ret = append(ret, ob)
					jData, _ = yaml.Marshal(ob)
				}
				jData, _ = json.Marshal(ret)
			}
		} else {
			ob, _ := jsonpath.JsonPathLookup(dict, path)
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
