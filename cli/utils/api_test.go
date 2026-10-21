/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMissingResourcesAreErrorsExceptDelete(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		if _, err := callRestAPI(server.URL, "/missing", method, nil, "", nil); err == nil {
			t.Errorf("%s silently accepted HTTP 404", method)
		}
	}
	if _, err := callRestAPI(server.URL, "/missing", http.MethodDelete, nil, "", nil); err != nil {
		t.Fatalf("idempotent delete failed: %v", err)
	}
}

func TestAPIRequestsHonorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := GetInNamespace(ctx, "http://127.0.0.1:1", "", "", "target", "", "", "edge", "factory")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("request did not honor context cancellation: %v", err)
	}
}
