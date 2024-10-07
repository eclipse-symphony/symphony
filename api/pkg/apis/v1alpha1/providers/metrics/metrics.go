/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

import (
	"time"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
)

const (
	ValidateRuleOperation             string = "ValidateRule"
	ApplyScriptOperation              string = "ApplyScript"
	ApplyYamlOperation                string = "ApplyYaml"
	ApplyCustomResource               string = "ApplyCustomResource"
	ReceiveDataChannelOperation       string = "ReceiveFromDataChannel"
	ReceiveErrorChannelOperation      string = "ReceiveFromErrorChannel"
	ConvertResourceDataBytesOperation string = "ConvertResourceDataToBytes"
	ObjectOperation                   string = "Object"
	ResourceOperation                 string = "Resource"
	PullChartOperation                string = "PullChart"
	LoadChartOperation                string = "LoadChart"
	HelmChartOperation                string = "HelmChart"
	HelmActionConfigOperation         string = "HelmActionConfig"
	HelmPropertiesOperation           string = "HelmProperties"
	K8SProjectorOperation             string = "K8SProjector"
	K8SDeploymentOperation            string = "K8SDeployment"
	K8SRemoveServiceOperation         string = "K8SRemoveService"
	K8SRemoveDeploymentOperation      string = "K8SRemoveDeployment"
	K8SRemoveNamespaceOperation       string = "K8SRemoveNamespace"
	ConfigMapOperation                string = "ConfigMap"
	IngressOperation                  string = "Ingress"
	IngressPropertiesOperation        string = "IngressProperties"

	ProcessOperation string = "Process"
	ApplyOperation   string = "Apply"

	GetOperationType      string = "Get"
	ValidateOperationType string = "Validate"
	RunOperationType      string = "Run"
	ApplyOperationType    string = "Apply"
)

// Metrics is a metrics tracker for a provider operation.
type Metrics struct {
	providerOperationLatency observability.Gauge
	providerOperationErrors  observability.Counter
}

func New() (*Metrics, error) {
	observable := observability.New(constants.API)

	providerOperationLatency, err := observable.Metrics.Gauge(
		"symphony_provider_operation_latency",
		"measure of overall latency for provider operation side",
	)
	if err != nil {
		return nil, err
	}

	providerOperationErrors, err := observable.Metrics.Counter(
		"symphony_provider_operation_errors",
		"count of errors in provider operation side",
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		providerOperationLatency: providerOperationLatency,
		providerOperationErrors:  providerOperationErrors,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
	if m == nil {
		return
	}

	m.providerOperationErrors.Close()
}

// ProviderOperationLatency tracks the overall provider target latency.
func (m *Metrics) ProviderOperationLatency(
	startTime time.Time,
	providerType string,
	operation string,
	operationType string,
	functionName string,
) {
	if m == nil {
		return
	}

	m.providerOperationLatency.Set(
		latency(startTime),
		Target(
			providerType,
			functionName,
			operation,
			operationType,
			v1alpha2.OK.String(),
		),
	)
}

// ProviderOperationErrors increments the count of errors for provider target.
func (m *Metrics) ProviderOperationErrors(
	providerType string,
	functionName string,
	operation string,
	operationType string,
	errorCode string,
) {
	if m == nil {
		return
	}

	m.providerOperationErrors.Add(
		1,
		Target(
			providerType,
			functionName,
			operation,
			operationType,
			errorCode,
		),
	)
}

// latency gets the time since the given start in milliseconds.
func latency(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}
