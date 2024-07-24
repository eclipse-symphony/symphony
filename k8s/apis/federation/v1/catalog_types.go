/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	commoncontainers "gopls-workspace/apis/containers/v1"
	k8smodel "gopls-workspace/apis/model/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type CatalogStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Catalog is the Schema for the catalogs API
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.CatalogSpec `json:"spec,omitempty"`
	Status CatalogStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CatalogList contains a list of Catalog
type CatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalog `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// CatalogContainer is the Schema for the CatalogContainer API
type CatalogContainer struct {
	commoncontainers.CommonContainer
}

// +kubebuilder:object:root=true
// CatalogContainerList contains a list of CatalogContainer
type CatalogContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogContainer `json:"items"`
}

var _ webhook.Validator = &CatalogContainer{}

var _ webhook.Defaulter = &CatalogContainer{}

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
	SchemeBuilder.Register(&CatalogContainer{}, &CatalogContainerList{})
}
