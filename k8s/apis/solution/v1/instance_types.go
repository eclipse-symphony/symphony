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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Targets",type=string,JSONPath=`.status.targets`
// +kubebuilder:printcolumn:name="Deployed",type=string,JSONPath=`.status.deployed`

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.InstanceSpec   `json:"spec,omitempty"`
	Status k8smodel.InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

func (i *Instance) GetStatus() k8smodel.InstanceStatus {
	return i.Status
}

func (i *Instance) SetStatus(status k8smodel.InstanceStatus) {
	i.Status = status
}

func (i *Instance) GetReconciliationPolicy() *k8smodel.ReconciliationPolicySpec {
	return i.Spec.ReconciliationPolicy
}
