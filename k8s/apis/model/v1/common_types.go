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
	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	"k8s.io/apimachinery/pkg/runtime"
)

// Defines a desired runtime component
// +kubebuilder:object:generate=true
type ComponentSpec struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   runtime.RawExtension `json:"properties,omitempty"`
	Routes       []model.RouteSpec    `json:"routes,omitempty"`
	Constraints  string               `json:"constraints,omitempty"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Skills       []string             `json:"skills,omitempty"`
}

// Defines the desired state of Target
// +kubebuilder:object:generate=true
type TargetSpec struct {
	DisplayName   string               `json:"displayName,omitempty"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
	Properties    map[string]string    `json:"properties,omitempty"`
	Components    []ComponentSpec      `json:"components,omitempty"`
	Constraints   string               `json:"constraints,omitempty"`
	Topologies    []model.TopologySpec `json:"topologies,omitempty"`
	ForceRedeploy bool                 `json:"forceRedeploy,omitempty"`
	Scope         string               `json:"scope,omitempty"`
	// Defines the version of a particular resource
	Version    string `json:"version,omitempty"`
	Generation string `json:"generation,omitempty"`
}

// +kubebuilder:object:generate=true
type SolutionSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Scope       string            `json:"scope,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Components  []ComponentSpec   `json:"components,omitempty"`
	// Defines the version of a particular resource
	Version string `json:"version,omitempty"`
}

// +kubebuilder:object:generate=true
type StageSpec struct {
	Name     string `json:"name,omitempty"`
	Contexts string `json:"contexts,omitempty"`
	Provider string `json:"provider,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Config        runtime.RawExtension `json:"config,omitempty"`
	StageSelector string               `json:"stageSelector,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs          runtime.RawExtension `json:"inputs,omitempty"`
	TriggeringStage string               `json:"triggeringStage,omitempty"`
}

// +kubebuilder:object:generate=true
type ActivationSpec struct {
	Campaign string `json:"campaign,omitempty"`
	Name     string `json:"name,omitempty"`
	Stage    string `json:"stage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs     runtime.RawExtension `json:"inputs,omitempty"`
	Generation string               `json:"generation,omitempty"`
}

// +kubebuilder:object:generate=true
type CampaignSpec struct {
	Name        string               `json:"name,omitempty"`
	FirstStage  string               `json:"firstStage,omitempty"`
	Stages      map[string]StageSpec `json:"stages,omitempty"`
	SelfDriving bool                 `json:"selfDriving,omitempty"`
}

// +kubebuilder:object:generate=true
type CatalogSpec struct {
	SiteId string `json:"siteId"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties runtime.RawExtension `json:"properties"`
	Metadata   map[string]string    `json:"metadata,omitempty"`
	ParentName string               `json:"parentName,omitempty"`
	ObjectRef  model.ObjectRef      `json:"objectRef,omitempty"`
	Generation string               `json:"generation,omitempty"`
}
