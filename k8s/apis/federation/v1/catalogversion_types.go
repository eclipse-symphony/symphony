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

type CatalogVersionStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CatalogVersion is the Schema for the catalogversions API
type CatalogVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.CatalogVersionSpec `json:"spec,omitempty"`
	Status CatalogVersionStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CatalogVersionList contains a list of CatalogVersion
type CatalogVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogVersion{}, &CatalogVersionList{})
}
