package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateSingleNumber(t *testing.T) {
	parser := NewParser("1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSingleNegativeNumber(t *testing.T) {
	parser := NewParser("-1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestEvaluateSingleDoubleNegativeNumber(t *testing.T) {
	parser := NewParser("--1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSinglePositiveNegativeNumber(t *testing.T) {
	parser := NewParser("+-1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestEvaluateSingleDoublePositiveNumber(t *testing.T) {
	parser := NewParser("++1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSingleNegativePositiveNumber(t *testing.T) {
	parser := NewParser("-+1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestAddition(t *testing.T) {
	parser := NewParser("1+2")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 3.0, val)
}
func TestSubtraction(t *testing.T) {
	parser := NewParser("1-2")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestMultiply(t *testing.T) {
	parser := NewParser("3*4")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 12.0, val)
}
func TestDivide(t *testing.T) {
	parser := NewParser("10/2")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 5.0, val)
}
func TestDivideZero(t *testing.T) {
	parser := NewParser("10/0")
	node := parser.expr()
	_, err := node.Eval()
	assert.NotNil(t, err)
}
func TestStringAddNumber(t *testing.T) {
	parser := NewParser("dog+1")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "dog1", val)
}
func TestNumberAddString(t *testing.T) {
	parser := NewParser("1+cat")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "1cat", val)
}
func TestStringAddString(t *testing.T) {
	parser := NewParser("dog+cat")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "dogcat", val)
}
func TestStringMinusString(t *testing.T) {
	parser := NewParser("crazydogs-dogs")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "crazy", val)
}
func TestStringMinusStringMiss(t *testing.T) {
	parser := NewParser("crazydogs-cats")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs", val)
}
func TestParentheses(t *testing.T) {
	parser := NewParser("3-(1+2)/(2+1)")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, 2.0, val)
}
func TestParenthesesWithString(t *testing.T) {
	parser := NewParser("dog+(32-10/2)")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "dog27", val)
}
func TestStringMultiply(t *testing.T) {
	parser := NewParser("dog*3")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "dogdogdog", val)
}
func TestNumberMultiplyString(t *testing.T) {
	parser := NewParser("3*dog")
	node := parser.expr()
	_, err := node.Eval()
	assert.NotNil(t, err)
}
func TestStringMultiplyNegative(t *testing.T) {
	parser := NewParser("dog*-3")
	node := parser.expr()
	_, err := node.Eval()
	assert.NotNil(t, err)
}
func TestStringDivide(t *testing.T) {
	parser := NewParser("dog/3")
	node := parser.expr()
	_, err := node.Eval()
	assert.NotNil(t, err)
}
func TestMixedExpressions(t *testing.T) {
	parser := NewParser("dog1+3")
	node := parser.expr()
	val, err := node.Eval()
	assert.Nil(t, err)
	assert.Equal(t, "dog13", val)
}
