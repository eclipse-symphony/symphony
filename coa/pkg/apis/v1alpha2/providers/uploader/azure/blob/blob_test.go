/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package blob

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
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
