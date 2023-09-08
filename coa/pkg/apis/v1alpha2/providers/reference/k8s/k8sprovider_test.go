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

package k8s

import (
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	testRedis := os.Getenv("TEST_K8S")
	if testRedis == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReferenceProvider{}
	err := provider.Init(K8sReferenceProviderConfig{})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testRedis := os.Getenv("TEST_K8S")
	symphonySolution := os.Getenv("SYMPHONY_SOLUTION")
	if testRedis == "" || symphonySolution == "" {
		t.Skip("Skipping because TEST_K8S or SYMPHONY_SOLUTION enviornment variable is not set")
	}
	provider := K8sReferenceProvider{}
	err := provider.Init(K8sReferenceProviderConfig{})
	assert.Nil(t, err)
	_, err = provider.Get(symphonySolution, "default", "solution.symphony", "solutions", "v1", "")
	assert.NotNil(t, err)
}
func TestK8sReferenceProviderConfigFromMapMapNil(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(nil)
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapEmpty(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{})
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapBadInClusterValue(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"inCluster": "bad",
	})
	assert.NotNil(t, err)
	cErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.BadConfig, cErr.State)
}
func TestK8sReferenceProviderConfigFromMap(t *testing.T) {
	_, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"configPath": "my-path",
		"inCluster":  "true",
	})
	assert.Nil(t, err)
}
func TestK8sReferenceProviderConfigFromMapEnvOverride(t *testing.T) {
	os.Setenv("my-path", "true-path")
	os.Setenv("my-name", "true-name")
	config, err := K8sReferenceProviderConfigFromMap(map[string]string{
		"name":       "$env:my-name",
		"configPath": "$env:my-path",
		"inCluster":  "true",
	})
	assert.Nil(t, err)
	assert.Equal(t, "true-path", config.ConfigPath)
	assert.Equal(t, "true-name", config.Name)
}
