/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ReconciliationPolicy_Active allows the controller to reconcile periodically
	ReconciliationPolicy_Active ReconciliationPolicyState = "active"
	// ReconciliationPolicy_Inactive disables periodic reconciliation
	ReconciliationPolicy_Inactive ReconciliationPolicyState = "inactive"
)

// +kubebuilder:validation:Enum=active;inactive;
type ReconciliationPolicyState string

func (r ReconciliationPolicyState) String() string {
	return string(r)
}

func (r ReconciliationPolicyState) IsActive() bool {
	return strings.ToLower(r.String()) == ReconciliationPolicy_Active.String()
}

func (r ReconciliationPolicyState) IsInActive() bool {
	return strings.ToLower(r.String()) == ReconciliationPolicy_Inactive.String()
}

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
	Generation    string               `json:"generation,omitempty"`
	Version       string               `json:"version,omitempty"`
	RootResource  string               `json:"rootResource,omitempty"`

	// Optional ReconcilicationPolicy to specify how target controller should reconcile.
	// Now only periodic reconciliation is supported. If the interval is 0, it will only reconcile
	// when the instance is created or updated.
	ReconciliationPolicy *ReconciliationPolicySpec `json:"reconciliationPolicy,omitempty"`
}

// +kubebuilder:object:generate=true
type InstanceSpec struct {
	DisplayName  string                       `json:"displayName,omitempty"`
	Scope        string                       `json:"scope,omitempty"`
	Parameters   map[string]string            `json:"parameters,omitempty"` //TODO: Do we still need this?
	Metadata     map[string]string            `json:"metadata,omitempty"`
	Solution     string                       `json:"solution"`
	Target       model.TargetSelector         `json:"target,omitempty"`
	Topologies   []model.TopologySpec         `json:"topologies,omitempty"`
	Pipelines    []model.PipelineSpec         `json:"pipelines,omitempty"`
	Arguments    map[string]map[string]string `json:"arguments,omitempty"`
	Generation   string                       `json:"generation,omitempty"`
	Version      string                       `json:"version,omitempty"`
	RootResource string                       `json:"rootResource,omitempty"`

	// Optional ReconcilicationPolicy to specify how target controller should reconcile.
	// Now only periodic reconciliation is supported. If the interval is 0, it will only reconcile
	// when the instance is created or updated.
	ReconciliationPolicy *ReconciliationPolicySpec `json:"reconciliationPolicy,omitempty"`
}

// +kubebuilder:object:generate=true
type TargetContainerSpec struct {
}

// +kubebuilder:object:generate=true
type InstanceContainerSpec struct {
}

// +kubebuilder:object:generate=true
type SolutionSpec struct {
	DisplayName  string            `json:"displayName,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Components   []ComponentSpec   `json:"components,omitempty"`
	Version      string            `json:"version,omitempty"`
	RootResource string            `json:"rootResource,omitempty"`
}

// +kubebuilder:object:generate=true
type SolutionContainerSpec struct {
}

// +kubebuilder:object:generate=true
type ScheduleSpec struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Zone string `json:"zone"`
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
}

// +kubebuilder:object:generate=true
type ActivationSpec struct {
	Campaign string `json:"campaign,omitempty"`
	Stage    string `json:"stage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs     runtime.RawExtension `json:"inputs,omitempty"`
	Generation string               `json:"generation,omitempty"`
}

// +kubebuilder:object:generate=true
type CampaignSpec struct {
	Name         string               `json:"name,omitempty"`
	FirstStage   string               `json:"firstStage,omitempty"`
	Stages       map[string]StageSpec `json:"stages,omitempty"`
	SelfDriving  bool                 `json:"selfDriving,omitempty"`
	Version      string               `json:"version,omitempty"`
	RootResource string               `json:"rootResource,omitempty"`
}

// +kubebuilder:object:generate=true
type CampaignContainerSpec struct {
}

// +kubebuilder:object:generate=true
type CatalogSpec struct {
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   runtime.RawExtension `json:"properties"`
	ParentName   string               `json:"parentName,omitempty"`
	ObjectRef    model.ObjectRef      `json:"objectRef,omitempty"`
	Generation   string               `json:"generation,omitempty"`
	Version      string               `json:"version,omitempty"`
	RootResource string               `json:"rootResource,omitempty"`
}

// +kubebuilder:object:generate=true
type CatalogContainerSpec struct {
}

// +kubebuilder:object:generate=true
type DeployableStatus struct {
	Properties         map[string]string        `json:"properties,omitempty"`
	ProvisioningStatus model.ProvisioningStatus `json:"provisioningStatus"`
	LastModified       metav1.Time              `json:"lastModified,omitempty"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus = DeployableStatus

// TargetStatus defines the observed state of Target
type TargetStatus = DeployableStatus

// +kubebuilder:object:generate=true
type ReconciliationPolicySpec struct {
	State ReconciliationPolicyState `json:"state"`
	// +kubebuilder:validation:MinLength=1
	Interval *string `json:"interval,omitempty"`
}
