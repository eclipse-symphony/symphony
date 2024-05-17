/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	k8smodel "github.com/eclipse-symphony/symphony/k8s/apis/model/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SolutionContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// SolutionContainer is the Schema for the SolutionContainer API
type SolutionContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.SolutionContainerSpec `json:"spec,omitempty"`
	Status SolutionContainerStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionContainerList contains a list of SolutionContainer
type SolutionContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SolutionContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SolutionContainer{}, &SolutionContainerList{})
}
