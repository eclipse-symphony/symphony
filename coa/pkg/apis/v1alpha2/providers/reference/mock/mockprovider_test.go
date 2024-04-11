/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{})
	assert.Nil(t, err)
}

// TestConfigData tests the ConfigData function
func TestConfigData(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)
	id := provider.ID()
	assert.Equal(t, "test", id)
	targetID := provider.TargetID()
	assert.Equal(t, "mock-target", targetID)
	referenceType := provider.ReferenceType()
	assert.Equal(t, "mock", referenceType)
	err = provider.Reconfigure(MockReferenceProviderConfig{})
	assert.Nil(t, err)
}

// TestCloneWithEmptyConfig	tests the Clone function with an empty config
func TestCloneWithEmptyConfig(t *testing.T) {
	provider := MockReferenceProvider{}
	_, err := provider.Clone(MockReferenceProviderConfig{})
	assert.Nil(t, err)
}

// TestClone tests the Clone function
func TestClone(t *testing.T) {
	provider := MockReferenceProvider{}
	config := MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": "def",
		},
	}
	_, err := provider.Clone(config)
	assert.Nil(t, err)
}

// TestGetValidKey tests the Get function with a valid key
func TestGetValidKey(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": "def",
		},
	})
	assert.Nil(t, err)
	val, err := provider.Get("abc", "default", "unknown", "abc", "v1", "")
	assert.Nil(t, err)
	assert.Equal(t, "def", val.(string))
}

func TestGetInvalidKey(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"abc": "def",
		},
	})
	assert.Nil(t, err)
	_, err = provider.Get("hij", "default", "unknown", "hij", "v1", "")
	assert.NotNil(t, err)
	coaE, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.NotFound, coaE.State)
}
func TestMockReferenceProviderConfigFromMapNil(t *testing.T) {
	_, err := MockReferenceProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestMockReferenceProviderConfigFromMapEmpty(t *testing.T) {
	_, err := MockReferenceProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestMockReferenceProviderConfigFromMap(t *testing.T) {
	config, err := MockReferenceProviderConfigFromMap(map[string]string{
		"name": "my-name",
		"key1": "value1",
	})
	assert.Nil(t, err)
	assert.Equal(t, "my-name", config.Name)
	assert.Equal(t, 1, len(config.Values))
	assert.Equal(t, "value1", config.Values["key1"])
}
func TestMockReferenceProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-name", "real-name")
	os.Setenv("my-value", "real-value")
	config, err := MockReferenceProviderConfigFromMap(map[string]string{
		"name": "$env:my-name",
		"key1": "$env:my-value",
	})
	assert.Nil(t, err)
	assert.Equal(t, "real-name", config.Name)
	assert.Equal(t, 1, len(config.Values))
	assert.Equal(t, "real-value", config.Values["key1"])
}

func TestMockReferenceProvider_Clone(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{})
	assert.Nil(t, err)

	clonedProvider, err := provider.Clone(nil)
	assert.Nil(t, err)
	assert.NotNil(t, clonedProvider)

	newConfig, err := MockReferenceProviderConfigFromMap(map[string]string{
		"name": "$env:my-name",
		"key1": "$env:my-value",
	})
	assert.Nil(t, err)
	clonedProvider, err = provider.Clone(newConfig)
	assert.Nil(t, err)
	assert.NotNil(t, clonedProvider)
}
