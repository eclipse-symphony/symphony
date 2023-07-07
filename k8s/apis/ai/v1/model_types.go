/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BindingSpec struct {
	Role       string            `json:"role"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type RouteSpec struct {
	Route    string            `json:"route"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Filters  []FilterSpec      `json:"filters,omitempty"`
}

// ModelSpec defines the desired state of Model
type ModelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DisplayName string            `json:"displayName,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Constraints string            `json:"constraints,omitempty"`
	Bindings    []BindingSpec     `json:"bindings,omitempty"`
	Routes      []RouteSpec       `json:"routes,omitempty"`
}

// ModelStatus defines the observed state of Model
type ModelStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Model is the Schema for the models API
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ModelList contains a list of Model
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

// func init() {
// 	SchemeBuilder.Register(&Model{}, &ModelList{})
// }
