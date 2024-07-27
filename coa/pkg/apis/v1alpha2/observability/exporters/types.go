/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package exporters

import (
	"io"
	// "log"
	// "os"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)

func NewTraceConsoleExporter(writer io.Writer) (sdktrace.SpanExporter, error) {
	if writer != nil {
		return stdout.New(
			stdout.WithPrettyPrint(),
			stdout.WithWriter(writer),
		)
	} else {
		return stdout.New(
			stdout.WithPrettyPrint(),
		)
	}
}

func NewLogConsoleExporter(writer io.Writer) (*stdoutlog.Exporter, error) {
	if writer != nil {
		return stdoutlog.New(
			stdoutlog.WithPrettyPrint(),
			stdoutlog.WithWriter(writer),
		)
	} else {
		return stdoutlog.New(
			stdoutlog.WithPrettyPrint(),
		)
	}
}

func NewZipkinExporter(url string, sampleRate string) (sdktrace.SpanExporter, error) {
	// return zipkin.New(url, zipkin.WithLogger(logger))
	return zipkin.New(url)
}
