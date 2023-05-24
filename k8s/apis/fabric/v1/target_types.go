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
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type ConstraintSpec struct {
	Key       string   `json:"key"`
	Qualifier string   `json:"qualifier,omitempty"`
	Operator  string   `json:"operator,omitempty"`
	Value     string   `json:"value,omitempty"`
	Values    []string `json:"values,omitempty"` //TODO: It seems kubebuilder has difficulties handling recursive defs. This is supposed to be an ConstraintSpec array
}

type FilterSpec struct {
	Direction  string            `json:"direction"`
	Type       string            `json:"type"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type RouteSpec struct {
	Name       string            `json:"name"`
	Route      string            `json:"route"`
	Properties map[string]string `json:"metadata,omitempty"`
	Filters    []FilterSpec      `json:"filters,omitempty"`
}

type ComponentSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   ComponentProperties `json:"properties,omitempty"`
	Routes       []RouteSpec         `json:"routes,omitempty"`
	Constraints  []ConstraintSpec    `json:"constraints,omitempty"`
	Dependencies []string            `json:"dependencies,omitempty"`
	Skills       []string            `json:"skills,omitempty"`
}

// +k8s:deepcopy-gen=false
type ComponentProperties = runtime.RawExtension
type BindingSpec struct {
	Role     string            `json:"role"`
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config,omitempty"`
}

type TopologySpec struct {
	Device   string            `json:"device,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
	Bindings []BindingSpec     `json:"bindings,omitempty"`
}

// TargetSpec defines the desired state of Target
type TargetSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	DisplayName   string            `json:"displayName,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Properties    map[string]string `json:"properties,omitempty"`
	Components    []ComponentSpec   `json:"components,omitempty"`
	Constraints   []ConstraintSpec  `json:"constraints,omitempty"`
	Topologies    []TopologySpec    `json:"topologies,omitempty"`
	ForceRedeploy bool              `json:"forceRedeploy,omitempty"`
}

type ErrorType struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ProvisioningStatus defines the state of the ARM resource for long running operations
type ProvisioningStatus struct {
	OperationID  string            `json:"operationId"`
	Status       string            `json:"status"`
	FailureCause string            `json:"failureCause,omitempty"`
	LogErrors    bool              `json:"logErrors,omitempty"`
	Error        ErrorType         `json:"error,omitempty"`
	Output       map[string]string `json:"output,omitempty"`
}

// TargetStatus defines the observed state of Target
type TargetStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Properties         map[string]string  `json:"properties,omitempty"`
	ProvisioningStatus ProvisioningStatus `json:"provisioningStatus"`
	LastModified       metav1.Time        `json:"lastModified,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.properties.status`

// Target is the Schema for the targets API
type Target struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetSpec   `json:"spec,omitempty"`
	Status TargetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TargetList contains a list of Target
type TargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Target `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Target{}, &TargetList{})
}
