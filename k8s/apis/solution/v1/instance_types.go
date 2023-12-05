/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	apimodel "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties         map[string]string           `json:"properties,omitempty"`
	ProvisioningStatus apimodel.ProvisioningStatus `json:"provisioningStatus"`
	LastModified       metav1.Time                 `json:"lastModified,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.properties.status`
// +kubebuilder:printcolumn:name="Targets",type=string,JSONPath=`.status.properties.targets`
// +kubebuilder:printcolumn:name="Deployed",type=string,JSONPath=`.status.properties.deployed`

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   apimodel.InstanceSpec `json:"spec,omitempty"`
	Status InstanceStatus        `json:"status,omitempty"`
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
