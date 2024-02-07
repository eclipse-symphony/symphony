/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
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

type SpanAttrValue struct {
	Type  string
	Value string
}

type SpanAttr struct {
	Key   string
	Value SpanAttrValue
}

func TestTracing(t *testing.T) {
	tracing := Tracing{
		Observability: observability.Observability{},
	}
	config := observability.ObservabilityConfig{}
	config.Pipelines = []observability.PipelineConfig{
		{
			Exporter: observability.ExporterConfig{
				Type:       "tracing.exporters.console",
				BackendUrl: "",
				Sampler: observability.SamplerConfig{
					SampleRate: "always",
				},
			},
		},
	}
	var buf v1alpha2.SafeBuffer
	tracing.Observability.Buffer = &buf
	err := tracing.Observability.Init(config)
	assert.Nil(t, err)

	successPath := "/success"
	successUrl := "http://localhost:7777" + successPath
	// lauch a HTTP server and validate console tracing in buffer
	go func() {
		requestHandler := func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case successPath:
				if string(ctx.Method()) == fasthttp.MethodGet {
					fmt.Fprintf(ctx, "Hello, World!")
				}
			default:
				ctx.Error("Unsupported path", fasthttp.StatusNotFound)
			}
		}

		_ = fasthttp.ListenAndServe(":7777", tracing.Tracing(requestHandler))
	}()

	// wait for server to start
	time.Sleep(3 * time.Second)

	ctx, span := observability.StartSpan("HTTP-Test-Client", context.Background(), &map[string]string{
		"method":      "TestTracing",
		"http.method": fasthttp.MethodGet,
		"http.url":    successUrl,
	})

	// GET request with parent span
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, fasthttp.MethodGet, successUrl, nil)
	observ_utils.PropagateSpanContextToHttpRequestHeader(req)
	assert.Nil(t, err)
	resp, err := client.Do(req)
	assert.Nil(t, err)

	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello, World!", string(respBody))

	// validate the buffer
	time.Sleep(3 * time.Second) // wait for the span to be written to the buffer
	spanContent := buf.String()
	var mapSpan map[string]interface{}
	json.Unmarshal([]byte(spanContent), &mapSpan)
	fmt.Println(spanContent)

	assert.Equal(t, span.SpanContext().TraceID().String(), mapSpan["SpanContext"].(map[string]interface{})["TraceID"])
	assert.Equal(t, span.SpanContext().TraceID().String(), mapSpan["Parent"].(map[string]interface{})["TraceID"])
	assert.Equal(t, span.SpanContext().SpanID().String(), mapSpan["Parent"].(map[string]interface{})["SpanID"])
	assert.Equal(t, successPath, mapSpan["Name"])

	attrs, err := json.Marshal(mapSpan["Attributes"])
	assert.Nil(t, err)
	spanAttrs := make([]SpanAttr, 0)
	json.Unmarshal(attrs, &spanAttrs)
	attrsMap := make(map[string]SpanAttrValue)
	for _, attr := range spanAttrs {
		attrsMap[attr.Key] = attr.Value
	}

	_, expectedVendor := observ_utils.GetVendor(successPath)
	assert.Equal(t, successPath, attrsMap["path"].Value)
	assert.Equal(t, expectedVendor, attrsMap["vendor"].Value)
	assert.Equal(t, fasthttp.MethodGet, attrsMap["method"].Value)
	assert.Equal(t, "200", attrsMap["status"].Value)
}
