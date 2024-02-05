/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"testing"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestBuildPipeline_WithInvalidJWTConfig(t *testing.T) {
	config := HttpBindingConfig{
		Port: 8080,
		Pipeline: []MiddlewareConfig{
			{
				Type: "middleware.http.jwt",
				// JWT.MustHave is a string array
				Properties: map[string]interface{}{
					"MustHave": "test",
				},
			},
		},
	}
	_, err := BuildPipeline(config, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
	assert.Equal(t, "incorrect jwt pipeline configuration format", coaError.Message)
}

func TestBuildPipeline_WithInvalidTracingConfig(t *testing.T) {
	config := HttpBindingConfig{
		Port: 8080,
		Pipeline: []MiddlewareConfig{
			{
				Type: "middleware.http.tracing",
				// Tracing.Pipeline is a string array
				Properties: map[string]interface{}{
					"pipeline": "test",
				},
			},
		},
	}
	_, err := BuildPipeline(config, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
	assert.Equal(t, "incorrect tracing pipeline configuration format", coaError.Message)
}

func TestBuildPipeline_WithUnknownType(t *testing.T) {
	config := HttpBindingConfig{
		Port: 8080,
		Pipeline: []MiddlewareConfig{
			{
				Type: "middleware.http.unknown",
				Properties: map[string]interface{}{
					"test": "test",
				},
			},
		},
	}
	_, err := BuildPipeline(config, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.BadConfig, coaError.State)
	assert.Equal(t, "middleware type 'middleware.http.unknown' is not recognized", coaError.Message)
}
