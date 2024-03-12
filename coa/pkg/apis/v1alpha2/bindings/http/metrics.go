/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"strings"
	"time"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Metrics struct {
	Observability observability.Observability
}

func (m Metrics) Metrics(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		startTime := time.Now()

		next(ctx)

		endpoint := ctx.UserValue(router.MatchedRoutePathParam)

		if endpoint != nil {
			endpointString := endpoint.(string)

			if ctx.Response.StatusCode() >= 400 {
				ApiOperationMetrics.ApiOperationErrors(
					endpointString,
					string(ctx.Method()),
					formatErrorCode(ctx.Response.StatusCode()),
				)
			} else {
				ApiOperationMetrics.ApiOperationLatency(
					startTime,
					endpointString,
					string(ctx.Method()),
				)
			}
		} else {
			ApiOperationMetrics.ApiOperationErrors(
				string(ctx.Path()),
				string(ctx.Method()),
				formatErrorCode(ctx.Response.StatusCode()),
			)
		}
	}
}

// set the default errorcodes for the http request
func formatErrorCode(httpCode int) string {
	errorCode := v1alpha2.State(httpCode).String()
	if strings.HasPrefix(errorCode, "Unknown State:") {
		if httpCode >= 400 && httpCode < 500 {
			return "Bad Request"
		} else {
			return "Internal Error"
		}
	}
	return errorCode
}
