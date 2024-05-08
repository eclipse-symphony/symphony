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

type CatalogContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CatalogContainer is the Schema for the catalogs API
type CatalogContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.CatalogContainerSpec `json:"spec,omitempty"`
	Status CatalogContainerStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CatalogContainerList contains a list of Catalog
type CatalogContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogContainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CatalogContainer{}, &CatalogContainerList{})
}
