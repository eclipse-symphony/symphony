/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createStagingVendor() StagingVendor {
	vendor := StagingVendor{}
	return vendor
}
func TestStagingEndpoints(t *testing.T) {
	vendor := createStagingVendor()
	vendor.Route = "staging"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}
func TestStagingInfo(t *testing.T) {
	vendor := createStagingVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
