/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1alpha2

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/stretchr/testify/assert"
)

func TestCOARequestDeepCopyInto(t *testing.T) {
	reqIn := &COARequest{
		Context:     context.TODO(),
		Method:      "GET",
		Route:       "/test",
		ContentType: "application/json",
		Body:        []byte("test"),
		Metadata:    map[string]string{"test": "test"},
		Parameters:  map[string]string{"test": "test"},
	}

	reqOut := &COARequest{}
	reqIn.DeepCopyInto(reqOut)

	assert.NotNil(t, reqOut.Context)
	assert.Equal(t, reqIn.Method, reqOut.Method)
	assert.Equal(t, reqIn.Route, reqOut.Route)
	assert.Equal(t, reqIn.ContentType, reqOut.ContentType)
	assert.Equal(t, reqIn.Body, reqOut.Body)
	assert.Equal(t, "test", reqOut.Metadata["test"])
	assert.Equal(t, "test", reqOut.Parameters["test"])
}

func TestCOARequestDeepCopy(t *testing.T) {
	reqIn := &COARequest{
		Context:     context.TODO(),
		Method:      "GET",
		Route:       "/test",
		ContentType: "application/json",
		Body:        []byte("test"),
		Metadata:    map[string]string{"test": "test"},
		Parameters:  map[string]string{"test": "test"},
	}

	reqOut := reqIn.DeepCopy()

	assert.NotNil(t, reqOut.Context)
	assert.Equal(t, reqIn.Method, reqOut.Method)
	assert.Equal(t, reqIn.Route, reqOut.Route)
	assert.Equal(t, reqIn.ContentType, reqOut.ContentType)
	assert.Equal(t, reqIn.Body, reqOut.Body)
	assert.Equal(t, "test", reqOut.Metadata["test"])
	assert.Equal(t, "test", reqOut.Parameters["test"])

	var reqIn2 *COARequest = nil
	reqOut2 := reqIn2.DeepCopy()
	assert.Nil(t, reqOut2)
}

func TestCOAResponseString(t *testing.T) {
	resp := COAResponse{
		ContentType: "application/json",
		Body:        []byte("test"),
		State:       OK,
		Metadata:    map[string]string{"test": "test"},
		RedirectUri: "http://test.com",
	}

	assert.Equal(t, "test", resp.String())

	resp.Body = nil
	assert.Equal(t, "", resp.String())
}

func TestCOAResponsePrint(t *testing.T) {
	resp := COAResponse{
		ContentType: "application/json",
		Body:        []byte("test"),
		State:       OK,
		Metadata:    map[string]string{"test": "test"},
		RedirectUri: "http://test.com",
	}

	resp.Println()

	resp.Body = nil
	resp.Println()
}

func TestCOARequests_MarshalAndUnmarshalJSON(t *testing.T) {
	actCtx := contexts.NewActivityLogContext("diagnosticResourceId", "diagnosticResourceCloudLocation", "resourceCloudId", "resourceCloudLocation", "edgeLocation", "operationName", "correlationId", "callerId", "resourceK8SId")
	diagCtx := contexts.NewDiagnosticLogContext("correlationId", "resourceId", "traceId", "spanId")
	ctx := contexts.PatchActivityLogContextToCurrentContext(actCtx, diagCtx)
	ctx = contexts.PatchDiagnosticLogContextToCurrentContext(diagCtx, ctx)
	req := COARequest{
		Body:    []byte("body"),
		Route:   "/test",
		Method:  "GET",
		Context: ctx,
	}
	req.Metadata = map[string]string{
		"metadata1": "metadata1-value",
	}
	req.Parameters = map[string]string{
		"param1": "param1-value",
	}

	data, err := req.MarshalJSON()
	assert.Nil(t, err)

	var req2 COARequest
	err = req2.UnmarshalJSON(data)
	assert.Nil(t, err)

	assert.True(t, COARequestEquals(&req, &req2))
}
