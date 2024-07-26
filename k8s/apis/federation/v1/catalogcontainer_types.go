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
// CatalogContainer is the Schema for the CatalogContainers API
type CatalogContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   common.ContainerSpec   `json:"spec,omitempty"`
	Status common.ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CatalogContainerList contains a list of CatalogContainer
type CatalogContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CatalogContainer `json:"items"`
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.

//+kubebuilder:webhook:path=/validate-federation-symphony-v1-catalogcontainer,mutating=false,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogcontainers,verbs=create;update;delete,versions=v1,name=vcatalogcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CatalogContainer{}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-federation-symphony-v1-catalogcontainer,mutating=true,failurePolicy=fail,sideEffects=None,groups=federation.symphony,resources=catalogcontainers,verbs=create;update,versions=v1,name=mcatalogcontainer.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CatalogContainer{}

func init() {
	SchemeBuilder.Register(&Catalog{}, &CatalogList{})
	SchemeBuilder.Register(&CatalogContainer{}, &CatalogContainerList{})
}
