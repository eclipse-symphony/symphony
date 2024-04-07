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
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
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

func K8SBindingSpecToAPIBindingSpec(binding k8smodel.BindingSpec) (model.BindingSpec, error) {
	bindingSpec := model.BindingSpec{}
	data, _ := json.Marshal(binding)
	err := json.Unmarshal(data, &bindingSpec)
	return bindingSpec, err
}

func K8STopologySpecToAPITopologySpec(topology k8smodel.TopologySpec) (model.TopologySpec, error) {
	topologySpec := model.TopologySpec{}
	data, _ := json.Marshal(topology)
	err := json.Unmarshal(data, &topologySpec)
	return topologySpec, err
}

func K8SPipelineSpecToAPIPipelineSpec(pipeline k8smodel.PipelineSpec) (model.PipelineSpec, error) {
	pipelineSpec := model.PipelineSpec{}
	data, _ := json.Marshal(pipeline)
	err := json.Unmarshal(data, &pipeline)
	return pipelineSpec, err
}

func K8SSidecarSpecToAPISidecarSpec(sidecar k8smodel.SidecarSpec) (model.SidecarSpec, error) {
	sidecarSpec := model.SidecarSpec{}
	data, _ := json.Marshal(sidecar)
	err := json.Unmarshal(data, &sidecarSpec)
	return sidecarSpec, err
}

func K8SComponentSpecToAPIComponentSpec(component k8smodel.ComponentSpec) (model.ComponentSpec, error) {
	componentSpec := model.ComponentSpec{}
	data, _ := json.Marshal(component)
	err := json.Unmarshal(data, &componentSpec)
	return componentSpec, err
}

func K8STargetToAPITargetState(target fabric_v1.Target) (model.TargetState, error) {
	ret := model.TargetState{
		ObjectMeta: model.ObjectMeta{
			Name:        target.ObjectMeta.Name,
			Namespace:   target.ObjectMeta.Namespace,
			Labels:      target.ObjectMeta.Labels,
			Annotations: target.ObjectMeta.Annotations,
		},
		Spec: &model.TargetSpec{
			Properties: target.Spec.Properties,
			Generation: target.Spec.Generation,
		},
	}

	var err error
	ret.Spec.Components = make([]model.ComponentSpec, len(target.Spec.Components))
	for i, c := range target.Spec.Components {
		ret.Spec.Components[i], err = K8SComponentSpecToAPIComponentSpec(c)
		if err != nil {
			return model.TargetState{}, err
		}
	}

	return ret, nil
}

func K8SInstanceToAPIInstanceState(instance solution_v1.Instance) (model.InstanceState, error) {
	ret := model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:        instance.ObjectMeta.Name,
			Namespace:   instance.ObjectMeta.Namespace,
			Labels:      instance.ObjectMeta.Labels,
			Annotations: instance.ObjectMeta.Annotations,
		},
		Spec: &model.InstanceSpec{
			Scope:       instance.Spec.Scope,
			Name:        instance.Spec.Name,
			DisplayName: instance.Spec.DisplayName,
			Solution:    instance.Spec.Solution,
			Target: model.TargetSelector{
				Name:     instance.Spec.Target.Name,
				Selector: instance.Spec.Target.Selector,
			},
			Parameters: instance.Spec.Parameters,
			Metadata:   instance.Spec.Metadata,
			Generation: instance.Spec.Generation,
			Version:    instance.Spec.Version,
		},
	}

	var err error
	ret.Spec.Topologies = make([]model.TopologySpec, len(instance.Spec.Topologies))
	for i, t := range instance.Spec.Topologies {
		ret.Spec.Topologies[i], err = K8STopologySpecToAPITopologySpec(t)
		if err != nil {
			return model.InstanceState{}, err
		}
	}

	ret.Spec.Pipelines = make([]model.PipelineSpec, len(instance.Spec.Pipelines))
	for i, p := range instance.Spec.Pipelines {
		ret.Spec.Pipelines[i], err = K8SPipelineSpecToAPIPipelineSpec(p)
		if err != nil {
			return model.InstanceState{}, err
		}
	}

	return ret, nil
}

func K8SSolutionToAPISolutionState(solution solution_v1.Solution) (model.SolutionState, error) {
	ret := model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:        solution.ObjectMeta.Name,
			Namespace:   solution.ObjectMeta.Namespace,
			Labels:      solution.ObjectMeta.Labels,
			Annotations: solution.ObjectMeta.Annotations,
		},
		Spec: &model.SolutionSpec{
			DisplayName: solution.Spec.DisplayName,
			Metadata:    solution.Spec.Metadata,
		},
	}

	var err error
	ret.Spec.Components = make([]model.ComponentSpec, len(solution.Spec.Components))
	for i, t := range solution.Spec.Components {
		ret.Spec.Components[i], err = K8SComponentSpecToAPIComponentSpec(t)
		if err != nil {
			return model.SolutionState{}, err
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

	targetStates := make([]model.TargetState, len(targets))
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

	return ret, err
}
