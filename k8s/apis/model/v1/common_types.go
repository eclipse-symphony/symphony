/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=true
type SidecarSpec struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

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
	Sidecars     []SidecarSpec        `json:"sidecars,omitempty"`
}

// Defines the desired state of Target
// +kubebuilder:object:generate=true
type TargetSpec struct {
	DisplayName   string               `json:"displayName,omitempty"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
	Scope         string               `json:"scope,omitempty"`
	Properties    map[string]string    `json:"properties,omitempty"`
	Components    []ComponentSpec      `json:"components,omitempty"`
	Constraints   string               `json:"constraints,omitempty"`
	Topologies    []model.TopologySpec `json:"topologies,omitempty"`
	ForceRedeploy bool                 `json:"forceRedeploy,omitempty"`
	// Defines the version of a particular resource
	Version    string `json:"version,omitempty"`
	Generation string `json:"generation,omitempty"`
}

// +kubebuilder:object:generate=true
type SolutionSpec struct {
	DisplayName string            `json:"displayName,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Components  []ComponentSpec   `json:"components,omitempty"`
	// Defines the version of a particular resource
	Version string `json:"version,omitempty"`
}

// +kubebuilder:object:generate=true
type ScheduleSpec struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Zone string `json:"zone"`
}

// +kubebuilder:object:generate=true
type ProxyConfigSpec struct {
	BaseUrl  string `json:"baseUrl,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// +kubebuilder:object:generate=true
type ProxySpec struct {
	Provider string          `json:"provider,omitempty"`
	Config   ProxyConfigSpec `json:"config,omitempty"`
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
	Schedule        *ScheduleSpec        `json:"schedule,omitempty"`
	Proxy           *ProxySpec           `json:"proxy,omitempty"`
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
	SiteId   string            `json:"siteId"`
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties runtime.RawExtension `json:"properties"`
	ParentName string               `json:"parentName,omitempty"`
	ObjectRef  model.ObjectRef      `json:"objectRef,omitempty"`
	Generation string               `json:"generation,omitempty"`
}
