/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var mcpLog = logger.NewLogger("coa.runtime")

const (
	mcpProtocolVersion = "2024-11-05"

	jsonRPCVersion        = "2.0"
	jsonRPCParseError     = -32700
	jsonRPCInvalidRequest = -32600
	jsonRPCMethodNotFound = -32601
	jsonRPCInvalidParams  = -32602
	jsonRPCInternalError  = -32603
)

// objectRoutes maps the object types exposed as MCP tools to their Symphony API
// registry routes.
var objectRoutes = map[string]string{
	"targets":          "/targets/registry",
	"solutions":        "/solutions",
	"solutionversions": "/solutionversions",
	"instances":        "/instances",
	"catalogs":         "/catalogs",
	"catalogversions":  "/catalogversions/registry",
	"campaigns":        "/campaigns",
	"campaignversions": "/campaignversions",
	"activations":      "/activations/registry",
	"devices":          "/devices",
}

// MCPVendor exposes Symphony operations to Model Context Protocol (MCP) clients
// over the Streamable HTTP transport. It reuses the existing HTTP binding: the
// transport is a single JSON-RPC endpoint (POST /<version>/mcp). Tools are
// backed by calls to the Symphony REST API.
//
// Tool calls are executed on behalf of the authenticated caller: the vendor
// forwards the caller's bearer token to the underlying REST API so every tool
// operation is subject to the same authentication and RBAC as a direct API
// call. The vendor never uses ambient/stored credentials.
//
// Configuration (vendor properties, optional):
//   - "baseUrl": base URL of the Symphony API the tools call. Defaults to the
//     site's current base URL.
type MCPVendor struct {
	vendors.Vendor
	apiBaseUrl string
}

func (o *MCPVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "MCP",
		Producer: "Microsoft",
	}
}

func (e *MCPVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}

	e.apiBaseUrl = e.Context.SiteInfo.CurrentSite.BaseUrl
	if config.Properties != nil {
		if v, ok := config.Properties["baseUrl"]; ok && v != "" {
			e.apiBaseUrl = coa_utils.ParseProperty(v)
		}
	}
	return nil
}

func (o *MCPVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "mcp"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods: []string{fasthttp.MethodPost, fasthttp.MethodGet},
			Route:   route,
			Version: o.Version,
			Handler: o.onMCP,
		},
	}
}

// ---- JSON-RPC types ----

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

func (c *MCPVendor) onMCP(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("MCP Vendor", request.Context, &map[string]string{
		"method": "onMCP",
	})
	defer span.End()
	mcpLog.InfofCtx(pCtx, "V (MCP): onMCP, method: %s", request.Method)

	// The Streamable HTTP transport uses GET to open a server-to-client SSE
	// stream. This server does not push server-initiated messages, so decline.
	if request.Method == fasthttp.MethodGet {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.MethodNotAllowed,
			Body:        []byte("SSE stream is not supported by this server"),
			ContentType: "text/plain",
		})
	}

	var rpcReq jsonRPCRequest
	if err := json.Unmarshal(request.Body, &rpcReq); err != nil {
		return observ_utils.CloseSpanWithCOAResponse(span, rpcErrorResponse(nil, jsonRPCParseError, "failed to parse JSON-RPC request", err.Error()))
	}

	// The caller's bearer token (forwarded by the HTTP binding) is used to
	// authorize any downstream tool operations, so tools act as the caller.
	authToken := ""
	if request.Metadata != nil {
		authToken = request.Metadata["Authorization"]
	}

	// Notifications (no id) require no response body.
	if len(rpcReq.ID) == 0 {
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.Accepted,
			ContentType: "application/json",
		})
	}

	switch rpcReq.Method {
	case "initialize":
		return observ_utils.CloseSpanWithCOAResponse(span, c.handleInitialize(rpcReq))
	case "ping":
		return observ_utils.CloseSpanWithCOAResponse(span, rpcResultResponse(rpcReq.ID, map[string]interface{}{}))
	case "tools/list":
		return observ_utils.CloseSpanWithCOAResponse(span, rpcResultResponse(rpcReq.ID, map[string]interface{}{"tools": toolDefinitions()}))
	case "tools/call":
		return observ_utils.CloseSpanWithCOAResponse(span, c.handleToolCall(pCtx, rpcReq, authToken))
	default:
		return observ_utils.CloseSpanWithCOAResponse(span, rpcErrorResponse(rpcReq.ID, jsonRPCMethodNotFound, fmt.Sprintf("method '%s' is not supported", rpcReq.Method), nil))
	}
}

func (c *MCPVendor) handleInitialize(rpcReq jsonRPCRequest) v1alpha2.COAResponse {
	protocolVersion := mcpProtocolVersion
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(rpcReq.Params, &params); err == nil && params.ProtocolVersion != "" {
		protocolVersion = params.ProtocolVersion
	}
	return rpcResultResponse(rpcReq.ID, map[string]interface{}{
		"protocolVersion": protocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "symphony-mcp",
			"version": c.Version,
		},
	})
}

func (c *MCPVendor) handleToolCall(ctx context.Context, rpcReq jsonRPCRequest, authToken string) v1alpha2.COAResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(rpcReq.Params, &params); err != nil {
		return rpcErrorResponse(rpcReq.ID, jsonRPCInvalidParams, "invalid tool call parameters", err.Error())
	}
	if params.Arguments == nil {
		params.Arguments = map[string]interface{}{}
	}

	text, err := c.dispatchTool(ctx, params.Name, params.Arguments, authToken)
	if err != nil {
		// Tool execution errors are reported inside the result with isError,
		// per the MCP spec, so the model can react to them.
		return rpcResultResponse(rpcReq.ID, toolResult(err.Error(), true))
	}
	return rpcResultResponse(rpcReq.ID, toolResult(text, false))
}

func (c *MCPVendor) dispatchTool(ctx context.Context, name string, args map[string]interface{}, authToken string) (string, error) {
	switch name {
	case "list_objects":
		objectType, route, err := resolveObjectRoute(args)
		if err != nil {
			return "", err
		}
		body, err := c.callAPI(ctx, http.MethodGet, route, queryParams(args), nil, authToken)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Objects of type '%s':\n%s", objectType, string(body)), nil
	case "get_object":
		objectType, route, err := resolveObjectRoute(args)
		if err != nil {
			return "", err
		}
		objName := argString(args, "name")
		if objName == "" {
			return "", fmt.Errorf("'name' is required")
		}
		body, err := c.callAPI(ctx, http.MethodGet, route+"/"+objName, queryParams(args), nil, authToken)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s '%s':\n%s", objectType, objName, string(body)), nil
	case "create_object":
		_, route, err := resolveObjectRoute(args)
		if err != nil {
			return "", err
		}
		objName := argString(args, "name")
		if objName == "" {
			return "", fmt.Errorf("'name' is required")
		}
		spec, ok := args["body"]
		if !ok {
			return "", fmt.Errorf("'body' is required")
		}
		payload, err := json.Marshal(spec)
		if err != nil {
			return "", fmt.Errorf("failed to serialize 'body': %v", err)
		}
		if _, err := c.callAPI(ctx, http.MethodPost, route+"/"+objName, queryParams(args), payload, authToken); err != nil {
			return "", err
		}
		return fmt.Sprintf("Created/updated '%s'", objName), nil
	case "delete_object":
		_, route, err := resolveObjectRoute(args)
		if err != nil {
			return "", err
		}
		objName := argString(args, "name")
		if objName == "" {
			return "", fmt.Errorf("'name' is required")
		}
		if _, err := c.callAPI(ctx, http.MethodDelete, route+"/"+objName, queryParams(args), nil, authToken); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deleted '%s'", objName), nil
	default:
		return "", fmt.Errorf("unknown tool '%s'", name)
	}
}

// ---- Symphony API access ----

func (c *MCPVendor) callAPI(ctx context.Context, method string, route string, params map[string]string, payload []byte, authToken string) ([]byte, error) {
	if c.apiBaseUrl == "" {
		return nil, fmt.Errorf("the MCP vendor is not configured with a Symphony API base URL")
	}
	if authToken == "" {
		return nil, fmt.Errorf("tool calls require an authenticated caller")
	}
	return c.callRestAPI(ctx, method, route, payload, authToken, params)
}

func (c *MCPVendor) callRestAPI(ctx context.Context, method string, route string, payload []byte, token string, params map[string]string) ([]byte, error) {
	rUrl := strings.TrimRight(c.apiBaseUrl, "/") + route
	req, err := http.NewRequestWithContext(ctx, method, rUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	if len(params) > 0 {
		query := req.URL.Query()
		for k, v := range params {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Symphony API returned [%d]: %s", resp.StatusCode, string(bodyBytes))
	}
	return bodyBytes, nil
}

// ---- helpers ----

func resolveObjectRoute(args map[string]interface{}) (string, string, error) {
	objectType := argString(args, "objectType")
	if objectType == "" {
		return "", "", fmt.Errorf("'objectType' is required")
	}
	route, ok := objectRoutes[objectType]
	if !ok {
		return "", "", fmt.Errorf("unsupported object type '%s'", objectType)
	}
	return objectType, route, nil
}

func queryParams(args map[string]interface{}) map[string]string {
	params := map[string]string{}
	if ns := argString(args, "namespace"); ns != "" {
		params["namespace"] = ns
	}
	return params
}

func argString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func rpcResultResponse(id json.RawMessage, result interface{}) v1alpha2.COAResponse {
	data, _ := json.Marshal(jsonRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	})
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

func rpcErrorResponse(id json.RawMessage, code int, message string, errData interface{}) v1alpha2.COAResponse {
	data, _ := json.Marshal(jsonRPCResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
			Data:    errData,
		},
	})
	return v1alpha2.COAResponse{
		State:       v1alpha2.OK,
		Body:        data,
		ContentType: "application/json",
	}
}

func toolResult(text string, isError bool) map[string]interface{} {
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": text},
		},
		"isError": isError,
	}
}

func objectTypeEnum() []string {
	types := make([]string, 0, len(objectRoutes))
	for t := range objectRoutes {
		types = append(types, t)
	}
	return types
}

func toolDefinitions() []map[string]interface{} {
	objectTypeProp := map[string]interface{}{
		"type":        "string",
		"description": "The Symphony object type.",
		"enum":        objectTypeEnum(),
	}
	nameProp := map[string]interface{}{
		"type":        "string",
		"description": "The name of the object.",
	}
	namespaceProp := map[string]interface{}{
		"type":        "string",
		"description": "Optional namespace of the object.",
	}
	return []map[string]interface{}{
		{
			"name":        "list_objects",
			"description": "List all Symphony objects of a given type.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"objectType": objectTypeProp,
					"namespace":  namespaceProp,
				},
				"required": []string{"objectType"},
			},
		},
		{
			"name":        "get_object",
			"description": "Get a single Symphony object of a given type by name.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"objectType": objectTypeProp,
					"name":       nameProp,
					"namespace":  namespaceProp,
				},
				"required": []string{"objectType", "name"},
			},
		},
		{
			"name":        "create_object",
			"description": "Create or update a Symphony object of a given type. 'body' is the full object definition.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"objectType": objectTypeProp,
					"name":       nameProp,
					"namespace":  namespaceProp,
					"body": map[string]interface{}{
						"type":        "object",
						"description": "The full object definition to create or update.",
					},
				},
				"required": []string{"objectType", "name", "body"},
			},
		},
		{
			"name":        "delete_object",
			"description": "Delete a Symphony object of a given type by name.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"objectType": objectTypeProp,
					"name":       nameProp,
					"namespace":  namespaceProp,
				},
				"required": []string{"objectType", "name"},
			},
		},
	}
}
