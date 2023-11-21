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

package observability

import (
	"fmt"

	"bytes"
	"context"

	v1alpha2 "github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	exporters "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/exporters"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	observ_utils "github.com/azure/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"go.opentelemetry.io/otel/attribute"
	resource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type ObservabilityConfig struct {
	Pipelines []PipelineConfig `json:"pipelines"`
}

type PipelineConfig struct {
	Exporter ExporterConfig `json:"exporter"`
}
type ExporterConfig struct {
	Type       string        `json:"type"`
	BackendUrl string        `json:"backendUrl"`
	Sampler    SamplerConfig `json:"sampler"`
}
type ProcessorConfig struct {
}
type ProviderConfig struct {
}
type SamplerConfig struct {
	SampleRate string `json:"sampleRate"`
}

type Observability struct {
	Tracer         trace.Tracer
	TracerProvider trace.TracerProvider
	Buffer         *bytes.Buffer
}

func StartSpan(name string, ctx context.Context, attributes *map[string]string) (context.Context, trace.Span) {
	span := observ_utils.SpanFromContext(ctx)
	if span != nil {
		childCtx, childSpan := otel.Tracer(name).Start(trace.ContextWithSpan(ctx, *span), name)
		setSpanAttributes(childSpan, attributes)
		return childCtx, childSpan
	} else {
		childCtx, childSpan := otel.Tracer(name).Start(ctx, name)
		setSpanAttributes(childSpan, attributes)
		return childCtx, childSpan
	}
}
func setSpanAttributes(span trace.Span, attributes *map[string]string) {
	if attributes != nil {
		for k, v := range *attributes {
			span.SetAttributes(attribute.String(k, v))
		}
	}
}
func EndSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()
}
func (o *Observability) Init(config ObservabilityConfig) error {
	for _, p := range config.Pipelines {
		err := o.createPipeline(p)
		if err != nil {
			return err
		}
	}
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)
	return nil
}
func (o *Observability) createPipeline(config PipelineConfig) error {
	err := o.createExporter(config.Exporter)
	if err != nil {
		return err
	}
	return nil
}
func (o *Observability) createExporter(config ExporterConfig) error {
	var exporter sdktrace.SpanExporter
	var err error
	switch config.Type {
	case v1alpha2.TracingExporterConsole:
		if o.Buffer == nil {
			exporter, err = exporters.NewConsoleExporter(nil)
		} else {
			exporter, err = exporters.NewConsoleExporter(o.Buffer)
		}
		if err != nil {
			return err
		}
	case v1alpha2.TracingExporterZipkin:
		exporter, err = exporters.NewZipkinExporter(config.BackendUrl, config.Sampler.SampleRate)
		if err != nil {
			return err
		}
	default:
		return v1alpha2.NewCOAError(nil, fmt.Sprintf("exporter type '%s' is not supported", config.Type), v1alpha2.BadConfig)
	}
	batcher := sdktrace.NewBatchSpanProcessor(exporter)
	//otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(batcher)))
	//res, _ := resource.New(context.TODO(), resource.WithAttributes(attribute.String("service.name", "Symphony API (PAI)")))
	//otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(batcher), sdktrace.WithResource(res)))
	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("Symphony API"),
		))))
	return nil
}
