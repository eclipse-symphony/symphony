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
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	secretmock "github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateSingleNumber(t *testing.T) {
	parser := NewParser("1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1", val)
}
func TestEvaluateNumberSpaceNumber(t *testing.T) {
	parser := NewParser("1 2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1", val)
}
func TestEvaluateDoubleDigitNumber(t *testing.T) {
	parser := NewParser("12")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "12", val)
}
func TestEvaluateSpace(t *testing.T) {
	parser := NewParser(" ")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, nil, val)
}
func TestEvaluateSurroundingSpaces(t *testing.T) {
	parser := NewParser("  abc  ")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc", val)
}
func TestSpacesInBetween(t *testing.T) {
	parser := NewParser("abc def")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc", val)
}
func TestEvaluateOpenSingleQuote(t *testing.T) {
	parser := NewParser("'abc def")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "'abc def", val)
}
func TestSingleQuotedAdExtra(t *testing.T) {
	parser := NewParser("'abc def'hij")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestNumberDotString(t *testing.T) {
	parser := NewParser("3.abc")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "3.abc", val)
}
func TestDot(t *testing.T) {
	parser := NewParser(".")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, ".", val)
}
func TestDotDot(t *testing.T) {
	parser := NewParser("..")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "..", val)
}
func TestAdd(t *testing.T) {
	parser := NewParser("+")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAdd(t *testing.T) {
	parser := NewParser("++")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAddAdd(t *testing.T) {
	parser := NewParser("+++")
	node, _ := parser.expr(false)
	val, err := node.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAddAddNumber(t *testing.T) {
	parser := NewParser("+++123")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "123", val)
}
func TestMinus(t *testing.T) {
	parser := NewParser("-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestMinusInQuote(t *testing.T) {
	parser := NewParser("'-'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusMinus(t *testing.T) {
	parser := NewParser("--")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusMinusMinus(t *testing.T) {
	parser := NewParser("---")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "--", val)
}
func TestAddMinus(t *testing.T) {
	// this is "positive negative nothing"
	parser := NewParser("+-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddMinusMinus(t *testing.T) {
	// this is "positive negative dash"
	parser := NewParser("+--")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusAddMinus(t *testing.T) {
	// this is nothing dash positive negative nothing
	parser := NewParser("-+-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusWord(t *testing.T) {
	parser := NewParser("-a")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-a", val)
}
func TestWordMinus(t *testing.T) {
	parser := NewParser("a-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a-", val)
}
func TestAddWord(t *testing.T) {
	parser := NewParser("+a")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestWordAdd(t *testing.T) {
	parser := NewParser("a+")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestDivideSingle(t *testing.T) {
	parser := NewParser("/")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "/", val)
}
func TestDvidieDivide(t *testing.T) {
	parser := NewParser("//")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "//", val)
}
func TestDvidieDivideDivide(t *testing.T) {
	parser := NewParser("///")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "///", val)
}
func TestUnderScore(t *testing.T) {
	parser := NewParser("_")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "_", val)
}
func TestAmpersand(t *testing.T) {
	parser := NewParser("&")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "&", val)
}
func TestAmpersandAmpersand(t *testing.T) {
	parser := NewParser("&&")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "&&", val)
}
func TestForwardSlash(t *testing.T) {
	parser := NewParser("\\")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "\\", val)
}
func TestDivideWord(t *testing.T) {
	parser := NewParser("/abc")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "/abc", val)
}
func TestWordDivide(t *testing.T) {
	parser := NewParser("abc/")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc/", val)
}
func TestPath(t *testing.T) {
	parser := NewParser("abc/def")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc/def", val)
}
func TestAbsolutePath(t *testing.T) {
	parser := NewParser("/abc/def")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def", val)
}
func TestPathWithQuery(t *testing.T) {
	parser := NewParser("/abc/def?parm=tok")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok", val)
}
func TestPathWithMultipleParams(t *testing.T) {
	parser := NewParser("/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok&foo=bar", val)
}
func TestUrl(t *testing.T) {
	parser := NewParser("http://abc.com/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com/abc/def?parm=tok&foo=bar", val)
}
func TestUrlWithPort(t *testing.T) {
	parser := NewParser("http://abc.com:8080/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:8080/abc/def?parm=tok&foo=bar", val)
}
func TestUrlWithPortAddition(t *testing.T) {
	parser := NewParser("http://abc.com:(8080+1)/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:8081/abc/def?parm=tok&foo=bar", val)
}

func TestEvaluateSingleNegativeNumber(t *testing.T) {
	parser := NewParser("-1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestEvaluateSingleDoubleNegativeNumber(t *testing.T) {
	parser := NewParser("--1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1", val)
}
func TestEvaluateSinglePositiveNegativeNumber(t *testing.T) {
	parser := NewParser("+-1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestEvaluateSingleDoublePositiveNumber(t *testing.T) {
	parser := NewParser("++1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1", val)
}
func TestEvaluateSingleNegativePositiveNumber(t *testing.T) {
	parser := NewParser("-+1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestAddition(t *testing.T) {
	parser := NewParser("1+2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "3", val)
}
func TestAdditions(t *testing.T) {
	parser := NewParser("1+2+3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6", val)
}
func TestFloat(t *testing.T) {
	parser := NewParser("6.3") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6.3", val)
}
func TestFloatAdd(t *testing.T) {
	parser := NewParser("6.3 + 3.4") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6.33.4", val)
}
func TestFloatAddInt(t *testing.T) {
	parser := NewParser("6.3 + 3") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6.33", val)
}
func TestVersionString(t *testing.T) {
	parser := NewParser("6.3.4")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6.3.4", val)
}
func TestVersionStringWithCalculation(t *testing.T) {
	parser := NewParser("6.(1+2).(5-1)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "6.3.4", val)
}
func TestSubtraction(t *testing.T) {
	parser := NewParser("1-2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestDash(t *testing.T) {
	parser := NewParser("1-a")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1-a", val)
}
func TestDashFloat(t *testing.T) {
	parser := NewParser("1-1.2.3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1-1.2.3", val)
}
func TestMultiply(t *testing.T) {
	parser := NewParser("3*4")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "12", val)
}
func TestStar(t *testing.T) {
	parser := NewParser("*")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "*", val)
}
func TestStarStar(t *testing.T) {
	parser := NewParser("**")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "**", val)
}
func TestNumberStar(t *testing.T) {
	parser := NewParser("123*")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "123*", val)
}
func TestStarNumber(t *testing.T) {
	// repeat (empty) 123 times
	parser := NewParser("*123")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestStringRepeat(t *testing.T) {
	// repeat (empty) 123 times
	parser := NewParser("abc*3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abcabcabc", val)
}
func TestStringStarStar(t *testing.T) {
	// repeat (empty) 123 times
	parser := NewParser("abc**")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc**", val)
}
func TestDivide(t *testing.T) {
	parser := NewParser("10/2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "5", val)
}
func TestDivideAdd(t *testing.T) {
	parser := NewParser("5/2+1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "3.5", val)
}
func TestDivideAddString(t *testing.T) {
	parser := NewParser("5/2+a")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "2.5a", val)
}
func TestDivideNegative(t *testing.T) {
	parser := NewParser("10/-2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "-5", val)
}
func TestDivideZero(t *testing.T) {
	parser := NewParser("10/0")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestStringAddNumber(t *testing.T) {
	parser := NewParser("dog+1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dog1", val)
}
func TestNumberAddString(t *testing.T) {
	parser := NewParser("1+cat")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "1cat", val)
}
func TestStringAddString(t *testing.T) {
	parser := NewParser("dog+cat")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dogcat", val)
}
func TestStringMinusString(t *testing.T) {
	parser := NewParser("crazydogs-dogs")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-dogs", val)
}
func TestStringMinusStringMiss(t *testing.T) {
	parser := NewParser("crazydogs-cats")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-cats", val)
}
func TestParentheses(t *testing.T) {
	parser := NewParser("3-(1+2)/(2+1)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "2", val)
}
func TestParenthesesWithString(t *testing.T) {
	parser := NewParser("dog+(32-10/2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dog27", val)
}
func TestStringMultiply(t *testing.T) {
	parser := NewParser("dog*3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dogdogdog", val)
}
func TestNumberMultiplyString(t *testing.T) {
	parser := NewParser("3*dog")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "3*dog", val)
}
func TestStringMultiplyNegative(t *testing.T) {
	parser := NewParser("dog*-3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dog*-3", val)
}
func TestStringMultiplyZero(t *testing.T) {
	parser := NewParser("dog*0")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestStringMultiplyFraction(t *testing.T) {
	parser := NewParser("dog*(5/2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dogdog", val)
}
func TestStringDivide(t *testing.T) {
	parser := NewParser("dog/3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dog/3", val)
}
func TestStringDivideDivide(t *testing.T) {
	parser := NewParser("10/2/2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "2.5", val)
}
func TestTimeString(t *testing.T) {
	parser := NewParser("'12:24:41 3/8/2023'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41 3/8/2023", val)
}
func TestTimeStringNoQuote(t *testing.T) {
	// this becomes unintuitive
	parser := NewParser("12:24:41 3/8/2023")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41/8/2023", val)
}
func TestUnderScores(t *testing.T) {
	parser := NewParser("a_b_c_d")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a_b_c_d", val)
}
func TestMixedExpressions(t *testing.T) {
	parser := NewParser("dog1+3")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "dog13", val)
}
func TestSecretSingleArg(t *testing.T) {
	parser := NewParser("$secret(abc)")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestScretNoProvider(t *testing.T) {
	parser := NewParser("$secret(abc,def)")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestSecret(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret(abc,def)")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc>>def", val)
}
func TestSecretWithExpression(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret(abc*2,def+4)")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abcabc>>def4", val)
}
func TestSecretRecursive(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret($secret(a,b), $secret(c,d))")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "a>>b>>c>>d", val)
}
func TestSecretRecursiveMixed(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$secret($secret(a,b)+c, $secret(c,d)+e)+f")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc>>c>>def", val)
}

func TestConfigSingleArg(t *testing.T) {
	parser := NewParser("$config(abc)")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestConfigNoProvider(t *testing.T) {
	parser := NewParser("$config(abc,def)")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestConfig(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config(abc,def)")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc::def", val)
}
func TestConfigWithExpression(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config(abc*2,def+4)")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abcabc::def4", val)
}
func TestConfigRecursive(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config($config(a,b), $config(c,d))")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "a::b::c::d", val)
}
func TestConfigRecursiveMixed(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config($config(a,b)+c, $config(c,d)+e)+f")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
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
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: configProvider, SecretProvider: secretProvider})
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc::c>>def", val)
}
func TestConfigWithQuotedStrings(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config('abc',\"def\")")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc::\"def\"", val)
}
func TestQuotedString(t *testing.T) {

	parser := NewParser("'abc def'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestQuotedStringAdd(t *testing.T) {
	parser := NewParser("'abc def'+' ghi jkl'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "abc def ghi jkl", val)
}
func TestEvaulateParamEmptySpec(t *testing.T) {
	parser := NewParser("$param(abc)")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
func TestString(t *testing.T) {
	parser := NewParser("docker.io")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io", val)
}
func TestDockerImage(t *testing.T) {
	parser := NewParser("docker.io/redis:6.0.5")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5", val)
}
func TestComplexExpression(t *testing.T) {
	parser := NewParser("docker.io/redis:6.0.5 + 678-9")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5678-9", val)
}
func TestDivideToFloat(t *testing.T) {
	parser := NewParser("9/2")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "4.5", val)
}
func TestDivideToFloatAddInt(t *testing.T) {
	parser := NewParser("9/2+35")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "39.5", val)
}
func TestDivideToFloatAddString(t *testing.T) {
	parser := NewParser("9/2+abc")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "4.5abc", val)
}
func TestParenthesis(t *testing.T) {
	parser := NewParser("(1+2)*(3+4+5)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "36", val)
}
func TestStringDivide2(t *testing.T) {
	parser := NewParser("prom/prometheus")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "prom/prometheus", val)
}
func TestWindowsPath(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("'c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle", val)
}
func TestComplexUrl(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("'https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}
func TestComplexUrlWithExpression(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("'https://manual-approval.azurewebsites.net:'+(442+1)+'/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>'")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}
func TestComplexUrlWithFunctionExpression(t *testing.T) {
	//create mock config provider
	configProvider := &mock.MockConfigProvider{}
	err := configProvider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	//create mock secret provider
	secretProvider := &secretmock.MockSecretProvider{}
	err = secretProvider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("'https://manual-approval.azurewebsites.net:'+(442+1+$secret(a,b))+'/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>'")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: secretProvider})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443a>>b/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}

func TestConfigCommaConfig(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("$config(abc,def),$config(ghi,jkl)")
	val, err := parser.Eval(utils.EvaluationContext{ConfigProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc::def,ghi::jkl", val)
}
func TestJson1(t *testing.T) {
	parser := NewParser("[{\"containerPort\":9090,\"protocol\":\"TCP\"}]")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"containerPort\":9090,\"protocol\":\"TCP\"}]", val)
}
func TestJson2(t *testing.T) {
	parser := NewParser("{\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "{\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}", val)
}
func TestIncompletePlus(t *testing.T) {
	parser := NewParser("a+")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestDashAtEnd(t *testing.T) {
	parser := NewParser("a-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "a-", val)
}
func TestDashFollowNumber(t *testing.T) {
	parser := NewParser("10-")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "10-", val)
}
func TestEvaulateInstance(t *testing.T) {
	parser := NewParser("$instance()")
	val, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Name: "instance-1",
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance-1", val)
}
func TestEvaulateParamNoComponent(t *testing.T) {
	parser := NewParser("$param(abc)")
	_, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamNoArgument(t *testing.T) {
	parser := NewParser("$param(a)")
	val, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "b", val)
}
func TestEvaulateParamArgumentOverride(t *testing.T) {
	parser := NewParser("$param(a)")
	val, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
				Arguments: map[string]map[string]string{
					"component-1": {
						"a": "new-value",
					},
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "new-value", val)
}
func TestEvaulateParamWrongComponentName(t *testing.T) {
	parser := NewParser("$param(a)")
	_, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
				Arguments: map[string]map[string]string{
					"component-1": {
						"a": "new-value",
					},
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
		Component: "component-2",
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamMissing(t *testing.T) {
	parser := NewParser("$param(d)")
	_, err := parser.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
				Arguments: map[string]map[string]string{
					"component-1": {
						"a": "new-value",
					},
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
		Component: "component-1",
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamExpressionArgumentOverride(t *testing.T) {
	parser := NewParser("$param(a)+$param(c)")
	node, _ := parser.expr(false)
	val, err := node.Eval(utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
				Arguments: map[string]map[string]string{
					"component-1": {
						"a": "new-value",
					},
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "new-valued", val)
}
func TestEvaluateDeployment(t *testing.T) {
	context := utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceSpec{
				Solution: "fake-solution",
				Arguments: map[string]map[string]string{
					"component-1": {
						"a": "new-value",
					},
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionSpec{
				Components: []model.ComponentSpec{
					{
						Name: "component-1",
						Parameters: map[string]string{
							"a": "b",
							"c": "d",
						},
						Properties: map[string]interface{}{
							"foo": "$param(a)",
							"bar": "$param(c) + ' ' + $param(a)",
						},
					},
				},
			},
		},
		Component: "component-1",
	}
	deployment, err := EvaluateDeployment(context)
	assert.Nil(t, err)
	assert.Equal(t, "new-value", deployment.Solution.Components[0].Properties["foo"])
	assert.Equal(t, "d new-value", deployment.Solution.Components[0].Properties["bar"])
}
func TestEqualNumbers(t *testing.T) {
	parser := NewParser("$equal(123, 123)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestEqualNumberString(t *testing.T) {
	parser := NewParser("$equal(123, '123')")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestEqualProperty(t *testing.T) {
	parser := NewParser("$equal(bar, $property(foo))")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestEvalProperty(t *testing.T) {
	parser := NewParser("$property(foo)")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "bar", val)
}
func TestEqualPropertyExpression(t *testing.T) {
	parser := NewParser("$equal(bar+2, $property(foo+1))")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo1": "bar2",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestPropertyAnd(t *testing.T) {
	parser := NewParser("$and($equal($property(foo), bar), $equal($property(book), title))")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo":  "bar",
			"book": "title",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestPropertyOr(t *testing.T) {
	parser := NewParser("$or($equal($property(foo), bar), $equal($property(foo), bar2))")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestPropertyOrFalse(t *testing.T) {
	parser := NewParser("$or($equal($property(foo), bar), $equal($property(foo), bar2))")
	val, err := parser.Eval(utils.EvaluationContext{
		Properties: map[string]string{
			"foo": "bar3",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestNot(t *testing.T) {
	parser := NewParser("$not(true)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestNotNot(t *testing.T) {
	parser := NewParser("$not($not(true))")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestGt(t *testing.T) {
	parser := NewParser("$gt(2, 1.0)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestGtEqual(t *testing.T) {
	parser := NewParser("$gt(2, 2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestGtNegative(t *testing.T) {
	parser := NewParser("$gt(2, 3)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestGe(t *testing.T) {
	parser := NewParser("$ge(2, 1.0)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestGeEqual(t *testing.T) {
	parser := NewParser("$ge(2, 2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestGeNegative(t *testing.T) {
	parser := NewParser("$ge(2, 3)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestLt(t *testing.T) {
	parser := NewParser("$lt(2, 3.0)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestLtEqual(t *testing.T) {
	parser := NewParser("$lt(2, 2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestLtNegative(t *testing.T) {
	parser := NewParser("$lt(2, 1)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestLe(t *testing.T) {
	parser := NewParser("$le(2, 3.0)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestLeEqual(t *testing.T) {
	parser := NewParser("$le(2, 2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestLeNegative(t *testing.T) {
	parser := NewParser("$le(2, 1)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestBetween(t *testing.T) {
	parser := NewParser("$between(2, 1, 3)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestBetweenNegative(t *testing.T) {
	parser := NewParser("$between(2, 3, 1)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestLongVersionNumber(t *testing.T) {

	parser := NewParser("0.2.0-20230627.2-develop")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0-20230627.2-develop", val)
}
func TestInputAnd(t *testing.T) {
	parser := NewParser("$and($equal($input(foo), bar), $equal($input(book), title))")
	val, err := parser.Eval(utils.EvaluationContext{
		Inputs: map[string]interface{}{
			"foo":  "bar",
			"book": "title",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestInputOr(t *testing.T) {
	parser := NewParser("$or($equal($input(foo), bar), $equal($input(foo), bar2))")
	val, err := parser.Eval(utils.EvaluationContext{
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestStringLiteral(t *testing.T) {
	parser := NewParser("stage-1")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "stage-1", val)
}
func TestIf(t *testing.T) {
	parser := NewParser("$if(true, stage-1, stage-2)")
	val, err := parser.Eval(utils.EvaluationContext{})
	assert.Nil(t, err)
	assert.Equal(t, "stage-1", val)
}
func TestIfLess(t *testing.T) {
	parser := NewParser("$if($lt($output(foo,bar),10), stage-1, stage-2)")
	val, err := parser.Eval(utils.EvaluationContext{
		Outputs: map[string]map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": 5,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "stage-1", val)
}
func TestIfLessNegative(t *testing.T) {
	parser := NewParser("$if($lt($output(foo,bar),10), stage-1, stage-2)")
	val, err := parser.Eval(utils.EvaluationContext{
		Outputs: map[string]map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": 11,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "stage-2", val)
}
func TestIfLessNegativeEmptyString(t *testing.T) {
	parser := NewParser("$if($lt($output(foo, bar),5),stage-1, '')")
	val, err := parser.Eval(utils.EvaluationContext{
		Outputs: map[string]map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": 11,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestOutputArray(t *testing.T) {
	parser := NewParser("$output(foo, bar)")
	val, err := parser.Eval(utils.EvaluationContext{
		Outputs: map[string]map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": []interface{}{"a", "b", "c"},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{"a", "b", "c"}, val)
}
func TestLeadingUnderScore(t *testing.T) {
	parser := NewParser("a__b")
	val, err := parser.Eval(utils.EvaluationContext{
		Outputs: map[string]map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": []interface{}{"a", "b", "c"},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "a__b", val)
}
func TestEvaulateValueRange(t *testing.T) {
	parser := NewParser("$and($gt($val(),5), $lt($val(),10))")
	val, err := parser.Eval(utils.EvaluationContext{
		Value: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}
func TestEvaulateValueRangeOutside(t *testing.T) {
	parser := NewParser("$and($gt($val(),5), $lt($val(),10))")
	val, err := parser.Eval(utils.EvaluationContext{
		Value: 16,
	})
	assert.Nil(t, err)
	assert.Equal(t, "false", val)
}
func TestValWithJsonPath(t *testing.T) {
	parser := NewParser("$val('$.foo.bar')")
	val, err := parser.Eval(utils.EvaluationContext{
		Value: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", val)
}
func TestValWithJsonPathArray(t *testing.T) {

	parser := NewParser("$val('$[?(@.foo.bar==\"baz1\")].foo.bar')")
	val, err := parser.Eval(utils.EvaluationContext{
		Value: []interface{}{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz1",
				},
			},
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz2",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz1", val)
}

func TestValWithJsonPathArrayBoolean(t *testing.T) {

	parser := NewParser("$equal($val('$[?(@.foo.bar==\"baz1\")].foo.bar'),'baz1')")
	val, err := parser.Eval(utils.EvaluationContext{
		Value: []interface{}{
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz1",
				},
			},
			map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz2",
				},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "true", val)
}

func TestMessedUpQuote(t *testing.T) {
	parser := NewParser("$val('$[?(@.foo.bar==\"baz1\")].foo.bar)'")
	_, err := parser.Eval(utils.EvaluationContext{})
	assert.NotNil(t, err)
}
