/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package conformance

import (
	"testing"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/providers/target/azure/adu"
	"github.com/stretchr/testify/assert"
)

func TestConformanceSuite(t *testing.T) {
	provider := &adu.ADUTargetProvider{}
	err := provider.Init(adu.ADUTargetProviderConfig{})
	assert.Nil(t, err)
	ConformanceSuite(t, provider)
}
