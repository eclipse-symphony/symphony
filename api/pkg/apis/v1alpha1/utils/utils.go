package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
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

func evaluateTargetCompatibility(target model.TargetSpec, component model.ComponentSpec) (int, error) {
	score1, err := evaluateConstraints(target.Constraints, component.Properties)
	if err != nil {
		return -1, err
	}
	if score1 == -1 {
		return score1, nil // target is rejected by component
	}
	if score1 < 0 {
		score1 = 0
	}
	score2, err := evaluateConstraints(component.Constraints, target.Properties)
	if err != nil {
		return -1, err
	}
	if score2 == -1 {
		return score2, nil // component is rejected by target
	}
	if score2 < 0 {
		score2 = 0
	}
	return score1 + score2, nil
}

func evaluateConstraints(constraints interface{}, properties map[string]string) (int, error) {
	data, _ := json.Marshal(constraints)
	var cons []model.ConstraintSpec
	json.Unmarshal(data, &cons)
	score := 0
	minScore := -4
	for _, c := range cons {
		score1, err := evaluateConstraint(c, properties)
		if err != nil {
			return -1, err
		}
		if score1 < 0 && score1 >= minScore {
			minScore = score1
		}
		if score1 >= 0 {
			score += score1
		}
	}
	if minScore > -4 {
		return minScore, nil
	} else {
		return score, nil
	}
}

// Evaluate a constraint. Returns:  -1 = denied; -2 = allowed; -3 = value presents but no decision is made; 0-n = preferred value
func evaluateConstraint(constraint model.ConstraintSpec, properties map[string]string) (int, error) {
	if constraint.Key == "" {
		return -1, errors.New("constraint is missing key")
	}
	if constraint.Value != "" {
		if v, ok := properties[constraint.Key]; ok && matchString(constraint.Value, v) {
			switch constraint.Qualifier {
			case Reject:
				return -1, nil
			case Prefer:
				return 1, nil
			case Must:
				return -2, nil
			case "":
				return -3, nil
			default:
				return -1, fmt.Errorf("constraint has invalid qualifier '%s'", constraint.Qualifier)
			}
		}
		switch constraint.Qualifier {
		case Reject:
			return -2, nil
		case Prefer:
			return 0, nil
		case Must:
			return -1, nil
		case "":
			return 0, nil
		default:
			return -1, fmt.Errorf("constraint has invalid qualifier '%s'", constraint.Qualifier)
		}

	} else if len(constraint.Values) > 0 && constraint.Operator != "" {
		scores := make([]int, len(constraint.Values))
		for i, c := range constraint.Values {
			var con model.ConstraintSpec
			err := json.Unmarshal([]byte(c), &con)
			if err != nil {
				return -1, err
			}
			scores[i], err = evaluateConstraint(con, properties)
			if err != nil {
				return -1, err
			}
		}
		switch constraint.Operator {
		case Any:
			switch constraint.Qualifier {
			case Must: // if any values are allowed, delayed or preferred
				for _, s := range scores {
					if s == -1 {
						return -1, nil
					}
				}
				return 0, nil
			case Prefer:
				sum := 0
				for _, s := range scores {
					if s == -1 {
						return -1, nil
					} else if s <= -2 {
						sum += 1
					} else {
						sum += s
					}
				}
				return sum, nil
			case Reject:
				for _, s := range scores {
					if s != 0 {
						return -1, nil
					}
				}
				return 0, nil
			case "":
				for _, s := range scores {
					if s == -1 {
						return -1, nil
					}
				}
				for _, s := range scores {
					if s == -3 || s == 0 {
						return -3, nil
					}
				}
				return 0, nil
			default:
				return -1, fmt.Errorf("constraint has invalid qualifier '%s'", constraint.Qualifier)
			}
		default:
			return -1, fmt.Errorf("unsupported operator '%s'", constraint.Operator)
		}

	}
	return -1, errors.New("incomplete constraint")
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
