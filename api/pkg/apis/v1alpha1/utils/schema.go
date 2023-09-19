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
	"fmt"
	"regexp"
	"strconv"

	coa_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type Rule struct {
	Type       string `json:"type,omitempty"`
	Required   bool   `json:"required,omitempty"`
	Pattern    string `json:"pattern,omitempty"`
	Expression string `json:"expression,omitempty"`
}
type Schema struct {
	Rules map[string]Rule `json:"rules,omitempty"`
}

type RuleResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

type SchemaResult struct {
	Valid  bool                  `json:"valid"`
	Errors map[string]RuleResult `json:"errors,omitempty"`
}

func (s *Schema) CheckProperties(properties map[string]interface{}, evaluationContext *coa_utils.EvaluationContext) (SchemaResult, error) {
	context := evaluationContext
	if context == nil {
		context = &coa_utils.EvaluationContext{}
	}
	ret := SchemaResult{Valid: true, Errors: make(map[string]RuleResult)}
	for k, v := range s.Rules {
		if v.Type != "" {
			if val, ok := properties[k]; ok {
				if v.Type == "int" {
					if _, err := strconv.Atoi(val.(string)); err != nil {
						ret.Valid = false
						ret.Errors[k] = RuleResult{Valid: false, Error: "property is not an int"}
					}
				} else if v.Type == "float" {
					if _, err := strconv.ParseFloat(val.(string), 64); err != nil {
						ret.Valid = false
						ret.Errors[k] = RuleResult{Valid: false, Error: "property is not a float"}
					}
				} else if v.Type == "bool" {
					if _, err := strconv.ParseBool(val.(string)); err != nil {
						ret.Valid = false
						ret.Errors[k] = RuleResult{Valid: false, Error: "property is not a bool"}
					}
				} else if v.Type == "uint" {
					if _, err := strconv.ParseUint(val.(string), 10, 64); err != nil {
						ret.Valid = false
						ret.Errors[k] = RuleResult{Valid: false, Error: "property is not a uint"}
					}
				} else if v.Type == "string" {
					// Do nothing
				} else {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: "unknown type"}
				}
			}
		}
		if v.Required {
			if _, ok := properties[k]; !ok {
				ret.Valid = false
				ret.Errors[k] = RuleResult{Valid: false, Error: "missing required property"}
			}
		}
		if v.Pattern != "" {
			if val, ok := properties[k]; ok {
				match, err := s.matchPattern(val.(string), v.Pattern)
				if err != nil {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: "error matching pattern: " + err.Error()}
				}
				if !match {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: fmt.Sprintf("property does not match pattern: %s", v.Pattern)}
				}
			}
		}
		if v.Expression != "" {
			if val, ok := properties[k]; ok {
				context.Value = val
				parser := NewParser(v.Expression)
				res, err := parser.Eval(*context)
				if err != nil {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: "error evaluating expression: " + err.Error()}
				}
				if res != "true" && res != "false" && res != true && res != false {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: "expression does not evaluate to boolean"}
				}
				if res != "true" && res != true {
					ret.Valid = false
					ret.Errors[k] = RuleResult{Valid: false, Error: fmt.Sprintf("property does not match expression: %s", v.Expression)}
				}
			}
		}
	}
	return ret, nil
}
func (s *Schema) matchPattern(value string, pattern string) (bool, error) {
	regexPattern := pattern
	switch pattern {
	case "<email>":
		regexPattern = `^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`
	case "<url>":
		regexPattern = `^https?://.*$`
	case "<uuid>":
		regexPattern = `^[a-f\d]{8}(-[a-f\d]{4}){4}[a-f\d]{8}$`
	case "<dns-label>":
		regexPattern = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	case "<dns-name>":
		regexPattern = `^([a-z0-9]([-a-z0-9]*[a-z0-9])?\.)+[a-z]{2,}$`
	case "<ip4>":
		regexPattern = `^(\d{1,3}\.){3}\d{1,3}$`
	case "<ip4-range>":
		regexPattern = `^(\d{1,3}\.){3}\d{1,3}-(\d{1,3}\.){3}\d{1,3}$`
	case "<port>":
		regexPattern = `^\d{1,5}$`
	case "<mac-address>":
		regexPattern = `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	case "<cidr>":
		regexPattern = `^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`
	case "<ip6>":
		regexPattern = `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`
	case "<ip6-range>":
		regexPattern = `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}-([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`
	}
	matched, err := regexp.MatchString(regexPattern, value)
	if err != nil {
		return false, err
	}
	return matched, nil
}
