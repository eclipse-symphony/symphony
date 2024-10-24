/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/valyala/fasthttp"
)

var (
	corsAllowHeaders     = "authorization,Content-Type"
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

type CORS struct {
	Properties map[string]interface{}
}

// CORS middleware to allow CORS. The middleware doesn't override existing headers in incoming requests
func (c CORS) CORS(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		for k, v := range c.Properties {
			ctx.Response.Header.Set(k, utils.FormatAsString(v))
		}
		if _, ok := c.Properties["Access-Control-Allow-Headers"]; !ok {
			ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		}
		if _, ok := c.Properties["Access-Control-Allow-Credentials"]; !ok {
			ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
		}
		if _, ok := c.Properties["Access-Control-Allow-Methods"]; !ok {
			ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		}
		if _, ok := c.Properties["Access-Control-Allow-Origin"]; !ok {
			ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)
		}
		next(ctx)
	}
}
