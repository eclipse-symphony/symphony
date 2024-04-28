/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

import (
	"gopls-workspace/constants"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
)

const (
	//validation type
	CreateOperationType string = "Create"
	UpdateOperationType string = "Update"
	//validation result
	ValidResource   string = "Valid"
	InvalidResource string = "Invalid"
	//resource type
	TargetResourceType   string = "Target"
	InstanceResourceType string = "Instance"
	CatalogResourceType  string = "Catalog"
	ModelResourceType    string = "Model"
	SkillResourceType    string = "Skill"
	DeviceResourceType   string = "Device"
)

// Metrics is a metrics tracker for a controller operation.
type Metrics struct {
	controllerValidationLatency observability.Histogram
}

func New() (*Metrics, error) {
	observable := observability.New(constants.K8S)

	controllerValidationLatency, err := observable.Metrics.Histogram(
		"symphony_controller_validation_latency",
		"measure of overall controller validate latency",
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		controllerValidationLatency: controllerValidationLatency,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
	if m == nil {
		return
	}
}

// ControllerValidationLatency tracks the overall Controller validation latency.
func (m *Metrics) ControllerValidationLatency(
	startTime time.Time,
	validationType string,
	validationResult string,
	resourceType string,
) {
	if m == nil {
		return
	}

	m.controllerValidationLatency.Add(
		latency(startTime),
		Deployment(
			validationType,
			validationResult,
			resourceType,
		),
	)
}

// Latency gets the time since the given start in milliseconds.
func latency(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}
