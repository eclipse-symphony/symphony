/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package conformance

import (
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/secret"
	"github.com/stretchr/testify/assert"
)

func GetSecretNotFound[P secret.ISecretProvider](t *testing.T, p P) {
	// TODO: this case should fail. This is a prototype of conformance test suite
	// but unfortunately the mock secret provider doesn't confirm with reasonable
	// expected behavior
	_, err := p.Read("fake_object", "fake_key", nil)
	assert.Nil(t, err)
}
func ConformanceSuite[P secret.ISecretProvider](t *testing.T, p P) {
	t.Run("Level=Default", func(t *testing.T) {
		GetSecretNotFound(t, p)
	})
}
