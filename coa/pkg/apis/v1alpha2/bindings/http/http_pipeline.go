/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"encoding/json"
	"fmt"
	"os"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type Middleware func(h fasthttp.RequestHandler) fasthttp.RequestHandler

type Pipeline struct {
	Handlers []Middleware
}

func BuildPipeline(config HttpBindingConfig, pubsubProvider pubsub.IPubSubProvider) (Pipeline, error) {
	ret := Pipeline{Handlers: make([]Middleware, 0)}
	for _, c := range config.Pipeline {
		switch c.Type {
		case "middleware.http.cors":
			cors := CORS{Properties: c.Properties}
			ret.Handlers = append(ret.Handlers, cors.CORS)
		case "middleware.http.trail":
			trail := Trail{}
			trail.SetPubSubProvider(pubsubProvider)
			ret.Handlers = append(ret.Handlers, trail.Trail)
		case "middleware.http.telemetry":
			enableAppInsight := os.Getenv("ENABLE_APP_INSIGHT")
			c.Properties["enabled"] = enableAppInsight == "true"
			if c.Properties["enabled"] == true {
				if os.Getenv("APP_INSIGHT_KEY") == "" {
					return ret, v1alpha2.NewCOAError(nil, "APP_INSIGHT_KEY is not set", v1alpha2.BadConfig)
				}
			}
			c.Properties["client"] = uuid.New().String()
			telemetry := Telemetry{Properties: c.Properties}
			ret.Handlers = append(ret.Handlers, telemetry.Telemetry)
		case "middleware.http.jwt":
			jwts := JWT{}
			jData, _ := json.Marshal(c.Properties)
			err := json.Unmarshal(jData, &jwts)
			if err != nil {
				return ret, v1alpha2.NewCOAError(nil, "incorrect jwt pipeline configuration format", v1alpha2.BadConfig)
			}
			if jwts.AuthHeader == "" {
				jwts.AuthHeader = "Authorization"
			}
			ret.Handlers = append(ret.Handlers, jwts.JWT)
		case "middleware.http.tracing":
			tracing := Tracing{
				Observability: observability.Observability{},
			}
			config := observability.ObservabilityConfig{}
			if p, ok := c.Properties["pipeline"]; ok {
				data, _ := json.Marshal(p)
				pipelines := make([]observability.PipelineConfig, 0)
				err := json.Unmarshal(data, &pipelines)
				if err != nil {
					return ret, v1alpha2.NewCOAError(nil, "incorrect tracing pipeline configuration format", v1alpha2.BadConfig)
				}
				config.Pipelines = pipelines
			}
			err := tracing.Observability.Init(config)
			if err != nil {
				return ret, v1alpha2.NewCOAError(nil, "failed to initialize tracing middleware", v1alpha2.InternalError)
			}
			ret.Handlers = append(ret.Handlers, tracing.Tracing)
		default:
			return ret, v1alpha2.NewCOAError(nil, fmt.Sprintf("middleware type '%s' is not recognized", c.Type), v1alpha2.BadConfig)
		}
	}
	return ret, nil
}

func (p Pipeline) Apply(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	for i := len(p.Handlers) - 1; i >= 0; i-- {
		handler = p.Handlers[i](handler)
	}
	return handler
}
