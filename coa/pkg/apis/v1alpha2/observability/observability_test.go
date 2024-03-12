/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package observability

import (
	"context"
	"testing"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStartSpan(t *testing.T) {
	ctx, span := StartSpan("Sample Vendor", context.Background(), &map[string]string{
		"method": "OnSample",
	})
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)

	requestCtx := &fasthttp.RequestCtx{}
	requestCtx.SetUserValue("paiSpanContextKey", &span)
	ctx, span = StartSpan("Sample Vendor", requestCtx, &map[string]string{
		"method": "OnSample",
	})
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
}

func TestEndSpan(t *testing.T) {
	EndSpan(context.Background())
	assert.True(t, true)
}

func TestConsolePipeline(t *testing.T) {
	ob := Observability{}
	err := ob.Init(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:       v1alpha2.TracingExporterConsole,
					BackendUrl: "",
					Sampler: SamplerConfig{
						SampleRate: "always",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}
func TestZipkinPipeline(t *testing.T) {
	ob := Observability{}
	err := ob.Init(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:       v1alpha2.TracingExporterZipkin,
					BackendUrl: "http://localhost:9411/api/v2/spans",
					Sampler: SamplerConfig{
						SampleRate: "always",
					},
				},
			},
		},
	})
	assert.Nil(t, err)
}

func TestMetricsOTLPgRPCPipeline(t *testing.T) {
	ob := Observability{}
	err := ob.InitMetric(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:         v1alpha2.MetricsExporterOTLPgRPC,
					CollectorUrl: "http://otel-collector.alice-springs.svc.cluster.local:4317",
				},
			},
		},
	})
	assert.Nil(t, err)
}

func TestTracingOTLPgRPCPipeline(t *testing.T) {
	ob := Observability{}
	err := ob.InitTrace(ObservabilityConfig{
		Pipelines: []PipelineConfig{
			{
				Exporter: ExporterConfig{
					Type:         v1alpha2.TracingExporterOTLPgRPC,
					CollectorUrl: "http://otel-collector.alice-springs.svc.cluster.local:4317",
				},
			},
		},
	})
	assert.Nil(t, err)
}
