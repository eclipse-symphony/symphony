package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config"
)

type Token int

const (
	EOF Token = iota
	NUMBER
	DOLLAR
	IDENT
	OPAREN
	CPAREN
	PLUS
	MINUS
	MULT
	DIV
	COMMA
	STRING
)

var opNames = map[Token]string{
	PLUS:  "+",
	MINUS: "-",
	MULT:  "*",
	DIV:   "/",
	COMMA: ",",
}

type Node interface {
	Eval(confiProvider config.IConfigProvider) (interface{}, error)
}

type NumberNode struct {
	Value float64
}

func (n *NumberNode) Eval(confiProvider config.IConfigProvider) (interface{}, error) {
	return n.Value, nil
}

type IdentifierNode struct {
	Value string
}

func (n *IdentifierNode) Eval(confiProvider config.IConfigProvider) (interface{}, error) {
	return n.Value, nil
}

type UnaryNode struct {
	Op   Token
	Expr Node
}

func (n *UnaryNode) Eval(confiProvider config.IConfigProvider) (interface{}, error) {
	switch n.Op {
	case PLUS:
		return n.Expr.Eval(confiProvider)
	case MINUS:
		val, err := n.Expr.Eval(confiProvider)
		if err != nil {
			return val, err
		}
		if v, ok := val.(float64); ok {
			return -v, nil
		}
		return nil, errors.New("can't apply '-' to non-numeric values")
	}
	return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
}

type BinaryNode struct {
	Op    Token
	Left  Node
	Right Node
}

func (n *BinaryNode) Eval(confiProvider config.IConfigProvider) (interface{}, error) {
	switch n.Op {
	case PLUS:
		lv, le := n.Left.Eval(confiProvider)
		if le != nil {
			return nil, le
		}
		rv, re := n.Right.Eval(confiProvider)
		if re != nil {
			return nil, re
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
		lv, le := n.Left.Eval(confiProvider)
		if le != nil {
			return nil, le
		}
		rv, re := n.Right.Eval(confiProvider)
		if re != nil {
			return nil, re
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			v := vl - vr
			return v, nil
		} else {
			return strings.ReplaceAll(fmt.Sprintf("%v", lv), fmt.Sprintf("%v", rv), ""), nil
		}
	case MULT:
		lv, le := n.Left.Eval(confiProvider)
		if le != nil {
			return nil, le
		}
		rv, re := n.Right.Eval(confiProvider)
		if re != nil {
			return nil, re
		}
		vl, okl := lv.(float64)
		vr, okr := rv.(float64)
		if okl && okr {
			v := vl * vr
			return v, nil
		} else {
			if !okl && okr && vr > 0 {
				return strings.Repeat(fmt.Sprintf("%v", lv), int(vr)), nil
			}
			return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
		}
	case DIV:
		lv, le := n.Left.Eval(confiProvider)
		if le != nil {
			return nil, le
		}
		rv, re := n.Right.Eval(confiProvider)
		if re != nil {
			return nil, re
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
			return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
		}
	}
	return nil, fmt.Errorf("operator '%s' is not allowed in this context", opNames[n.Op])
}

type FunctionNode struct {
	Name string
	Args []Node
}

func (n *FunctionNode) Eval(confiProvider config.IConfigProvider) (interface{}, error) {
	switch n.Name {
	case "params":
		if len(n.Args) == 1 {
			return n.Args[0].Eval(confiProvider)
		}
		return nil, fmt.Errorf("$params() expects 1 argument, fount %d", len(n.Args))
	case "config":
		if len(n.Args) == 2 {
			if confiProvider == nil {
				return nil, errors.New("a config provider is needed to evaluate $config()")
			}
			obj, err := n.Args[0].Eval(confiProvider)
			if err != nil {
				return nil, err
			}
			field, err := n.Args[1].Eval(confiProvider)
			if err != nil {
				return nil, err
			}
			return confiProvider.Get(obj.(string), field.(string))
		}
		return nil, fmt.Errorf("$params() expects 2 arguments, fount %d", len(n.Args))
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
	s.Init(strings.NewReader(text))
	s.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings
	p := &Parser{
		s: &s,
	}
	p.next()
	return p
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
	case '+':
		return PLUS
	case '-':
		return MINUS
	case '*':
		return MULT
	case '/':
		return DIV
	case ',':
		return COMMA
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
		expr := p.expr()
		p.match(CPAREN)
		return expr
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
		default:
			return node
		}
	}
}

func (p *Parser) expr() Node {
	node := p.factor()
	for {
		switch p.token {
		case PLUS:
			p.next()
			node = &BinaryNode{PLUS, node, p.factor()}
		case MINUS:
			p.next()
			node = &BinaryNode{MINUS, node, p.factor()}
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
		args = append(args, p.expr())
		if p.token == COMMA {
			p.next()
		}
	}
	p.match(CPAREN)
	return &FunctionNode{name, args}
}
