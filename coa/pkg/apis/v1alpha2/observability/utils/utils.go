/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultSamplingRate   = 1e-4
	paiFastHTTPContextKey = "paiSpanContextKey"
)

func GetTraceSamplingRate(rate string) float64 {
	f, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return defaultSamplingRate
	}
	return f
}

func IsTracingEnabled(rate string) bool {
	return GetTraceSamplingRate(rate) != 0
}

func AddAttributesToSpan(span trace.Span, attributes map[string]string) {
	if span == nil {
		return
	}

	for k, v := range attributes {
		span.SetAttributes(attribute.String(k, v))
	}
}

func PropagateSpanContextToHttpRequestHeader(req *http.Request) {
	// https://www.w3.org/TR/trace-context/#traceparent-header
	if req == nil {
		return
	}

	propagator := propagation.TraceContext{}
	propagator.Inject(req.Context(), propagation.HeaderCarrier(req.Header))
}

func SpanToFastHTTPContext(ctx *fasthttp.RequestCtx, span *trace.Span) {
	ctx.SetUserValue(paiFastHTTPContextKey, span)
}

func GetVendor(apiPath string) (string, string) {
	if apiPath == "" {
		return "", ""
	}

	// Split up to 4 delimiters in '/v1.0/state/statestore/key' to get component api type and value
	var tokens = strings.SplitN(apiPath, "/", 4)
	if len(tokens) < 3 {
		return "", ""
	}

	// return 'state', 'statestore' from the parsed tokens in apiComponent type
	return tokens[1], tokens[2]
}

// func SpanContextToW3CString(sc trace.SpanContext) string {
// 	return fmt.Sprintf("%x-%s-%s-%x",
// 		[]byte{0},
// 		sc.TraceID().String(),
// 		sc.SpanID().String(),
// 		[]byte{sc.TraceFlags()})
// }

func SpanFromContext(ctx context.Context) *trace.Span {
	if reqCtx, ok := ctx.(*fasthttp.RequestCtx); ok {
		val := reqCtx.UserValue(paiFastHTTPContextKey)
		if val == nil {
			return nil
		}
		return val.(*trace.Span)
	}

	return nil
}

// func TraceStateToW3CString(sc trace.SpanContext) string {
// 	return sc.TraceState().String()
// }

func UpdateSpanStatusFromCOAResponse(span trace.Span, resp v1alpha2.COAResponse) {
	var code codes.Code = codes.Unset
	if resp.State == v1alpha2.OK {
		code = codes.Ok
	} else {
		code = codes.Error
	}
	msg := ""
	if code == codes.Error {
		msg = string(resp.Body)
	}
	span.SetStatus(code, msg)
	//span.SetAttributes(attribute.String("a", "b"))
}

func CloseSpanWithCOAResponse(span trace.Span, resp v1alpha2.COAResponse) v1alpha2.COAResponse {
	UpdateSpanStatusFromCOAResponse(span, resp)
	span.End()
	return resp
}
func CloseSpanWithError(span trace.Span, err *error) {
	if err != nil && *err != nil {
		span.SetStatus(codes.Error, (*err).Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// GetFunctionName returns the name of the function that called it
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	function := runtime.FuncForPC(pc).Name()
	slice := strings.Split(function, "/")
	index := 0 // Default index in case len(slice) == 0
	if len(slice) > 0 {
		index = len(slice) - 1
	}
	funcName := slice[index]
	return funcName
}
