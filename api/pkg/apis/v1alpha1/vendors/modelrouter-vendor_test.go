/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	sym_mgr "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub/memory"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createModelRouterVendor(properties map[string]string) (*ModelRouterVendor, error) {
	pubsubProvider := memory.InMemoryPubSubProvider{}
	pubsubProvider.Init(memory.InMemoryPubSubConfig{})
	vendor := ModelRouterVendor{}
	err := vendor.Init(vendors.VendorConfig{
		Properties: properties,
		Managers:   []managers.ManagerConfig{},
	}, []managers.IManagerFactroy{
		&sym_mgr.SymphonyManagerFactory{},
	}, map[string]map[string]providers.IProvider{}, &pubsubProvider)
	return &vendor, err
}

func TestModelRouterVendorInit(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints":       `[{"name":"openai","url":"https://api.openai.com","key":"sk-test"}]`,
		"defaultEndpoint": "openai",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vendor.Endpoints))
	assert.Equal(t, "openai", vendor.DefaultEndpoint)
	assert.Equal(t, "https://api.openai.com", vendor.Endpoints["openai"].URL)
}

func TestModelRouterVendorInitSingleEndpointDefaults(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints": `[{"name":"local","url":"http://localhost:8000"}]`,
	})
	assert.Nil(t, err)
	assert.Equal(t, "local", vendor.DefaultEndpoint)
}

func TestModelRouterVendorInitResolvesEnvKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-from-env")
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints":       `[{"name":"openai","url":"https://api.openai.com","key":"$env:OPENAI_API_KEY"}]`,
		"defaultEndpoint": "openai",
	})
	assert.Nil(t, err)
	assert.Equal(t, "sk-from-env", vendor.Endpoints["openai"].Key)
}

func TestModelRouterVendorInitInvalidJSON(t *testing.T) {
	_, err := createModelRouterVendor(map[string]string{
		"endpoints": `not-json`,
	})
	assert.NotNil(t, err)
}

func TestModelRouterVendorInitMissingURL(t *testing.T) {
	_, err := createModelRouterVendor(map[string]string{
		"endpoints": `[{"name":"broken"}]`,
	})
	assert.NotNil(t, err)
}

func TestModelRouterVendorGetInfo(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{})
	assert.Nil(t, err)
	info := vendor.GetInfo()
	assert.Equal(t, "ModelRouter", info.Name)
}

func TestModelRouterVendorGetEndpoints(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{})
	assert.Nil(t, err)
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 4, len(endpoints))
	assert.Equal(t, "modelrouter/chat/completions", endpoints[0].Route)
}

func TestModelRouterVendorProxyChatCompletions(t *testing.T) {
	var receivedAuth string
	var receivedPath string
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedPath = r.URL.Path
		receivedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-123"}`))
	}))
	defer server.Close()

	endpoints, _ := json.Marshal([]ModelEndpoint{{Name: "test", URL: server.URL, Key: "sk-abc"}})
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints":       string(endpoints),
		"defaultEndpoint": "test",
	})
	assert.Nil(t, err)

	handler := vendor.onProxy("/v1/chat/completions")
	resp := handler(v1alpha2.COARequest{
		Context:     context.Background(),
		Method:      fasthttp.MethodPost,
		Body:        []byte(`{"model":"gpt-4"}`),
		ContentType: "application/json",
		Parameters:  map[string]string{},
	})

	assert.Equal(t, v1alpha2.OK, resp.State)
	assert.Equal(t, "Bearer sk-abc", receivedAuth)
	assert.Equal(t, "/v1/chat/completions", receivedPath)
	assert.Equal(t, `{"model":"gpt-4"}`, string(receivedBody))
	assert.Equal(t, `{"id":"chatcmpl-123"}`, string(resp.Body))
}

func TestModelRouterVendorProxyEndpointSelection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer server.Close()

	endpoints, _ := json.Marshal([]ModelEndpoint{
		{Name: "a", URL: "http://invalid.invalid"},
		{Name: "b", URL: server.URL},
	})
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints": string(endpoints),
	})
	assert.Nil(t, err)

	handler := vendor.onProxy("/v1/models")
	resp := handler(v1alpha2.COARequest{
		Context:    context.Background(),
		Method:     fasthttp.MethodGet,
		Parameters: map[string]string{"endpoint": "b"},
	})
	assert.Equal(t, v1alpha2.OK, resp.State)
}

func TestModelRouterVendorProxyStatusPassthrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	endpoints, _ := json.Marshal([]ModelEndpoint{{Name: "test", URL: server.URL}})
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints": string(endpoints),
	})
	assert.Nil(t, err)

	handler := vendor.onProxy("/v1/chat/completions")
	resp := handler(v1alpha2.COARequest{
		Context:    context.Background(),
		Method:     fasthttp.MethodPost,
		Parameters: map[string]string{},
	})
	assert.Equal(t, v1alpha2.State(http.StatusTooManyRequests), resp.State)
	assert.Equal(t, `{"error":"rate limited"}`, string(resp.Body))
}

func TestModelRouterVendorProxyNoEndpoints(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{})
	assert.Nil(t, err)

	handler := vendor.onProxy("/v1/chat/completions")
	resp := handler(v1alpha2.COARequest{
		Context:    context.Background(),
		Method:     fasthttp.MethodPost,
		Parameters: map[string]string{},
	})
	assert.Equal(t, v1alpha2.BadRequest, resp.State)
}

func TestModelRouterVendorProxyUnknownEndpoint(t *testing.T) {
	vendor, err := createModelRouterVendor(map[string]string{
		"endpoints": `[{"name":"a","url":"http://localhost"}]`,
	})
	assert.Nil(t, err)

	handler := vendor.onProxy("/v1/chat/completions")
	resp := handler(v1alpha2.COARequest{
		Context:    context.Background(),
		Method:     fasthttp.MethodPost,
		Parameters: map[string]string{"endpoint": "missing"},
	})
	assert.Equal(t, v1alpha2.BadRequest, resp.State)
}
