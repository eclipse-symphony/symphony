package utils

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	secretmock "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateSingleNumber(t *testing.T) {
	parser := NewParser("1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSingleNegativeNumber(t *testing.T) {
	parser := NewParser("-1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestEvaluateSingleDoubleNegativeNumber(t *testing.T) {
	parser := NewParser("--1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSinglePositiveNegativeNumber(t *testing.T) {
	parser := NewParser("+-1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestEvaluateSingleDoublePositiveNumber(t *testing.T) {
	parser := NewParser("++1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, val)
}
func TestEvaluateSingleNegativePositiveNumber(t *testing.T) {
	parser := NewParser("-+1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestAddition(t *testing.T) {
	parser := NewParser("1+2")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3.0, val)
}
func TestSubtraction(t *testing.T) {
	parser := NewParser("1-2")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, -1.0, val)
}
func TestMultiply(t *testing.T) {
	parser := NewParser("3*4")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 12.0, val)
}
func TestDivide(t *testing.T) {
	parser := NewParser("10/2")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 5.0, val)
}
func TestDivideZero(t *testing.T) {
	parser := NewParser("10/0")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestStringAddNumber(t *testing.T) {
	parser := NewParser("dog+1")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "dog1", val)
}
func TestNumberAddString(t *testing.T) {
	parser := NewParser("1+cat")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "1cat", val)
}
func TestStringAddString(t *testing.T) {
	parser := NewParser("dog+cat")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "dogcat", val)
}
func TestStringMinusString(t *testing.T) {
	parser := NewParser("crazydogs-dogs")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "crazy", val)
}
func TestStringMinusStringMiss(t *testing.T) {
	parser := NewParser("crazydogs-cats")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs", val)
}
func TestParentheses(t *testing.T) {
	parser := NewParser("3-(1+2)/(2+1)")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2.0, val)
}
func TestParenthesesWithString(t *testing.T) {
	parser := NewParser("dog+(32-10/2)")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "dog27", val)
}
func TestStringMultiply(t *testing.T) {
	parser := NewParser("dog*3")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "dogdogdog", val)
}
func TestNumberMultiplyString(t *testing.T) {
	parser := NewParser("3*dog")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestStringMultiplyNegative(t *testing.T) {
	parser := NewParser("dog*-3")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestStringDivide(t *testing.T) {
	parser := NewParser("dog/3")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestMixedExpressions(t *testing.T) {
	parser := NewParser("dog1+3")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "dog13", val)
}
func TestSecretSingleArg(t *testing.T) {
	parser := NewParser("$secret(abc)")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestScretNoProvider(t *testing.T) {
	parser := NewParser("$secret(abc,def)")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestSecret(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret(abc,def)")
	node := parser.expr()
	val, err := node.Eval(nil, provider)
	assert.Nil(t, err)
	assert.Equal(t, "abc>>def", val)
}
func TestSecretWithExpression(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret(abc*2,def+4)")
	node := parser.expr()
	val, err := node.Eval(nil, provider)
	assert.Nil(t, err)
	assert.Equal(t, "abcabc>>def4", val)
}
func TestSecretRecursive(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret($secret(a,b), $secret(c,d))")
	node := parser.expr()
	val, err := node.Eval(nil, provider)
	assert.Nil(t, err)
	assert.Equal(t, "a>>b>>c>>d", val)
}
func TestSecretRecursiveMixed(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret($secret(a,b)+c, $secret(c,d)+e)+f")
	node := parser.expr()
	val, err := node.Eval(nil, provider)
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc>>c>>def", val)
}

func TestConfigSingleArg(t *testing.T) {
	parser := NewParser("$config(abc)")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestConfigNoProvider(t *testing.T) {
	parser := NewParser("$config(abc,def)")
	node := parser.expr()
	_, err := node.Eval(nil, nil)
	assert.NotNil(t, err)
}
func TestConfig(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config(abc,def)")
	node := parser.expr()
	val, err := node.Eval(provider, nil)
	assert.Nil(t, err)
	assert.Equal(t, "abc::def", val)
}
func TestConfigWithExpression(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config(abc*2,def+4)")
	node := parser.expr()
	val, err := node.Eval(provider, nil)
	assert.Nil(t, err)
	assert.Equal(t, "abcabc::def4", val)
}
func TestConfigRecursive(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config($config(a,b), $config(c,d))")
	node := parser.expr()
	val, err := node.Eval(provider, nil)
	assert.Nil(t, err)
	assert.Equal(t, "a::b::c::d", val)
}
func TestConfigRecursiveMixed(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config($config(a,b)+c, $config(c,d)+e)+f")
	node := parser.expr()
	val, err := node.Eval(provider, nil)
	assert.Nil(t, err)
	assert.Equal(t, "a::bc::c::def", val)
}
func TestConfigSecretMix(t *testing.T) {
	//create mock config provider
	configProvider := &mock.MockConfigProvider{}
	err := configProvider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	//create mock secret provider
	secretProvider := &secretmock.MockSecretProvider{}
	err = secretProvider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config($secret(a,b)+c, $secret(c,d)+e)+f")
	node := parser.expr()
	val, err := node.Eval(configProvider, secretProvider)
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc::c>>def", val)
}
func TestConfigWithQuotedStrings(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config('abc',\"def\")")
	node := parser.expr()
	val, err := node.Eval(provider, nil)
	assert.Nil(t, err)
	assert.Equal(t, "abc::def", val)
}
func TestQuotedString(t *testing.T) {

	parser := NewParser("'abc def'")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestQuotedStringAdd(t *testing.T) {
	parser := NewParser("'abc def'+' ghi jkl'")
	node := parser.expr()
	val, err := node.Eval(nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "abc def ghi jkl", val)
}
