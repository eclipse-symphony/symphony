/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package observability

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	v1alpha2 "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	exporters "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/exporters"
	coacontexts "github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"go.opentelemetry.io/otel/attribute"
	resource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

const defaultExporterTimeout = 300

type ObservabilityConfig struct {
	Pipelines   []PipelineConfig `json:"pipelines"`
	ServiceName string           `json:"serviceName"`
}

type PipelineConfig struct {
	Exporter ExporterConfig `json:"exporter"`
}

type ExporterConfig struct {
	Type         string        `json:"type"`
	BackendUrl   string        `json:"backendUrl"`
	Sampler      SamplerConfig `json:"sampler"`
	CollectorUrl string        `json:"collectorUrl"`
	Temporality  bool          `json:"temporality"`
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
	Buffer         *v1alpha2.SafeBuffer // should be a thread safe buffer
	Metrics        *metrics
}

// New returns a new instance of the observability.
func New(symphonyProject string) Observability {
	return Observability{
		Metrics: &metrics{
			provider: otel.GetMeterProvider(),
			meter:    otel.GetMeterProvider().Meter(symphonyProject),
		},
	}
}

func populateSpanContextToDiagnosticLogContext(span trace.Span, parent context.Context) context.Context {
	if span == nil {
		return parent
	}
	traceId := ""
	spanId := ""
	if span.SpanContext().IsValid() && span.SpanContext().TraceID().IsValid() {
		traceId = span.SpanContext().TraceID().String()
	}
	if span.SpanContext().IsValid() && span.SpanContext().SpanID().IsValid() {
		spanId = span.SpanContext().SpanID().String()
	}
	return coacontexts.PopulateTraceAndSpanToDiagnosticLogContext(traceId, spanId, parent)
}

func StartSpan(name string, ctx context.Context, attributes *map[string]string) (context.Context, trace.Span) {
	span := observ_utils.SpanFromContext(ctx)
	if span != nil {
		childCtx, childSpan := otel.Tracer(name).Start(trace.ContextWithSpan(ctx, *span), name)
		childCtx = coacontexts.InheritActivityLogContextFromOriginalContext(ctx, childCtx)
		childCtx = coacontexts.InheritDiagnosticLogContextFromOriginalContext(ctx, childCtx)
		childCtx = populateSpanContextToDiagnosticLogContext(childSpan, childCtx)
		setSpanAttributes(childSpan, attributes)
		return childCtx, childSpan
	} else {
		childCtx, childSpan := otel.Tracer(name).Start(ctx, name)
		childCtx = coacontexts.InheritActivityLogContextFromOriginalContext(ctx, childCtx)
		childCtx = coacontexts.InheritDiagnosticLogContextFromOriginalContext(ctx, childCtx)
		childCtx = populateSpanContextToDiagnosticLogContext(childSpan, childCtx)
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
	coacontexts.ClearTraceAndSpanFromDiagnosticLogContext(&ctx)
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
			populateMicrosoftResourceId(),
		))))
	return nil
}

func (o *Observability) InitTrace(config ObservabilityConfig) error {
	var traceExporters []sdktrace.TracerProviderOption
	var exporter sdktrace.SpanExporter
	var err error

	for _, p := range config.Pipelines {
		switch p.Exporter.Type {
		case v1alpha2.TracingExporterConsole:
			if o.Buffer == nil {
				exporter, err = exporters.NewConsoleExporter(nil)
			} else {
				exporter, err = exporters.NewConsoleExporter(o.Buffer)
			}
			if err != nil {
				return err
			}

			batcher := sdktrace.NewBatchSpanProcessor(exporter)

			traceExporters = append(
				traceExporters,
				sdktrace.WithSpanProcessor(batcher),
			)
		case v1alpha2.TracingExporterZipkin:
			exporter, err = exporters.NewZipkinExporter(p.Exporter.BackendUrl, p.Exporter.Sampler.SampleRate)
			if err != nil {
				return err
			}

			batcher := sdktrace.NewBatchSpanProcessor(exporter)

			traceExporters = append(
				traceExporters,
				sdktrace.WithSpanProcessor(batcher),
			)
		case v1alpha2.TracingExporterOTLPgRPC:
			ctx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*defaultExporterTimeout,
			)
			defer cancel()

			te, err := otlptracegrpc.New(
				ctx,
				otlptracegrpc.WithInsecure(),
				otlptracegrpc.WithEndpoint(p.Exporter.CollectorUrl),
			)
			if err != nil {
				return v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.BadConfig)
			}

			traceExporters = append(
				traceExporters,
				sdktrace.WithBatcher(te),
			)
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("exporter type '%s' is not supported", p.Exporter.Type), v1alpha2.BadConfig)
		}
	}

	sdkTraceOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.ServiceName),
				populateMicrosoftResourceId(),
			),
		),
	}
	sdkTraceOpts = append(
		sdkTraceOpts,
		traceExporters...,
	)

	traceProvider := sdktrace.NewTracerProvider(
		sdkTraceOpts...,
	)

	otel.SetTracerProvider(traceProvider)

	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	otel.SetTextMapPropagator(propagator)

	// temporary solution to clean up the meter provider until we have hooks for shutdown sequence in symphony
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan

		traceProvider.Shutdown(context.Background())
	}()

	return nil
}

func (o *Observability) InitMetric(config ObservabilityConfig) error {
	var metricExporters []sdkmetric.Option

	for _, p := range config.Pipelines {
		switch p.Exporter.Type {
		case v1alpha2.MetricsExporterOTLPgRPC:
			ctx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*defaultExporterTimeout,
			)
			defer cancel()

			var otlpmetricgrpcOptions []otlpmetricgrpc.Option

			otlpmetricgrpcOptions = append(
				otlpmetricgrpcOptions,
				otlpmetricgrpc.WithEndpoint(p.Exporter.CollectorUrl),
				otlpmetricgrpc.WithInsecure(),
			)

			if p.Exporter.Temporality {
				otlpmetricgrpcOptions = append(
					otlpmetricgrpcOptions,
					otlpmetricgrpc.WithTemporalitySelector(genevaTemporality),
				)
			}

			exporter, err := otlpmetricgrpc.New(
				ctx,
				otlpmetricgrpcOptions...,
			)
			if err != nil {
				return v1alpha2.NewCOAError(nil, err.Error(), v1alpha2.BadConfig)
			}

			metricExporters = append(
				metricExporters,
				sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
					exporter,
					sdkmetric.WithInterval(1*time.Second),
				)),
			)
		default:
			return v1alpha2.NewCOAError(nil, fmt.Sprintf("exporter type '%s' is not supported", p.Exporter.Type), v1alpha2.BadConfig)
		}
	}

	sdkMetricOpts := []sdkmetric.Option{
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.ServiceName),
				populateMicrosoftResourceId(),
			),
		),
	}
	sdkMetricOpts = append(sdkMetricOpts, metricExporters...)

	meterProvider := sdkmetric.NewMeterProvider(
		// TODO: perhaps adding some k8s information here
		sdkMetricOpts...,
	)

	otel.SetMeterProvider(meterProvider)

	return nil
}

func (o *Observability) Shutdown(ctx context.Context) error {
	if mp, ok := otel.GetMeterProvider().(v1alpha2.Terminable); ok {
		return mp.Shutdown(ctx)
	}
	if tp, ok := otel.GetTracerProvider().(v1alpha2.Terminable); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}

// Geneva only supports delta temporality.
func genevaTemporality(ik sdkmetric.InstrumentKind) metricdata.Temporality {
	switch ik {
	default:
		return metricdata.DeltaTemporality
	}
}

func populateMicrosoftResourceId() attribute.KeyValue {
	rid, ok := os.LookupEnv("EXTENSION_RESOURCEID")
	if !ok {
		return attribute.KeyValue{}
	}

	return attribute.KeyValue{
		Key:   "microsoft.resourceId",
		Value: attribute.StringValue(rid),
	}
}
