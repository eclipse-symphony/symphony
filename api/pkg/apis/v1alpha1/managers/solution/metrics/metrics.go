/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

import (
	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
)

const (
	GetSummaryOperation string = "GetSummary"

	GetOperationType string = "Get"
)

// Metrics is a metrics tracker for an api operation.
type Metrics struct {
	apiComponentCount observability.Gauge
}

func New() (*Metrics, error) {
	observable := observability.New(constants.API)

	apiComponentCount, err := observable.Metrics.Gauge(
		"aio_orc_api_component_count",
		"count of components in API operation",
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		apiComponentCount: apiComponentCount,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
}

// ApiComponentCount gets the total count of components for an API operation.
func (m *Metrics) ApiComponentCount(
	componentCount int,
	operation string,
	operationType string,
) {
	if m == nil {
		return
	}

	m.apiComponentCount.Set(
		float64(componentCount),
		Deployment(
			operation,
			operationType,
		),
	)
}
