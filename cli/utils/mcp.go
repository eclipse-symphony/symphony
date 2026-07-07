/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
)

// mcpRoute is the Symphony API route exposing the MCP (Model Context Protocol)
// Streamable HTTP endpoint.
const mcpRoute = "/mcp"

// MCPTool describes a tool advertised by the Symphony MCP server.
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// mcpRPC performs a single JSON-RPC call against the Symphony MCP endpoint,
// authenticating as the given user so the server executes the request under the
// caller's identity.
func mcpRPC(url string, username string, password string, method string, params interface{}) (json.RawMessage, error) {
	token, err := Login(url, username, password)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, err
	}
	resp, err := callRestAPI(url, mcpRoute, "POST", payload, token, nil)
	if err != nil {
		return nil, err
	}
	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %v", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP error [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// MCPListTools returns the tools advertised by the Symphony MCP server.
func MCPListTools(url string, username string, password string) ([]MCPTool, error) {
	result, err := mcpRPC(url, username, password, "tools/list", nil)
	if err != nil {
		return nil, err
	}
	var listResult struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := json.Unmarshal(result, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %v", err)
	}
	return listResult.Tools, nil
}

// MCPCallTool invokes a tool on the Symphony MCP server and returns its textual
// result. Tool-level failures (the server's "isError" results) are returned as
// a Go error so callers can surface them to the model or the user.
func MCPCallTool(url string, username string, password string, name string, arguments map[string]interface{}) (string, error) {
	if arguments == nil {
		arguments = map[string]interface{}{}
	}
	result, err := mcpRPC(url, username, password, "tools/call", map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	})
	if err != nil {
		return "", err
	}
	var callResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(result, &callResult); err != nil {
		return "", fmt.Errorf("failed to parse tool result: %v", err)
	}
	text := ""
	for _, c := range callResult.Content {
		text += c.Text
	}
	if callResult.IsError {
		if text == "" {
			return "", errors.New("tool call failed")
		}
		return "", errors.New(text)
	}
	return text, nil
}
