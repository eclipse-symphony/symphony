/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestInitWithMap(t *testing.T) {
	provider := HTTPReferenceProvider{}
	err := provider.InitWithMap(map[string]string{
		"name": "test",
		"url":  "http://localhost",
	})
	assert.Nil(t, err)
}

func TestProviderProperties(t *testing.T) {
	provider := HTTPReferenceProvider{
		Config: HTTPReferenceProviderConfig{
			Name:     "test",
			Url:      "http://localhost",
			TargetID: "target",
		},
	}
	assert.Equal(t, "test", provider.ID())
	assert.Equal(t, "target", provider.TargetID())
	assert.Equal(t, "v1alpha2.ReferenceHTTP", provider.ReferenceType())

	err := provider.Reconfigure(nil)
	assert.Nil(t, err)

	context := &contexts.ManagerContext{}
	provider.SetContext(context)
	assert.Equal(t, context, provider.Context)
}

func TestHTTPReferenceProviderGetandList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	properties := map[string]string{
		"name":   "test",
		"url":    ts.URL,
		"target": "target",
	}

	provider := &HTTPReferenceProvider{}
	err := provider.InitWithMap(properties)
	assert.Nil(t, err)

	config, err := HTTPReferenceProviderConfigFromMap(properties)
	assert.Nil(t, err)
	err = provider.Init(config)
	assert.Nil(t, err)

	_, err = provider.Get("id", "namespace", "group", "kind", "version", "ref")
	assert.Nil(t, err)

	_, err = provider.List("labelSelector", "fieldSelector", "namespace", "group", "kind", "version", "ref")
	assert.Nil(t, err)

	clonedProvider, err := provider.Clone(config)
	assert.Nil(t, err)
	assert.IsType(t, &HTTPReferenceProvider{}, clonedProvider)
}
