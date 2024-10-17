/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

// Defines an error in the ARM resource for long running operations
// +kubebuilder:object:generate=true
type ErrorType struct {
	Code    string        `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Target  string        `json:"target,omitempty"`
	Details []TargetError `json:"details,omitempty"`
}

// Defines an error for symphony target
// +kubebuilder:object:generate=true
type TargetError struct {
	Code    string           `json:"code,omitempty"`
	Message string           `json:"message,omitempty"`
	Target  string           `json:"target,omitempty"`
	Details []ComponentError `json:"details,omitempty"`
}

// Defines an error for components defined in symphony
// +kubebuilder:object:generate=true
type ComponentError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Target  string `json:"target,omitempty"`
}

// Defines the state of the ARM resource for long running operations
// +kubebuilder:object:generate=true
type ProvisioningStatus struct {
	OperationID     string            `json:"operationId"`
	Status          string            `json:"status"`
	PercentComplete int64             `json:"percentComplete,omitempty"`
	FailureCause    string            `json:"failureCause,omitempty"`
	LogErrors       bool              `json:"logErrors,omitempty"`
	Error           ErrorType         `json:"error,omitempty"`
	Output          map[string]string `json:"output,omitempty"`
}
