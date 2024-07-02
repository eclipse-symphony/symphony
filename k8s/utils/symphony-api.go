/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"context"
	"encoding/json"
	"gopls-workspace/constants"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	apimodel "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"

	fabric_v1 "gopls-workspace/apis/fabric/v1"
	k8smodel "gopls-workspace/apis/model/v1"
	solution_v1 "gopls-workspace/apis/solution/v1"
)

var (
	SymphonyAPIAddressBase = os.Getenv(constants.SymphonyAPIUrlEnvName)
)

type (
	ApiClient interface {
		api_utils.SummaryGetter
		api_utils.Dispatcher
	}

	DeploymentResources struct {
		Instance         solution_v1.Instance
		Solution         solution_v1.Solution
		TargetList       fabric_v1.TargetList
		TargetCandidates []fabric_v1.Target
	}
)

func K8SSidecarSpecToAPISidecarSpec(sidecar k8smodel.SidecarSpec) (apimodel.SidecarSpec, error) {
	sidecarSpec := apimodel.SidecarSpec{}
	data, _ := json.Marshal(sidecar)
	var err error
	err = json.Unmarshal(data, &sidecarSpec)
	if err != nil {
		return apimodel.SidecarSpec{}, err
	}
	sidecarSpec.Properties = make(map[string]interface{})
	err = json.Unmarshal(sidecar.Properties.Raw, &sidecarSpec.Properties)
	return sidecarSpec, err
}

func K8SComponentSpecToAPIComponentSpec(component k8smodel.ComponentSpec) (apimodel.ComponentSpec, error) {
	componentSpec := apimodel.ComponentSpec{}
	data, _ := json.Marshal(component)
	var err error
	err = json.Unmarshal(data, &componentSpec)
	if err != nil {
		return apimodel.ComponentSpec{}, err
	}
	componentSpec.Properties = make(map[string]interface{})
	err = json.Unmarshal(component.Properties.Raw, &componentSpec.Properties)
	return componentSpec, err
}

func K8STargetToAPITargetState(target fabric_v1.Target) (apimodel.TargetState, error) {
	ret := apimodel.TargetState{
		ObjectMeta: apimodel.ObjectMeta{
			Name:        target.ObjectMeta.Name,
			Namespace:   target.ObjectMeta.Namespace,
			Labels:      target.ObjectMeta.Labels,
			Annotations: target.ObjectMeta.Annotations,
		},
		Spec: &apimodel.TargetSpec{
			DisplayName:   target.Spec.DisplayName,
			Metadata:      target.Spec.Metadata,
			Scope:         target.Spec.Scope,
			Properties:    target.Spec.Properties,
			Constraints:   target.Spec.Constraints,
			ForceRedeploy: target.Spec.ForceRedeploy,
			Topologies:    target.Spec.Topologies,
		},
	}

	var err error
	ret.Spec.Components = make([]apimodel.ComponentSpec, len(target.Spec.Components))
	for i, c := range target.Spec.Components {
		ret.Spec.Components[i], err = K8SComponentSpecToAPIComponentSpec(c)
		if err != nil {
			return apimodel.TargetState{}, err
		}
	}

	return ret, nil
}

func K8SInstanceToAPIInstanceState(instance solution_v1.Instance) (apimodel.InstanceState, error) {
	ret := apimodel.InstanceState{
		ObjectMeta: apimodel.ObjectMeta{
			Name:        instance.ObjectMeta.Name,
			Namespace:   instance.ObjectMeta.Namespace,
			Labels:      instance.ObjectMeta.Labels,
			Annotations: instance.ObjectMeta.Annotations,
		},
		Spec: &apimodel.InstanceSpec{
			Scope:       instance.Spec.Scope,
			DisplayName: instance.Spec.DisplayName,
			Solution:    instance.Spec.Solution,
			Target:      instance.Spec.Target,
			Parameters:  instance.Spec.Parameters,
			Metadata:    instance.Spec.Metadata,
			Topologies:  instance.Spec.Topologies,
			Pipelines:   instance.Spec.Pipelines,
		},
	}

	return ret, nil
}

func K8SSolutionToAPISolutionState(solution solution_v1.Solution) (apimodel.SolutionState, error) {
	ret := apimodel.SolutionState{
		ObjectMeta: apimodel.ObjectMeta{
			Name:        solution.ObjectMeta.Name,
			Namespace:   solution.ObjectMeta.Namespace,
			Labels:      solution.ObjectMeta.Labels,
			Annotations: solution.ObjectMeta.Annotations,
		},
		Spec: &apimodel.SolutionSpec{
			DisplayName: solution.Spec.DisplayName,
			Metadata:    solution.Spec.Metadata,
		},
	}

	var err error
	ret.Spec.Components = make([]apimodel.ComponentSpec, len(solution.Spec.Components))
	for i, t := range solution.Spec.Components {
		ret.Spec.Components[i], err = K8SComponentSpecToAPIComponentSpec(t)
		if err != nil {
			return apimodel.SolutionState{}, err
		}
	}
	return ret, nil

}

func matchString(src string, target string) bool {
	if strings.Contains(src, "*") || strings.Contains(src, "%") {
		p := strings.ReplaceAll(src, "*", ".*")
		p = strings.ReplaceAll(p, "%", ".")
		re := regexp.MustCompile(p)
		return re.MatchString(target)
	} else {
		return src == target
	}
}

func MatchTargets(instance solution_v1.Instance, targets fabric_v1.TargetList) []fabric_v1.Target {
	ret := make(map[string]fabric_v1.Target)
	if instance.Spec.Target.Name != "" {
		for _, t := range targets.Items {
			if matchString(instance.Spec.Target.Name, t.ObjectMeta.Name) {
				ret[t.ObjectMeta.Name] = t
			}
		}
	}
	if len(instance.Spec.Target.Selector) > 0 {
		for _, t := range targets.Items {
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
	slice := make([]fabric_v1.Target, 0, len(ret))
	for _, v := range ret {
		slice = append(slice, v)
	}
	return slice
}

func CreateSymphonyDeploymentFromTarget(target fabric_v1.Target, namespace string) (apimodel.DeploymentSpec, error) {
	targetState, err := K8STargetToAPITargetState(target)
	if err != nil {
		return apimodel.DeploymentSpec{}, err
	}

	var ret apimodel.DeploymentSpec
	ret, err = api_utils.CreateSymphonyDeploymentFromTarget(targetState, namespace)
	ret.Hash = HashObjects(DeploymentResources{
		TargetCandidates: []fabric_v1.Target{target},
	})

	ret.Generation = strconv.Itoa(int(target.ObjectMeta.Generation))

	return ret, err
}

func CreateSymphonyDeployment(ctx context.Context, instance solution_v1.Instance, solution solution_v1.Solution, targets []fabric_v1.Target, objectNamespace string) (apimodel.DeploymentSpec, error) {
	instanceState, err := K8SInstanceToAPIInstanceState(instance)
	if err != nil {
		return apimodel.DeploymentSpec{}, err
	}

	solutionState, err := K8SSolutionToAPISolutionState(solution)
	if err != nil {
		return apimodel.DeploymentSpec{}, err
	}

	targetStates := make([]apimodel.TargetState, len(targets))
	for i, t := range targets {
		targetStates[i], err = K8STargetToAPITargetState(t)
		if err != nil {
			return apimodel.DeploymentSpec{}, err
		}
	}

	var ret apimodel.DeploymentSpec
	ret, err = api_utils.CreateSymphonyDeployment(instanceState, solutionState, targetStates, nil, objectNamespace)
	ret.Hash = HashObjects(DeploymentResources{
		Instance:         instance,
		Solution:         solution,
		TargetCandidates: targets,
	})

	ret.Generation = strconv.Itoa(int(instance.ObjectMeta.Generation))

	return ret, err
}

func NeedWatchInstance(instance solution_v1.Instance) bool {
	var interval time.Duration = 30
	if instance.Spec.ReconciliationPolicy != nil && instance.Spec.ReconciliationPolicy.Interval != nil {
		parsedInterval, err := time.ParseDuration(*instance.Spec.ReconciliationPolicy.Interval)
		if err != nil {
			parsedInterval = 30
		}
		interval = parsedInterval
	}

	if instance.Spec.ReconciliationPolicy != nil && instance.Spec.ReconciliationPolicy.State.IsInActive() || interval == 0 {
		return false
	}

	return true
}

func ReplaceLastSeperator(name string, seperatorBefore string, seperatorAfter string) string {
	i := strings.LastIndex(name, seperatorBefore)
	if i == -1 {
		return name
	}
	return name[:i] + seperatorAfter + name[i+1:]
}
