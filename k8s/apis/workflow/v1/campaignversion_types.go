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

// CampaignVersionStatus defines the observed state of CampaignVersion
type CampaignVersionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true

// CampaignVersion is the Schema for the campaignversions API
type CampaignVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec k8smodel.CampaignVersionSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CampaignVersionList contains a list of CampaignVersion
type CampaignVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CampaignVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CampaignVersion{}, &CampaignVersionList{})
}
