/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
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
