/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestDoHTTPRequestRetriesOn5xx(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "boom")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	assert.Nil(t, err)
	body, err := DoHTTPRequest(nil, req, 1, "test op")
	assert.Nil(t, err)
	assert.Equal(t, "ok", string(body))
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestDoHTTPRequestNoRetryOn4xx(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "missing")
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	assert.Nil(t, err)
	_, err = DoHTTPRequest(nil, req, 3, "test op")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "404")
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
	// 4xx maps to the corresponding COA state
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.NotFound, coaErr.State)
}

func TestDoHTTPRequestNoRetriesConfigured(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "boom")
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	assert.Nil(t, err)
	_, err = DoHTTPRequest(nil, req, 0, "test op")
	assert.NotNil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}
