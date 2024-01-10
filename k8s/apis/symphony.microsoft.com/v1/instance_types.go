/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Defines the observed state of Instance
type InstanceStatus struct {
	Properties         map[string]string        `json:"properties,omitempty"`
	ProvisioningStatus model.ProvisioningStatus `json:"provisioningStatus"`
	LastModified       metav1.Time              `json:"lastModified,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.properties.status`
//+kubebuilder:printcolumn:name="Targets",type=string,JSONPath=`.status.properties.targets`
//+kubebuilder:printcolumn:name="Deployed",type=string,JSONPath=`.status.properties.deployed`

// Defines an Instance resource
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   apimodel.InstanceSpec `json:"spec,omitempty"`
	Status InstanceStatus        `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Defines a list of Instances
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
