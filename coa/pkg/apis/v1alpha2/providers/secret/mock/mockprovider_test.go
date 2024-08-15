/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret/conformance"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	provider := MockSecretProvider{}
	err := provider.InitWithMap(
		map[string]string{
			"name": "test",
		},
	)
	assert.Nil(t, err)
}

func TestID(t *testing.T) {
	provider := MockSecretProvider{}
	provider.Init(MockSecretProviderConfig{
		Name: "name",
	})
	assert.Equal(t, "name", provider.ID())
}

func TestSetContext(t *testing.T) {
	provider := MockSecretProvider{}
	provider.Init(MockSecretProviderConfig{
		Name: "name",
	})
	provider.SetContext(&contexts.ManagerContext{})
	assert.NotNil(t, provider.Context)
}

func TestGet(t *testing.T) {
	provider := MockSecretProvider{}
	err := provider.Init(MockSecretProviderConfig{})
	assert.Nil(t, err)
	val, err := provider.Read("obj", "field", nil)
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
