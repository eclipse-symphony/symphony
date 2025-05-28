/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	k8smodel "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +k8s:deepcopy-gen=false
type ComponentProperties = runtime.RawExtension

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// Target is the Schema for the targets API
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.TargetSpec   `json:"spec,omitempty"`
	Status k8smodel.TargetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TargetList contains a list of Target
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Target `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Target{}, &TargetList{})
}

func (i *Target) GetStatus() k8smodel.TargetStatus {
	return i.Status
}

func (i *Target) SetStatus(status k8smodel.TargetStatus) {
	i.Status = status
}

func (i *Target) GetReconciliationPolicy() *k8smodel.ReconciliationPolicySpec {
	return i.Spec.ReconciliationPolicy
}
