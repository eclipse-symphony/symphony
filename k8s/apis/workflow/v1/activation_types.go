/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	k8smodel "gopls-workspace/apis/model/v1"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ActivationStatus struct {
	Status               v1alpha2.State `json:"status,omitempty"`
	StatusMessage        string         `json:"statusMessage,omitempty"`
	ActivationGeneration string         `json:"activationGeneration,omitempty"`
	UpdateTime           string         `json:"updateTime,omitempty"`
	StageHistory         []StageStatus  `json:"stageHistory,omitempty"`
}

type StageStatus struct {
	Stage     string `json:"stage,omitempty"`
	NextStage string `json:"nextStage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs runtime.RawExtension `json:"inputs,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Outputs       runtime.RawExtension `json:"outputs,omitempty"`
	Status        v1alpha2.State       `json:"status,omitempty"`
	StatusMessage string               `json:"statusMessage,omitempty"`
	ErrorMessage  string               `json:"errorMessage,omitempty"`
	IsActive      bool                 `json:"isActive,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.statusMessage`
// Activation is the Schema for the activations API
type Activation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.ActivationSpec `json:"spec,omitempty"`
	Status ActivationStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CampaignList contains a list of Activation
type ActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Activation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Activation{}, &ActivationList{})
}
