/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/config/mock"
	secretmock "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestEvaluateSingleNumber(t *testing.T) {
	parser := NewParser("1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1", val)
}
func TestEvaluateSingleNumberExpression(t *testing.T) {
	parser := NewParser("${{1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), val)
}
func TestEvaluateNumberSpaceNumber(t *testing.T) {
	parser := NewParser("1 2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1 2", val)
}
func TestEvaluateNumberExpressionSpaceNumberExpression(t *testing.T) {
	parser := NewParser("${{1}} ${{2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1 2", val)
}
func TestEvaluateDoubleDigitNumber(t *testing.T) {
	parser := NewParser("12")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "12", val)
}
func TestEvaluateDoubleDigitNumberExpression(t *testing.T) {
	parser := NewParser("${{12}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(12), val)
}
func TestEvaluateSpace(t *testing.T) {
	parser := NewParser(" ")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, " ", val)
}
func TestEvaluateSpaceExpression(t *testing.T) {
	// a null expression is evaluated to nil
	parser := NewParser("${{ }}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, nil, val)
}
func TestEvaluateSurroundingSpaces(t *testing.T) {
	parser := NewParser("  abc  ")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "  abc  ", val)
}
func TestEvaluateExpressionSurroundingSpaces(t *testing.T) {
	parser := NewParser("${{  abc  }}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc", val)
}
func TestSpacesInBetween(t *testing.T) {
	parser := NewParser("abc def")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestExpressionSpacesInBetween(t *testing.T) {
	// TODO: this behavior may be unintuitive. In this case, both abc and def are
	// recognized as identifiers, and the expression is evaluated to the first identifier
	// found, which is abc. This should change when we switch to a better syntax.
	// See bug #90.
	parser := NewParser("${{abc def}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc", val)
}
func TestEvaluateOpenSingleQuote(t *testing.T) {
	parser := NewParser("'abc def")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "'abc def", val)
}
func TestEvaluateOpenSingleQuoteInExpression(t *testing.T) {
	parser := NewParser("${{'abc def}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "'abc def", val)
}
func TestSingleQuotedAdExtra(t *testing.T) {
	parser := NewParser("'abc def'hij")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "'abc def'hij", val)
}
func TestSingleQuotedAdExtraInExpression(t *testing.T) {
	// TODO: see comments on TestExpressionSpacesInBetween()
	parser := NewParser("${{'abc def'hij}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestNumberDotString(t *testing.T) {
	parser := NewParser("3.abc")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3.abc", val)
}
func TestNumberDotStringInExpression(t *testing.T) {
	parser := NewParser("${{3.abc}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3.abc", val)
}
func TestDot(t *testing.T) {
	parser := NewParser(".")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, ".", val)
}
func TestDotInExpression(t *testing.T) {
	parser := NewParser("${{.}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, ".", val)
}
func TestDotDot(t *testing.T) {
	parser := NewParser("..")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "..", val)
}
func TestDotDotInExpression(t *testing.T) {
	parser := NewParser("${{..}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "..", val)
}
func TestAdd(t *testing.T) {
	parser := NewParser("+")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+", val)
}
func TestAddInExpression(t *testing.T) {
	// this is considered empty + empty, which is empty
	parser := NewParser("${{+}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAdd(t *testing.T) {
	parser := NewParser("++")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "++", val)
}
func TestAddAddInExpression(t *testing.T) {
	parser := NewParser("${{++}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAddAdd(t *testing.T) {
	parser := NewParser("+++")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+++", val)
}
func TestAddAddAddInExpression(t *testing.T) {
	parser := NewParser("${{+++}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddAddAddNumber(t *testing.T) {
	parser := NewParser("+++123")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+++123", val)
}
func TestAddAddAddNumberInExpression(t *testing.T) {
	parser := NewParser("${{+++123}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(123), val)
}
func TestMinus(t *testing.T) {
	parser := NewParser("-")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusInExpression(t *testing.T) {
	parser := NewParser("${{-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestMinusInQuote(t *testing.T) {
	parser := NewParser("'-'")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "'-'", val)
}
func TestMinusInQuoteInExpression(t *testing.T) {
	parser := NewParser("${{'-'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusMinus(t *testing.T) {
	parser := NewParser("--")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "--", val)
}
func TestMinusMinusInExpression(t *testing.T) {
	parser := NewParser("${{--}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusMinusMinus(t *testing.T) {
	parser := NewParser("---")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "---", val)
}
func TestMinusMinusMinusInExpression(t *testing.T) {
	parser := NewParser("${{---}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "--", val)
}
func TestAddMinus(t *testing.T) {
	parser := NewParser("+-")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+-", val)
}
func TestAddMinusInExpression(t *testing.T) {
	// this is "positive negative nothing"
	parser := NewParser("${{+-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "", val)
}
func TestAddMinusMinus(t *testing.T) {
	parser := NewParser("+--")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+--", val)
}
func TestAddMinusMinusInExpression(t *testing.T) {
	// this is "positive negative dash"
	parser := NewParser("${{+--}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusAddMinus(t *testing.T) {
	parser := NewParser("-+-")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-+-", val)
}
func TestMinusAddMinusInExpression(t *testing.T) {
	// this is nothing dash positive negative nothing
	parser := NewParser("${{-+-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-", val)
}
func TestMinusWord(t *testing.T) {
	parser := NewParser("-a")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-a", val)
}
func TestMinusWordInExpression(t *testing.T) {
	parser := NewParser("${{-a}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-a", val)
}
func TestWordMinus(t *testing.T) {
	parser := NewParser("a-")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a-", val)
}
func TestWordMinusInExpression(t *testing.T) {
	parser := NewParser("${{a-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a-", val)
}
func TestAddWord(t *testing.T) {
	parser := NewParser("+a")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+a", val)
}
func TestAddWordInExpression(t *testing.T) {
	parser := NewParser("${{+a}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestWordAdd(t *testing.T) {
	parser := NewParser("a+")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a+", val)
}
func TestWordAddInExpression(t *testing.T) {
	parser := NewParser("${{a+}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestDivideSingle(t *testing.T) {
	parser := NewParser("/")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/", val)
}
func TestDivideSingleInExpression(t *testing.T) {
	parser := NewParser("${{/}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/", val)
}
func TestDvidieDivide(t *testing.T) {
	parser := NewParser("//")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "//", val)
}
func TestDvidieDivideInExpression(t *testing.T) {
	parser := NewParser("${{//}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "//", val)
}
func TestDvidieDivideDivide(t *testing.T) {
	parser := NewParser("///")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "///", val)
}
func TestDvidieDivideDivideInExpression(t *testing.T) {
	parser := NewParser("${{///}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "///", val)
}
func TestUnderScore(t *testing.T) {
	parser := NewParser("_")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "_", val)
}
func TestUnderScoreInExpression(t *testing.T) {
	parser := NewParser("${{_}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "_", val)
}
func TestAmpersand(t *testing.T) {
	parser := NewParser("&")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "&", val)
}
func TestAmpersandInExpression(t *testing.T) {
	parser := NewParser("${{&}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "&", val)
}
func TestAmpersandAmpersand(t *testing.T) {
	parser := NewParser("&&")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "&&", val)
}
func TestAmpersandAmpersandInExpression(t *testing.T) {
	parser := NewParser("${{&&}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "&&", val)
}
func TestForwardSlash(t *testing.T) {
	parser := NewParser("\\")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "\\", val)
}
func TestForwardSlashInExpression(t *testing.T) {
	parser := NewParser("${{\\}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "\\", val)
}
func TestDivideWord(t *testing.T) {
	parser := NewParser("/abc")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc", val)
}
func TestDivideWordInExpression(t *testing.T) {
	parser := NewParser("${{/abc}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc", val)
}
func TestWordDivide(t *testing.T) {
	parser := NewParser("abc/")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc/", val)
}
func TestWordDivideInExpression(t *testing.T) {
	parser := NewParser("${{abc/}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc/", val)
}
func TestPath(t *testing.T) {
	parser := NewParser("abc/def")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc/def", val)
}
func TestPathInExpression(t *testing.T) {
	parser := NewParser("${{abc/def}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc/def", val)
}
func TestAbsolutePath(t *testing.T) {
	parser := NewParser("/abc/def")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def", val)
}
func TestAbsolutePathInExpression(t *testing.T) {
	parser := NewParser("${{/abc/def}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def", val)
}
func TestPathWithQuery(t *testing.T) {
	parser := NewParser("/abc/def?parm=tok")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok", val)
}
func TestPathWithQueryInExpression(t *testing.T) {
	parser := NewParser("${{/abc/def?parm=tok}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok", val)
}
func TestPathWithMultipleParams(t *testing.T) {
	parser := NewParser("/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok&foo=bar", val)
}
func TestPathWithMultipleParamsInExpression(t *testing.T) {
	parser := NewParser("${{/abc/def?parm=tok&foo=bar}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/abc/def?parm=tok&foo=bar", val)
}
func TestUrl(t *testing.T) {
	parser := NewParser("http://abc.com/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com/abc/def?parm=tok&foo=bar", val)
}
func TestUrlInExpression(t *testing.T) {
	parser := NewParser("${{http://abc.com/abc/def?parm=tok&foo=bar}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com/abc/def?parm=tok&foo=bar", val)
}
func TestUrlWithPort(t *testing.T) {
	parser := NewParser("http://abc.com:8080/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:8080/abc/def?parm=tok&foo=bar", val)
}
func TestUrlWithPortInExpression(t *testing.T) {
	parser := NewParser("${{http://abc.com:8080/abc/def?parm=tok&foo=bar}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:8080/abc/def?parm=tok&foo=bar", val)
}
func TestUrlWithPortAddition(t *testing.T) {
	parser := NewParser("http://abc.com:${{8080+1}}/abc/def?parm=tok&foo=bar")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "http://abc.com:8081/abc/def?parm=tok&foo=bar", val)
}

func TestEvaluateSingleNegativeNumber(t *testing.T) {
	parser := NewParser("-1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestEvaluateSingleNegativeNumberInExpression(t *testing.T) {
	parser := NewParser("${{-1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), val)
}
func TestEvaluateSingleNegativeNumberToStrInExpression(t *testing.T) {
	parser := NewParser("${{$str(-1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-1", val)
}
func TestEvaluateSingleDoubleNegativeNumber(t *testing.T) {
	parser := NewParser("--1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "--1", val)
}
func TestEvaluateSingleDoubleNegativeNumberInExpression(t *testing.T) {
	parser := NewParser("${{--1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), val)
}
func TestEvaluateSinglePositiveNegativeNumber(t *testing.T) {
	parser := NewParser("+-1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "+-1", val)
}
func TestEvaluateSinglePositiveNegativeNumberInExpression(t *testing.T) {
	parser := NewParser("${{+-1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), val)
}
func TestEvaluateSingleDoublePositiveNumber(t *testing.T) {
	parser := NewParser("++1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "++1", val)
}
func TestEvaluateSingleDoublePositiveNumberInExpresion(t *testing.T) {
	parser := NewParser("${{++1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), val)
}
func TestEvaluateSingleNegativePositiveNumber(t *testing.T) {
	parser := NewParser("-+1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "-+1", val)
}
func TestEvaluateSingleNegativePositiveNumberInExpression(t *testing.T) {
	parser := NewParser("${{-+1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), val)
}
func TestAddition(t *testing.T) {
	parser := NewParser("1+2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1+2", val)
}
func TestAdditionInExpression(t *testing.T) {
	parser := NewParser("${{1+2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(3), val)
}
func TestAdditions(t *testing.T) {
	parser := NewParser("1+2+3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1+2+3", val)
}
func TestAdditionsInExpression(t *testing.T) {
	parser := NewParser("${{1+2+3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(6), val)
}
func TestFloat(t *testing.T) {
	parser := NewParser("6.3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3", val)
}
func TestFloatInExpression(t *testing.T) {
	// floats are treated as string because they are a common format for version numbers
	// we mostly concern with configuration values instead of arithmetic calculations
	parser := NewParser("${{6.3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3", val)
}
func TestFloatAdd(t *testing.T) {
	parser := NewParser("6.3 + 3.4")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3 + 3.4", val)
}
func TestFloatAddInExpression(t *testing.T) {
	parser := NewParser("${{6.3 + 3.4}}") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.33.4", val)
}
func TestFloatAddInt(t *testing.T) {
	parser := NewParser("6.3 + 3") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3 + 3", val)
}
func TestCreateFloatInExpression(t *testing.T) {
	// it's possible that an expression can be evaluated into a float
	// as the result of a calculation
	parser := NewParser("${{1/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(0.5), val)
}
func TestFloatAddIntInExpression(t *testing.T) {
	parser := NewParser("${{6.3 + 3}}") // floats are treated as string
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.33", val)
}
func TestVersionString(t *testing.T) {
	parser := NewParser("6.3.4")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3.4", val)
}
func TestVersionStringInExpression(t *testing.T) {
	parser := NewParser("${{6.3.4}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3.4", val)
}
func TestVersionStringWithCalculation(t *testing.T) {
	parser := NewParser("6.(1+2).(5-1)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.(1+2).(5-1)", val)
}
func TestVersionStringWithCalculationInExpression(t *testing.T) {
	parser := NewParser("${{6.(1+2).(5-1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "6.3.4", val)
}
func TestSubtraction(t *testing.T) {
	parser := NewParser("1-2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1-2", val)
}
func TestSubtractionInExpression(t *testing.T) {
	parser := NewParser("${{1-2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), val)
}
func TestDash(t *testing.T) {
	parser := NewParser("1-a")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1-a", val)
}
func TestDashInExpression(t *testing.T) {
	parser := NewParser("${{1-a}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1-a", val)
}
func TestDashFloat(t *testing.T) {
	parser := NewParser("1-1.2.3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1-1.2.3", val)
}
func TestDashFloatInExpression(t *testing.T) {
	parser := NewParser("${{1-1.2.3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1-1.2.3", val)
}
func TestMultiply(t *testing.T) {
	parser := NewParser("3*4")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3*4", val)
}
func TestMultiplyInExpression(t *testing.T) {
	parser := NewParser("${{3*4}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(12), val)
}
func TestStar(t *testing.T) {
	parser := NewParser("*")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "*", val)
}
func TestStarInExpression(t *testing.T) {
	parser := NewParser("${{*}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "*", val)
}
func TestStarStar(t *testing.T) {
	parser := NewParser("**")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "**", val)
}
func TestStarStarInExpression(t *testing.T) {
	parser := NewParser("${{**}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "**", val)
}
func TestNumberStar(t *testing.T) {
	parser := NewParser("123*")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "123*", val)
}
func TestNumberStarInExpression(t *testing.T) {
	parser := NewParser("${{123*}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "123*", val)
}
func TestStarNumber(t *testing.T) {
	parser := NewParser("*123")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "*123", val)
}
func TestStarNumberInExpression(t *testing.T) {
	parser := NewParser("${{*123}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "*123", val)
}
func TestStringStarStar(t *testing.T) {
	parser := NewParser("abc**")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc**", val)
}
func TestStringStarStarInExpression(t *testing.T) {
	parser := NewParser("${{abc**}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc**", val)
}
func TestDivide(t *testing.T) {
	parser := NewParser("10/2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/2", val)
}
func TestDivideInExpression(t *testing.T) {
	parser := NewParser("${{10/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(5), val)
}
func TestDivideStringNumberInExpression(t *testing.T) {
	// 10 is string as it's wrapped in single quotes
	// this is no longer a division, but a string concatenation
	// with '10', '/' and '2'
	parser := NewParser("${{'10'/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/2", val)
}
func TestDivideAdd(t *testing.T) {
	parser := NewParser("5/2+1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "5/2+1", val)
}
func TestDivideAddInExpression(t *testing.T) {
	parser := NewParser("${{5/2+1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(3.5), val)
}
func TestDivideAddString(t *testing.T) {
	parser := NewParser("5/2+a")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "5/2+a", val)
}
func TestDivideAddStringInExpression(t *testing.T) {
	parser := NewParser("${{5/2+a}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5a", val)
}
func TestFloatMinus(t *testing.T) {
	parser := NewParser("5/2-5/2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "5/2-5/2", val)
}
func TestFloatMinusInExpression(t *testing.T) {
	parser := NewParser("${{5/2-5/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(0), val)
}
func TestFloatMinus2(t *testing.T) {
	parser := NewParser("2.5-2.5")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5-2.5", val)
}
func TestFloatMinus2InExpression(t *testing.T) {
	// Note we treat floats as strings, as it's a common format for version numbers
	// This differs from the above test, in which floats are the result of a calculation (5/2)
	parser := NewParser("${{2.5-2.5}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5-2.5", val)
}
func TestFloatMultiply(t *testing.T) {
	parser := NewParser("5/2*5/2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "5/2*5/2", val)
}
func TestFloatMultiplyInExpression(t *testing.T) {
	parser := NewParser("${{5/2*5/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(6.25), val)
}
func TestFloatMultiply2(t *testing.T) {
	parser := NewParser("2.5*2.5")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5*2.5", val)
}
func TestFloatMultiply2InExpression(t *testing.T) {
	// Note we treat floats as strings, as it's a common format for version numbers
	// This differs from the above test, in which floats are the result of a calculation (5/2)
	parser := NewParser("${{2.5*2.5}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5*2.5", val)
}
func TestFloatDivide(t *testing.T) {
	parser := NewParser("(5/2)/(5/2)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "(5/2)/(5/2)", val)
}
func TestFloatDivideInExpression(t *testing.T) {
	parser := NewParser("${{(5/2)/(5/2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), val)
}
func TestFloatDivide2(t *testing.T) {
	parser := NewParser("2.5/2.5")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5/2.5", val)
}
func TestFloatDivide2InExpression(t *testing.T) {
	// Note we treat floats as strings, as it's a common format for version numbers
	// This differs from the above test, in which floats are the result of a calculation (5/2)
	parser := NewParser("${{2.5/2.5}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5/2.5", val)
}
func TestFloatDot(t *testing.T) {
	parser := NewParser("5/2.")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "5/2.", val)
}
func TestFloatDotInExpression(t *testing.T) {
	parser := NewParser("${{5/2.}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5.", val)
}
func TestFloatDot2(t *testing.T) {
	parser := NewParser("2.5.")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5.", val)
}
func TestFloatDot2InExpression(t *testing.T) {
	parser := NewParser("${{2.5.}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "2.5.", val)
}
func TestDivideNegative(t *testing.T) {
	parser := NewParser("10/-2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/-2", val)
}
func TestDivideNegativeInExpression(t *testing.T) {
	parser := NewParser("${{10/-2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(-5), val)
}
func TestDivideZero(t *testing.T) {
	parser := NewParser("10/0")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/0", val)
}
func TestDivideZeroInExpression(t *testing.T) {
	parser := NewParser("${{10/0}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/0", val) // can't divide, original string is returned
}
func TestStringAddNumber(t *testing.T) {
	parser := NewParser("dog+1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog+1", val)
}
func TestStringAddNumberInExpression(t *testing.T) {
	parser := NewParser("${{dog+1}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog1", val)
}
func TestNumberAddString(t *testing.T) {
	parser := NewParser("1+cat")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1+cat", val)
}
func TestNumberAddStringInExpression(t *testing.T) {
	parser := NewParser("${{1+cat}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "1cat", val)
}
func TestStringAddString(t *testing.T) {
	parser := NewParser("dog+cat")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog+cat", val)
}
func TestStringAddStringInExpression(t *testing.T) {
	parser := NewParser("${{dog+cat}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dogcat", val)
}
func TestStringMinusString(t *testing.T) {
	parser := NewParser("crazydogs-dogs")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-dogs", val)
}
func TestStringMinusStringInExpression(t *testing.T) {
	// In original design, you can use string subtraction to remove a substring from another string
	// This is no longer supported, as it's not intuitive
	parser := NewParser("${{crazydogs-dogs}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-dogs", val)
}
func TestStringMinusStringMiss(t *testing.T) {
	parser := NewParser("crazydogs-cats")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-cats", val)
}
func TestStringMinusStringMissInExpression(t *testing.T) {
	parser := NewParser("${{crazydogs-cats}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "crazydogs-cats", val)
}
func TestParentheses(t *testing.T) {
	parser := NewParser("3-(1+2)/(2+1)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3-(1+2)/(2+1)", val)
}
func TestParenthesesInExpression(t *testing.T) {
	parser := NewParser("${{3-(1+2)/(2+1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(2), val)
}
func TestParenthesesWithString(t *testing.T) {
	parser := NewParser("dog+(32-10/2)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog+(32-10/2)", val)
}
func TestParenthesesWithStringInExpression(t *testing.T) {
	parser := NewParser("${{dog+(32-10/2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog27", val)
}
func TestStringMultiply(t *testing.T) {
	parser := NewParser("dog*3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*3", val)
}
func TestStringMultiplyInExpression(t *testing.T) {
	parser := NewParser("${{dog*3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*3", val)
}
func TestNumberMultiplyString(t *testing.T) {
	parser := NewParser("3*dog")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3*dog", val)
}
func TestNumberMultiplyStringInExpression(t *testing.T) {
	parser := NewParser("${{3*dog}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "3*dog", val)
}
func TestStringMultiplyNegative(t *testing.T) {
	parser := NewParser("dog*-3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*-3", val)
}
func TestStringMultiplyNegativeInExpression(t *testing.T) {
	parser := NewParser("${{dog*-3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*-3", val)
}
func TestStringMultiplyZero(t *testing.T) {
	parser := NewParser("dog*0")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*0", val)
}
func TestStringMultiplyZeroInExpression(t *testing.T) {
	parser := NewParser("${{dog*0}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*0", val)
}
func TestStringMultiplyFraction(t *testing.T) {
	parser := NewParser("dog*(5/2)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*(5/2)", val)
}
func TestStringMultiplyFractionInExpression(t *testing.T) {
	parser := NewParser("${{dog*(5/2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog*2.5", val)
}
func TestStringDivide(t *testing.T) {
	parser := NewParser("dog/3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog/3", val)
}
func TestStringDivideInExpression(t *testing.T) {
	parser := NewParser("${{dog/3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog/3", val)
}
func TestStringDivideDivide(t *testing.T) {
	parser := NewParser("10/2/2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10/2/2", val)
}
func TestStringDivideDivideInExpression(t *testing.T) {
	parser := NewParser("${{10/2/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(2.5), val)
}
func TestTimeString(t *testing.T) {
	parser := NewParser("12:24:41 3/8/2023")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41 3/8/2023", val)
}
func TestTimeStringInExpression(t *testing.T) {
	parser := NewParser("${{'12:24:41 3/8/2023'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41 3/8/2023", val)
}
func TestTimeStringNoQuote(t *testing.T) {
	parser := NewParser("12:24:41 3/8/2023")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41 3/8/2023", val)
}
func TestTimeStringNoQuoteInExpression(t *testing.T) {
	// this becomes a bit unintuitive
	parser := NewParser("${{12:24:41 3/8/2023}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "12:24:41/8/2023", val)
}
func TestUnderScores(t *testing.T) {
	parser := NewParser("a_b_c_d")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a_b_c_d", val)
}
func TestUnderScoresInExpression(t *testing.T) {
	parser := NewParser("${{a_b_c_d}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a_b_c_d", val)
}
func TestMixedExpressions(t *testing.T) {
	parser := NewParser("dog1+3")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog1+3", val)
}
func TestMixedExpressionsInExpression(t *testing.T) {
	parser := NewParser("${{dog1+3}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "dog13", val)
}
func TestSecretSingleArg(t *testing.T) {
	parser := NewParser("${{$secret(abc)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}
func TestScretNoProvider(t *testing.T) {
	parser := NewParser("${{$secret(abc,def)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}
func TestSecret(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$secret(abc,def)}}")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc>>def", val)
}
func TestSecretWithExpression(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$secret(abc*2,def+4)}}")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "abc*2>>def4", val)
}
func TestSecretRecursive(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$secret($secret(a,b), $secret(c,d))}}")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "a>>b>>c>>d", val)
}
func TestSecretRecursiveMixed(t *testing.T) {
	//create mock secret provider
	provider := &secretmock.MockSecretProvider{}
	err := provider.Init(secretmock.MockSecretProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$secret($secret(a,b)+c, $secret(c,d)+e)+f}}")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: provider})
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc>>c>>def", val)
}

func TestConfigSingleArg(t *testing.T) {
	parser := NewParser("${{$config(abc)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}
func TestConfigNoProvider(t *testing.T) {
	parser := NewParser("${{$config(abc,def)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}

func TestConfigInExpression(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("[{\"name\":\"port${{$config(line-config-$instance(), SERVICE_PORT)}}\",\"port\": ${{$config(line-config-$instance(), SERVICE_PORT)}},\"targetPort\":5000}]")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1",
				},
				Spec: &model.InstanceSpec{},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"name\":\"portline-config-instance1::SERVICE_PORT\",\"port\": line-config-instance1::SERVICE_PORT,\"targetPort\":5000}]", val)
}

func TestConfigObjectInExpression(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config('<' + 'line-config-' + $instance() + '>', \"\")}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance1",
				},
				Spec: &model.InstanceSpec{},
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "<line-config-instance1>::\"\"", val)
}

func TestConfig(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config(abc,def)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc::def", val)
}
func TestConfigWithExpression(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config(abc*2,def+4)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc*2::def4", val)
}
func TestConfigRecursive(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config($config(a,b), $config(c,d))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a::b::c::d", val)
}
func TestConfigRecursiveMixed(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config($config(a,b)+c, $config(c,d)+e)+f}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
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

	parser := NewParser("${{$config($secret(a,b)+c, $secret(c,d)+e)+f}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: configProvider,
		SecretProvider: secretProvider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a>>bc::c>>def", val)
}
func TestConfigWithQuotedStrings(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config('abc',\"def\")}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc::\"def\"", val)
}
func TestQuotedString(t *testing.T) {

	parser := NewParser("${{'abc def'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc def", val)
}
func TestQuotedStringAdd(t *testing.T) {
	parser := NewParser("${{'abc def'+' ghi jkl'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc def ghi jkl", val)
}
func TestEvaulateParamEmptySpec(t *testing.T) {
	parser := NewParser("${{$param(abc)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}
func TestString(t *testing.T) {
	parser := NewParser("docker.io")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io", val)
}
func TestStringInExpression(t *testing.T) {
	parser := NewParser("${{docker.io}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io", val)
}
func TestDockerImage(t *testing.T) {
	parser := NewParser("docker.io/redis:6.0.5")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5", val)
}
func TestDockerImageInExpression(t *testing.T) {
	parser := NewParser("${{docker.io/redis:6.0.5}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5", val)
}
func TestComplexExpression(t *testing.T) {
	parser := NewParser("docker.io/redis:6.0.5 + 678-9")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5 + 678-9", val)
}
func TestComplexExpressionInExpression(t *testing.T) {
	parser := NewParser("${{docker.io/redis:6.0.5 + 678-9}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "docker.io/redis:6.0.5678-9", val)
}
func TestDivideToFloat(t *testing.T) {
	parser := NewParser("9/2")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "9/2", val)
}
func TestDivideToFloatInExpression(t *testing.T) {
	parser := NewParser("${{9/2}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(4.5), val)
}
func TestDivideToFloatAddInt(t *testing.T) {
	parser := NewParser("9/2+35")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "9/2+35", val)
}
func TestDivideToFloatAddIntInExpression(t *testing.T) {
	parser := NewParser("${{9/2+35}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(39.5), val)
}
func TestDivideToFloatAddString(t *testing.T) {
	parser := NewParser("9/2+abc")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "9/2+abc", val)
}
func TestDivideToFloatAddStringInExpression(t *testing.T) {
	parser := NewParser("${{9/2+abc}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "4.5abc", val)
}
func TestParenthesis(t *testing.T) {
	parser := NewParser("(1+2)*(3+4+5)")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "(1+2)*(3+4+5)", val)
}
func TestParenthesisInExpression(t *testing.T) {
	parser := NewParser("${{(1+2)*(3+4+5)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(36), val)
}
func TestStringDivide2(t *testing.T) {
	parser := NewParser("prom/prometheus")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "prom/prometheus", val)
}
func TestStringDivide2InExpression(t *testing.T) {
	parser := NewParser("${{prom/prometheus}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "prom/prometheus", val)
}
func TestWindowsPath(t *testing.T) {
	parser := NewParser("c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle", val)
}
func TestWindowsPathInExpression(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("${{'c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle", val)
}
func TestComplexUrl(t *testing.T) {
	parser := NewParser("https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}
func TestComplexUrlInExpression(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("${{'https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>'}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}
func TestComplexUrlWithExpression(t *testing.T) {
	// The parser can't parse this string correctly. The '' around the string stops the parsing and returns the string as it is
	parser := NewParser("https://manual-approval.azurewebsites.net:${{442+1}}/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
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
	parser := NewParser("https://manual-approval.azurewebsites.net:${{442+1+$secret(a,b)}}/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>")
	val, err := parser.Eval(utils.EvaluationContext{SecretProvider: secretProvider})
	assert.Nil(t, err)
	assert.Equal(t, "https://manual-approval.azurewebsites.net:443a>>b/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<sig>", val)
}

func TestConfigCommaConfig(t *testing.T) {
	//create mock config provider
	provider := &mock.MockConfigProvider{}
	err := provider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	parser := NewParser("${{$config(abc,def),$config(ghi,jkl)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context:        ctx,
		ConfigProvider: provider,
	})
	assert.Nil(t, err)
	assert.Equal(t, "abc::def,ghi::jkl", val)
}
func TestJson1(t *testing.T) {
	parser := NewParser("[{\"containerPort\":9090,\"protocol\":\"TCP\"}]")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"containerPort\":9090,\"protocol\":\"TCP\"}]", val)
}
func TestJson1InExpression(t *testing.T) {
	parser := NewParser("${{[{\"containerPort\":9090,\"protocol\":\"TCP\"}]}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "[{\"containerPort\":9090,\"protocol\":\"TCP\"}]", val)
}
func TestJson2(t *testing.T) {
	parser := NewParser("{\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "{\"requests\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"}}", val)
}
func TestIncompletePlus(t *testing.T) {
	parser := NewParser("${{a+}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a", val)
}
func TestDashAtEnd(t *testing.T) {
	parser := NewParser("${{a-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "a-", val)
}
func TestDashFollowNumber(t *testing.T) {
	parser := NewParser("${{10-}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "10-", val)
}
func TestEvaulateInstance(t *testing.T) {
	parser := NewParser("${{$instance()}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				ObjectMeta: model.ObjectMeta{
					Name: "instance-1",
				},
				Spec: &model.InstanceSpec{},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "instance-1", val)
}
func TestEvaulateParamNoComponent(t *testing.T) {
	parser := NewParser("${{$param(abc)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamNoArgument(t *testing.T) {
	parser := NewParser("${{$param(a)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "b", val)
}
func TestEvaulateParamArgumentOverride(t *testing.T) {
	parser := NewParser("${{$param(a)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "b", val)
}
func TestEvaulateParamWrongComponentName(t *testing.T) {
	parser := NewParser("${{$param(a)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
		Component: "component-2",
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamMissing(t *testing.T) {
	parser := NewParser("${{$param(d)}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
		Component: "component-1",
	})
	assert.NotNil(t, err)
}
func TestEvaulateParamExpressionArgumentOverride(t *testing.T) {
	parser := NewParser("${{$param(a)+$param(c)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
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
		},
		Component: "component-1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "bd", val)
}
func TestEvaluateDeployment(t *testing.T) {
	context := utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "component-1",
							Parameters: map[string]string{
								"a": "b",
								"c": "d",
							},
							Properties: map[string]interface{}{
								"foo": "${{$param(a)}}",
								"bar": "${{$param(c) + ' ' + $param(a)}}",
							},
						},
					},
				},
			},
		},
		Component: "component-1",
	}
	deployment, err := EvaluateDeployment(context)
	assert.Nil(t, err)
	assert.Equal(t, "b", deployment.Solution.Spec.Components[0].Properties["foo"])
	assert.Equal(t, "d b", deployment.Solution.Spec.Components[0].Properties["bar"])
}

func TestEvaluateDeploymentMetadata(t *testing.T) {
	context := utils.EvaluationContext{
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "component-1",
							Parameters: map[string]string{
								"a": "b",
								"c": "d",
							},
							Metadata: map[string]string{
								"foo": "${{$param(a)}}",
								"bar": "${{$param(c) + ' ' + $param(a)}}",
							},
							Properties: map[string]interface{}{
								"foo": "${{$param(a)}}",
								"bar": "${{$param(c) + ' ' + $param(a)}}",
							},
						},
					},
				},
			},
		},
		Component: "component-1",
	}
	deployment, err := EvaluateDeployment(context)
	assert.Nil(t, err)
	assert.Equal(t, "b", deployment.Solution.Spec.Components[0].Properties["foo"])
	assert.Equal(t, "d b", deployment.Solution.Spec.Components[0].Properties["bar"])
	assert.Equal(t, "b", deployment.Solution.Spec.Components[0].Metadata["foo"])
	assert.Equal(t, "d b", deployment.Solution.Spec.Components[0].Metadata["bar"])
}
func TestEvaluateDeploymentConfig(t *testing.T) {
	configProvider := &mock.MockConfigProvider{}
	err := configProvider.Init(mock.MockConfigProviderConfig{})
	assert.Nil(t, err)

	context := utils.EvaluationContext{
		ConfigProvider: configProvider,
		DeploymentSpec: model.DeploymentSpec{
			Instance: model.InstanceState{
				Spec: &model.InstanceSpec{
					Solution: "fake-solution",
				},
			},
			SolutionName: "fake-solution",
			Solution: model.SolutionState{
				Spec: &model.SolutionSpec{
					Components: []model.ComponentSpec{
						{
							Name: "component-1",
							Properties: map[string]interface{}{
								"foo": "${{$config(a,b)}}",
								"bar": "${{$config(c,d)}}",
							},
						},
					},
				},
			},
		},
		Component: "component-1",
	}
	deployment, err := EvaluateDeployment(context)
	assert.Nil(t, err)
	assert.Equal(t, "a::b", deployment.Solution.Spec.Components[0].Properties["foo"])
	assert.Equal(t, "c::d", deployment.Solution.Spec.Components[0].Properties["bar"])
}
func TestEqualNumbers(t *testing.T) {
	parser := NewParser("${{$equal(123, 123)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestEqualNumberString(t *testing.T) {
	parser := NewParser("${{$equal(123, '123')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestEqualProperty(t *testing.T) {
	parser := NewParser("${{$equal(bar, $property(foo))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestEvalProperty(t *testing.T) {
	parser := NewParser("${{$property(foo)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "bar", val)
}
func TestEqualPropertyExpression(t *testing.T) {
	parser := NewParser("${{$equal(bar+2, $property(foo+1))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo1": "bar2",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestPropertyAnd(t *testing.T) {
	parser := NewParser("${{$and($equal($property(foo), bar), $equal($property(book), title))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo":  "bar",
			"book": "title",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestPropertyOr(t *testing.T) {
	parser := NewParser("${{$or($equal($property(foo), bar), $equal($property(foo), bar2))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestPropertyOrFalse(t *testing.T) {
	parser := NewParser("${{$or($equal($property(foo), bar), $equal($property(foo), bar2))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Properties: map[string]string{
			"foo": "bar3",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestNot(t *testing.T) {
	parser := NewParser("${{$not(true)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestNotNot(t *testing.T) {
	parser := NewParser("${{$not($not(true))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestGt(t *testing.T) {
	parser := NewParser("${{$gt(2, 1.0)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestGtEqual(t *testing.T) {
	parser := NewParser("${{$gt(2, 2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestGtNegative(t *testing.T) {
	parser := NewParser("${{$gt(2, 3)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestGe(t *testing.T) {
	parser := NewParser("${{$ge(2, 1.0)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestGeEqual(t *testing.T) {
	parser := NewParser("${{$ge(2, 2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestGeNegative(t *testing.T) {
	parser := NewParser("${{$ge(2, 3)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestLt(t *testing.T) {
	parser := NewParser("${{$lt(2, 3.0)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestLtEqual(t *testing.T) {
	parser := NewParser("${{$lt(2, 2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestLtNegative(t *testing.T) {
	parser := NewParser("${{$lt(2, 1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}

func TestLe(t *testing.T) {
	parser := NewParser("${{$le(2, 3.0)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestLeEqual(t *testing.T) {
	parser := NewParser("${{$le(2, 2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestLeNegative(t *testing.T) {
	parser := NewParser("${{$le(2, 1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestBetween(t *testing.T) {
	parser := NewParser("${{$between(2, 1, 3)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestBetweenNegative(t *testing.T) {
	parser := NewParser("${{$between(2, 3, 1)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestLongVersionNumber(t *testing.T) {
	parser := NewParser("${{0.2.0-20230627.2-develop}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0-20230627.2-develop", val)
}
func TestInputAnd(t *testing.T) {
	parser := NewParser("${{$and($equal($input(foo), bar), $equal($input(book), title))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Inputs: map[string]interface{}{
			"foo":  "bar",
			"book": "title",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestInputOr(t *testing.T) {
	parser := NewParser("${{$or($equal($input(foo), bar), $equal($input(foo), bar2))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Inputs: map[string]interface{}{
			"foo": "bar",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestStringLiteral(t *testing.T) {
	parser := NewParser("stage-1")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "stage-1", val)
}
func TestStringLiteralDoubleUnderScore(t *testing.T) {
	parser := NewParser("__status")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "__status", val)
}
func TestIf(t *testing.T) {
	parser := NewParser("${{$if(true, stage-1, stage-2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "stage-1", val)
}
func TestIfLess(t *testing.T) {
	parser := NewParser("${{$if($lt($output(foo,bar),10), stage-1, stage-2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	parser := NewParser("${{$if($lt($output(foo,bar),10), stage-1, stage-2)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	parser := NewParser("${{$if($lt($output(foo, bar),5),stage-1, '')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	parser := NewParser("${{$output(foo, bar)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	parser := NewParser("${{a__b}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	parser := NewParser("${{$and($gt($val(),5), $lt($val(),10))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Value:   6,
	})
	assert.Nil(t, err)
	assert.Equal(t, true, val)
}
func TestEvaulateValueRangeOutside(t *testing.T) {
	parser := NewParser("${{$and($gt($val(),5), $lt($val(),10))}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Value:   16,
	})
	assert.Nil(t, err)
	assert.Equal(t, false, val)
}
func TestValWithJsonPath(t *testing.T) {
	parser := NewParser("${{$val('$.foo.bar')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Value: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", val)
}
func TestValWithProperty(t *testing.T) {
	parser := NewParser("${{$val(foo)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Value: map[string]interface{}{
			"foo": "baz",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", val)
}
func TestValWithContextProperty(t *testing.T) {
	parser := NewParser("${{$context(foo)}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
		Value: map[string]interface{}{
			"foo": "baz",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "baz", val)
}
func TestValWithJsonPathArray(t *testing.T) {

	parser := NewParser("${{$val('$[?(@.foo.bar==\"baz1\")].foo.bar')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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

func TestContextWithJsonPathArray(t *testing.T) {

	parser := NewParser("${{$context('$[?(@.foo.bar==\"baz1\")].foo.bar')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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

	parser := NewParser("${{$equal($val('$[?(@.foo.bar==\"baz1\")].foo.bar'),'baz1')}}")
	val, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
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
	assert.Equal(t, true, val)
}

func TestMessedUpQuote(t *testing.T) {
	// Note the messed up quote in the json path
	parser := NewParser("${{$val('$[?(@.foo.bar==\"baz1\")].foo.bar)'}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}

func TestStrangeString(t *testing.T) {
	parser := NewParser("${{~pg~edges~ffr4~adapter~collector-ffr4}}")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "~pg~edges~ffr4~adapter~collector-ffr4", output)
}
func TestTwoDolloars(t *testing.T) {
	parser := NewParser("/$2")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/$2", output)
}
func TestReqularExps(t *testing.T) {
	parser := NewParser("/api(/|$)(.*)")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "/api(/|$)(.*)", output)
}

func TestJsonPathSimple(t *testing.T) {
	parser := NewParser("$.store.books[*]")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$.store.books[*]", output)
}
func TestJsonPathSlice(t *testing.T) {
	parser := NewParser("$.store.books[-2:]")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$.store.books[-2:]", output)
}
func TestJsonPathConditional(t *testing.T) {
	parser := NewParser("$.store.books[?(@author=='Nigel Rees')]")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$.store.books[?(@author=='Nigel Rees')]", output)
}
func TestJsonPathRegularExpression(t *testing.T) {
	parser := NewParser("$.store.books[?(@author=~ /^Nigel|Waugh$/  )]")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$.store.books[?(@author=~ /^Nigel|Waugh$/  )]", output)
}

func TestJsonPathComplex(t *testing.T) {
	parser := NewParser("$.store.books[?(@.sections[*]=='s1' || @.sections[*]=='s2' )]")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$.store.books[?(@.sections[*]=='s1' || @.sections[*]=='s2' )]", output)
}
func TestInvalidExpression1(t *testing.T) {
	parser := NewParser("${{half-open")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "${{half-open", output)
}
func TestInvalidExpression2(t *testing.T) {
	parser := NewParser("${missing-one-opening-bracket}}")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "${missing-one-opening-bracket}}", output)
}
func TestInvalidExpression3(t *testing.T) {
	parser := NewParser("${{missing-one-closing-bracket}")
	output, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.Nil(t, err)
	assert.Equal(t, "${{missing-one-closing-bracket}", output)
}
func TestRecursiveUnsupported(t *testing.T) {
	// note we don't support recursive expressions
	parser := NewParser("${{${{recursive}}}}")
	_, err := parser.Eval(utils.EvaluationContext{
		Context: ctx,
	})
	assert.NotNil(t, err)
}
