package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type ContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:generate=true
type ContainerSpec struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Container is the Schema for the Containers API
type CommonContainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerSpec   `json:"spec,omitempty"`
	Status ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ContainerList contains a list of Container
type CommonContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CommonContainer `json:"items"`
}
