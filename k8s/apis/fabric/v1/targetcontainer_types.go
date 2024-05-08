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

// TargetContainerStatus defines the observed state of Target
type TargetContainerStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Target is the Schema for the targets API
type TargetContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.TargetContainerSpec `json:"spec,omitempty"`
	Status TargetContainerStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TargetList contains a list of Target
type TargetContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TargetContainer{}, &TargetContainerList{})
}
