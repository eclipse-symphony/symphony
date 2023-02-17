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

package customvision

import (
	"os"
	"strings"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
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
	assert.True(t, strings.Contains(obj.(string), "blob.core.windows.net:443"))
}
