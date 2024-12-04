/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package solution

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
)

func PlanForDeployment(deployment model.DeploymentSpec, state model.DeploymentState) (model.DeploymentPlan, error) {
	ret := model.DeploymentPlan{
		Steps: make([]model.DeploymentStep, 0),
	}
	for _, c := range state.Components {
		for _, t := range state.Targets {
			key := fmt.Sprintf("%s::%s", c.Name, t.Name) //TODO: this assumes provider/component keys don't contain "::"
			if v, ok := state.TargetComponent[key]; ok {
				role := c.Type
				if role == "" {
					role = "instance"
				}
				action := model.ComponentUpdate
				if strings.HasPrefix(v, "-") {
					action = model.ComponentDelete
				}
				index := ret.FindLastTargetRole(t.Name, c.Type)
				if index < 0 || !ret.CanAppendToStep(index, c) {
					ret.Steps = append(ret.Steps, model.DeploymentStep{
						Target:  t.Name,
						Role:    role,
						IsFirst: index < 0,
						Components: []model.ComponentStep{
							{
								Action:    action,
								Component: c,
							},
						},
					})
				} else {
					ret.Steps[index].Components = append(ret.Steps[index].Components, model.ComponentStep{
						Action:    action,
						Component: c,
					})
				}
			}
		}
	}
	return ret.RevisedForDeletion(), nil
}

func NewDeploymentState(deployment model.DeploymentSpec) (model.DeploymentState, error) {
	ret := model.DeploymentState{
		Components:      make([]model.ComponentSpec, 0),
		Targets:         make([]model.TargetDesc, 0),
		TargetComponent: make(map[string]string),
	}

	components, err := sortByDepedencies(deployment.Solution.Spec.Components)
	if err != nil {
		return ret, err
	}

	for _, component := range components {
		ret.Components = append(ret.Components, component)

		providers := findComponentProviders(component.Name, deployment)
		for k, v := range providers {
			found := false
			for _, t := range ret.Targets {
				if t.Name == k {
					found = true
					break
				}
			}
			if !found {
				ret.Targets = append(ret.Targets, model.TargetDesc{Name: k, Spec: *v.Spec})
			}
			t := component.Type
			if t == "" {
				t = "instance"
			}
			ret.TargetComponent[fmt.Sprintf("%s::%s", component.Name, k)] = t //TODO: this assumes provider/component keys don't contain "::"
		}
	}

	sort.Sort(model.ByTargetName(ret.Targets)) //sort target by name for easier testing

	return ret, nil
}
func MergeDeploymentStates(previous model.DeploymentState, current model.DeploymentState) model.DeploymentState {
	// merge components
	for _, c := range previous.Components {
		found := false
		for _, cc := range current.Components {
			if cc.Name == c.Name {
				found = true
				break
			}
		}
		if !found {
			current.Components = append(current.Components, c)
		}
	}
	// merge targets
	for _, t := range previous.Targets {
		found := false
		for _, tt := range current.Targets {
			if tt.Name == t.Name {
				found = true
				break
			}
		}
		if !found {
			current.Targets = append(current.Targets, t)
		}
	}
	// merge state matrix
	for k, v := range previous.TargetComponent {
		if _, ok := current.TargetComponent[k]; !ok {
			if !strings.HasPrefix(v, "-") {
				current.TargetComponent[k] = "-" + v
			}
		}
	}
	return current
}
func findComponentProviders(component string, deployment model.DeploymentSpec) map[string]model.TargetState {
	ret := make(map[string]model.TargetState)
	for k, v := range deployment.Assignments {
		if v != "" {
			if strings.Contains(v, "{"+component+"}") {
				if t, ok := deployment.Targets[k]; ok {
					ret[k] = t
				}
			}
		}
	}
	return ret
}
