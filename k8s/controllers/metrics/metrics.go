/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package metrics

import (
	"gopls-workspace/constants"
	"os"
	"time"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
)

type (
	ReconciliationType   string
	ReconciliationResult string
	ResourceType         string
	OperationStatus      string
)

const (
	// reconciliation type
	CreateOperationType ReconciliationType = "Create"
	UpdateOperationType ReconciliationType = "Update"
	DeleteOperationType ReconciliationType = "Delete"
	// resource type
	TargetResourceType   ResourceType = "Target"
	InstanceResourceType ResourceType = "Instance"
	// reconciliation result
	ReconcileSuccessResult ReconciliationResult = "Succeeded"
	ReconcileFailedResult  ReconciliationResult = "Failed"
	// operation status
	StatusNoOp         OperationStatus = "NoOp"
	StatusUpdateFailed OperationStatus = "StatusUpdateFailed"
	// deployment operation status
	DeploymentQueued           OperationStatus = "DeploymentQueued"
	DeploymentStatusPolled     OperationStatus = "DeploymentStatusPolled"
	DeploymentSucceeded        OperationStatus = "DeploymentSucceeded"
	DeploymentFailed           OperationStatus = "DeploymentFailed"
	DeploymentTimedOut         OperationStatus = "DeploymentTimedOut"
	GetDeploymentSummaryFailed OperationStatus = "GetDeploymentSummaryFailed"
	QueueDeploymentFailed      OperationStatus = "QueueDeploymentFailed"
)

// Metrics is a metrics tracker for a controller operation.
type Metrics struct {
	controllerReconcileLatency observability.Histogram
}

func New() (*Metrics, error) {
	observable := observability.New(constants.K8S)

	controllerReconcileLatency, err := observable.Metrics.Histogram(
		"symphony_controller_reconcile_latency",
		"measure of overall latency for controller operation side",
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		controllerReconcileLatency: controllerReconcileLatency,
	}, nil
}

// Close closes all metrics.
func (m *Metrics) Close() {
	if m == nil {
		return
	}
}

// ControllerReconcileLatency tracks the overall Controller reconcile latency.
func (m *Metrics) ControllerReconcileLatency(
	startTime time.Time,
	reconcilationType ReconciliationType,
	reconcilationResult ReconciliationResult,
	resourceType ResourceType,
	operationStatus OperationStatus,
) {
	if m == nil {
		return
	}

	chartVersion := os.Getenv("CHART_VERSION")
	m.controllerReconcileLatency.Add(
		latency(startTime),
		Deployment(
			reconcilationType,
			reconcilationResult,
			resourceType,
			operationStatus,
			chartVersion,
		),
	)
}

// Latency gets the time since the given start in milliseconds.
func latency(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}
