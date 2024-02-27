/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package customvision

import (
	"os"
	"strings"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := CustomVisionReferenceProvider{}
	err := provider.Init(CustomVisionReferenceProviderConfig{})
	assert.Nil(t, err)
}

func TestCustomVisionReferenceProviderConfigFromMapNil(t *testing.T) {
	_, err := CustomVisionReferenceProviderConfigFromMap(nil)
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}

func TestCustomVisionReferenceProviderConfigFromMapEmpty(t *testing.T) {
	_, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}

func TestCustomVisionReferenceProviderConfigFromMapNoKey(t *testing.T) {
	_, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"retries":       "1",
		"retryInterval": "2",
	})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}

func TestCustomVisionReferenceProviderConfigFromMapInvalidRetries(t *testing.T) {
	_, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"key":           "my-key",
		"retries":       "abc",
		"retryInterval": "2",
	})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}

func TestCustomVisionReferenceProviderConfigFromMapEmptyRetries(t *testing.T) {
	config, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"key":           "my-key",
		"retries":       "",
		"retryInterval": "2",
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, config.Retries)
}

func TestCustomVisionReferenceProviderConfigFromMapInvalidRetryInterval(t *testing.T) {
	_, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"key":           "my-key",
		"retries":       "3",
		"retryInterval": "def",
	})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}

func TestCustomVisionReferenceProviderConfigFromMapEmptyRetryInterval(t *testing.T) {
	config, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"key":           "my-key",
		"retries":       "3",
		"retryInterval": "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 5, config.RetryInterval)
}

func TestCustomVisionReferenceProviderConfigFromMap(t *testing.T) {
	config, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "my-name",
		"key":           "my-key",
		"retries":       "33",
		"retryInterval": "55",
	})
	assert.Nil(t, err)
	assert.Equal(t, 33, config.Retries)
	assert.Equal(t, 55, config.RetryInterval)
}

func TestCustomVisionReferenceProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-name", "real-name")
	os.Setenv("my-key", "real-key")
	os.Setenv("my-platform", "real-platform")
	os.Setenv("my-flavor", "real-flavor")
	config, err := CustomVisionReferenceProviderConfigFromMap(map[string]string{
		"name":          "$env:my-name",
		"key":           "$env:my-key",
		"retries":       "33",
		"retryInterval": "55",
	})
	assert.Nil(t, err)
	assert.Equal(t, "real-name", config.Name)
	assert.Equal(t, "real-key", config.APIKey)
	assert.Equal(t, 33, config.Retries)
	assert.Equal(t, 55, config.RetryInterval)
}

// https://jessetest.cognitiveservices.azure.com/customvision/v3.3/Training/projects/0ade741f-cf53-4449-bdc2-e1b1f33a5a20/iterations/1ad3644a-eb5b-43f6-bf4a-8393fcfb547b/export?flavor=Linux&platform=DockerFile
// 1483681c96874612b97a3c67baeaaef5
func TestGet(t *testing.T) {
	apiKey := os.Getenv("TEST_CV_API_KEY")
	cvProject := os.Getenv("TEST_CV_PROJECT")
	cvEndpoint := os.Getenv("TEST_CV_ENDPOINT")
	cvIteration := os.Getenv("TEST_CV_ITERATION")
	if apiKey == "" || cvProject == "" || cvEndpoint == "" || cvIteration == "" {
		t.Skip("Skipping becuase TEST_CV_API_KEY, TEST_CV_PROJECT, TEST_CV_ENDPOINT or TEST_CV_ITERATION environment variable is not set.")
	}
	provider := CustomVisionReferenceProvider{}
	err := provider.Init(CustomVisionReferenceProviderConfig{
		APIKey: apiKey,
	})
	assert.Nil(t, err)
	obj, err := provider.Get(cvProject, cvEndpoint, "", "", cvIteration, "")
	assert.Nil(t, err)
	exports := obj.([]Export)
	assert.True(t, strings.Contains(exports[0].DownloadUri, "blob.core.windows.net:443"))
}
