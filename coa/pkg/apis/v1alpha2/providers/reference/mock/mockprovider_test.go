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

package mock

import (
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockReferenceProvider{}
	err := provider.Init(MockReferenceProviderConfig{})
	assert.Nil(t, err)
}

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
