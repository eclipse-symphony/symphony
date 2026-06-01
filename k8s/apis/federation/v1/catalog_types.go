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
// Catalog is the Schema for the Catalogs API
type Catalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.ContainerSpec   `json:"spec,omitempty"`
	Status common.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CatalogList contains a list of Catalog
type CatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalog `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update;delete,versions=v1,name=vcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Catalog{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalog,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogs,verbs=create;update,versions=v1,name=mcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Catalog{}

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
}
