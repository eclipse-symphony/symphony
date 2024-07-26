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
// CampaignContainer is the Schema for the CampaignContainers API
type CampaignContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.ContainerSpec   `json:"spec,omitempty"`
	Status common.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CampaignContainerList contains a list of CampaignContainer
type CampaignContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CampaignContainer `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-workflow-symphony-v1-campaigncontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigncontainers,verbs=create;update;delete,versions=v1,name=vcampaigncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CampaignContainer{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workflow-symphony-v1-campaigncontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=workflow.symphony,resources=campaigncontainers,verbs=create;update,versions=v1,name=mcampaigncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CampaignContainer{}

func init() {
	SchemeBuilder.Register(&CampaignContainer{}, &CampaignContainerList{})
}
