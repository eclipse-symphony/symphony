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

package http

import (
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
			ctx.Response.Header.Set(k, v.(string))
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
