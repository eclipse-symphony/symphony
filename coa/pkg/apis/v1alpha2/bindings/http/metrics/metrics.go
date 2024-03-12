/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

import (
	"time"

	"github.com/eclipse-symphony/symphony/coa/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
)

// Metrics is a metrics tracker for an api operation.
type Metrics struct {
	apiOperationLatency observability.Histogram
	apiOperationErrors  observability.Counter
}

func New() (*Metrics, error) {
	observable := observability.New(constants.API)

	apiOperationLatency, err := observable.Metrics.Histogram(
		"aio_orc_api_operation_latency",
		"measure of overall latency for API operation side",
	)
	if err != nil {
		return nil, err
	}

	apiOperationErrors, err := observable.Metrics.Counter(
		"aio_orc_api_operation_errors",
		"count of errors in API operation side",
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		apiOperationLatency: apiOperationLatency,
		apiOperationErrors:  apiOperationErrors,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
	if m == nil {
		return
	}

	m.apiOperationErrors.Close()
}

// ApiOperationLatency tracks the overall API operation latency.
func (m *Metrics) ApiOperationLatency(
	startTime time.Time,
	operation string,
	operationType string,
) {
	if m == nil {
		return
	}

	m.apiOperationLatency.Add(
		latency(startTime),
		Deployment(
			operation,
			operationType,
		),
	)
}

// ApiOperationErrors increments the count of errors for API operation.
func (m *Metrics) ApiOperationErrors(
	operation string,
	operationType string,
	errorCode string,
) {
	if m == nil {
		return
	}

	m.apiOperationErrors.Add(
		1,
		Deployment(
			operation,
			operationType,
		),
		Error(
			errorCode,
		),
	)
}

// Latency gets the time since the given start in milliseconds.
func latency(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}
