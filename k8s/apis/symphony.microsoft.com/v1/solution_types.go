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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Defines the desired state of Solution
type SolutionSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Components  []ComponentSpec   `json:"components,omitempty"`
	// Defines the version of a particular resource
	Version string `json:"version,omitempty"`
}

// Defines the observed state of Solution
type SolutionStatus struct {
	Properties map[string]string `json:"properties,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Defines a Solution resource
type Solution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SolutionSpec   `json:"spec,omitempty"`
	Status SolutionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Defines a list of Solutions
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Solution{}, &SolutionList{})
}
