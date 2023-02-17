/*
MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE
*/

package utils

import (
	"context"
	"strconv"
	"strings"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
func CloseSpanWithError(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
