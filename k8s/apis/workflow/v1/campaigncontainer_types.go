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

type CampaignContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CampaignContainer is the Schema for the campaigns API
type CampaignContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.CampaignContainerSpec `json:"spec,omitempty"`
	Status CampaignContainerStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CampaignContainerList contains a list of Campaign
type CampaignContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CampaignContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CampaignContainer{}, &CampaignContainerList{})
}
