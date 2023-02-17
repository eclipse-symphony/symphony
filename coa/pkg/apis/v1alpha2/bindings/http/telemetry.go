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
	"fmt"
	"os"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/valyala/fasthttp"
)

type Telemetry struct {
	Properties map[string]interface{}
}

var client appinsights.TelemetryClient
var initialized bool

func initClient(properties map[string]interface{}) {
	instrumentationKey := os.Getenv("APP_INSIGHT_KEY")
	if instrumentationKey == "" {
		instrumentationKey = "0be0a36e-6e0a-4544-a453-a237fd25cf64"
	}
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)
	if batchSize, ok := properties["maxBatchSize"]; ok {
		telemetryConfig.MaxBatchSize = int(batchSize.(float64))
	} else {
		telemetryConfig.MaxBatchSize = 8192
	}
	if batchInterval, ok := properties["maxBatchInterval"]; ok {
		telemetryConfig.MaxBatchInterval = time.Duration(int(batchInterval.(float64))) * time.Second
	} else {
		telemetryConfig.MaxBatchInterval = 2 * time.Second
	}
	client = appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	initialized = true
}

// CORS middleware to allow CORS. The middleware doesn't override existing headers in incoming requests
func (c Telemetry) Telemetry(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if v, ok := c.Properties["enabled"]; !ok || v != true {
			next(ctx)
			return
		}
		if !initialized {
			initClient(c.Properties)
		}
		if initialized {
			path := string(ctx.Path())
			method := string(ctx.Method())
			eventId := fmt.Sprintf("%s-%s", path, method)
			event := appinsights.NewEventTelemetry(eventId)
			event.Properties["client"] = fmt.Sprintf("%v", c.Properties["client"])
			next(ctx)
			event.Properties["status"] = fmt.Sprintf("%v", ctx.Response.StatusCode())
			client.Track(event)
		} else {
			next(ctx)
		}
	}
}
