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

package model

import (
	"strings"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
)

type DeploymentPlan struct {
	Steps []DeploymentStep
}
type DeploymentStep struct {
	Target     string
	Components []ComponentStep
	Role       string
	IsFirst    bool
}
type ComponentStep struct {
	Action    string        `json:"action"`
	Component ComponentSpec `json:"component"`
}

type TargetDesc struct {
	Name string
	Spec TargetSpec
}
type ByTargetName []TargetDesc

func (p ByTargetName) Len() int           { return len(p) }
func (p ByTargetName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p ByTargetName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type DeploymentState struct {
	Components      []ComponentSpec
	Targets         []TargetDesc
	TargetComponent map[string]string
}

func (s DeploymentStep) PrepareResultMap() map[string]ComponentResultSpec {
	ret := make(map[string]ComponentResultSpec)
	for _, c := range s.Components {
		ret[c.Component.Name] = ComponentResultSpec{
			Status:  v1alpha2.Untouched,
			Message: "",
		}
	}
	return ret
}
func (s DeploymentStep) GetComponents() []ComponentSpec {
	ret := make([]ComponentSpec, 0)
	for _, c := range s.Components {
		ret = append(ret, c.Component)
	}
	return ret
}
func (s DeploymentStep) GetUpdatedComponents() []ComponentSpec {
	ret := make([]ComponentSpec, 0)
	for _, c := range s.Components {
		if c.Action == "update" {
			ret = append(ret, c.Component)
		}
	}
	return ret
}
func (s DeploymentStep) GetDeletedComponents() []ComponentSpec {
	ret := make([]ComponentSpec, 0)
	for _, c := range s.Components {
		if c.Action == "delete" {
			ret = append(ret, c.Component)
		}
	}
	return ret
}
func (s DeploymentStep) GetUpdatedComponentSteps() []ComponentStep {
	ret := make([]ComponentStep, 0)
	for _, c := range s.Components {
		if c.Action == "update" {
			ret = append(ret, c)
		}
	}
	return ret
}
func (t *DeploymentState) MarkRemoveAll() {
	for k, v := range t.TargetComponent {
		if !strings.HasPrefix(v, "-") {
			t.TargetComponent[k] = "-" + v
		}
	}
}
func (t *DeploymentState) ClearAllRemoved() {
	for k, v := range t.TargetComponent {
		if strings.HasPrefix(v, "-") {
			delete(t.TargetComponent, k)
		}
	}
}
func (p DeploymentPlan) FindLastTargetRole(target, role string) int {
	for i := len(p.Steps) - 1; i >= 0; i-- {
		if p.Steps[i].Role == role && p.Steps[i].Target == target {
			return i
		}
	}
	return -1
}
func (p DeploymentPlan) CanAppendToStep(step int, component ComponentSpec) bool {
	canAppend := true
	for _, d := range component.Dependencies {
		resolved := false
		for j := 0; j <= step; j++ {
			for _, c := range p.Steps[j].Components {
				if c.Component.Name == d && c.Action == "update" {
					resolved = true
					break
				}
			}
			if resolved {
				break
			}
		}
		if !resolved {
			return false
		}
	}
	return canAppend
}
func (p DeploymentPlan) RevisedForDeletion() DeploymentPlan {
	ret := DeploymentPlan{
		Steps: make([]DeploymentStep, 0),
	}
	// create a stack to save deleted steps
	deletedSteps := make([]DeploymentStep, 0)

	for _, s := range p.Steps {
		deleted := s.GetDeletedComponents()
		all := s.GetComponents()
		if len(deleted) == 0 {
			ret.Steps = append(ret.Steps, s)
		} else if len(deleted) == len(all) {
			// add this step to the deleted steps stack
			deletedSteps = append(deletedSteps, s)
		} else {
			//split the steps into two steps, one with updated only, one with deleted only
			ret.Steps = append(ret.Steps, makeUpdateStep(s))
			deletedSteps = append(deletedSteps, makeReversedDeletionStep(s))
		}
	}
	for i := len(deletedSteps) - 1; i >= 0; i-- {
		ret.Steps = append(ret.Steps, deletedSteps[i])
	}
	return ret
}
func makeUpdateStep(step DeploymentStep) DeploymentStep {
	ret := DeploymentStep{
		Target:     step.Target,
		Components: make([]ComponentStep, 0),
		Role:       step.Role,
		IsFirst:    step.IsFirst,
	}
	for _, c := range step.Components {
		if c.Action == "update" {
			ret.Components = append(ret.Components, c)
		}
	}
	return ret
}
func makeReversedDeletionStep(step DeploymentStep) DeploymentStep {
	ret := DeploymentStep{
		Target:     step.Target,
		Components: make([]ComponentStep, 0),
		Role:       step.Role,
		IsFirst:    step.IsFirst,
	}
	for i := len(step.Components) - 1; i >= 0; i-- {
		if step.Components[i].Action == "delete" {
			ret.Components = append(ret.Components, step.Components[i])
		}
	}
	return ret
}
