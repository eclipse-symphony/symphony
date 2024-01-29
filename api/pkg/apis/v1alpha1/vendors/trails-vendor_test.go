/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	mockledger "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/ledger/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createTrailsVendor(route string) TrailsVendor {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	_ = ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	vendor := TrailsVendor{}
	_ = vendor.Init(vendors.VendorConfig{
		Route: route,
		Type:  "vendors.trails",
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name:       "trails-manager",
				Type:       "managers.symphony.trails",
				Properties: map[string]string{},
				Providers: map[string]managers.ProviderConfig{
					"mock": {
						Type:   "providers.ledger.mock",
						Config: mockledger.MockLedgerProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"trails-manager": {
			"mock": ledgerProvider,
		},
	}, nil)
	return vendor
}

func TestTrailsVendorInit(t *testing.T) {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)

	vendor := TrailsVendor{}
	err = vendor.Init(vendors.VendorConfig{
		Type: "vendors.trails",
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name:       "trails-manager",
				Type:       "managers.symphony.trails",
				Properties: map[string]string{},
				Providers: map[string]managers.ProviderConfig{
					"mock": {
						Type:   "providers.ledger.mock",
						Config: mockledger.MockLedgerProviderConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"trails-manager": {
			"mock": ledgerProvider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestTrailsVendorInitFail(t *testing.T) {
	ledgerProvider := &mockledger.MockLedgerProvider{}
	err := ledgerProvider.Init(mockledger.MockLedgerProviderConfig{})
	assert.Nil(t, err)

	vendor := TrailsVendor{}
	// no trails manager
	err = vendor.Init(vendors.VendorConfig{
		Type: "vendors.trails",
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"trails-manager": {
			"mock": ledgerProvider,
		},
	}, nil)
	assert.NotNil(t, err)
	coaError := err.(v1alpha2.COAError)
	assert.Equal(t, v1alpha2.MissingConfig, coaError.State)
	assert.Equal(t, "trails manager is not supplied", coaError.Message)
}

func TestTrailsVendorGetInfo(t *testing.T) {
	vendor := createTrailsVendor("")
	info := vendor.GetInfo()
	assert.Equal(t, "Trails", info.Name)
	assert.Equal(t, "Microsoft", info.Producer)
}

func TestTrailsVendorEndopints(t *testing.T) {
	vendor := createTrailsVendor("")
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))

	vendor = createTrailsVendor("trails")
	endpoints = vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, "trails", endpoints[0].Route)
}

func TestTrailsVendorOnTrails_PostEmptyArrayAsBody(t *testing.T) {
	vendor := createTrailsVendor("")
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    []byte("[]"),
		Context: context.Background(),
	}
	response := vendor.onTrails(*request)
	assert.Equal(t, v1alpha2.OK, response.State)
}

func TestTrailsVendorOnTrails_PostErrorArrayAsBody(t *testing.T) {
	vendor := createTrailsVendor("")
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    []byte("[error]"),
		Context: context.Background(),
	}
	response := vendor.onTrails(*request)
	assert.Equal(t, v1alpha2.InternalError, response.State)
}

func TestTrailsVendorOnTrails_PostTrailsArrayAsBody(t *testing.T) {
	trails := []v1alpha2.Trail{
		{
			Origin:  "site1",
			Catalog: "catalog1",
			Type:    "solutions.solution.symphony/v1",
			Properties: map[string]interface{}{
				"post": model.CatalogSpec{
					Name: "test1",
				},
			},
		},
		{
			Origin:  "site2",
			Catalog: "catalog2",
			Type:    "solutions.solution.symphony/v1",
			Properties: map[string]interface{}{
				"post": model.CatalogSpec{
					Name: "test2",
				},
			},
		},
	}
	data, err := json.Marshal(trails)
	assert.Nil(t, err)

	vendor := createTrailsVendor("")
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodPost,
		Body:    data,
		Context: context.Background(),
	}
	response := vendor.onTrails(*request)
	assert.Equal(t, v1alpha2.OK, response.State)
}

func TestTrailsVendorOnTrails_Get(t *testing.T) {
	vendor := createTrailsVendor("")
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	response := vendor.onTrails(*request)
	assert.Equal(t, v1alpha2.MethodNotAllowed, response.State)
}
