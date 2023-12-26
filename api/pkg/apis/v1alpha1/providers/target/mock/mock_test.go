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

// TestKubectlTargetProviderConfigFromMapNil tests that passing nil to KubectlTargetProviderConfigFromMap returns a valid config
func TestInit(t *testing.T) {
	targetProvider := &MockTargetProvider{}
	err := targetProvider.Init(MockTargetProviderConfig{})
	assert.Nil(t, err)
}
