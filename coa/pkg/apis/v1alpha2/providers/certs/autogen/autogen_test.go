/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package autogen

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

// TestSetContext tests that the SetContext method returns an error
func TestSetContext(t *testing.T) {
	provider := AutoGenCertProvider{}
	err := provider.SetContext(contexts.ManagerContext{})
	assert.NotNil(t, err)
}

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
