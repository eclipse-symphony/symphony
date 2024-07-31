/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"os"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalDuration(t *testing.T) {
	duration, err := UnmarshalDuration("1")
	assert.Nil(t, err)
	assert.Equal(t, 1*time.Nanosecond, duration)

	duration, err = UnmarshalDuration("1.0")
	assert.Nil(t, err)
	assert.Equal(t, 1*time.Nanosecond, duration)

	duration, err = UnmarshalDuration("\"1h\"")
	assert.Nil(t, err)
	assert.Equal(t, 1*time.Hour, duration)

	duration, err = UnmarshalDuration("1h") // 1h is not a valid json, "1h" is a valid json
	assert.NotNil(t, err)
	assert.Equal(t, 1*time.Millisecond, duration)

	duration, err = UnmarshalDuration("\"invalid\"")
	assert.NotNil(t, err)
	assert.Equal(t, 1*time.Microsecond, duration)

	duration, err = UnmarshalDuration("true") // bool type
	assert.NotNil(t, err)
	assert.Equal(t, 1*time.Microsecond, duration)
	assert.EqualError(t, err, "invalid duration format")
}

func TestParseProperty(t *testing.T) {
	assert.Equal(t, "abc", ParseProperty("abc"))
	assert.Equal(t, "", ParseProperty("$env:abc"))
	os.Setenv("abc", "def")
	assert.Equal(t, "def", ParseProperty("$env:abc"))
}

type TestSecretProvider struct {
}

func (t *TestSecretProvider) Init(config providers.IProviderConfig) error {
	return nil
}

func (t *TestSecretProvider) Get(object string, field string, localContext interface{}) (string, error) {
	return "test", nil
}

type TestExtConfigProvider struct {
}

func (t *TestExtConfigProvider) Init(config providers.IProviderConfig) error {
	return nil
}

func (t *TestExtConfigProvider) Get(object string, field string, overrides []string, localContext interface{}) (interface{}, error) {
	return "test", nil
}

func (t *TestExtConfigProvider) GetObject(object string, overrides []string, localContext interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"test": "test"}, nil
}

func TestEvaluationContextClone(t *testing.T) {
	secretProvider := &TestSecretProvider{}
	configProvider := &TestExtConfigProvider{}

	ec := &EvaluationContext{
		Properties: map[string]string{
			"abc": "def",
		},
		Inputs: map[string]interface{}{
			"abc": "def",
		},
		Outputs: map[string]map[string]interface{}{
			"abc": {
				"def": "ghi",
			},
		},
		ConfigProvider: configProvider,
		SecretProvider: secretProvider,
	}
	clone := ec.Clone()
	assert.NotNil(t, clone)
	assert.Equal(t, ec.ConfigProvider, clone.ConfigProvider) // ref equals
	assert.Equal(t, ec.SecretProvider, clone.SecretProvider) // ref equals

	var ec2 *EvaluationContext = nil
	assert.Nil(t, ec2.Clone())
}
