/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	k8smodel "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SolutionVersionStatus defines the observed state of SolutionVersion
type SolutionVersionStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// SolutionVersion is the Schema for the solutionversions API
type SolutionVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.SolutionVersionSpec `json:"spec,omitempty"`
	Status SolutionVersionStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionVersionList contains a list of SolutionVersion
type SolutionVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SolutionVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SolutionVersion{}, &SolutionVersionList{})
}
