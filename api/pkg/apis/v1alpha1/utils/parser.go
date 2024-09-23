/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

type Token int

const (
	EOF Token = iota
	NUMBER
	INT
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

type IntNode struct {
	Value int64
}

func (n *IntNode) Eval(context utils.EvaluationContext) (interface{}, error) {
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
			if v, ok := val.(int64); ok {
				return -v, nil
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
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("operator '%s' is not allowed in this context", opNames[n.Op]), v1alpha2.BadConfig)
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
		return formatFloats(lv, rv, ""), nil
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
		return formatFloats(lv, rv, "-"), nil
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
		return formatFloats(lv, rv, "*"), nil
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
		return formatFloats(lv, rv, "/"), nil
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
		return formatFloats(lv, rv, "."), nil
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

	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("operator '%s' is not allowed in this context", opNames[n.Op]), v1alpha2.BadConfig)
}

type FunctionNode struct {
	Name string
	Args []Node
}

func readProperty(properties map[string]string, key string) (string, error) {
	if v, ok := properties[key]; ok {
		return v, nil
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("property %s is not found", key), v1alpha2.BadConfig)
}
func readPropertyInterface(properties map[string]interface{}, key string) (interface{}, error) {
	if v, ok := properties[key]; ok {
		return v, nil
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("property %s is not found", key), v1alpha2.BadConfig)
}
func readArgument(deployment model.DeploymentSpec, component string, key string) (string, error) {
	components := deployment.Solution.Spec.Components
	for _, c := range components {
		if c.Name == component {
			if v, ok := c.Parameters[key]; ok {
				return v, nil
			}
		}
	}
	return "", v1alpha2.NewCOAError(nil, fmt.Sprintf("parameter %s is not found on component %s", key, component), v1alpha2.BadConfig)
}

func toIntIfPossible(f float64) interface{} {
	i := int64(f)
	if float64(i) == f {
		return i
	}
	return f
}

func formatFloats(left interface{}, right interface{}, operator string) interface{} {
	var lv_f, rv_f float64
	var okl, okr bool
	if lv_i, ok := left.(int64); ok {
		lv_f = float64(lv_i)
		okl = true
	} else {
		lv_f, okl = left.(float64)
	}
	if rv_i, ok := right.(int64); ok {
		rv_f = float64(rv_i)
		okr = true
	} else {
		rv_f, okr = right.(float64)
	}
	if okl && okr {
		switch operator {
		case "":
			return toIntIfPossible(lv_f + rv_f)
		case "-":
			return toIntIfPossible(lv_f - rv_f)
		case "*":
			return toIntIfPossible(lv_f * rv_f)
		case "/":
			if rv_f != 0 {
				return toIntIfPossible(lv_f / rv_f)
			} else {
				lv_str := strconv.FormatFloat(lv_f, 'f', -1, 64)
				rv_str := strconv.FormatFloat(rv_f, 'f', -1, 64)
				return fmt.Sprintf("%v%s%v", lv_str, operator, rv_str)
			}
		case ".":
			lv_str := strconv.FormatFloat(lv_f, 'f', -1, 64)
			rv_str := strconv.FormatFloat(rv_f, 'f', -1, 64)
			return fmt.Sprintf("%v%s%v", lv_str, operator, rv_str)
		default:
			return fmt.Errorf("operator '%s' is not allowed in this context", operator)
		}
	} else if okl {
		lv_str := strconv.FormatFloat(lv_f, 'f', -1, 64)
		return fmt.Sprintf("%v%s%v", lv_str, operator, right)
	} else if okr {
		rv_str := strconv.FormatFloat(rv_f, 'f', -1, 64)
		return fmt.Sprintf("%v%s%v", left, operator, rv_str)
	} else {
		return fmt.Sprintf("%v%s%v", left, operator, right)
	}
}

func (n *FunctionNode) Eval(context utils.EvaluationContext) (interface{}, error) {
	switch n.Name {
	case "param":
		if len(n.Args) == 1 {
			if context.Component == "" {
				return nil, v1alpha2.NewCOAError(nil, "a component name is needed to evaluate $param()", v1alpha2.BadConfig)
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
			return nil, v1alpha2.NewCOAError(nil, "deployment spec is not found", v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$params() expects 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
	case "property":
		if len(n.Args) == 1 {
			if context.Properties == nil || len(context.Properties) == 0 {
				return nil, v1alpha2.NewCOAError(nil, "a property collection is needed to evaluate $property()", v1alpha2.BadConfig)
			}
			key, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			property, err := readProperty(context.Properties, FormatAsString(key))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$property() expects 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
	case "input":
		if len(n.Args) == 1 {
			if context.Inputs == nil || len(context.Inputs) == 0 {
				return nil, errors.New("an input collection is needed to evaluate $input()")
			}
			key, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			property, err := readPropertyInterface(context.Inputs, FormatAsString(key))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$input() expects 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
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
			if _, ok := context.Outputs[FormatAsString(step)]; !ok {
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("step %s is not found in output collection", FormatAsString(step)), v1alpha2.BadConfig)
			}
			property, err := readPropertyInterface(context.Outputs[FormatAsString(step)], FormatAsString(key))
			if err != nil {
				return nil, err
			}
			return property, nil
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$output() expects 2 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$equal() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$and() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$or() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
	case "not":
		if len(n.Args) == 1 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			return notBool(val)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$not() expects 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
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
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val2), v1alpha2.BadConfig)
			}
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val1), v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$gt() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val2), v1alpha2.BadConfig)
			}
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val1), v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$ge() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$if() expects 3 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$in() expects at least 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val2), v1alpha2.BadConfig)
			}
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val1), v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$lt() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
					return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val3), v1alpha2.BadConfig)
				}
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val2), v1alpha2.BadConfig)
			}
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val1), v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$le() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
				return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val2), v1alpha2.BadConfig)
			}
			return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a valid number", val1), v1alpha2.BadConfig)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$le() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
					overlays = append(overlays, FormatAsString(overlay))
				}
			}

			return context.ConfigProvider.Get(FormatAsString(obj), FormatAsString(field), overlays, context)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$config() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
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
			return context.SecretProvider.Get(FormatAsString(obj), FormatAsString(field), context)
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$secret() expects 2 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
	case "instance":
		if len(n.Args) == 0 {
			if deploymentSpec, ok := context.DeploymentSpec.(model.DeploymentSpec); ok {
				return deploymentSpec.Instance.ObjectMeta.Name, nil
			}
			return nil, errors.New("deployment spec is not found")
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$instance() expects 0 arguments, found %d", len(n.Args)), v1alpha2.BadConfig)
	case "val", "context":
		if len(n.Args) == 0 {
			return context.Value, nil
		}
		if len(n.Args) == 1 {
			obj, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			path := FormatAsString(obj)
			if strings.HasPrefix(path, "$") || strings.HasPrefix(path, "{$") {
				result, err := JsonPathQuery(context.Value, FormatAsString(obj))
				if err != nil {
					return nil, err
				}
				return result, nil
			} else {
				if mobj, ok := context.Value.(map[string]interface{}); ok {
					if v, ok := mobj[path]; ok {
						return v, nil
					} else {
						return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("key %s is not found in context value", path), v1alpha2.BadConfig)
					}
				} else {
					return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("context value '%v' is not a map", context.Value), v1alpha2.BadConfig)
				}
			}
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$val() or $context() expects 0 or 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
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
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$json() expects 1 argument, fount %d", len(n.Args)), v1alpha2.BadConfig)
	case "str":
		if len(n.Args) == 1 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("%v", val), nil
		}
		return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("$str() expects 1 argument, found %d", len(n.Args)), v1alpha2.BadConfig)
	}
	return nil, v1alpha2.NewCOAError(nil, fmt.Sprintf("invalid function name: '%s'", n.Name), v1alpha2.BadConfig)
}

type Parser struct {
	Segments     []string
	OriginalText string
}

type ExpressionParser struct {
	s     *scanner.Scanner
	token Token
	text  string
}

func NewParser(text string) *Parser {
	re := regexp.MustCompile(`(\${{.*?}})`)
	loc := re.FindAllStringIndex(text, -1)

	segments := make([]string, 0, len(loc)*2+1)
	start := 0
	for _, l := range loc {
		if start != l[0] {
			segments = append(segments, text[start:l[0]])
		}
		segments = append(segments, text[l[0]:l[1]])
		start = l[1]
	}
	if start < len(text) {
		segments = append(segments, text[start:])
	}

	p := &Parser{
		Segments: segments,
	}
	return p
}

func (p *Parser) Eval(context utils.EvaluationContext) (interface{}, error) {
	results := make([]interface{}, 0)
	for _, s := range p.Segments {
		if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
			text := s[3 : len(s)-2]
			parser := newExpressionParser(text)
			n, err := parser.Eval(context)
			if err != nil {
				return nil, err
			}
			results = append(results, n)
		} else {
			results = append(results, s)
		}
	}
	if len(results) == 1 {
		return results[0], nil
	}
	//join the results as string
	var ret interface{}
	for _, v := range results {
		if ret == nil {
			ret = fmt.Sprintf("%v", v)
		} else {
			ret = fmt.Sprintf("%v%v", ret, v)
		}
	}
	return ret, nil
}

func newExpressionParser(text string) *ExpressionParser {
	var s scanner.Scanner // TODO: this is mostly used to scan go code, we should use a custom scanner
	s.Init(strings.NewReader(strings.TrimSpace(text)))
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanStrings | scanner.ScanInts
	p := &ExpressionParser{
		s:    &s,
		text: text,
	}
	p.next()
	return p
}

func (p *ExpressionParser) Eval(context utils.EvaluationContext) (interface{}, error) {
	var ret interface{}
	for {
		n, err := p.expr(false)
		if err != nil {
			return nil, err
		}
		if _, ok := n.(*NullNode); !ok {
			v, r := n.Eval(context)
			if r != nil {
				return "", r
			}
			if vt, ok := v.([]string); ok {
				if ret == nil {
					ret = vt
				} else if vr, o := ret.([]string); o {
					vr = append(vr, vt...)
					ret = vr
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else if vt, ok := v.([]interface{}); ok {
				if ret == nil {
					ret = vt
				} else if vr, o := ret.([]interface{}); o {
					vr = append(vr, vt...)
					ret = vr
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else if vt, ok := v.(map[string]interface{}); ok {
				if ret == nil {
					ret = vt
				} else if vr, o := ret.(map[string]interface{}); o {
					for k, v := range vt {
						vr[k] = v
					}
					ret = vr
				} else {
					jData, _ := json.Marshal(v)
					ret = fmt.Sprintf("%v%v", ret, string(jData))
				}
			} else {
				if ret == nil {
					ret = v
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

func (p *ExpressionParser) next() {
	p.token = p.scan()
}

func (p *ExpressionParser) scan() Token {
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
	if _, err := strconv.ParseInt(p.text, 10, 64); err == nil {
		return INT
	}

	if _, err := strconv.ParseFloat(p.text, 64); err == nil {
		return NUMBER
	}
	return IDENT
}

func (p *ExpressionParser) match(t Token) error {
	if p.token == t {
		p.next()
	} else {
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("expected %T, got %s", t, p.text), v1alpha2.BadConfig)
	}
	return nil
}

func (p *ExpressionParser) primary() (Node, error) {
	switch p.token {
	case INT:
		v, _ := strconv.ParseInt(p.text, 10, 64)
		p.next()
		return &IntNode{v}, nil
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

func (p *ExpressionParser) factor() (Node, error) {
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

func (p *ExpressionParser) expr(inFunc bool) (Node, error) {
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

func (p *ExpressionParser) function() (Node, error) {
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
			return nil, v1alpha2.NewCOAError(nil, "invalid argument", v1alpha2.BadConfig)
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
		for ic, c := range deploymentSpec.Solution.Spec.Components {

			val, err := evalProperties(context, c.Metadata)
			if err != nil {
				log.ErrorfCtx(context.Context, " (Parser): Evaluate deployment failed: %v", err)
				return deploymentSpec, err
			}
			if val != nil {
				metadata, ok := val.(map[string]string)
				if !ok {
					err := v1alpha2.NewCOAError(nil, "metadata must be a map", v1alpha2.BadConfig)
					log.ErrorfCtx(context.Context, " (Parser): Evaluate deployment failed: %v", err)
					return deploymentSpec, err
				}
				stringMap := make(map[string]string)
				for k, v := range metadata {
					stringMap[k] = fmt.Sprintf("%v", v)
				}
				deploymentSpec.Solution.Spec.Components[ic].Metadata = stringMap
			}

			val, err = evalProperties(context, c.Properties)
			if err != nil {
				log.ErrorfCtx(context.Context, " (Parser): Evaluate deployment failed: %v", err)
				return deploymentSpec, err
			}
			props, ok := val.(map[string]interface{})
			if !ok {
				err := v1alpha2.NewCOAError(nil, "properties must be a map", v1alpha2.BadConfig)
				log.ErrorfCtx(context.Context, " (Parser): Evaluate deployment failed: %v", err)
				return deploymentSpec, err
			}
			deploymentSpec.Solution.Spec.Components[ic].Properties = props
		}
		log.DebugCtx(context.Context, " (Parser): Evaluate deployment completed.")
		return deploymentSpec, nil
	}

	err := errors.New("deployment spec is not found")
	log.ErrorfCtx(context.Context, " (Parser): Evaluate deployment failed: %v", err)
	return model.DeploymentSpec{}, err
}

func compareInterfaces(a, b interface{}) bool {
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		switch a.(type) {
		case int, int8, int16, int32, int64:
			return a.(int64) == b.(int64)
		case uint, uint8, uint16, uint32, uint64:
			return a.(uint64) == b.(uint64)
		case float32, float64:
			return math.Abs(a.(float64)-b.(float64)) < 1e-9
		case string:
			return a.(string) == b.(string)
		case bool:
			return a.(bool) == b.(bool)
		}
	}
	if aState, ok := a.(v1alpha2.State); ok {
		a = int(aState)
	}
	if bState, ok := b.(v1alpha2.State); ok {
		b = int(bState)
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
func andBools(a, b interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		if bBool, ok := toBool(b); ok {
			return aBool && bBool, nil
		}
		return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a boolean value", b), v1alpha2.BadConfig)
	}
	return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a boolean value", a), v1alpha2.BadConfig)
}
func orBools(a, b interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		if bBool, ok := toBool(b); ok {
			return aBool || bBool, nil
		}
		return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a boolean value", b), v1alpha2.BadConfig)
	}
	return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a boolean value", a), v1alpha2.BadConfig)
}
func notBool(a interface{}) (bool, error) {
	if aBool, ok := toBool(a); ok {
		return !aBool, nil
	}
	return false, v1alpha2.NewCOAError(nil, fmt.Sprintf("%v is not a boolean value", a), v1alpha2.BadConfig)
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
				log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
				return nil, err
			}
			p[k] = FormatAsString(val)
		}
	case map[string]interface{}:
		for k, v := range p {
			val, err := evalProperties(context, v)
			if err != nil {
				log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
				return nil, err
			}
			p[k] = val
		}
	case []interface{}:
		for i, v := range p {
			val, err := evalProperties(context, v)
			if err != nil {
				log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
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
				log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
				return nil, err
			}
			jsBytes, err := json.Marshal(modified)
			if err != nil {
				log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
				return nil, err
			}
			return string(jsBytes), nil
		}
		parser := NewParser(p)
		val, err := parser.Eval(context)
		if err != nil {
			log.ErrorfCtx(context.Context, " (Parser): Evaluate properties failed: %v", err)
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
