/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package states

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/yalp/jsonpath"
)

func JsonPathMatch(jsonData interface{}, path string, target string) bool {
	// var data interface{}
	// if err := json.Unmarshal(jsonData, &data); err != nil {
	// 	return false
	// }
	res, err := jsonpath.Read(jsonData, path)
	if err != nil {
		return false
	}
	return res.(string) == target
}

func MatchFilter(entry StateEntry, filterType string, filterValue string) (bool, error) {
	var dict map[string]interface{}
	j, _ := json.Marshal(entry.Body)
	err := json.Unmarshal(j, &dict)
	if err != nil {
		err = v1alpha2.NewCOAError(nil, "failed to unmarshal state entry when applying filter", v1alpha2.InternalError)
		return false, err
	}
	switch filterType {
	case "label":
		if dict["metadata"] != nil {
			metadata, ok := dict["metadata"].(map[string]interface{})
			if ok {
				if metadata["labels"] != nil {
					labels, ok := metadata["labels"].(map[string]interface{})
					if ok {
						var match bool
						match, err = InMemoryFilter(labels, filterValue)
						if err != nil {
							return false, err
						}
						return match, nil
					}
				}
			}
		}
	case "field":
		var match bool
		match, err = InMemoryFilter(dict, filterValue)
		if err != nil {
			return false, err
		}
		return match, nil
	case "status":
		if dict["status"] != nil {
			var statusdict map[string]interface{}
			j, _ := json.Marshal(dict["status"])
			err = json.Unmarshal(j, &statusdict)
			if err != nil {
				err = v1alpha2.NewCOAError(nil, "failed to unmarshal status field when applying filter", v1alpha2.InternalError)
				return false, err
			}

			if v, e := utils.JsonPathQuery(statusdict, filterValue); e != nil || v == nil {
				return false, nil
			}
			return true, nil

		}
	case "spec":
		if dict["spec"] != nil {
			spec, ok := dict["spec"].(map[string]interface{})
			if ok {
				if v, e := utils.JsonPathQuery(spec, filterValue); e != nil || v == nil {
					return false, nil
				}
				return true, nil
			}
			return false, nil
		}
	default:
		return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter type '%s' is not supported", filterType), v1alpha2.BadRequest)
	}
	return false, nil
}
func InMemoryFilter(entity map[string]interface{}, filter string) (bool, error) {
	parts := strings.Split(filter, ",")
	for _, part := range parts {
		match, err := inMemoryFilterSingleKey(entity, part)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

func inMemoryFilterSingleKey(entity map[string]interface{}, filter string) (bool, error) {
	if strings.Index(filter, "!=") > 0 {
		parts := strings.Split(filter, "!=")
		if len(parts) == 2 {
			dict, key, err := traceDownField(entity, parts[0])
			if err != nil {
				// TODO: bad config could happen if the field is not set for the entry, suppress for now to avoid confusion
				if v1alpha2.IsBadConfig(err) {
					return false, nil
				}
				return false, err
			}
			if dict[key] != nil {
				if dict[key] != parts[1] {
					return true, nil
				}
			}
			return false, nil
		} else {
			return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter '%s' is not a valid selector", filter), v1alpha2.BadRequest)
		}
	} else if strings.Index(filter, "==") > 0 {
		parts := strings.Split(filter, "==")
		if len(parts) == 2 {
			dict, key, err := traceDownField(entity, parts[0])
			if err != nil {
				// TODO: bad config could happen if the field is not set for the entry, suppress for now to avoid confusion
				if v1alpha2.IsBadConfig(err) {
					return false, nil
				}
				return false, err
			}
			if dict[key] != nil {
				if dict[key] == parts[1] {
					return true, nil
				}
			}
			return false, nil
		} else {
			return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter '%s' is not a valid selector", filter), v1alpha2.BadRequest)
		}
	} else if strings.Index(filter, "=") > 0 {
		parts := strings.Split(filter, "=")
		if len(parts) == 2 {
			dict, key, err := traceDownField(entity, parts[0])
			if err != nil {
				// TODO: bad config could happen if the field is not set for the entry, suppress for now to avoid confusion
				if v1alpha2.IsBadConfig(err) {
					return false, nil
				}
				return false, err
			}
			if dict[key] != nil {
				if dict[key] == parts[1] {
					return true, nil
				}
			}
			return false, nil
		} else {
			return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter '%s' is not a valid selector", filter), v1alpha2.BadRequest)
		}
	} else {
		return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter '%s' is not a valid selector", filter), v1alpha2.BadRequest)
	}
}

func traceDownField(entity map[string]interface{}, filter string) (map[string]interface{}, string, error) {
	if !strings.Contains(filter, ".") {
		return entity, filter, nil
	}
	parts := strings.Split(filter, ".")
	if v, ok := entity[parts[0]]; ok {
		var dict = make(map[string]interface{})
		j, _ := json.Marshal(v)
		err := json.Unmarshal(j, &dict)
		if err != nil {
			return nil, filter, err
		}
		return traceDownField(dict, strings.Join(parts[1:], "."))
	} else {
		return nil, filter, v1alpha2.NewCOAError(nil, fmt.Sprintf("filter '%s' is not a valid selector", filter), v1alpha2.BadConfig)
	}
}
