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
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/itchyny/gojq"
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

// Define the struct
type ObjectInfo struct {
	Name         string
	SummaryId    string
	SummaryJobId string
}

func IsNotFound(err error) bool {
	if apiError, ok := err.(APIError); ok {
		return apiError.Code == v1alpha2.NotFound
	}
	return v1alpha2.IsNotFound(err)
}

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

func MergeCollection_StringAny(cols ...map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for _, col := range cols {
		for k, v := range col {
			ret[k] = v
		}
	}
	return ret
}

func DeepCopyCollection(originalCols map[string]interface{}, excludeKeys ...string) map[string]interface{} {
	ret := make(map[string]interface{})
	if originalCols == nil {
		return ret
	}
	for k, v := range originalCols {
		if len(excludeKeys) > 0 && ContainsString(excludeKeys, k) {
			continue
		}
		ret[k] = v
	}
	return ret
}

func DeepCopyCollectionWithPrefixExclude(originalCols map[string]interface{}, prefixExcludes ...string) map[string]interface{} {
	ret := make(map[string]interface{})
	if originalCols == nil {
		return ret
	}
	for k, v := range originalCols {
		exclude := false
		for _, prefix := range prefixExcludes {
			if strings.HasPrefix(k, prefix) {
				exclude = true
				break
			}
		}
		if !exclude {
			ret[k] = v
		}
	}
	return ret
}

func ToJsonString(obj interface{}) string {
	json, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(json)
}

func GenerateKeyLockName(strs ...string) string {
	ret := ""
	for i, str := range strs {
		if i == 0 {
			ret += str
		} else {
			ret += ("&&" + str)
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
	var jData []byte
	var err error
	if path != "" {
		if path == "first_embedded" {
			path = "$.spec.components[0].properties.embedded"
		}
		if isArray {
			rawData, _ := json.Marshal(obj)
			dict := make([]map[string]interface{}, 0)
			json.Unmarshal(rawData, &dict)
			if format == "yaml" {
				ret := make([]byte, 0)
				for i, item := range dict {
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
							jData, err = yaml.Marshal(o)
							if err != nil {
								return nil, err
							}
						}
					} else {
						jData, err = yaml.Marshal(ob)
						if err != nil {
							return nil, err
						}
					}
					if i > 0 {
						ret = append(ret, []byte("---\n")...)
					}
					ret = append(ret, jData...)
				}
				jData = ret
			} else {
				ret := make([]interface{}, 0)
				for _, item := range dict {
					ob, _ := oJsonpath.JsonPathLookup(item, path)
					ret = append(ret, ob)
				}
				jData, err = json.Marshal(ret)
				if err != nil {
					return nil, err
				}
			}
		} else {
			ob, _ := oJsonpath.JsonPathLookup(obj, path)
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
						jData, err = yaml.Marshal(o)
						if err != nil {
							return nil, err
						}
					}
				} else {
					jData, err = yaml.Marshal(ob)
					if err != nil {
						return nil, err
					}
				}
			} else {
				jData, err = json.Marshal(ob)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		if format == "yaml" {
			jData, err = yaml.Marshal(obj)
		} else {
			jData, err = json.Marshal(obj)
		}
	}
	return jData, err
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

func isAlphanum(query string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(query)
}

func JsonParseProperty(properties interface{}, fieldPath string) (any, bool) {
	s := formatPathForNestedJsonField(fieldPath)
	query, err := gojq.Parse(s)
	if err != nil {
		return nil, false
	}

	var value any
	iter := query.Run(properties)
	for {
		result, ok := iter.Next()
		if !ok {
			// iterator terminates
			break
		}
		if err, ok := result.(error); ok {
			fmt.Println(err)
			return nil, false
		}
		value = result
	}
	return value, value != nil
}

func formatPathForNestedJsonField(s string) string {
	if len(s) == 0 {
		return s
	}

	// if the string contains "`", it means it is a string with jq syntax and needs to be unquoted
	if s[0] == '`' {
		val, err := strconv.Unquote(s)
		if err != nil {
			return ""
		}
		return val
	} else {
		return fmt.Sprintf(".%s", strconv.Quote(s))
	}
}

func ConvertReferenceToObjectName(name string) string {
	return ConvertReferenceToObjectNameHelper(name)
}

func ConvertObjectNameToReference(name string) string {
	index := strings.LastIndex(name, constants.ResourceSeperator)
	if index == -1 {
		return name
	}
	return name[:index] + constants.ReferenceSeparator + name[index+len(constants.ResourceSeperator):]
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

type FailedDeployment struct {
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
}

func DetermineObjectTerminalStatus(objectMeta model.ObjectMeta, status model.DeployableStatusV2) bool {
	return status.Generation == int(objectMeta.ObjGeneration) && (status.Status == "Succeeded" || status.Status == "Failed")
}

// Once status report is enabled in standalone mode, we need to use object status rather than summary to check the deployment status
func FilterIncompleteDeploymentUsingStatus(ctx context.Context, apiclient *ApiClient, namespace string, objectNames []string, isInstance bool, username string, password string) ([]string, []FailedDeployment) {
	remainingObjects := make([]string, 0)
	failedDeployments := make([]FailedDeployment, 0)
	var err error
	var objectMeta model.ObjectMeta
	var status model.DeployableStatusV2
	for _, objectName := range objectNames {
		if isInstance {
			var state model.InstanceState
			state, err = (*apiclient).GetInstance(ctx, objectName, namespace, username, password)
			objectMeta = state.ObjectMeta
			status = state.Status
		} else {
			var state model.TargetState
			state, err = (*apiclient).GetTarget(ctx, objectName, namespace, username, password)
			objectMeta = state.ObjectMeta
			status = state.Status
		}
		// TODO: check error code
		if err != nil {
			remainingObjects = append(remainingObjects, objectName)
			continue
		}
		if !DetermineObjectTerminalStatus(objectMeta, status) {
			remainingObjects = append(remainingObjects, objectName)
		} else if status.Status == "Failed" {
			targetErrors := make([]string, 0)
			for _, result := range status.ProvisioningStatus.Error.Details {
				targetErrors = append(targetErrors, fmt.Sprintf("%s: \"%s\"", result.Target, result.Message))
			}
			failedDeployments = append(failedDeployments, FailedDeployment{Name: objectName, Message: strings.Join(targetErrors, "; ")})
		}
	}
	return remainingObjects, failedDeployments
}

func FilterIncompleteDeploymentUsingSummary(ctx context.Context, apiclient *ApiClient, namespace string, objects []ObjectInfo, isInstance bool, username string, password string) ([]ObjectInfo, []FailedDeployment) {
	remainingObjects := make([]ObjectInfo, 0)
	failedDeployments := make([]FailedDeployment, 0)
	var err error
	for _, object := range objects {
		var key string
		var nameKey string
		if isInstance {
			key = object.SummaryId
			nameKey = object.Name
		} else {
			key = GetTargetRuntimeKey(object.SummaryId)
			nameKey = GetTargetRuntimeKey(object.Name)
		}
		jobId := object.SummaryJobId
		var summary *model.SummaryResult
		summary, err = (*apiclient).GetSummary(ctx, key, nameKey, namespace, username, password)
		// TODO: summary.Summary.JobID may be empty in standalone
		if err != nil || summary == nil {
			remainingObjects = append(remainingObjects, object)
			continue
		}

		if jobId == "" {
			jobId = "-1"
		}
		jobIdInt, err := strconv.Atoi(jobId)
		if err != nil {
			log.DebugfCtx(ctx, "Failed to convert jobId %s to int: %s", jobId, err.Error())
			remainingObjects = append(remainingObjects, object)
			continue
		}
		summaryJobIdInt, err := strconv.Atoi(summary.Summary.JobID)
		if err != nil {
			log.DebugfCtx(ctx, "Failed to convert summaryJobId %s to int: %s", summary.Summary.JobID, err.Error())
			remainingObjects = append(remainingObjects, object)
			continue
		}
		log.DebugfCtx(ctx, "Getting job id as %s from resource and %s from summary", jobId, summary.Summary.JobID)

		if err == nil && summary.State == model.SummaryStateDone && summaryJobIdInt > jobIdInt {
			if !summary.Summary.AllAssignedDeployed {
				errMsg := summary.Summary.GenerateErrorMessage()
				log.DebugfCtx(ctx, "Summary for %s is not fully deployed with error %s", object.Name, errMsg)
				failedDeployments = append(failedDeployments, FailedDeployment{Name: object.Name, Message: errMsg})
			}
			log.DebugfCtx(ctx, "Object for %s is done: with remainingObjects: %d and failedDeployments: %d.", object.Name, len(remainingObjects), len(failedDeployments))
			continue
		}
		remainingObjects = append(remainingObjects, object)
	}
	return remainingObjects, failedDeployments
}

func FilterIncompleteDelete(ctx context.Context, apiclient *ApiClient, namespace string, objectNames []string, isInstance bool, username string, password string) []string {
	remainingObjects := make([]string, 0)
	var err error
	for _, objectName := range objectNames {
		if isInstance {
			_, err = (*apiclient).GetInstance(ctx, objectName, namespace, username, password)

		} else {
			_, err = (*apiclient).GetTarget(ctx, objectName, namespace, username, password)
		}
		//if err != nil && IsNotFound(err) {
		if err != nil && strings.Contains(err.Error(), "Not Found") {
			continue
		}
		remainingObjects = append(remainingObjects, objectName)
	}
	return remainingObjects
}
