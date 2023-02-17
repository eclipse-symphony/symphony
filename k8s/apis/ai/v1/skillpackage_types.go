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

// SkillPackageSpec defines the desired state of SkillPackage
type SkillPackageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DisplayName string            `json:"displayName,omitempty"`
	Skill       string            `json:"skill"`
	Properties  map[string]string `json:"properties,omitempty"`
	Constraints []ConstraintSpec  `json:"constraints,omitempty"`
	Routes      []RouteSpec       `json:"routes,omitempty"`
}

// SkillPackageStatus defines the observed state of SkillPackage
type SkillPackageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SkillPackage is the Schema for the skillpackages API
type SkillPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkillPackageSpec   `json:"spec,omitempty"`
	Status SkillPackageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SkillPackageList contains a list of SkillPackage
type SkillPackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkillPackage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkillPackage{}, &SkillPackageList{})
}
