/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ActionState indicates current state of action progress.
type ActionState string

const (
	// SucceededActionState indicates that action is successfully completed.
	SucceededActionState ActionState = "Succeeded"

	// InProgressActionState indicates that action is currently in progress.
	InProgressActionState ActionState = "InProgress"

	// FailedActionState indicates that performed action failed.
	FailedActionState ActionState = "Failed"
)

// ActionResult contains result of performing an action for a resource.
type ActionResult struct {
	// Status indicates current state of action progress.
	Status ActionState `json:"status"`

	// OperationID is the unique identifier for tracking this action.
	OperationID string `json:"operationID,omitempty"`

	// Error indicates the error that occurred for a failed attempt at performing action.
	Error *ProvisioningError `json:"error,omitempty"`

	// Output of the action if succeeds.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Output runtime.RawExtension `json:"output,omitempty"`
}

// ProvisioningError captures the error details when provisioning has failed.
type ProvisioningError struct {
	// Message contains the string suitable for logging and human consumption.
	Message string `json:"message,omitempty"`

	// Code contains any error code associated with the message.
	Code string `json:"code,omitempty"`

	// AdditionalInfo contains error info.
	AdditionalInfo []TypedErrorInfo `json:"additionalInfo,omitempty"`
}

// TypedErrorInfo captures the additional error info details when provisioning has failed.
type TypedErrorInfo struct {
	// Type contains ErrorInfo.
	Type string `json:"type,omitempty"`

	// Info contains category.
	Info JSONInfo `json:"info,omitempty"`
}

// JSONInfo captures category, recommended URL and troubleshooting URL for error when provisioning has failed.
type JSONInfo struct {
	// Category contains any extra information relevant to the error.
	Category string `json:"category,omitempty"`

	// RecommendedAction contains action user can take relevant to the error.
	RecommendedAction string `json:"recommendedAction,omitempty"`

	// TroubleshootingURL contains link to the troubleshooting steps.
	TroubleshootingURL string `json:"troubleshootingURL,omitempty"`
}

// ParentReference refers to the parent's namespace, api group and resource name.
type ParentReference struct {
	// Resource kind
	Kind string `json:"kind,omitempty"`
	// Resource name
	Name string `json:"name,omitempty"`
	// API group
	APIgroup string `json:"apiGroup,omitempty"`
	// Namespace
	Namespace string `json:"namespace,omitempty"`
}

// GetNamespacedName returns the Namespaced Name for the resource reference.
func (r *ParentReference) GetNamespacedName() *types.NamespacedName {
	if r == nil {
		return new(types.NamespacedName)
	}
	return &types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}
