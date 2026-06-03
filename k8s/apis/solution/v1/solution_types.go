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
// Solution is the Schema for the Solutions API
type Solution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.ContainerSpec   `json:"spec,omitempty"`
	Status common.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SolutionList contains a list of Solution
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solution `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-solution-symphony-v1-solution,mutating=false,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update;delete,versions=v1,name=vsolution.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Solution{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-solution-symphony-v1-solution,mutating=true,failurePolicy=fail,sideEffects=None,groups=solution.symphony,resources=solutions,verbs=create;update,versions=v1,name=msolution.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Solution{}

func init() {
	SchemeBuilder.Register(&Solution{}, &SolutionList{})
}
