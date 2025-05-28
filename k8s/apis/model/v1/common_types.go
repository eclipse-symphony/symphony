/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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

// +kubebuilder:object:generate=true
type ContainerStatus struct {
	Properties map[string]string `json:"properties"`
}

// +kubebuilder:object:generate=true
type ContainerSpec struct {
}

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

// UnmarshalJSON customizes the JSON unmarshalling for SidecarSpec
func (s *SidecarSpec) UnmarshalJSON(data []byte) error {
	type Alias SidecarSpec
	aux := &struct {
		Properties json.RawMessage `json:"properties,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Properties = runtime.RawExtension{Raw: aux.Properties}

	return nil
}

// MarshalJSON customizes the JSON marshalling for SidecarSpec
func (s SidecarSpec) MarshalJSON() ([]byte, error) {
	type Alias SidecarSpec
	return json.Marshal(&struct {
		Properties json.RawMessage `json:"properties,omitempty"`
		*Alias
	}{
		Properties: json.RawMessage(s.Properties.Raw),
		Alias:      (*Alias)(&s),
	})
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

// UnmarshalJSON customizes the JSON unmarshalling for ComponentSpec
func (c *ComponentSpec) UnmarshalJSON(data []byte) error {
	type Alias ComponentSpec
	aux := &struct {
		Properties json.RawMessage `json:"properties,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Properties = runtime.RawExtension{Raw: aux.Properties}

	return nil
}

// MarshalJSON customizes the JSON marshalling for ComponentSpec
func (c ComponentSpec) MarshalJSON() ([]byte, error) {
	type Alias ComponentSpec
	return json.Marshal(&struct {
		Properties json.RawMessage `json:"properties,omitempty"`
		*Alias
	}{
		Properties: json.RawMessage(c.Properties.Raw),
		Alias:      (*Alias)(&c),
	})
}

// Defines the desired state of Target
// +kubebuilder:object:generate=true
type TargetSpec struct {
	DisplayName   string               `json:"displayName,omitempty"`
	Metadata      map[string]string    `json:"metadata,omitempty"`
	Scope         string               `json:"scope,omitempty"`
	SolutionScope string               `json:"solutionScope,omitempty"`
	Properties    map[string]string    `json:"properties,omitempty"`
	Components    []ComponentSpec      `json:"components,omitempty"`
	Constraints   string               `json:"constraints,omitempty"`
	Topologies    []model.TopologySpec `json:"topologies,omitempty"`
	ForceRedeploy bool                 `json:"forceRedeploy,omitempty"`
	IsDryRun      bool                 `json:"isDryRun,omitempty"`

	// Optional ReconcilicationPolicy to specify how target controller should reconcile.
	// Now only periodic reconciliation is supported. If the interval is 0, it will only reconcile
	// when the instance is created or updated.
	ReconciliationPolicy *ReconciliationPolicySpec `json:"reconciliationPolicy,omitempty"`
}

func (c TargetSpec) DeepEquals(other TargetSpec) bool {
	if c.DisplayName != other.DisplayName {
		return false
	}
	if !model.StringMapsEqual(c.Metadata, other.Metadata, nil) {
		return false
	}
	if c.Scope != other.Scope {
		return false
	}
	if c.SolutionScope != other.SolutionScope {
		return false
	}
	if !model.StringMapsEqual(c.Properties, other.Properties, nil) {
		return false
	}
	if !reflect.DeepEqual(c.Components, other.Components) {
		return false
	}
	if c.Constraints != other.Constraints {
		return false
	}
	if !model.SlicesEqual(c.Topologies, other.Topologies) {
		return false
	}
	if c.ForceRedeploy != other.ForceRedeploy {
		return false
	}
	if c.IsDryRun != other.IsDryRun {
		return false
	}

	// check reconciliation policy
	if c.ReconciliationPolicy == nil {
		return other.ReconciliationPolicy == nil
	}

	if other.ReconciliationPolicy == nil {
		return false
	}

	return c.ReconciliationPolicy.DeepEquals(*other.ReconciliationPolicy)
}

// +kubebuilder:object:generate=true
type InstanceSpec struct {
	DisplayName string               `json:"displayName,omitempty"`
	Scope       string               `json:"scope,omitempty"`
	Parameters  map[string]string    `json:"parameters,omitempty"` //TODO: Do we still need this?
	Metadata    map[string]string    `json:"metadata,omitempty"`
	Solution    string               `json:"solution"`
	Target      model.TargetSelector `json:"target,omitempty"`
	Topologies  []model.TopologySpec `json:"topologies,omitempty"`
	Pipelines   []model.PipelineSpec `json:"pipelines,omitempty"`
	IsDryRun    bool                 `json:"isDryRun,omitempty"`
	ActiveState model.ActiveState    `json:"activeState,omitempty"`

	// Optional ReconcilicationPolicy to specify how target controller should reconcile.
	// Now only periodic reconciliation is supported. If the interval is 0, it will only reconcile
	// when the instance is created or updated.
	ReconciliationPolicy *ReconciliationPolicySpec `json:"reconciliationPolicy,omitempty"`
}

func (c InstanceSpec) DeepEquals(other InstanceSpec) bool {
	if c.DisplayName != other.DisplayName {
		return false
	}

	if c.Scope != other.Scope {
		return false
	}

	if !model.StringMapsEqual(c.Parameters, other.Parameters, nil) {
		return false
	}
	if !model.StringMapsEqual(c.Metadata, other.Metadata, nil) {
		return false
	}

	if c.Solution != other.Solution {
		return false
	}

	if c.ActiveState != other.ActiveState {
		return false
	}

	equal, err := c.Target.DeepEquals(other.Target)
	if err != nil {
		return false
	}

	if !equal {
		return false
	}

	if !model.SlicesEqual(c.Topologies, other.Topologies) {
		return false
	}

	if !model.SlicesEqual(c.Pipelines, other.Pipelines) {
		return false
	}

	if c.IsDryRun != other.IsDryRun {
		return false
	}

	// check reconciliation policy
	if c.ReconciliationPolicy == nil {
		return other.ReconciliationPolicy == nil
	}

	if other.ReconciliationPolicy == nil {
		return false
	}

	return c.ReconciliationPolicy.DeepEquals(*other.ReconciliationPolicy)
}

// +kubebuilder:object:generate=true
type InstanceHistorySpec struct {
	// Snapshot of the instance
	DisplayName          string                    `json:"displayName,omitempty"`
	Scope                string                    `json:"scope,omitempty"`
	Parameters           map[string]string         `json:"parameters,omitempty"` //TODO: Do we still need this?
	Metadata             map[string]string         `json:"metadata,omitempty"`
	Solution             SolutionSpec              `json:"solution"`
	SolutionId           string                    `json:"solutionId"`
	Target               TargetSpec                `json:"target,omitempty"`
	TargetId             string                    `json:"targetId,omitempty"`
	TargetSelector       map[string]string         `json:"targetSelector,omitempty"`
	Topologies           []model.TopologySpec      `json:"topologies,omitempty"`
	Pipelines            []model.PipelineSpec      `json:"pipelines,omitempty"`
	IsDryRun             bool                      `json:"isDryRun,omitempty"`
	ReconciliationPolicy *ReconciliationPolicySpec `json:"reconciliationPolicy,omitempty"`

	// Add rootresoure to the instance history spec
	RootResource string `json:"rootResource,omitempty"`
}

func (c InstanceHistorySpec) DeepEquals(other InstanceHistorySpec) bool {
	if c.DisplayName != other.DisplayName {
		return false
	}
	if c.Scope != other.Scope {
		return false
	}
	if !model.StringMapsEqual(c.Parameters, other.Parameters, nil) {
		return false
	}
	if !model.StringMapsEqual(c.Metadata, other.Metadata, nil) {
		return false
	}
	if !c.Solution.DeepEquals(other.Solution) {
		return false
	}
	if c.SolutionId != other.SolutionId {
		return false
	}
	if !c.Target.DeepEquals(other.Target) {
		return false
	}
	if c.TargetId != other.TargetId {
		return false
	}
	if !model.StringMapsEqual(c.TargetSelector, other.TargetSelector, nil) {
		return false
	}
	if !model.SlicesEqual(c.Topologies, other.Topologies) {
		return false
	}
	if !model.SlicesEqual(c.Pipelines, other.Pipelines) {
		return false
	}
	if c.IsDryRun != other.IsDryRun {
		return false
	}
	if c.RootResource != other.RootResource {
		return false
	}
	// check reconciliation policy
	if c.ReconciliationPolicy == nil {
		return other.ReconciliationPolicy == nil
	}
	if other.ReconciliationPolicy == nil {
		return false
	}
	return c.ReconciliationPolicy.DeepEquals(*other.ReconciliationPolicy)
}

// +kubebuilder:object:generate=true
type SolutionSpec struct {
	DisplayName  string            `json:"displayName,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Components   []ComponentSpec   `json:"components,omitempty"`
	Version      string            `json:"version,omitempty"`
	RootResource string            `json:"rootResource,omitempty"`
}

func (c SolutionSpec) DeepEquals(other SolutionSpec) bool {
	if c.DisplayName != other.DisplayName {
		return false
	}
	if !model.StringMapsEqual(c.Metadata, other.Metadata, nil) {
		return false
	}
	if !reflect.DeepEqual(c.Components, other.Components) {
		return false
	}
	if c.Version != other.Version {
		return false
	}
	return c.RootResource == other.RootResource
}

// +kubebuilder:object:generate=true
type SolutionContainerSpec struct {
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
	Schedule        string               `json:"schedule,omitempty"`
	Target          string               `json:"target,omitempty"`
}

// UnmarshalJSON customizes the JSON unmarshalling for StageSpec
func (s *StageSpec) UnmarshalJSON(data []byte) error {
	type Alias StageSpec
	aux := &struct {
		Config json.RawMessage `json:"config,omitempty"`
		Inputs json.RawMessage `json:"inputs,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Config = runtime.RawExtension{Raw: aux.Config}
	s.Inputs = runtime.RawExtension{Raw: aux.Inputs}

	// validate if Schedule meet RFC 3339
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return fmt.Errorf("invalid timestamp format: %v", err)
		}
	}
	return nil
}

// MarshalJSON customizes the JSON marshalling for StageSpec
func (s StageSpec) MarshalJSON() ([]byte, error) {
	type Alias StageSpec
	if s.Schedule != "" {
		if _, err := time.Parse(time.RFC3339, s.Schedule); err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %v", err)
		}
	}
	return json.Marshal(&struct {
		Config json.RawMessage `json:"config,omitempty"`
		Inputs json.RawMessage `json:"inputs,omitempty"`
		*Alias
	}{
		Config: json.RawMessage(s.Config.Raw),
		Inputs: json.RawMessage(s.Inputs.Raw),
		Alias:  (*Alias)(&s),
	})
}

// +kubebuilder:object:generate=true
type ActivationSpec struct {
	Campaign string `json:"campaign,omitempty"`
	Stage    string `json:"stage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Inputs runtime.RawExtension `json:"inputs,omitempty"`
}

// UnmarshalJSON customizes the JSON unmarshalling for ActivationSpec
func (a *ActivationSpec) UnmarshalJSON(data []byte) error {
	type Alias ActivationSpec
	aux := &struct {
		Inputs json.RawMessage `json:"inputs,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	a.Inputs = runtime.RawExtension{Raw: aux.Inputs}

	return nil
}

// MarshalJSON customizes the JSON marshalling for ActivationSpec
func (a ActivationSpec) MarshalJSON() ([]byte, error) {
	type Alias ActivationSpec
	return json.Marshal(&struct {
		Inputs json.RawMessage `json:"inputs,omitempty"`
		*Alias
	}{
		Inputs: json.RawMessage(a.Inputs.Raw),
		Alias:  (*Alias)(&a),
	})
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
	CatalogType string            `json:"catalogType"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Properties   runtime.RawExtension `json:"properties"`
	ParentName   string               `json:"parentName,omitempty"`
	ObjectRef    model.ObjectRef      `json:"objectRef,omitempty"`
	Version      string               `json:"version,omitempty"`
	RootResource string               `json:"rootResource,omitempty"`
}

// UnmarshalJSON customizes the JSON unmarshalling for CatalogSpec
func (c *CatalogSpec) UnmarshalJSON(data []byte) error {
	type Alias CatalogSpec
	aux := &struct {
		Properties json.RawMessage `json:"properties"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Properties = runtime.RawExtension{Raw: aux.Properties}

	return nil
}

// MarshalJSON customizes the JSON marshalling for CatalogSpec
func (c CatalogSpec) MarshalJSON() ([]byte, error) {
	type Alias CatalogSpec
	return json.Marshal(&struct {
		Properties json.RawMessage `json:"properties"`
		*Alias
	}{
		Properties: json.RawMessage(c.Properties.Raw),
		Alias:      (*Alias)(&c),
	})
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

func (c DeployableStatus) DeepEquals(other DeployableStatus) bool {
	if !model.StringMapsEqual(c.Properties, other.Properties, nil) {
		return false
	}
	if !reflect.DeepEqual(c.ProvisioningStatus, other.ProvisioningStatus) {
		return false
	}
	if !reflect.DeepEqual(c.LastModified, other.LastModified) {
		return false
	}
	return true
}

// +kubebuilder:object:generate=true
type DeployableStatusV2 struct {
	ProvisioningStatus   model.ProvisioningStatus `json:"provisioningStatus"`
	LastModified         metav1.Time              `json:"lastModified,omitempty"`
	Deployed             int                      `json:"deployed,omitempty"`
	Targets              int                      `json:"targets,omitempty"` // missing in typespec
	Status               string                   `json:"status,omitempty"`
	StatusDetails        string                   `json:"statusDetails,omitempty"`
	RunningJobId         int                      `json:"runningJobId,omitempty"`
	ExpectedRunningJobId int                      `json:"expectedRunningJobId,omitempty"`
	Generation           int                      `json:"generation,omitempty"`
	TargetStatuses       []TargetDeployableStatus `json:"targetStatuses,omitempty"`
	Properties           map[string]string        `json:"properties,omitempty"`
}

func (c DeployableStatusV2) DeepEquals(other DeployableStatusV2) bool {
	if !model.StringMapsEqual(c.Properties, other.Properties, nil) {
		return false
	}
	if !reflect.DeepEqual(c.ProvisioningStatus, other.ProvisioningStatus) {
		return false
	}
	if c.LastModified != other.LastModified {
		return false
	}
	if c.Deployed != other.Deployed {
		return false
	}
	if c.Targets != other.Targets {
		return false
	}
	if c.Status != other.Status {
		return false
	}
	if c.StatusDetails != other.StatusDetails {
		return false
	}
	if c.RunningJobId != other.RunningJobId {
		return false
	}
	if c.ExpectedRunningJobId != other.ExpectedRunningJobId {
		return false
	}
	if c.Generation != other.Generation {
		return false
	}
	if !model.SlicesEqual(c.TargetStatuses, other.TargetStatuses) {
		return false
	}
	return true
}

// +kubebuilder:object:generate=true
type TargetDeployableStatus struct {
	Name              string                      `json:"name,omitempty"`
	Status            string                      `json:"status,omitempty"`
	ComponentStatuses []ComponentDeployableStatus `json:"componentStatuses,omitempty"`
}

func (c TargetDeployableStatus) DeepEquals(other model.IDeepEquals) (bool, error) {
	otherC, ok := other.(TargetDeployableStatus)
	if !ok {
		return false, errors.New("parameter is not a TargetDeployableStatus type")
	}
	if c.Name != otherC.Name {
		return false, nil
	}
	if c.Status != otherC.Status {
		return false, nil
	}
	if !model.SlicesEqual(c.ComponentStatuses, otherC.ComponentStatuses) {
		return false, nil
	}
	return true, nil
}

// +kubebuilder:object:generate=true
type ComponentDeployableStatus struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

func (c ComponentDeployableStatus) DeepEquals(other model.IDeepEquals) (bool, error) {
	otherC, ok := other.(ComponentDeployableStatus)
	if !ok {
		return false, errors.New("parameter is not a ComponentDeployableStatus type")
	}
	if c.Name != otherC.Name {
		return false, nil
	}
	if c.Status != otherC.Status {
		return false, nil
	}
	return true, nil
}

func (c *DeployableStatusV2) GetComponentStatus(targetName string, componentName string) string {
	if c == nil {
		return ""
	}
	for _, targetStatus := range c.TargetStatuses {
		if targetStatus.Name == targetName {
			for _, componentStatus := range targetStatus.ComponentStatuses {
				if componentStatus.Name == componentName {
					return componentStatus.Status
				}
			}
		}
	}
	return ""
}

func (c *DeployableStatusV2) SetTargetStatus(targetName string, status string) {
	if c == nil {
		return
	}
	for i, targetStatus := range c.TargetStatuses {
		if targetStatus.Name == targetName {
			c.TargetStatuses[i].Status = status
			return
		}
	}
	c.TargetStatuses = append(c.TargetStatuses, TargetDeployableStatus{
		Name:   targetName,
		Status: status,
	})
}

func (c *DeployableStatusV2) GetTargetStatus(targetName string) string {
	if c == nil {
		return ""
	}
	for _, targetStatus := range c.TargetStatuses {
		if targetStatus.Name == targetName {
			return targetStatus.Status
		}
	}
	return ""
}

func (c *DeployableStatusV2) SetComponentStatus(targetName string, componentName string, status string) {
	if c == nil {
		return
	}
	foundTarget := false
	foundComponent := false
	for i, targetStatus := range c.TargetStatuses {
		if targetStatus.Name == targetName {
			for j, componentStatus := range targetStatus.ComponentStatuses {
				if componentStatus.Name == componentName {
					c.TargetStatuses[i].ComponentStatuses[j].Status = status
					return
				}
			}
			if !foundComponent {
				c.TargetStatuses[i].ComponentStatuses = append(c.TargetStatuses[i].ComponentStatuses, ComponentDeployableStatus{
					Name:   componentName,
					Status: status,
				})
				return
			}
		}
	}
	if !foundTarget {
		c.TargetStatuses = append(c.TargetStatuses, TargetDeployableStatus{
			Name: targetName,
			ComponentStatuses: []ComponentDeployableStatus{
				{
					Name:   componentName,
					Status: status,
				},
			},
		})
	}
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus = DeployableStatusV2

// TargetStatus defines the observed state of Target
type TargetStatus = DeployableStatusV2

// InstanceHistoryStatus defines the observed state of Solution
type InstanceHistoryStatus = DeployableStatusV2

// +kubebuilder:object:generate=true
type ReconciliationPolicySpec struct {
	State ReconciliationPolicyState `json:"state"`
	// +kubebuilder:validation:MinLength=1
	Interval *string `json:"interval,omitempty"`
}

func (c ReconciliationPolicySpec) DeepEquals(other ReconciliationPolicySpec) bool {
	if c.State != other.State {
		return false
	}
	if *c.Interval != *other.Interval {
		return false
	}
	return true
}
