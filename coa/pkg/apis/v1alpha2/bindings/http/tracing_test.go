/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceStateFromHeader(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "vendorname1=opaqueValue1,vendorname2=opaqueValue2")
	state := tracestateFromRequest(request)
	val := state.Get("vendorname1")
	assert.Equal(t, "opaqueValue1", val)
	val = state.Get("vendorname2")
	assert.Equal(t, "opaqueValue2", val)
}

func TestTraceStateFromHeaderNil(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}

	state := tracestateFromRequest(request)
	assert.Nil(t, state)
}

func TestTraceStateFromHeaderEmpty(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "")
	state := tracestateFromRequest(request)
	assert.Nil(t, state)
}

func TestTraceStateFromHeaderBadFormat(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "as;lkgjdgasgasgjsdkgjas;ldgkj;kjg")
	state := tracestateFromRequest(request)
	assert.Nil(t, state)
}
func TestTraceStateFromHeaderConfusedFormat1(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "a=b=c")
	state := tracestateFromRequest(request)
	assert.Nil(t, state)
}
func TestTraceStateFromHeaderConfusedFormat2(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "a==c")
	state := tracestateFromRequest(request)
	assert.Nil(t, state)
}
func TestTraceStateFromHeaderWithSpaces(t *testing.T) {
	request := &fasthttp.Request{
		Header: fasthttp.RequestHeader{},
	}
	request.Header.Add("tracestate", "  		vendorname1 = opaqueValue1, vendorname2 = opaqueValue2 ")
	state := tracestateFromRequest(request)
	val := state.Get("vendorname1")
	assert.Equal(t, "opaqueValue1", val)
	val = state.Get("vendorname2")
	assert.Equal(t, "opaqueValue2", val)
}

func TestSpanContextFromW3CString1(t *testing.T) {
	ctx, ok := SpanContextFromW3CString("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	assert.True(t, ok)
	assert.Equal(t, "00f067aa0ba902b7", ctx.SpanID().String())
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", ctx.TraceID().String())
	assert.Equal(t, trace.TraceFlags(0x1), ctx.TraceFlags())
}
func TestSpanContextFromW3CString2(t *testing.T) {
	ctx, ok := SpanContextFromW3CString("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
	assert.True(t, ok)
	assert.Equal(t, "00f067aa0ba902b7", ctx.SpanID().String())
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", ctx.TraceID().String())
	assert.Equal(t, trace.TraceFlags(0x0), ctx.TraceFlags())
}
func TestSpanContextFromW3CStringInvalidTraceId(t *testing.T) {
	_, ok := SpanContextFromW3CString("00-00000000000000000000000000000000-00f067aa0ba902b7-00")
	assert.False(t, ok)
}

func TestSpanContextFromW3CStringInvalidParentId(t *testing.T) {
	_, ok := SpanContextFromW3CString("00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-00")
	assert.False(t, ok)
}
func TestSpanContextFromW3CStringEmpty(t *testing.T) {
	_, ok := SpanContextFromW3CString("")
	assert.False(t, ok)
}
