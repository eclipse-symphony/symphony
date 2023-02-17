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
	"encoding/json"
	"fmt"
	"os"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	observability "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type Middleware func(h fasthttp.RequestHandler) fasthttp.RequestHandler

type Pipeline struct {
	Handlers []Middleware
}

func BuildPipeline(config HttpBindingConfig) (Pipeline, error) {
	ret := Pipeline{Handlers: make([]Middleware, 0)}
	for _, c := range config.Pipeline {
		switch c.Type {
		case "middleware.http.cors":
			cors := CORS{Properties: c.Properties}
			ret.Handlers = append(ret.Handlers, cors.CORS)
		case "middleware.http.telemetry":
			enableAppInsight := os.Getenv("ENABLE_APP_INSIGHT")
			c.Properties["enabled"] = enableAppInsight == "true"
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
