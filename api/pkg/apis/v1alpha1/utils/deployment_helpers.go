/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	coa_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
)

func MatchTargets(instance model.InstanceState, targets []model.TargetState) []model.TargetState {
	ret := make(map[string]model.TargetState)
	if instance.Spec.Target.Name != "" {
		for _, t := range targets {
			targetName := ConvertReferenceToObjectName(instance.Spec.Target.Name)
			if matchString(targetName, t.ObjectMeta.Name) {
				ret[t.ObjectMeta.Name] = t
			}
		}
	}

	if len(instance.Spec.Target.Selector) > 0 {
		for _, t := range targets {
			fullMatch := true
			for k, v := range instance.Spec.Target.Selector {
				if tv, ok := t.Spec.Properties[k]; !ok || !matchString(v, tv) {
					fullMatch = false
				}
			}

			if fullMatch {
				ret[t.ObjectMeta.Name] = t
			}
		}
	}

	slice := make([]model.TargetState, 0, len(ret))
	for _, v := range ret {
		slice = append(slice, v)
	}

	return slice
}

func CreateSymphonyDeploymentFromTarget(ctx context.Context, target model.TargetState, namespace string) (model.DeploymentSpec, error) {
	key := GetTargetRuntimeKey(target.ObjectMeta.Name)
	scope := target.Spec.Scope
	if scope == "" {
		scope = constants.DefaultScope
	}

	ret := model.DeploymentSpec{
		ObjectNamespace: namespace,
	}
	solutionversion := model.SolutionVersionState{
		ObjectMeta: model.ObjectMeta{
			Name:      key,
			Namespace: target.ObjectMeta.Namespace,
		},
		Spec: &model.SolutionVersionSpec{
			DisplayName: key,
			Components:  make([]model.ComponentSpec, 0),
		},
	}

	for _, component := range target.Spec.Components {
		var c model.ComponentSpec
		data, _ := json.Marshal(component)
		err := json.Unmarshal(data, &c)

		if err != nil {
			return ret, err
		}
		solutionversion.Spec.Components = append(solutionversion.Spec.Components, c)
	}

	targets := make(map[string]model.TargetState)
	targets[target.ObjectMeta.Name] = target

	instance := model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      key,
			Namespace: target.ObjectMeta.Namespace,
		},
		Spec: &model.InstanceSpec{
			Scope:    scope,
			DisplayName: key,
			SolutionVersion: key,
			Target: model.TargetSelector{
				Name: target.ObjectMeta.Name,
			},
		},
	}
	// TODO: is this a good way to set guid for deployment?
	instance.ObjectMeta.SetGuid(target.ObjectMeta.GetGuid())

	ret.SolutionVersion = solutionversion
	ret.Instance = instance
	ret.Targets = targets
	ret.SolutionVersionName = key
	// set the target generation to the deployment
	ret.Generation = target.ObjectMeta.ETag
	assignments, err := AssignComponentsToTargets(ctx, ret.SolutionVersion.Spec.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}
	ret.IsDryRun = target.Spec.IsDryRun

	return ret, nil
}

// Add target-runtime prefix to notify the object is a target.
func GetTargetRuntimeKey(guid string) string {
	return fmt.Sprintf("%s-%s", constants.TargetRuntimePrefix, guid)
}

func ConstructSummaryId(name string, guid string) string {
	if guid != "" {
		return fmt.Sprintf("%s-%s", name, guid)
	}
	return name
}

func CreateSymphonyDeployment(ctx context.Context, instance model.InstanceState, solutionversion model.SolutionVersionState, targets []model.TargetState, devices []model.DeviceState, namespace string) (model.DeploymentSpec, error) {
	ret := model.DeploymentSpec{
		ObjectNamespace: namespace,
	}
	ret.Generation = instance.ObjectMeta.ETag

	// convert targets
	sTargets := make(map[string]model.TargetState)
	for _, t := range targets {
		sTargets[t.ObjectMeta.Name] = t
	}

	//TODO: handle devices
	ret.SolutionVersion = solutionversion
	ret.Targets = sTargets
	ret.Instance = instance
	ret.SolutionVersionName = solutionversion.ObjectMeta.Name
	ret.Instance.ObjectMeta.Name = instance.ObjectMeta.Name
	ret.Instance.ObjectMeta.SetGuid(instance.ObjectMeta.GetGuid())

	assignments, err := AssignComponentsToTargets(ctx, ret.SolutionVersion.Spec.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}
	ret.IsDryRun = instance.Spec.IsDryRun
	ret.IsInActive = instance.Spec.ActiveState == model.ActiveState_Inactive

	return ret, nil
}

func AssignComponentsToTargets(ctx context.Context, components []model.ComponentSpec, targets map[string]model.TargetState) (map[string]string, error) {
	//TODO: evaluate constraints
	ret := make(map[string]string)
	for key, target := range targets {
		ret[key] = ""
		for _, component := range components {
			match := true
			if component.Constraints != "" {
				parser := NewParser(component.Constraints)
				val, err := parser.Eval(coa_utils.EvaluationContext{
					Properties: target.Spec.Properties,
					Context:    ctx,
				})
				if err != nil {
					// append the error message with the component constraint expression
					errMsg := fmt.Sprintf("%s in constraint expression: %s", err.Error(), component.Constraints)
					return ret, v1alpha2.NewCOAError(nil, errMsg, v1alpha2.TargetPropertyNotFound)
				}
				match = (val == "true" || val == true)
			}
			if match {
				ret[key] += "{" + component.Name + "}"
			}
		}
	}

	return ret, nil
}
