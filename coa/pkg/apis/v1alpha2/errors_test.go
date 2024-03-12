/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	coaError := COAError{
		State:      InternalError,
		Message:    "Mock Error Message",
		InnerError: errors.New("Mock Inner Error"),
	}
	errStr := coaError.Error()
	assert.Equal(t, "Internal Error: Mock Error Message (caused by: Mock Inner Error)", errStr)

	coaError = COAError{
		State:      InternalError,
		InnerError: errors.New("Mock Inner Error"),
	}
	errStr = coaError.Error()
	assert.Equal(t, "Mock Inner Error", errStr)

	coaError = COAError{
		State:   InternalError,
		Message: "Mock Error Message",
	}
	errStr = coaError.Error()
	assert.Equal(t, "Internal Error: Mock Error Message", errStr)
}

func TestFromError(t *testing.T) {
	coaError := FromError(errors.New("Mock Error"))
	assert.Equal(t, "Mock Error", coaError.Message)
	assert.Equal(t, InternalError, coaError.State)
}

func TestFromHTTPResponseCode(t *testing.T) {
	mapCodeToErrorInfo := map[int]map[string]interface{}{
		400: {
			"state":   BadRequest,
			"message": "Bad Request",
		},
		403: {
			"state":   Unauthorized,
			"message": "Unauthorized",
		},
		404: {
			"state":   NotFound,
			"message": "Not Found",
		},
		405: {
			"state":   MethodNotAllowed,
			"message": "Method Not Allowed",
		},
		409: {
			"state":   Conflict,
			"message": "Conflict",
		},
		500: {
			"state":   InternalError,
			"message": "Internal Server Error",
		},
	}
	for code, errorInfo := range mapCodeToErrorInfo {
		coaError := FromHTTPResponseCode(code, []byte(errorInfo["message"].(string)))
		assert.Equal(t, errorInfo["state"].(State), coaError.State)
		assert.Equal(t, errorInfo["message"].(string), coaError.Message)
	}
}

func TestNewCOAError(t *testing.T) {
	coaError := NewCOAError(errors.New("Mock Error"), "Mock Error Message", InternalError)
	assert.Equal(t, "Mock Error", coaError.InnerError.Error())
	assert.Equal(t, "Mock Error Message", coaError.Message)
	assert.Equal(t, InternalError, coaError.State)
}

func TestIsNotFound(t *testing.T) {
	assert.False(t, IsNotFound(errors.New("Mock Error")))
	assert.True(t, IsNotFound(NewCOAError(errors.New("Mock Error"), "Mock Error Message", NotFound)))
}

func TestIsDelayed(t *testing.T) {
	assert.False(t, IsDelayed(errors.New("Mock Error")))
	assert.True(t, IsDelayed(NewCOAError(errors.New("Mock Error"), "Mock Error Message", Delayed)))
}
func TestInnerError_SimpleError(t *testing.T) {
	testErr := NewCOAError(nil, "This is an error msg", InternalError)
	expected := "Internal Error: This is an error msg"
	assert.Equal(t, expected, testErr.Error())
}

func TestInnerError_StandardInnerError(t *testing.T) {
	testErr := NewCOAError(fmt.Errorf("standard inner error"), "This is an error msg", InternalError)
	expected := "Internal Error: This is an error msg (caused by: standard inner error)"
	assert.Equal(t, expected, testErr.Error())
}

func TestInnerError_COAInnerError(t *testing.T) {
	innerErr := NewCOAError(nil, "This is an inner error msg", InternalError)
	testErr := NewCOAError(innerErr, "This is an error msg", InternalError)
	expected := "Internal Error: This is an error msg (caused by: Internal Error: This is an inner error msg)"
	assert.Equal(t, expected, testErr.Error())
}
