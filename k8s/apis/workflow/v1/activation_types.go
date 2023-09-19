/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package v1

import (
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	k8smodel "github.com/azure/symphony/k8s/apis/model/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ActivationStatus struct {
	Stage     string `json:"stage"`
	NextStage string `json:"nextStage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs runtime.RawExtension `json:"inputs,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Outputs              runtime.RawExtension `json:"outputs,omitempty"`
	Status               v1alpha2.State       `json:"status,omitempty"`
	ErrorMessage         string               `json:"errorMessage,omitempty"`
	IsActive             bool                 `json:"isActive,omitempty"`
	ActivationGeneration string               `json:"activationGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Next Stage",type=string,JSONPath=`.status.nextStage`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// Activation is the Schema for the activations API
type Activation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   k8smodel.ActivationSpec `json:"spec,omitempty"`
	Status ActivationStatus        `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CampaignList contains a list of Activation
type ActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Activation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Activation{}, &ActivationList{})
}
