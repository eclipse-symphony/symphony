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

// SkillSpec defines the desired state of Skill
type SkillSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Nodes       []NodeSpec        `json:"nodes"`
	Properties  map[string]string `json:"properties,omitempty"`
	Bindings    []BindingSpec     `json:"bindings,omitempty"`
	Edges       []EdgeSpec        `json:"edges"`
}

type NodeSpec struct {
	Id             string            `json:"id"`
	NodeType       string            `json:"type"`
	Name           string            `json:"name"`
	Configurations map[string]string `json:"configurations,omitempty"`
	Inputs         []RouteSpec       `json:"inputs,omitempty"`
	Outputs        []RouteSpec       `json:"outputs,omitempty"`
	Model          string            `json:"model,omitempty"`
}

type ConnectionSpec struct {
	Node  string `json:"node"`
	Route string `json:"route"`
}

type EdgeSpec struct {
	Source ConnectionSpec `json:"source"`
	Target ConnectionSpec `json:"target"`
}

// SkillStatus defines the observed state of Skill
type SkillStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	Properties map[string]string `json:"properties,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Skill is the Schema for the skills API
type Skill struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkillSpec   `json:"spec,omitempty"`
	Status SkillStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SkillList contains a list of Skill
type SkillList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Skill `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Skill{}, &SkillList{})
}
