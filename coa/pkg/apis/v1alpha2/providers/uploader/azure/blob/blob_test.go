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

package blob

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := AzureBlobUploader{}
	err := provider.Init(AzureBlobUploaderConfig{
		Name: "test",
	})
	assert.Nil(t, err)
}
func TestProbe(t *testing.T) {
	provider := AzureBlobUploader{}
	err := provider.Init(AzureBlobUploaderConfig{
		Name:      "test",
		Account:   "voestore",
		Container: "snapshots",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", provider.ID())
	_, e := provider.Upload("test.txt", []byte("This is a text"))
	assert.NotNil(t, e)
	//assert.Equal(t, "https://voestore.blob.core.windows.net/snapshots/test.txt", st)
}
func TestInitWithMap(t *testing.T) {
	provider := AzureBlobUploader{}
	err := provider.InitWithMap(
		map[string]string{
			"name":      "test",
			"account":   "voestore",
			"container": "snapshots",
		},
	)
	assert.Nil(t, err)
}

func TestSetContext(t *testing.T) {
	provider := AzureBlobUploader{}
	provider.Init(AzureBlobUploaderConfig{
		Name:      "test",
		Account:   "voestore",
		Container: "snapshots",
	})
	provider.SetContext(&contexts.ManagerContext{})
	assert.NotNil(t, provider.Context)
}
