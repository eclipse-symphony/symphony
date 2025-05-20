/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

const (
	id        = "testid"
	namespace = "testns"
	group     = "testgroup"
	kind      = "testkind"
	version   = "testversion"
)

func TestHttpProviderInit(t *testing.T) {
	provider := HTTPReporter{}
	properties := map[string]string{
		"name": "test",
		"url":  "http://localhost:8080",
	}
	err := provider.InitWithMap(properties)
	assert.Nil(t, err)
}

func TestHttpProviderProcess(t *testing.T) {
	ts := InitializeMockSymphonyAPI(t)
	provider := HTTPReporter{}
	properties := map[string]string{
		"name": "test",
		"url":  ts.URL + "/",
	}
	err := provider.InitWithMap(properties)
	assert.Nil(t, err)
	err = provider.Report(id, namespace, group, kind, version, map[string]string{}, false)
	assert.Nil(t, err)
}

func TestHttpProviderClone(t *testing.T) {
	provider := HTTPReporter{}
	properties := map[string]string{
		"name": "test",
		"url":  "http://localhost:8080",
	}
	err := provider.InitWithMap(properties)
	provider.SetContext(&contexts.ManagerContext{})
	assert.Nil(t, err)
	clone, err := provider.Clone(nil)
	assert.Nil(t, err)
	assert.Equal(t, provider.Config, clone.(*HTTPReporter).Config)

	config := HTTPReporterConfig{}
	clone, err = provider.Clone(config)
	assert.Nil(t, err)
	assert.Equal(t, config, clone.(*HTTPReporter).Config)
}

func InitializeMockSymphonyAPI(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, id, r.URL.Query().Get("id"))
		assert.Equal(t, namespace, r.URL.Query().Get("namespace"))
		assert.Equal(t, group, r.URL.Query().Get("group"))
		assert.Equal(t, kind, r.URL.Query().Get("kind"))
		assert.Equal(t, version, r.URL.Query().Get("version"))

		json.NewEncoder(w).Encode(response)
	}))
	return ts
}
