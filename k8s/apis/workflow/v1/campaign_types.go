/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	common "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Campaign is the Schema for the Campaigns API
type Campaign struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.ContainerSpec   `json:"spec,omitempty"`
	Status common.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CampaignList contains a list of Campaign
type CampaignList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Campaign `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaign,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update;delete,versions=v1,name=vcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Campaign{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaign,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigns,verbs=create;update,versions=v1,name=mcampaign.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Campaign{}

func init() {
	SchemeBuilder.Register(&Campaign{}, &CampaignList{})
}
