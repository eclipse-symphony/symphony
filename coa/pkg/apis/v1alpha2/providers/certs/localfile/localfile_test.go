/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package localfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertGetInvalidFile(t *testing.T) {
	provider := LocalCertFileProvider{}
	err := provider.Init(LocalCertFileProviderConfig{
		Name:     "test",
		CertFile: "a",
		KeyFile:  "b",
	})
	assert.Nil(t, err)
	_, _, err = provider.GetCert("localhost")
	assert.NotNil(t, err)
}

func TestID(t *testing.T) {
	provider := LocalCertFileProvider{}
	err := provider.Init(LocalCertFileProviderConfig{
		Name:     "test",
		CertFile: "a",
		KeyFile:  "b",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", provider.ID())
}

// TestCertGetTestFile tests the GetCert function with a valid cert file
func TestCertGetTestFile(t *testing.T) {
	provider := LocalCertFileProvider{}
	err := provider.Init(LocalCertFileProviderConfig{
		Name:     "test",
		CertFile: "test_cert.crt",
		KeyFile:  "test_key.key",
	})
	assert.Nil(t, err)
	_, _, err = provider.GetCert("localhost")
	assert.Nil(t, err)
}
