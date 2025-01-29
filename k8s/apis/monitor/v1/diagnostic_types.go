/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DiagnosticSpec defines the desired state of Diagnostic
type DiagnosticSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Diagnostic. Edit diagnostic_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// DiagnosticStatus defines the observed state of Diagnostic
type DiagnosticStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Diagnostic is the Schema for the diagnostics API
type Diagnostic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosticSpec   `json:"spec,omitempty"`
	Status DiagnosticStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiagnosticList contains a list of Diagnostic
type DiagnosticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Diagnostic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Diagnostic{}, &DiagnosticList{})
}
