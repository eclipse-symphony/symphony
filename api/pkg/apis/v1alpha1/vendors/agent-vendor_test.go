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
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/reference"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	refmock "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reference/mock"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/reporter/http"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func creatAgentVendor() AgentVendor {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{
		Values: map[string]interface{}{
			"testId": "testValue",
		},
	})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := reference.ReferenceManager{}

	manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.reference":     "reference",
			"providers.volatilestate": "memory",
			"providers.reporter":      "report",
		},
	}, map[string]providers.IProvider{
		"reference": &referenceProvider,
		"memory":    &stateProvider,
		"report":    &MockReporter{},
	})

	vendor := AgentVendor{
		ReferenceManager: &manager,
	}
	return vendor
}

func TestAgentVendorInit(t *testing.T) {
	referenceProvider := refmock.MockReferenceProvider{}
	referenceProvider.Init(refmock.MockReferenceProviderConfig{})
	stateProvider := memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	reporterProvider := http.HTTPReporter{}
	reporterProvider.Init(http.HTTPReporterConfig{})
	vendor := AgentVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: map[string]string{
			"test": "true",
		},
		Managers: []managers.ManagerConfig{
			{
				Name: "reference-manager",
				Type: "managers.symphony.reference",
				Properties: map[string]string{
					"providers.reference":     "reference",
					"providers.volatilestate": "mem-state",
					"providers.reporter":      "report",
				},
				Providers: map[string]managers.ProviderConfig{
					"reference": {
						Type:   "providers.reference.mock",
						Config: refmock.MockReferenceProviderConfig{},
					},
					"mem-state": {
						Type:   "providers.state.memory",
						Config: memorystate.MemoryStateProviderConfig{},
					},
					"report": {
						Type:   "providers.reporter.http",
						Config: http.HTTPReporterConfig{},
					},
				},
			},
		},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{
		"reference-manager": {
			"reference": &referenceProvider,
			"mem-state": &stateProvider,
			"report":    &reporterProvider,
		},
	}, nil)
	assert.Nil(t, err)
}

func TestAgentEndpoints(t *testing.T) {
	vendor := creatAgentVendor()
	vendor.Route = "agent"
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 2, len(endpoints))
}

func TestAgentInfo(t *testing.T) {
	vendor := creatAgentVendor()
	vendor.Version = "1.0"
	info := vendor.GetInfo()
	assert.NotNil(t, info)
	assert.Equal(t, "1.0", info.Version)
}

func TestApplyConfig(t *testing.T) {
	vendor := creatAgentVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}
	res := vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.MethodNotAllowed, res.State)

	request.Method = fasthttp.MethodPost
	config := managers.ProviderConfig{
		Type:   "providers.reference.customvision",
		Config: managers.ProviderConfig{},
	}
	data, err := json.Marshal(config)
	assert.Nil(t, err)
	request.Body = data
	res = vendor.onConfig(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
}

func TestPostReference(t *testing.T) {
	vendor := creatAgentVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodDelete,
		Context: context.Background(),
	}
	res := vendor.onReference(*request)
	assert.Equal(t, v1alpha2.MethodNotAllowed, res.State)

	request.Method = fasthttp.MethodPost
	request.Parameters = map[string]string{
		"namespace": "test",
		"kind":      "kind",
		"version":   "version",
		"group":     "group",
		"id":        "id",
		"overwrite": "false",
	}
	body := map[string]string{
		"property1": "test1",
		"property2": "test2",
	}
	data, err := json.Marshal(body)
	assert.Nil(t, err)
	request.Body = data
	res = vendor.onReference(*request)
	assert.Equal(t, v1alpha2.OK, res.State)

	request.Parameters = map[string]string{
		"id": "uploaderror",
	}
	res = vendor.onReference(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)

	bodyErr := "property1"
	data, err = json.Marshal(bodyErr)
	assert.Nil(t, err)
	request.Body = data
	res = vendor.onReference(*request)
	assert.Equal(t, v1alpha2.InternalError, res.State)
}

func TestGetReference(t *testing.T) {
	vendor := creatAgentVendor()
	request := &v1alpha2.COARequest{
		Method:  fasthttp.MethodGet,
		Context: context.Background(),
	}

	request.Parameters = map[string]string{
		"id": "testId",
	}
	res := vendor.onReference(*request)
	assert.Equal(t, v1alpha2.OK, res.State)

	request.Parameters = map[string]string{
		"namespace": "scope",
		"kind":      "kind",
		"version":   "version",
		"group":     "group",
		"platform":  "platform",
		"flavor":    "flavor",
		"iteration": "iteration",
		"alias":     "alias",
		"ref":       "reference",
		"id":        "testId",
		"instance":  "testId",
	}
	res = vendor.onReference(*request)
	assert.Equal(t, v1alpha2.OK, res.State)
}

type MockReporter struct{}

func (r *MockReporter) Init(config providers.IProviderConfig) error {
	return nil
}
func (r *MockReporter) Report(id string, namespace string, group string, kind string, version string, properties map[string]string, overwrite bool) error {
	if id == "uploaderror" {
		return &json.SyntaxError{}
	}
	return nil
}
