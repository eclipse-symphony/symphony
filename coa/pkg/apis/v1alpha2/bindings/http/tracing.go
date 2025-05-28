/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var log = logger.NewLogger("coa.runtime")

const (
	traceparentHeader   = "traceparent"
	tracestateHeader    = "tracestate"
	paiHeaderPrefix     = "pai-"
	paiSpanNameInternal = "pai-spanname"
	maxVersion          = 0
)

type Tracing struct {
	Observability observability.Observability
	Buffer        *v1alpha2.SafeBuffer // this is not used but should be a thread safe buffer
}

func (t Tracing) Tracing(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Request.URI().Path())
		ctx, span := startTracingClientSpanFromHTTPContext(ctx, path)

		next(ctx)
		if span.SpanContext().IsSampled() {
			observ_utils.AddAttributesToSpan(span, userDefinedHTTPHeaders(ctx))
			spanAttr := spanAttributesMapFromHTTPContext(ctx)
			observ_utils.AddAttributesToSpan(span, spanAttr)
			if sname, ok := spanAttr[paiSpanNameInternal]; ok {
				span.SetName(sname)
			}
		}
		UpdateSpanStatusFromHTTPStatus(span, ctx.Response.StatusCode())
		span.End()
	}
}

func startTracingClientSpanFromHTTPContext(ctx *fasthttp.RequestCtx, spanName string) (*fasthttp.RequestCtx, trace.Span) {
	sc, ok := SpanContextFromRequest(&ctx.Request)
	if ok {
		cCtx := trace.ContextWithRemoteSpanContext(ctx, sc)

		_, span := otel.Tracer("test").Start(cCtx, spanName, trace.WithSpanKind(trace.SpanKindClient)) //TODO: Is "test" appropriate？

		observ_utils.SpanToFastHTTPContext(ctx, &span)

		return ctx, span
	} else {
		_, cSpan := otel.Tracer("test").Start(ctx, spanName)
		observ_utils.SpanToFastHTTPContext(ctx, &cSpan)
		return ctx, cSpan
	}
}

func SpanContextFromRequest(req *fasthttp.Request) (sc trace.SpanContext, ok bool) {
	h, ok := getRequestHeader(req, traceparentHeader)
	if !ok {
		return trace.SpanContext{}, false
	}
	sc, ok = SpanContextFromW3CString(h)
	if ok {
		state := tracestateFromRequest(req)
		if state != nil {
			return sc.WithTraceState(*state), ok
		} else {
			return sc, ok
		}
	}
	return sc, ok
}

func getRequestHeader(req *fasthttp.Request, name string) (string, bool) {
	s := string(req.Header.Peek(textproto.CanonicalMIMEHeaderKey(name)))
	if s == "" {
		return "", false
	}
	return s, true
}

func SpanContextFromW3CString(h string) (sc trace.SpanContext, ok bool) {
	// See https://www.w3.org/TR/trace-context/#traceparent-header-field-values
	if h == "" {
		return trace.SpanContext{}, false
	}
	sections := strings.Split(h, "-")
	if len(sections) < 4 {
		return trace.SpanContext{}, false
	}

	if len(sections[0]) != 2 { // this is supposed to be 00 for now. Could be 00-fe.
		return trace.SpanContext{}, false
	}
	ver, err := hex.DecodeString(sections[0])
	if err != nil {
		return trace.SpanContext{}, false
	}
	version := int(ver[0])
	if version > maxVersion {
		return trace.SpanContext{}, false
	}

	if version == 0 && len(sections) != 4 { // version 0 expects 4 fields: version, trace-id, parent-id, and trace-flags
		return trace.SpanContext{}, false
	}

	traceId, err := trace.TraceIDFromHex(sections[1])
	if err != nil {
		return trace.SpanContext{}, false
	}

	spanId, err := trace.SpanIDFromHex(sections[2])
	if err != nil {
		return trace.SpanContext{}, false
	}

	flags, err := hex.DecodeString(sections[3])
	if err != nil || len(flags) < 1 {
		return trace.SpanContext{}, false
	}
	return trace.SpanContext{}.WithSpanID(spanId).WithTraceID(traceId).WithTraceFlags(trace.TraceFlags(flags[0])), true
}

func tracestateFromRequest(req *fasthttp.Request) *trace.TraceState {
	h, _ := getRequestHeader(req, tracestateHeader)
	return TraceStateFromW3CString(h)
}

func userDefinedHTTPHeaders(reqCtx *fasthttp.RequestCtx) map[string]string {
	var m = map[string]string{}

	reqCtx.Request.Header.VisitAll(func(key []byte, value []byte) {
		k := strings.ToLower(string(key))
		if strings.HasPrefix(k, paiHeaderPrefix) {
			m[k] = string(value)
		}
	})

	return m
}

func TraceStateFromW3CString(h string) *trace.TraceState {
	if h == "" {
		return nil
	}

	state := trace.TraceState{}

	pairs := strings.Split(h, ",")

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil
		}
		state, _ = state.Insert(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
	}
	return &state
}

func spanAttributesMapFromHTTPContext(ctx *fasthttp.RequestCtx) map[string]string {
	// Span Attribute reference https://github.com/open-telemetry/opentelemetry-specification/tree/master/specification/trace/semantic_conventions
	path := string(ctx.Request.URI().Path())
	method := string(ctx.Request.Header.Method())
	statusCode := ctx.Response.StatusCode()

	var m = map[string]string{}
	_, vendor := observ_utils.GetVendor(path)

	m["vendor"] = vendor
	m["path"] = path
	m["method"] = method
	m["status"] = fmt.Sprint(statusCode)
	return m
}

func UpdateSpanStatusFromHTTPStatus(span trace.Span, code int) {
	if span != nil {
		code, msg := traceStatusFromHTTPCode(code)
		span.SetStatus(code, msg)
	}
}

func traceStatusFromHTTPCode(httpCode int) (codes.Code, string) {
	var code codes.Code = codes.Unset
	var msg string = ""
	switch httpCode {
	case http.StatusUnauthorized:
		code = codes.Error
		msg = "401 - Unauthenticated"
	case http.StatusForbidden:
		code = codes.Error
		msg = "403 - Forbidden"
	case http.StatusNotFound:
		code = codes.Error
		msg = "404 - Not Found"
	case http.StatusTooManyRequests:
		code = codes.Error
		msg = "429 - Too Many Requests"
	case http.StatusNotImplemented:
		code = codes.Error
		msg = "501 - Not Implemented"
	case http.StatusServiceUnavailable:
		code = codes.Error
		msg = "503 - Service Unavailable"
	case http.StatusGatewayTimeout:
		code = codes.Error
		msg = "504 - Gateway Timeout"
	}

	if code == codes.Unset {
		if httpCode >= 100 && httpCode < 300 {
			code = codes.Ok
		} else if httpCode >= 300 && httpCode < 400 {
			code = codes.Unset
		} else if httpCode >= 400 && httpCode < 500 {
			code = codes.Error
			msg = "Bad Request"
		} else if httpCode >= 500 {
			code = codes.Error
			msg = "Internal Error"
		}
	}

	return code, msg
}
