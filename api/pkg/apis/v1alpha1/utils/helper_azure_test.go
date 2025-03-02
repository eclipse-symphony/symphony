//go:build azure

/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertAzureSolutionVersionReferenceToObjectName(t *testing.T) {
	var azureSolutionVersionRef = "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourcegroups/xingdlitest/providers/Private.Edge/targets/target3/solutions/sol3/versions/ver1"
	objName, success := ConvertAzureSolutionVersionReferenceToObjectName(azureSolutionVersionRef)
	assert.Equal(t, "target3-v-sol3-v-ver1", objName)
	assert.True(t, success)
}

func TestConvertAzureSolutionVersionReferenceToObjectNameWithInvalidReference(t *testing.T) {
	var azureSolutionVersionRef = "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourceGroups/xingdlitest/providers/Private.Edge/targets/target3/solutions/sol3/versions"
	objName, success := ConvertAzureSolutionVersionReferenceToObjectName(azureSolutionVersionRef)
	assert.Equal(t, "", objName)
	assert.False(t, success)
}

func TestConvertAzureTargetReferenceToObjectName(t *testing.T) {
	var azureTargetRef = "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourcegroups/xingdlitest/providers/Private.Edge/targets/target3"
	objName, success := ConvertAzureTargetReferenceToObjectName(azureTargetRef)
	assert.Equal(t, "target3", objName)
	assert.True(t, success)
}

func TestConvertAzureTargetReferenceToObjectNameWithInvalidReference(t *testing.T) {
	var azureTargetRef = "/subscriptions/973d15c6-6c57-447e-b9c6-6d79b5b784ab/resourcegroups/xingdlitest/providers/Private.Edge/targets"
	objName, success := ConvertAzureTargetReferenceToObjectName(azureTargetRef)
	assert.Equal(t, "", objName)
	assert.False(t, success)
}
