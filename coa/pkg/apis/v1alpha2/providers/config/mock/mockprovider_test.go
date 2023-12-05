/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	provider := MockConfigProvider{}
	err := provider.Init(MockConfigProviderConfig{})
	assert.Nil(t, err)
}
func TestGet(t *testing.T) {
	provider := MockConfigProvider{}
	err := provider.Init(MockConfigProviderConfig{})
	assert.Nil(t, err)
	val, err := provider.Get("obj", "field", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "obj::field", val)
}
