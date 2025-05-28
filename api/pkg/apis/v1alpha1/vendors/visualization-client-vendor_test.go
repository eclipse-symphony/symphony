/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
)

func createVisualizationClientVendor(t *testing.T) VisualizationClientVendor {
	pubSubProvider := memory.InMemoryPubSubProvider{}
	pubSubProvider.Init(memory.InMemoryPubSubConfig{Name: "test"})
	vendor := VisualizationClientVendor{}
	err := vendor.Init(vendors.VendorConfig{}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, nil, &pubSubProvider)
	assert.Nil(t, err)
	return vendor
}

func TestVisualizationClientEndpoints(t *testing.T) {
	vendor := createVisualizationClientVendor(t)
	vendor.Route = "vis-client"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
}

func TestVisualizationClientInfo(t *testing.T) {
	vendor := createVisualizationClientVendor(t)
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}
