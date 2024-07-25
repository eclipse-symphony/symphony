/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	commoncontainers "gopls-workspace/apis/containers/v1"
	k8smodel "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// SolutionStatus defines the observed state of Solution
type SolutionStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Solution is the Schema for the solutions API
type Solution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.SolutionSpec `json:"spec,omitempty"`
	Status SolutionStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionList contains a list of Solution
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solution `json:"items"`
}

// +kubebuilder:object:root=true
// SolutionContainer is the Schema for the SolutionContainer API
type SolutionContainer struct {
	commoncontainers.CommonContainer
}

// +kubebuilder:object:root=true
// SolutionContainerList contains a list of SolutionContainer
type SolutionContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SolutionContainer `json:"items"`
}

var _ webhook.Validator = &SolutionContainer{}

var _ webhook.Defaulter = &SolutionContainer{}

func init() {
	SchemeBuilder.Register(&Solution{}, &SolutionList{})
	SchemeBuilder.Register(&SolutionContainer{}, &SolutionContainerList{})
}
