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

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	if testK8s == "" {
		t.Skip("Skipping because TEST_K8S enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	testK8s := os.Getenv("TEST_K8S")
	symphonyDevice := os.Getenv("SYMPHONY_DEVICE")
	if testK8s == "" || symphonyDevice == "" {
		t.Skip("Skipping because TEST_K8S or SYMPHONY_DEVICE enviornment variable is not set")
	}
	provider := K8sReporter{}
	err := provider.Init(K8sReporterConfig{})
	assert.Nil(t, err)
	err = provider.Report(symphonyDevice, "default", "fabric.symphony", "devices", "v1", map[string]string{
		"a": "ccc",
		"b": "ddd",
	}, false)
	assert.Nil(t, err)
}
