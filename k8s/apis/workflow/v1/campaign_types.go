/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	k8smodel "gopls-workspace/apis/model/v1"

	commoncontainers "gopls-workspace/apis/containers/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// CampaignStatus defines the observed state of Campaign
type CampaignStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true

// Campaign is the Schema for the campaigns API
type Campaign struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec k8smodel.CampaignSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CampaignList contains a list of Campaign
type CampaignList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Campaign `json:"items"`
}

type CampaignContainer struct {
	commoncontainers.CommonContainer
}

var _ webhook.Validator = &CampaignContainer{}

var _ webhook.Defaulter = &CampaignContainer{}

func init() {
	SchemeBuilder.Register(&Campaign{}, &CampaignList{})
	SchemeBuilder.Register(&CampaignContainer{})
}
