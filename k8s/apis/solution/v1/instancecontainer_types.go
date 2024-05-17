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

type InstanceContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// InstanceContainer is the Schema for the InstanceContainers API
type InstanceContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.InstanceContainerSpec `json:"spec,omitempty"`
	Status InstanceContainerStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// InstanceContainer1List contains a list of InstanceContainer
type InstanceContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstanceContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InstanceContainer{}, &InstanceContainerList{})
}
