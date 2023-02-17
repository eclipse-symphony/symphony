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

type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}
type RouteSpec struct {
	Route      string            `json:"route"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

type ConstraintSpec struct {
	Key       string   `json:"key"`
	Qualifier string   `json:"qualifier,omitempty"`
	Operator  string   `json:"operator,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"` //TODO: It seems kubebuilder has difficulties handling recursive defs. This is supposed to be an ConstraintSpec array
}

type ComponentSpec struct {
	Name         string            `json:"name"`
	Type         string            `json:"type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
	Routes       []RouteSpec       `json:"routes,omitempty"`
	Constraints  []ConstraintSpec  `json:"constraints,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Skills       []string          `json:"skills,omitempty"`
}

// SolutionSpec defines the desired state of Solution
type SolutionSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	DisplayName string            `json:"displayName,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Components  []ComponentSpec   `json:"components,omitempty"`
}

// SolutionStatus defines the observed state of Solution
type SolutionStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties map[string]string `json:"properties,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Solution is the Schema for the solutions API
type Solution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SolutionSpec   `json:"spec,omitempty"`
	Status SolutionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SolutionList contains a list of Solution
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Solution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Solution{}, &SolutionList{})
}
