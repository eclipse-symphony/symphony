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
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Metrics struct {
	Observability observability.Observability
}

func (m Metrics) Metrics(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		startTime := time.Now().UTC()

		next(ctx)

		endpoint := ctx.UserValue(router.MatchedRoutePathParam)
		actCtx := contexts.ParseActivityLogContextFromHttpRequestHeader(ctx)

		var endpointString string
		if endpoint != nil {
			endpointString = utils.FormatAsString(endpoint)
		} else {
			endpointString = string(ctx.Path())
		}
		ApiOperationMetrics.ApiOperationStatus(
			*actCtx,
			endpointString,
			string(ctx.Method()),
			ctx.Response.StatusCode(),
			formatErrorCode(ctx.Response.StatusCode()),
		)

		ApiOperationMetrics.ApiOperationLatency(
			startTime,
			endpointString,
			string(ctx.Method()),
			ctx.Response.StatusCode(),
			formatErrorCode(ctx.Response.StatusCode()),
		)
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
