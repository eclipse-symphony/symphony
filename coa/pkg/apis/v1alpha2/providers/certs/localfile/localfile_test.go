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
