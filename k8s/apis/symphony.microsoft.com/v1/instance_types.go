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

type PipelineSpec struct {
	Name       string            `json:"name"`
	Skill      string            `json:"skill"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type VersionSpec struct {
	Solution   string `json:"solution"`
	Percentage int    `json:"percentage"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	DisplayName          string            `json:"displayName,omitempty"`
	Scope                string            `json:"scope,omitempty"`
	Parameters           map[string]string `json:"parameters,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	Solution             string            `json:"solution"`
	Versions             []VersionSpec     `json:"versions,omitempty"`
	Target               TargetSelector    `json:"target,omitempty"`
	Topologies           []TopologySpec    `json:"topologies,omitempty"`
	Pipelines            []PipelineSpec    `json:"pipelines,omitempty"`
	OptOutReconciliation bool              `json:"optOutReconciliation,omitempty"`
	Version              string            `json:"version,omitempty"`
}
type TargetSelector struct {
	Name     string            `json:"name,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties         map[string]string  `json:"properties,omitempty"`
	ProvisioningStatus ProvisioningStatus `json:"provisioningStatus"`
	LastModified       metav1.Time        `json:"lastModified,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.properties.status`
//+kubebuilder:printcolumn:name="Targets",type=string,JSONPath=`.status.properties.targets`
//+kubebuilder:printcolumn:name="Deployed",type=string,JSONPath=`.status.properties.deployed`

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
