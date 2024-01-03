/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package autogen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertGeneration(t *testing.T) {
	provider := AutoGenCertProvider{}
	err := provider.Init(AutoGenCertProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)
	_, _, err = provider.GetCert("localhost")
	assert.Nil(t, err)
}

func TestID(t *testing.T) {
	provider := AutoGenCertProvider{}
	err := provider.Init(AutoGenCertProviderConfig{
		Name: "test",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", provider.ID())
}
