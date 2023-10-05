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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
)

type Token int

const (
	EOF Token = iota
	NUMBER
	DOLLAR
	IDENT
	OPAREN
	CPAREN
	OBRACKET
	CBRACKET
	OCURLY
	CCURLY
	PLUS
	MINUS
	MULT
	DIV
	COMMA
	PERIOD
	COLON
	QUESTION
	EQUAL
	STRING
	RUNON
	AMPHERSAND
	SLASH
	TILDE
)

var opNames = map[Token]string{
	PLUS:       "+",
	MINUS:      "-",
	MULT:       "*",
	DIV:        "/",
	SLASH:      "\\",
	COMMA:      ",",
	PERIOD:     ".",
	COLON:      ":",
	QUESTION:   "?",
	EQUAL:      "=",
	AMPHERSAND: "&",
	TILDE:      "~",
}

type Node interface {
	Eval(context utils.EvaluationContext) (interface{}, error)
}

type NumberNode struct {
	Value float64
}

func (n *NumberNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	return n.Value, nil
}

type IdentifierNode struct {
	Value string
}

func removeQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	first := s[0]
	last := s[len(s)-1]
	if first == '\'' && last == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

func (n *IdentifierNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	return removeQuotes(n.Value), nil
}

type NullNode struct {
}

func (n *NullNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	return "", nil
}

type UnaryNode struct {
	Op   Token
	Expr Node
}

func (n *UnaryNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	switch n.Op {
	case PLUS:
		if n.Expr != nil {
			return n.Expr.Eval(context)
		}
		return "", nil
	case MINUS:
		if n.Expr != nil {
			val, err := n.Expr.Eval(context)
			if err != nil {
				return val, err
			}
			if v, ok := val.(float64); ok {
				return -v, nil
			}
			return fmt.Sprintf("-%v", val), nil
		}
		return "", nil
	case OBRACKET:
		val, err := n.Expr.Eval(context)
		if err != nil {
			return val, err
		}
		return fmt.Sprintf("[%v]", val), nil
	case OCURLY:
		val, err := n.Expr.Eval(context)
		if err != nil {
			return val, err
		}
		return fmt.Sprintf("{%v}", val), nil
	}
	return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
}

type BinaryNode struct {
	Op    Token
	Left  Node
	Right Node
}

func (n *BinaryNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	switch n.Op {
	case PLUS:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		lv_f, okl := lv.(float64)
		rv_f, okr := rv.(float64)
		if okl && okr {
			v := lv_f + rv_f
			return v, nil
		} else if okl {
			lv_str := strconv.FormatFloat(lv_f, 'f', -1, 64)
			return fmt.Sprintf("%v%v", lv_str, rv), nil
		} else if okr {
			rv_str := strconv.FormatFloat(rv_f, 'f', -1, 64)
			return fmt.Sprintf("%v%v", lv, rv_str), nil
		} else {
			return fmt.Sprintf("%v%v", lv, rv), nil
		}
	case MINUS:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			v := vl - vr
			return v, nil
		} else {
			return fmt.Sprintf("%v-%v", lv, rv), nil
		}
	case COMMA:
		lv, le := n.Left.Eval(context)
		if le != nil {
			return nil, le
		}
		rv, re := n.Right.Eval(context)
		if re != nil {
			return nil, re
		}
		return fmt.Sprintf("%v,%v", lv, rv), nil
	case MULT:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			v := vl * vr
			return v, nil
		} else {
			if !okl && okr {
				if vr > 0 {
					return strings.Repeat(fmt.Sprintf("%v", lv), int(vr)), nil
				} else if vr == 0 {
					return "", nil
				} else {
					return fmt.Sprintf("%v*%v", lv, rv), nil
				}
			} else {
				return fmt.Sprintf("%v*%v", lv, rv), nil
			}
		}
	case DIV:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			if vr != 0 {
				v := vl / vr
				return v, nil
			} else {
				return nil, errors.New("divide by zero")
			}
		} else {
			return fmt.Sprintf("%v/%v", lv, rv), nil
		}
	case SLASH:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v\\%v", lv, rv), nil
	case PERIOD:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			return fmt.Sprintf("%s.%s", strconv.FormatFloat(vl, 'f', -1, 64), strconv.FormatFloat(vr, 'f', -1, 64)), nil
		} else if okl {
			return fmt.Sprintf("%s.%v", strconv.FormatFloat(vl, 'f', -1, 64), rv), nil
		} else if okr {
			return fmt.Sprintf("%v.%s", lv, strconv.FormatFloat(vr, 'f', -1, 64)), nil
		}
		return fmt.Sprintf("%v.%v", lv, rv), nil
	case COLON:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v:%v", lv, rv), nil
	case QUESTION:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v?%v", lv, rv), nil
	case EQUAL:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v=%v", lv, rv), nil
	case AMPHERSAND:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v&%v", lv, rv), nil
	case TILDE:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v~%v", lv, rv), nil
	case RUNON:
		var lv interface{} = ""
		var le error
		if n.Left != nil {
			lv, le = n.Left.Eval(context)
			if le != nil {
				return nil, le
			}
		}
		var rv interface{} = ""
		var re error
		if n.Right != nil {
			rv, re = n.Right.Eval(context)
			if re != nil {
				return nil, re
			}
		}
		return fmt.Sprintf("%v%v", lv, rv), nil
	}

	return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
}

type FunctionNode struct {
	Name string
	Args []Node
}

func readProperty(properties map[string]string, key string) (string, error) {
	if v, ok := properties[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("property %s is not found", key)
}
func readPropertyInterface(properties map[string]interface{}, key string) (interface{}, error) {
	if v, ok := properties[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("property %s is not found", key)
}
func readArgument(deployment model.DeploymentSpec, component string, key string) (string, error) {

	arguments := deployment.Instance.Arguments
	if ca, ok := arguments[component]; ok {
		if a, ok := ca[key]; ok {
			return a, nil
		}
	}
	components := deployment.Solution.Components
	for _, c := range components {
		if c.Name == component {
			if v, ok := c.Parameters[key]; ok {
				return v, nil
			}
		}
	}
	return "", fmt.Errorf("parameter %s is not found on component %s", key, component)
}

func (n *FunctionNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	switch n.Name {
	case "param":
		if len(n.Args) == 1 {
			if context.Component == "" {
				return nil, errors.New("a component name is needed to evaluate $param()")
			}
			key, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			if deploymentSpec, ok := context.DeploymentSpec.(model.DeploymentSpec); ok {
				argument, err := readArgument(deploymentSpec, context.Component, key.(string))
				if err != nil {
					return nil, err
				}
				return argument, nil
			}
			return nil, errors.New("deployment spec is not found")
		}
		return nil, fmt.Errorf("$params() expects 1 argument, found %d", len(n.Args))
	case "property":
		if len(n.Args) == 1 {
			if context.Properties == nil || len(context.Properties) == 0 {
				return nil, errors.New("a property collection is needed to evaluate $property()")
			}
			key, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			property, err := readProperty(context.Properties, key.(string))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, fmt.Errorf("$property() expects 1 argument, found %d", len(n.Args))
	case "input":
		if len(n.Args) == 1 {
			if context.Inputs == nil || len(context.Inputs) == 0 {
				return nil, errors.New("an input collection is needed to evaluate $input()")
			}
			key, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			property, err := readPropertyInterface(context.Inputs, key.(string))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, fmt.Errorf("$input() expects 1 argument, found %d", len(n.Args))
	case "output":
		if len(n.Args) == 2 {
			if context.Outputs == nil || len(context.Outputs) == 0 {
				//return nil, errors.New("an output collection is needed to evaluate $output()")
				return "", nil
			}
			step, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			key, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			if _, ok := context.Outputs[step.(string)]; !ok {
				return nil, fmt.Errorf("step %s is not found in output collection", step.(string))
			}
			property, err := readPropertyInterface(context.Outputs[step.(string)], key.(string))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, fmt.Errorf("$output() expects 2 argument, found %d", len(n.Args))
	case "equal":
		if len(n.Args) == 2 {
			v1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			v2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			return compareInterfaces(v1, v2), nil
		}
		return nil, fmt.Errorf("$equal() expects 2 arguments, found %d", len(n.Args))
	case "and":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			return andBools(val1, val2)
		}
		return nil, fmt.Errorf("$and() expects 2 arguments, found %d", len(n.Args))
	case "or":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			return orBools(val1, val2)
		}
		return nil, fmt.Errorf("$or() expects 2 arguments, found %d", len(n.Args))
	case "not":
		if len(n.Args) == 1 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			return notBool(val)
		}
		return nil, fmt.Errorf("$not() expects 1 argument, found %d", len(n.Args))
	case "gt":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			if fVal1, ok1 := toNumber(val1); ok1 {
				if fVal2, ok2 := toNumber((val2)); ok2 {
					return fVal1 > fVal2, nil
				}
				return nil, fmt.Errorf("%v is not a valid number", val2)
			}
			return nil, fmt.Errorf("%v is not a valid number", val1)
		}
		return nil, fmt.Errorf("$gt() expects 2 arguments, found %d", len(n.Args))
	case "ge":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			if fVal1, ok1 := toNumber(val1); ok1 {
				if fVal2, ok2 := toNumber((val2)); ok2 {
					return fVal1 >= fVal2, nil
				}
				return nil, fmt.Errorf("%v is not a valid number", val2)
			}
			return nil, fmt.Errorf("%v is not a valid number", val1)
		}
		return nil, fmt.Errorf("$ge() expects 2 arguments, found %d", len(n.Args))
	case "if":
		if len(n.Args) == 3 {
			cond, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			if fmt.Sprintf("%v", cond) == "true" {
				return n.Args[1].Eval(context)
			} else {
				return n.Args[2].Eval(context)
			}
		}
		return nil, fmt.Errorf("$if() expects 3 arguments, found %d", len(n.Args))
	case "in":
		if len(n.Args) >= 2 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			for i := 1; i < len(n.Args); i++ {
				v, err := n.Args[i].Eval(context)
				if err != nil {
					return nil, err
				}
				if fmt.Sprintf("%v", val) == fmt.Sprintf("%v", v) {
					return true, nil
				}
			}
			return false, nil
		}
		return nil, fmt.Errorf("$in() expects at least 2 arguments, found %d", len(n.Args))
	case "lt":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			if fVal1, ok1 := toNumber(val1); ok1 {
				if fVal2, ok2 := toNumber((val2)); ok2 {
					return fVal1 < fVal2, nil
				}
				return nil, fmt.Errorf("%v is not a valid number", val2)
			}
			return nil, fmt.Errorf("%v is not a valid number", val1)
		}
		return nil, fmt.Errorf("$lt() expects 2 arguments, found %d", len(n.Args))
	case "between":
		if len(n.Args) == 3 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			val3, err := n.Args[2].Eval(context)
			if err != nil {
				return nil, err
			}
			if fVal1, ok1 := toNumber(val1); ok1 {
				if fVal2, ok2 := toNumber((val2)); ok2 {
					if fVal3, ok2 := toNumber((val3)); ok2 {
						return fVal1 >= fVal2 && fVal1 <= fVal3, nil
					}
					return nil, fmt.Errorf("%v is not a valid number", val3)
				}
				return nil, fmt.Errorf("%v is not a valid number", val2)
			}
			return nil, fmt.Errorf("%v is not a valid number", val1)
		}
		return nil, fmt.Errorf("$le() expects 2 arguments, found %d", len(n.Args))
	case "le":
		if len(n.Args) == 2 {
			val1, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			val2, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			if fVal1, ok1 := toNumber(val1); ok1 {
				if fVal2, ok2 := toNumber((val2)); ok2 {
					return fVal1 <= fVal2, nil
				}
				return nil, fmt.Errorf("%v is not a valid number", val2)
			}
			return nil, fmt.Errorf("%v is not a valid number", val1)
		}
		return nil, fmt.Errorf("$le() expects 2 arguments, found %d", len(n.Args))
	case "config":
		if len(n.Args) >= 2 {
			if context.ConfigProvider == nil {
				return nil, errors.New("a config provider is needed to evaluate $config()")
			}
			obj, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			field, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}

			var overlays []string
			if len(n.Args) > 2 {
				for i := 2; i < len(n.Args); i++ {
					overlay, err := n.Args[i].Eval(context)
					if err != nil {
						return nil, err
					}
					overlays = append(overlays, overlay.(string))
				}
			}

			return context.ConfigProvider.Get(obj.(string), field.(string), overlays, context)
		}
		return nil, fmt.Errorf("$config() expects 2 arguments, found %d", len(n.Args))
	case "secret":
		if len(n.Args) == 2 {
			if context.SecretProvider == nil {
				return nil, errors.New("a secret provider is needed to evaluate $secret()")
			}
			obj, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			field, err := n.Args[1].Eval(context)
			if err != nil {
				return nil, err
			}
			return context.SecretProvider.Get(obj.(string), field.(string))
		}
		return nil, fmt.Errorf("$secret() expects 2 arguments, found %d", len(n.Args))
	case "instance":
		if len(n.Args) == 0 {
			if deploymentSpec, ok := context.DeploymentSpec.(model.DeploymentSpec); ok {
				return deploymentSpec.Instance.Name, nil
			}
			return nil, errors.New("deployment spec is not found")
		}
		return nil, fmt.Errorf("$instance() expects 0 arguments, found %d", len(n.Args))
	case "val", "context":
		if len(n.Args) == 0 {
			return context.Value, nil
		}
		if len(n.Args) == 1 {
			obj, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			path := obj.(string)
			if strings.HasPrefix(path, "$") || strings.HasPrefix(path, "{$") {
				result, err := JsonPathQuery(context.Value, obj.(string))
				if err != nil {
					return nil, err
				}
				return result, nil
			} else {
				if mobj, ok := context.Value.(map[string]interface{}); ok {
					if v, ok := mobj[path]; ok {
						return v, nil
					} else {
						return nil, fmt.Errorf("key %s is not found in context value", path)
					}
				} else {
					return nil, fmt.Errorf("context value '%v' is not a map", context.Value)
				}
			}
		}
		return nil, fmt.Errorf("$val() or $context() expects 0 or 1 argument, found %d", len(n.Args))
	case "json":
		if len(n.Args) == 1 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			jData, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}
			return string(jData), nil
		}
		return nil, fmt.Errorf("$json() expects 1 argument, fount %d", len(n.Args))
	}
	return nil, fmt.Errorf("invalid function name: '%s'", n.Name)
}

type Parser struct {
	s            *scanner.Scanner
	token        Token
	text         string
	OriginalText string
}

func NewParser(text string) *Parser {
	var s scanner.Scanner // TODO: this is mostly used to scan go code, we should use a custom scanner
	s.Init(strings.NewReader(strings.TrimSpace(text)))
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanStrings | scanner.ScanInts
	p := &Parser{
		s:            &s,
		text:         text,
		OriginalText: strings.TrimSpace(text),
	}
	p.next()
	return p
}

func (p *Parser) Eval(context utils.EvaluationContext) (interface{}, error) {
	var ret interface{}
	for {
		n, err := p.expr(false)
		if err != nil {
			return p.OriginalText, nil //can't be interpreted as an expression, return the original text
		}
		if _, ok := n.(*NullNode); !ok {
			v, r := n.Eval(context)
			if r != nil {
				return "", r
			}
			if vt, ok := v.([]string); ok {
				if ret == nil {
					ret = vt
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else if vt, ok := v.([]interface{}); ok {
				if ret == nil {
					ret = vt
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else if vt, ok := v.(map[string]interface{}); ok {
				if ret == nil {
					ret = vt
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else {
				if ret == nil {
					ret = fmt.Sprintf("%v", v)
				} else {
					ret = fmt.Sprintf("%v%v", ret, v)
				}
			}
		} else {
			return ret, nil
		}
		p.next()
	}
}

func (p *Parser) next() {
	p.token = p.scan()
}

func (p *Parser) scan() Token {
	tok := p.s.Scan()
	p.text = p.s.TokenText()
	switch tok {
	case scanner.EOF:
		return EOF
	case scanner.Float:
		return NUMBER
	case scanner.Ident:
		return IDENT
	case '$':
		return DOLLAR
	case '(':
		return OPAREN
	case ')':
		return CPAREN
	case '[':
		return OBRACKET
	case ']':
		return CBRACKET
	case '{':
		return OCURLY
	case '}':
		return CCURLY
	case '+':
		return PLUS
	case '-':
		return MINUS
	case '*':
		return MULT
	case '/':
		return DIV
	case '\\':
		return SLASH
	case ',':
		return COMMA
	case '.':
		return PERIOD
	case ':':
		return COLON
	case '?':
		return QUESTION
	case '=':
		return EQUAL
	case '&':
		return AMPHERSAND
	case '~':
		return TILDE
	}
	if _, err := strconv.ParseFloat(p.text, 64); err == nil {
		return NUMBER
	}
	return IDENT
}

func (p *Parser) match(t Token) error {
	if p.token == t {
		p.next()
	} else {
		return fmt.Errorf("expected %T, got %s", t, p.text)
	}
	return nil
}

func (p *Parser) primary() (Node, error) {
	switch p.token {
	case NUMBER:
		v, _ := strconv.ParseFloat(p.text, 64)
		p.next()
		return &NumberNode{v}, nil
	case DOLLAR:
		return p.function()
	case OPAREN:
		p.next()
		node, err := p.expr(false)
		if err != nil {
			return nil, err
		}
		expr := node
		if err := p.match(CPAREN); err != nil {
			return nil, err
		}
		return expr, nil
	case OBRACKET:
		p.next()
		node, err := p.expr(false)
		if err != nil {
			return nil, err
		}
		bexpr := node
		if err := p.match(CBRACKET); err != nil {
			return nil, err
		}
		return &UnaryNode{OBRACKET, bexpr}, nil
	case OCURLY:
		p.next()
		node, err := p.expr(false)
		if err != nil {
			return nil, err
		}
		cexpr := node
		if err := p.match(CCURLY); err != nil {
			return nil, err
		}
		return &UnaryNode{OCURLY, cexpr}, nil
	case PLUS:
		p.next()
		node, err := p.primary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{PLUS, node}, nil
	case MINUS:
		p.next()
		node, err := p.primary()
		if err != nil {
			return nil, err
		}
		return &UnaryNode{MINUS, node}, nil
	case IDENT:
		v := p.text
		p.next()
		return &IdentifierNode{v}, nil
	}
	return nil, nil
}

func (p *Parser) factor() (Node, error) {
	node, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.token {
		case MULT:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{MULT, node, n}
		case DIV:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{DIV, node, n}
		case SLASH:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{SLASH, node, n}
		case PERIOD:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{PERIOD, node, n}
		case COLON:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{COLON, node, n}
		case QUESTION:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{QUESTION, node, n}
		case EQUAL:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{EQUAL, node, n}
		case TILDE:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{TILDE, node, n}
		case AMPHERSAND:
			p.next()
			n, err := p.primary()
			if err != nil {
				return nil, err
			}
			node = &BinaryNode{AMPHERSAND, node, n}
		default:
			return node, nil
		}
	}
}

func (p *Parser) expr(inFunc bool) (Node, error) {
	node, err := p.factor()
	if node == nil || err != nil {
		return &NullNode{}, err
	}
	for {
		switch p.token {
		case PLUS:
			p.next()
			f, err := p.factor()
			if err != nil {
				return &NullNode{}, err
			}
			node = &BinaryNode{PLUS, node, f}
		case MINUS:
			p.next()
			f, err := p.factor()
			if err != nil {
				return &NullNode{}, err
			}
			node = &BinaryNode{MINUS, node, f}
		case COMMA:
			if !inFunc {
				p.next()
				f, err := p.factor()
				if err != nil {
					return &NullNode{}, err
				}
				node = &BinaryNode{COMMA, node, f}
			} else {
				return node, nil
			}
		case OPAREN:
			p.next()
			node, err := p.expr(false)
			if err != nil {
				return nil, err
			}
			expr := node
			if err := p.match(CPAREN); err != nil {
				return nil, err
			}
			return expr, nil
		default:
			return node, nil
		}
	}
}

func (p *Parser) function() (Node, error) {
	err := p.match(DOLLAR)
	if err != nil {
		return nil, err
	}
	name := p.text
	err = p.match(IDENT)
	if err != nil {
		return nil, err
	}
	err = p.match(OPAREN)
	if err != nil {
		return nil, err
	}
	args := []Node{}
	for p.token != CPAREN {
		node, err := p.expr(true)
		if err != nil {
			return nil, err
		}
		if _, ok := node.(*NullNode); ok {
			return nil, fmt.Errorf("invalid argument")
		}
		args = append(args, node)
		if p.token == COMMA {
			p.next()
		}
	}
	err = p.match(CPAREN)
	if err != nil {
		return nil, err
	}
	return &FunctionNode{name, args}, nil
}

func EvaluateDeployment(context utils.EvaluationContext) (model.DeploymentSpec, error) {
	if deploymentSpec, ok := context.DeploymentSpec.(model.DeploymentSpec); ok {
		for ic, c := range deploymentSpec.Solution.Components {

			val, err := evalProperties(context, c.Metadata)
			if err != nil {
				return deploymentSpec, err
			}
			if val != nil {
				metadata, ok := val.(map[string]string)
				if !ok {
					return deploymentSpec, fmt.Errorf("metadata must be a map")
				}
				stringMap := make(map[string]string)
				for k, v := range metadata {
					stringMap[k] = fmt.Sprintf("%v", v)
				}
				deploymentSpec.Solution.Components[ic].Metadata = stringMap
			}

			val, err = evalProperties(context, c.Properties)
			if err != nil {
				return deploymentSpec, err
			}
			props, ok := val.(map[string]interface{})
			if !ok {
				return deploymentSpec, fmt.Errorf("properties must be a map")
			}
			deploymentSpec.Solution.Components[ic].Properties = props
		}
		return deploymentSpec, nil
	}
	return model.DeploymentSpec{}, errors.New("deployment spec is not found")
}
func compareInterfaces(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
func andBools(a, b interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		if bBool, ok := toBool(b); ok {
			return aBool && bBool, nil
		}
		return false, fmt.Errorf("%v is not a boolean value", b)
	}
	return false, fmt.Errorf("%v is not a boolean value", a)
}
func orBools(a, b interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		if bBool, ok := toBool(b); ok {
			return aBool || bBool, nil
		}
		return false, fmt.Errorf("%v is not a boolean value", b)
	}
	return false, fmt.Errorf("%v is not a boolean value", a)
}
func notBool(a interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		return !aBool, nil
	}
	return false, fmt.Errorf("%v is not a boolean value", a)
}
func toBool(val interface{}) (bool, bool) {
	switch val := val.(type) {
	case bool:
		return val, true
	case string:
		boolVal, err := strconv.ParseBool(val)
		if err == nil {
			return boolVal, true
		}
	}
	return false, false
}
func toNumber(val interface{}) (float64, bool) {
	num, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
	if err == nil {
		return num, true
	}
	return 0, false
}
func evalProperties(context utils.EvaluationContext, properties interface{}) (interface{}, error) {
	switch p := properties.(type) {
	case map[string]string:
		for k, v := range p {
			val, err := evalProperties(context, v)
			if err != nil {
				return nil, err
			}
			p[k] = FormatAsString(val)
		}
	case map[string]interface{}:
		for k, v := range p {
			val, err := evalProperties(context, v)
			if err != nil {
				return nil, err
			}
			p[k] = val
		}
	case []interface{}:
		for i, v := range p {
			val, err := evalProperties(context, v)
			if err != nil {
				return nil, err
			}
			p[i] = val
		}
	case string:
		var js interface{}
		err := json.Unmarshal([]byte(p), &js)
		if err == nil {
			modified, err := enumerateProperties(js, context)
			if err != nil {
				return nil, err
			}
			jsBytes, err := json.Marshal(modified)
			if err != nil {
				return nil, err
			}
			return string(jsBytes), nil
		}
		parser := NewParser(p)
		val, err := parser.Eval(context)
		if err != nil {
			return nil, err
		}
		properties = val
	}
	return properties, nil
}

func enumerateProperties(js interface{}, context utils.EvaluationContext) (interface{}, error) {
	switch v := js.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if strVal, ok := val.(string); ok {
				parser := NewParser(strVal)
				val, err := parser.Eval(context)
				if err != nil {
					return nil, err
				}
				v[key] = val
			} else {
				nestedProps, err := enumerateProperties(val, context)
				if err != nil {
					return nil, err
				}
				v[key] = nestedProps
			}
		}
	case []interface{}:
		for i, val := range v {
			nestedProps, err := enumerateProperties(val, context)
			if err != nil {
				return nil, err
			}
			v[i] = nestedProps
		}
	}
	return js, nil
}
