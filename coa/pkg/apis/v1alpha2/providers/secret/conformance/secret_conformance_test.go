/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package conformance

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/mock"
	"github.com/stretchr/testify/assert"
)

func TestConformanceGetSecretNotFound(t *testing.T) {
	provider := &mock.MockSecretProvider{}
	err := provider.Init(mock.MockSecretProviderConfig{})
	assert.Nil(t, err)
	GetSecretNotFound(t, provider)
}

func TestConformanceSuite(t *testing.T) {
	provider := &mock.MockSecretProvider{}
	err := provider.Init(mock.MockSecretProviderConfig{})
	assert.Nil(t, err)
	ConformanceSuite(t, provider)
}
