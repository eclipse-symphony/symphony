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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InstanceHistory is the Schema for the instancehistories API
type InstanceHistory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.InstanceHistorySpec   `json:"spec,omitempty"`
	Status k8smodel.InstanceHistoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceHistoryList contains a list of InstanceHistory
type InstanceHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InstanceHistory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InstanceHistory{}, &InstanceHistoryList{})
}
