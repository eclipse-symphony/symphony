/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package model

import "time"

type DeployableStatus struct {
	Properties         map[string]string  `json:"properties,omitempty"`
	ProvisioningStatus ProvisioningStatus `json:"provisioningStatus"`
	LastModified       time.Time          `json:"lastModified,omitempty"`
}

type DeployableStatusV2 struct {
	ProvisioningStatus   ProvisioningStatus       `json:"provisioningStatus"`
	LastModified         time.Time                `json:"lastModified,omitempty"`
	Deployed             int                      `json:"deployed,omitempty"`
	Targets              int                      `json:"targets,omitempty"`
	Status               string                   `json:"status,omitempty"`
	StatusDetails        string                   `json:"statusDetails,omitempty"`
	RunningJobId         int                      `json:"runningJobId,omitempty"`
	ExpectedRunningJobId int                      `json:"expectedRunningJobId,omitempty"`
	Generation           int                      `json:"generation,omitempty"`
	TargetStatuses       []TargetDeployableStatus `json:"targetStatuses,omitempty"`
	Properties           map[string]string        `json:"properties,omitempty"`
}

type TargetDeployableStatus struct {
	Name              string                      `json:"name,omitempty"`
	Status            string                      `json:"status,omitempty"`
	ComponentStatuses []ComponentDeployableStatus `json:"componentStatuses,omitempty"`
}

type ComponentDeployableStatus struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

func (c *ComponentDeployableStatus) DeepCopy() *ComponentDeployableStatus {
	if c == nil {
		return nil
	}
	out := new(ComponentDeployableStatus)
	out.Name = c.Name
	out.Status = c.Status
	return out
}

func (t *TargetDeployableStatus) DeepCopy() *TargetDeployableStatus {
	if t == nil {
		return nil
	}
	out := new(TargetDeployableStatus)
	out.Name = t.Name
	out.Status = t.Status
	out.ComponentStatuses = make([]ComponentDeployableStatus, len(t.ComponentStatuses))
	for i := range t.ComponentStatuses {
		out.ComponentStatuses[i] = *t.ComponentStatuses[i].DeepCopy()
	}
	return out
}

func (d *DeployableStatusV2) DeepCopy() *DeployableStatusV2 {
	if d == nil {
		return nil
	}
	out := new(DeployableStatusV2)
	out.Deployed = d.Deployed
	out.ExpectedRunningJobId = d.ExpectedRunningJobId
	out.Generation = d.Generation
	out.LastModified = d.LastModified
	out.ProvisioningStatus = d.ProvisioningStatus
	out.RunningJobId = d.RunningJobId
	out.Status = d.Status
	out.StatusDetails = d.StatusDetails
	out.TargetStatuses = make([]TargetDeployableStatus, len(d.TargetStatuses))
	for i := range d.TargetStatuses {
		out.TargetStatuses[i] = *d.TargetStatuses[i].DeepCopy()
	}
	out.Properties = make(map[string]string)
	for k, v := range d.Properties {
		out.Properties[k] = v
	}
	return out
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
