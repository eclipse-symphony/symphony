/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	cc "gopls-workspace/apis/containers/v1"
	k8smodel "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// SolutionStatus defines the observed state of Solution
type SolutionStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Solution is the Schema for the solutions API
type Solution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.SolutionSpec `json:"spec,omitempty"`
	Status SolutionStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionList contains a list of Solution
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solution `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// SolutionContainer is the Schema for the SolutionContainers API
type SolutionContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   cc.ContainerSpec   `json:"spec,omitempty"`
	Status cc.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionContainerList contains a list of SolutionContainer
type SolutionContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SolutionContainer `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solutioncontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutioncontainers,verbs=create;update;delete,versions=v1,name=vsolutioncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &SolutionContainer{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-solutioncontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutioncontainers,verbs=create;update,versions=v1,name=msolutioncontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &SolutionContainer{}

func init() {
	SchemeBuilder.Register(&Solution{}, &SolutionList{})
	SchemeBuilder.Register(&SolutionContainer{}, &SolutionContainerList{})
}
