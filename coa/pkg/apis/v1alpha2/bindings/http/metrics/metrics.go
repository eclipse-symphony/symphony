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
	"github.com/eclipse-symphony/symphony/coa/pkg/logger/contexts"
)

// Metrics is a metrics tracker for an api operation.
type Metrics struct {
	apiOperationLatency observability.Gauge
	apiOperationStatus  observability.Counter
}

func New() (*Metrics, error) {
	observable := observability.New(constants.API)

	apiOperationLatency, err := observable.Metrics.Gauge(
		constants.APIOperationLatency,
		constants.APIOperationLatencyDescription,
	)
	if err != nil {
		return nil, err
	}

	apiOperationStatus, err := observable.Metrics.Counter(
		constants.APIOperationStatus,
		constants.APIOperationStatusDescription,
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		apiOperationLatency: apiOperationLatency,
		apiOperationStatus:  apiOperationStatus,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
	if m == nil {
		return
	}

	m.apiOperationLatency.Close()
	m.apiOperationStatus.Close()
}

// ApiOperationLatency tracks the overall API operation latency.
func (m *Metrics) ApiOperationLatency(
	startTime time.Time,
	operation string,
	operationType string,
	statusCode int,
	formatStatusCode string,
) {
	if m == nil {
		return
	}

	m.apiOperationLatency.Set(
		latency(startTime),
		Deployment(
			operation,
			operationType,
		),
		Status(
			statusCode,
			formatStatusCode,
		),
	)
}

// ApiOperationStatus increments the count of status code for API operation.
func (m *Metrics) ApiOperationStatus(
	context contexts.ActivityLogContext,
	operation string,
	operationType string,
	statusCode int,
	formatStatusCode string,
) {
	if m == nil {
		return
	}

	customerResourceId := context.GetResourceCloudId()
	locationId := context.GetResourceCloudLocation()

	m.apiOperationStatus.Add(
		1,
		SLI(
			customerResourceId,
			locationId,
		),
		Deployment(
			operation,
			operationType,
		),
		Status(
			statusCode,
			formatStatusCode,
		),
	)
}

// latency gets the time since the given start in milliseconds.
func latency(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}
