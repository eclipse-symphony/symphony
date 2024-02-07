/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"errors"
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
	assert.Equal(t, "Mock Error Message (Mock Inner Error)", errStr)
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
