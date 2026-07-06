/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func createMCPVendor(baseUrl string) *MCPVendor {
	vendor := &MCPVendor{}
	_ = vendor.Init(vendors.VendorConfig{
		Type:  "vendors.mcp",
		Route: "mcp",
		Properties: map[string]string{
			"baseUrl": baseUrl,
		},
	}, nil, nil, nil)
	return vendor
}

func rpcCall(t *testing.T, vendor *MCPVendor, method string, id string, params interface{}) jsonRPCResponse {
	return rpcCallWithToken(t, vendor, method, id, params, "Bearer tok")
}

func rpcCallWithToken(t *testing.T, vendor *MCPVendor, method string, id string, params interface{}, token string) jsonRPCResponse {
	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		assert.Nil(t, err)
		rawParams = b
	}
	body, err := json.Marshal(jsonRPCRequest{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage(id),
		Method:  method,
		Params:  rawParams,
	})
	assert.Nil(t, err)
	var metadata map[string]string
	if token != "" {
		metadata = map[string]string{"Authorization": token}
	}
	resp := vendor.onMCP(v1alpha2.COARequest{
		Context:  context.Background(),
		Method:   fasthttp.MethodPost,
		Body:     body,
		Metadata: metadata,
	})
	var rpcResp jsonRPCResponse
	err = json.Unmarshal(resp.Body, &rpcResp)
	assert.Nil(t, err)
	return rpcResp
}

func TestMCPGetInfo(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	info := vendor.GetInfo()
	assert.Equal(t, "MCP", info.Name)
}

func TestMCPGetEndpoints(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	endpoints := vendor.GetEndpoints()
	assert.Equal(t, 1, len(endpoints))
	assert.Equal(t, "mcp", endpoints[0].Route)
	assert.Contains(t, endpoints[0].Methods, fasthttp.MethodPost)
	assert.Contains(t, endpoints[0].Methods, fasthttp.MethodGet)
}

func TestMCPInitialize(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "initialize", "1", map[string]interface{}{
		"protocolVersion": "2024-11-05",
	})
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, "2024-11-05", result["protocolVersion"])
	assert.NotNil(t, result["capabilities"])
	assert.NotNil(t, result["serverInfo"])
}

func TestMCPPing(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "ping", "1", nil)
	assert.Nil(t, resp.Error)
}

func TestMCPToolsList(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "tools/list", "1", nil)
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	tools := result["tools"].([]interface{})
	assert.Equal(t, 4, len(tools))
	names := map[string]bool{}
	for _, tl := range tools {
		names[tl.(map[string]interface{})["name"].(string)] = true
	}
	assert.True(t, names["list_objects"])
	assert.True(t, names["get_object"])
	assert.True(t, names["create_object"])
	assert.True(t, names["delete_object"])
}

func TestMCPUnknownMethod(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "bogus/method", "1", nil)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, jsonRPCMethodNotFound, resp.Error.Code)
}

func TestMCPParseError(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := vendor.onMCP(v1alpha2.COARequest{
		Context: context.Background(),
		Method:  fasthttp.MethodPost,
		Body:    []byte("not-json"),
	})
	var rpcResp jsonRPCResponse
	err := json.Unmarshal(resp.Body, &rpcResp)
	assert.Nil(t, err)
	assert.NotNil(t, rpcResp.Error)
	assert.Equal(t, jsonRPCParseError, rpcResp.Error.Code)
}

func TestMCPNotificationNoBody(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	body, _ := json.Marshal(jsonRPCRequest{
		JSONRPC: jsonRPCVersion,
		Method:  "notifications/initialized",
	})
	resp := vendor.onMCP(v1alpha2.COARequest{
		Context: context.Background(),
		Method:  fasthttp.MethodPost,
		Body:    body,
	})
	assert.Equal(t, v1alpha2.Accepted, resp.State)
	assert.Empty(t, resp.Body)
}

func TestMCPGetNotAllowed(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := vendor.onMCP(v1alpha2.COARequest{
		Context: context.Background(),
		Method:  fasthttp.MethodGet,
	})
	assert.Equal(t, v1alpha2.MethodNotAllowed, resp.State)
}

func TestMCPToolCallListObjects(t *testing.T) {
	// Fake Symphony API: /users/auth returns a token, /targets/registry lists.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/auth":
			assert.Equal(t, http.MethodPost, r.Method)
			_, _ = w.Write([]byte(`{"accessToken":"tok"}`))
		case "/targets/registry":
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			_, _ = w.Write([]byte(`[{"metadata":{"name":"t1"}}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	vendor := createMCPVendor(server.URL)
	resp := rpcCall(t, vendor, "tools/call", "1", map[string]interface{}{
		"name": "list_objects",
		"arguments": map[string]interface{}{
			"objectType": "targets",
		},
	})
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, false, result["isError"])
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	assert.Contains(t, text, "t1")
}

func TestMCPToolCallCreateObject(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/auth":
			_, _ = w.Write([]byte(`{"accessToken":"tok"}`))
		case "/solutions/s1":
			assert.Equal(t, http.MethodPost, r.Method)
			b := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(b)
			receivedBody = string(b)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	vendor := createMCPVendor(server.URL)
	resp := rpcCall(t, vendor, "tools/call", "1", map[string]interface{}{
		"name": "create_object",
		"arguments": map[string]interface{}{
			"objectType": "solutions",
			"name":       "s1",
			"body":       map[string]interface{}{"spec": map[string]interface{}{"displayName": "s1"}},
		},
	})
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, false, result["isError"])
	assert.Contains(t, receivedBody, "displayName")
}

func TestMCPToolCallInvalidObjectType(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "tools/call", "1", map[string]interface{}{
		"name": "list_objects",
		"arguments": map[string]interface{}{
			"objectType": "bogus",
		},
	})
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, true, result["isError"])
}

func TestMCPToolCallUnknownTool(t *testing.T) {
	vendor := createMCPVendor("http://localhost")
	resp := rpcCall(t, vendor, "tools/call", "1", map[string]interface{}{
		"name":      "no_such_tool",
		"arguments": map[string]interface{}{},
	})
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, true, result["isError"])
}

func TestMCPToolCallRequiresCallerToken(t *testing.T) {
	// Without a forwarded caller token, tool calls that reach the API must fail.
	vendor := createMCPVendor("http://localhost")
	resp := rpcCallWithToken(t, vendor, "tools/call", "1", map[string]interface{}{
		"name": "list_objects",
		"arguments": map[string]interface{}{
			"objectType": "targets",
		},
	}, "")
	assert.Nil(t, resp.Error)
	result := resp.Result.(map[string]interface{})
	assert.Equal(t, true, result["isError"])
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	assert.Contains(t, text, "authenticated caller")
}
