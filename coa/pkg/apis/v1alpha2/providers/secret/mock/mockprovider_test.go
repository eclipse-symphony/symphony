/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/secret/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
}
func TestGet(t *testing.T) {
	provider := MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
	val, err := provider.Get("obj", "field")
	assert.Nil(t, err)
	assert.Equal(t, "obj>>field", val)
}

func TestConformanceGetSecretNotFound(t *testing.T) {
	provider := &MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
	conformance.GetSecretNotFound(t, provider)
}

func TestConformanceSuite(t *testing.T) {
	provider := &MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
	conformance.ConformanceSuite(t, provider)
}
