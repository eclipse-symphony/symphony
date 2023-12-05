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

// SkillPackageStatus defines the observed state of SkillPackage
type SkillPackageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// SkillPackage is the Schema for the skillpackages API
type SkillPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   apimodel.SkillPackageSpec `json:"spec,omitempty"`
	Status SkillPackageStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SkillPackageList contains a list of SkillPackage
type SkillPackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkillPackage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkillPackage{}, &SkillPackageList{})
}
