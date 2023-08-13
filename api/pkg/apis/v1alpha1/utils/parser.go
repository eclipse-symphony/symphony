package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret"
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
}

type EvaluationContext struct {
	ConfigProvider config.IConfigProvider
	SecretProvider secret.ISecretProvider
	DeploymentSpec model.DeploymentSpec
	Properties     map[string]string
	Inputs         map[string]interface{}
	Outputs        map[string]map[string]interface{}
	Component      string
}

type Node interface {
	Eval(context EvaluationContext) (interface{}, error)
}

type NumberNode struct {
	Value float64
}

func (n *NumberNode) Eval(context EvaluationContext) (interface{}, error) {
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

func (n *IdentifierNode) Eval(context EvaluationContext) (interface{}, error) {
	return removeQuotes(n.Value), nil
}

type NullNode struct {
}

func (n *NullNode) Eval(context EvaluationContext) (interface{}, error) {
	return "", nil
}

type UnaryNode struct {
	Op   Token
	Expr Node
}

func (n *UnaryNode) Eval(context EvaluationContext) (interface{}, error) {
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

func (n *BinaryNode) Eval(context EvaluationContext) (interface{}, error) {
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
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			v := vl + vr
			return v, nil
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

func (n *FunctionNode) Eval(context EvaluationContext) (interface{}, error) {
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
			argument, err := readArgument(context.DeploymentSpec, context.Component, key.(string))
			if err != nil {
				return nil, err
			}
			return argument, nil
		}
		return nil, fmt.Errorf("$params() expects 1 argument, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$property() expects 1 argument, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$input() expects 1 argument, fount %d", len(n.Args))
	case "output":
		if len(n.Args) == 2 {
			if context.Outputs == nil || len(context.Outputs) == 0 {
				return nil, errors.New("an output collection is needed to evaluate $output()")
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
		return nil, fmt.Errorf("$output() expects 2 argument, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$equal() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$and() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$or() expects 2 arguments, fount %d", len(n.Args))
	case "not":
		if len(n.Args) == 1 {
			val, err := n.Args[0].Eval(context)
			if err != nil {
				return nil, err
			}
			return notBool(val)
		}
		return nil, fmt.Errorf("$not() expects 1 argument, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$gt() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$ge() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$if() expects 3 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$lt() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$le() expects 2 arguments, fount %d", len(n.Args))
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
		return nil, fmt.Errorf("$le() expects 2 arguments, fount %d", len(n.Args))
	case "config":
		if len(n.Args) == 2 {
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
			return context.ConfigProvider.Get(obj.(string), field.(string))
		}
		return nil, fmt.Errorf("$config() expects 2 arguments, fount %d", len(n.Args))
	case "secret":
		if len(n.Args) == 2 {
			if context.SecretProvider == nil {
				return nil, errors.New("a secret provider is needed to evaluate $config()")
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
		return nil, fmt.Errorf("$secret() expects 2 arguments, fount %d", len(n.Args))
	case "instance":
		if len(n.Args) == 0 {
			return context.DeploymentSpec.Instance.Name, nil
		}
		return nil, fmt.Errorf("$instance() expects 0 arguments, fount %d", len(n.Args))
	}
	return nil, fmt.Errorf("invalid function name: '%s'", n.Name)
}

type Parser struct {
	s     *scanner.Scanner
	token Token
	text  string
}

func NewParser(text string) *Parser {
	var s scanner.Scanner
	s.Init(strings.NewReader(strings.TrimSpace(text)))
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanStrings | scanner.ScanInts
	p := &Parser{
		s: &s,
	}
	p.next()
	return p
}

func (p *Parser) Eval(context EvaluationContext) (interface{}, error) {
	var ret interface{}
	for {
		n := p.expr(false)
		if _, ok := n.(*NullNode); !ok {
			v, r := n.Eval(context)
			if r != nil {
				return "", r
			}
			if _, ok := v.([]string); ok {
				if ret == nil {
					ret = v
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
	}
	if _, err := strconv.ParseFloat(p.text, 64); err == nil {
		return NUMBER
	}
	return IDENT
}

func (p *Parser) match(t Token) {
	if p.token == t {
		p.next()
	} else {
		panic(fmt.Sprintf("expected %T, got %s", t, p.text))
	}
}

func (p *Parser) primary() Node {
	switch p.token {
	case NUMBER:
		v, _ := strconv.ParseFloat(p.text, 64)
		p.next()
		return &NumberNode{v}
	case DOLLAR:
		return p.function()
	case OPAREN:
		p.next()
		expr := p.expr(false)
		p.match(CPAREN)
		return expr
	case OBRACKET:
		p.next()
		bexpr := p.expr(false)
		p.match(CBRACKET)
		return &UnaryNode{OBRACKET, bexpr}
	case OCURLY:
		p.next()
		cexpr := p.expr(false)
		p.match(CCURLY)
		return &UnaryNode{OCURLY, cexpr}
	case PLUS:
		p.next()
		return &UnaryNode{PLUS, p.primary()}
	case MINUS:
		p.next()
		return &UnaryNode{MINUS, p.primary()}
	case IDENT:
		v := p.text
		p.next()
		return &IdentifierNode{v}
	}
	return nil
}

func (p *Parser) factor() Node {
	node := p.primary()
	for {
		switch p.token {
		case MULT:
			p.next()
			node = &BinaryNode{MULT, node, p.primary()}
		case DIV:
			p.next()
			node = &BinaryNode{DIV, node, p.primary()}
		case SLASH:
			p.next()
			node = &BinaryNode{SLASH, node, p.primary()}
		case PERIOD:
			p.next()
			node = &BinaryNode{PERIOD, node, p.primary()}
		case COLON:
			p.next()
			node = &BinaryNode{COLON, node, p.primary()}
		case QUESTION:
			p.next()
			node = &BinaryNode{QUESTION, node, p.primary()}
		case EQUAL:
			p.next()
			node = &BinaryNode{EQUAL, node, p.primary()}
		case AMPHERSAND:
			p.next()
			node = &BinaryNode{AMPHERSAND, node, p.primary()}
		default:
			return node
		}
	}
}

func (p *Parser) expr(inFunc bool) Node {
	node := p.factor()
	if node == nil {
		return &NullNode{}
	}
	for {
		switch p.token {
		case PLUS:
			p.next()
			node = &BinaryNode{PLUS, node, p.factor()}
		case MINUS:
			p.next()
			node = &BinaryNode{MINUS, node, p.factor()}
		case COMMA:
			if !inFunc {
				p.next()
				node = &BinaryNode{COMMA, node, p.factor()}
			} else {
				return node
			}
		default:
			return node
		}
	}
}

func (p *Parser) function() Node {
	p.match(DOLLAR)
	name := p.text
	p.match(IDENT)
	p.match(OPAREN)
	args := []Node{}
	for p.token != CPAREN {
		args = append(args, p.expr(true))
		if p.token == COMMA {
			p.next()
		}
	}
	p.match(CPAREN)
	return &FunctionNode{name, args}
}

func EvaluateDeployment(context EvaluationContext) (model.DeploymentSpec, error) {
	ret := context.DeploymentSpec
	for ic, c := range context.DeploymentSpec.Solution.Components {
		val, err := evalProperties(context, c.Properties)
		if err != nil {
			return ret, err
		}
		props, ok := val.(map[string]interface{})
		if !ok {
			return ret, fmt.Errorf("properties must be a map")
		}
		ret.Solution.Components[ic].Properties = props

	}
	return ret, nil
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
func evalProperties(context EvaluationContext, properties interface{}) (interface{}, error) {
	switch properties.(type) {
	case map[string]interface{}:
		for k, v := range properties.(map[string]interface{}) {
			val, err := evalProperties(context, v)
			if err != nil {
				return nil, err
			}
			properties.(map[string]interface{})[k] = val
		}
	case []interface{}:
		for i, v := range properties.([]interface{}) {
			val, err := evalProperties(context, v)
			if err != nil {
				return nil, err
			}
			properties.([]interface{})[i] = val
		}
	case string:
		parser := NewParser(properties.(string))
		val, err := parser.Eval(context)
		if err != nil {
			return nil, err
		}
		properties = val
	}
	return properties, nil
}
