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
	federation_v1 "gopls-workspace/apis/federation/v1"
	k8smodel "gopls-workspace/apis/model/v1"
	solutionversion_v1 "gopls-workspace/apis/solution/v1"
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
		Instance         solutionversion_v1.Instance
		SolutionVersion         solutionversion_v1.SolutionVersion
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
	if sidecar.Properties.Raw != nil {
		sidecarSpec.Properties = make(map[string]interface{})
		err = json.Unmarshal(sidecar.Properties.Raw, &sidecarSpec.Properties)
	}
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
	if component.Properties.Raw != nil {
		componentSpec.Properties = make(map[string]interface{})
		err = json.Unmarshal(component.Properties.Raw, &componentSpec.Properties)
	}

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
			SolutionVersionScope: target.Spec.SolutionVersionScope,
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

func K8SInstanceToAPIInstanceState(instance solutionversion_v1.Instance) (apimodel.InstanceState, error) {
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
			SolutionVersion:    instance.Spec.SolutionVersion,
			Target:      instance.Spec.Target,
			Parameters:  instance.Spec.Parameters,
			Metadata:    instance.Spec.Metadata,
			Topologies:  instance.Spec.Topologies,
			Pipelines:   instance.Spec.Pipelines,
		},
	}

	return ret, nil
}

func K8SCatalogVersionToAPICatalogVersionState(catalogversion federation_v1.CatalogVersion) (apimodel.CatalogVersionState, error) {
	ret := apimodel.CatalogVersionState{
		ObjectMeta: apimodel.ObjectMeta{
			Name:        catalogversion.ObjectMeta.Name,
			Namespace:   catalogversion.ObjectMeta.Namespace,
			Labels:      catalogversion.ObjectMeta.Labels,
			Annotations: catalogversion.ObjectMeta.Annotations,
		},
		Spec: &apimodel.CatalogVersionSpec{
			CatalogType:  catalogversion.Spec.CatalogType,
			Metadata:     catalogversion.Spec.Metadata,
			ParentName:   catalogversion.Spec.ParentName,
			ObjectRef:    catalogversion.Spec.ObjectRef,
			Version:      catalogversion.Spec.Version,
			RootResource: catalogversion.Spec.RootResource,
		},
	}

	if catalogversion.Spec.Properties.Raw != nil {
		ret.Spec.Properties = make(map[string]interface{})
		err := json.Unmarshal(catalogversion.Spec.Properties.Raw, &catalogversion.Spec.Properties)
		if err != nil {
			return apimodel.CatalogVersionState{}, err
		}
	}

	return ret, nil
}

func K8SSolutionVersionToAPISolutionVersionState(solutionversion solutionversion_v1.SolutionVersion) (apimodel.SolutionVersionState, error) {
	ret := apimodel.SolutionVersionState{
		ObjectMeta: apimodel.ObjectMeta{
			Name:        solutionversion.ObjectMeta.Name,
			Namespace:   solutionversion.ObjectMeta.Namespace,
			Labels:      solutionversion.ObjectMeta.Labels,
			Annotations: solutionversion.ObjectMeta.Annotations,
		},
		Spec: &apimodel.SolutionVersionSpec{
			DisplayName: solutionversion.Spec.DisplayName,
			Metadata:    solutionversion.Spec.Metadata,
		},
	}

	var err error
	ret.Spec.Components = make([]apimodel.ComponentSpec, len(solutionversion.Spec.Components))
	for i, t := range solutionversion.Spec.Components {
		ret.Spec.Components[i], err = K8SComponentSpecToAPIComponentSpec(t)
		if err != nil {
			return apimodel.SolutionVersionState{}, err
		}
	}
	return ret, nil

}

func ContainsString(slice []string, target string) bool {
	if slice == nil {
		return false
	}
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
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

func MatchTargets(instance solutionversion_v1.Instance, targets fabric_v1.TargetList) []fabric_v1.Target {
	ret := make(map[string]fabric_v1.Target)
	if instance.Spec.Target.Name != "" {
		for _, t := range targets.Items {
			if matchString(instance.Spec.Target.Name, t.ObjectMeta.Name) {
				ret[t.ObjectMeta.Name] = t
			} else {
				// azure case
				if t.Annotations[constants.AzureResourceIdKey] != "" && matchString(strings.ToLower(instance.Spec.Target.Name), strings.ToLower(t.Annotations[constants.AzureResourceIdKey])) {
					ret[t.ObjectMeta.Name] = t
				}
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

func CreateSymphonyDeploymentFromTarget(ctx context.Context, target fabric_v1.Target, namespace string) (apimodel.DeploymentSpec, error) {
	targetState, err := K8STargetToAPITargetState(target)
	if err != nil {
		return apimodel.DeploymentSpec{}, err
	}

	var ret apimodel.DeploymentSpec
	ret, err = api_utils.CreateSymphonyDeploymentFromTarget(ctx, targetState, namespace)
	ret.Hash = HashObjects(DeploymentResources{
		TargetCandidates: []fabric_v1.Target{target},
	})

	ret.Generation = strconv.Itoa(int(target.ObjectMeta.Generation))
	ret.IsDryRun = target.Spec.IsDryRun

	return ret, err
}

func CreateSymphonyDeployment(ctx context.Context, instance solutionversion_v1.Instance, solutionversion solutionversion_v1.SolutionVersion, targets []fabric_v1.Target, objectNamespace string) (apimodel.DeploymentSpec, error) {
	instanceState, err := K8SInstanceToAPIInstanceState(instance)
	if err != nil {
		return apimodel.DeploymentSpec{}, err
	}

	solutionversionState, err := K8SSolutionVersionToAPISolutionVersionState(solutionversion)
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
	ret, err = api_utils.CreateSymphonyDeployment(ctx, instanceState, solutionversionState, targetStates, nil, objectNamespace)
	ret.Hash = HashObjects(DeploymentResources{
		Instance:         instance,
		SolutionVersion:         solutionversion,
		TargetCandidates: targets,
	})

	ret.Generation = strconv.Itoa(int(instance.ObjectMeta.Generation))
	ret.IsDryRun = instance.Spec.IsDryRun
	ret.IsInActive = instance.Spec.ActiveState == apimodel.ActiveState_Inactive

	return ret, err
}

func NeedWatchInstance(instance solutionversion_v1.Instance) bool {
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
