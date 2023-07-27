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

package utils

import (
	"encoding/json"
	symphonyv1 "gopls-workspace/apis/symphony.microsoft.com/v1"
	"regexp"
	"strings"

	symphony "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
	api_utils "github.com/azure/symphony/api/pkg/apis/v1alpha1/utils" //TODO: Eventually, most logic here should be moved into this
)

const (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
)

func MatchTargets(instance symphonyv1.Instance, targets symphonyv1.TargetList) []symphonyv1.Target {
	ret := make(map[string]symphonyv1.Target)
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
	slice := make([]symphonyv1.Target, 0, len(ret))
	for _, v := range ret {
		slice = append(slice, v)
	}
	return slice
}

func CreateSymphonyDeploymentFromTarget(target symphonyv1.Target) (symphony.DeploymentSpec, error) {
	ret := symphony.DeploymentSpec{}
	// create solution
	scope := target.Spec.Scope
	if scope == "" {
		scope = "default"
	}

	solution := symphony.SolutionSpec{
		DisplayName: "target-runtime",
		Scope:       scope,
		Components:  make([]symphony.ComponentSpec, 0),
		Metadata:    make(map[string]string, 0),
	}

	for k, v := range target.Spec.Metadata {
		solution.Metadata[k] = v
	}

	for _, component := range target.Spec.Components {
		var c symphony.ComponentSpec
		data, _ := json.Marshal(component)
		err := json.Unmarshal(data, &c)
		if err != nil {
			return ret, err
		}
		solution.Components = append(solution.Components, c)
	}

	// create targets
	targets := make(map[string]symphony.TargetSpec)
	var t symphony.TargetSpec
	data, _ := json.Marshal(target.Spec)
	err := json.Unmarshal(data, &t)
	if err != nil {
		return ret, err
	}

	targets[target.ObjectMeta.Name] = t

	// create instance
	instance := symphony.InstanceSpec{
		Name:        "target-runtime",
		DisplayName: "target-runtime-" + target.ObjectMeta.Name,
		Scope:       scope,
		Solution:    "target-runtime",
		Target: symphony.TargetSelector{
			Name: target.ObjectMeta.Name,
		},
	}

	ret.Solution = solution
	ret.Instance = instance
	ret.Targets = targets
	ret.SolutionName = "target-runtime"
	assignments, err := api_utils.AssignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)

	for k, v := range assignments {
		ret.Assignments[k] = v
	}

	return ret, nil
}

func CreateSymphonyDeployment(instance symphonyv1.Instance, solution symphonyv1.Solution, targets []symphonyv1.Target) (symphony.DeploymentSpec, error) {
	ret := symphony.DeploymentSpec{}
	// convert instance
	var sInstance symphony.InstanceSpec
	data, _ := json.Marshal(instance.Spec)
	err := json.Unmarshal(data, &sInstance)
	if err != nil {
		return ret, err
	}

	sInstance.Name = instance.ObjectMeta.Name
	sInstance.Scope = instance.Spec.Scope
	if sInstance.Scope == "" {
		sInstance.Scope = "default"
	}

	// convert solution
	var sSolution symphony.SolutionSpec
	data, _ = json.Marshal(solution.Spec)
	err = json.Unmarshal(data, &sSolution)
	if err != nil {
		return ret, err
	}

	sSolution.DisplayName = solution.ObjectMeta.Name
	sSolution.Scope = solution.ObjectMeta.Namespace

	// convert targets
	sTargets := make(map[string]symphony.TargetSpec)
	for _, t := range targets {
		var target symphony.TargetSpec
		data, _ = json.Marshal(t.Spec)
		err = json.Unmarshal(data, &target)
		if err != nil {
			return ret, err
		}
		sTargets[t.ObjectMeta.Name] = target
	}

	//TODO: handle devices
	ret.Solution = sSolution
	ret.Targets = sTargets
	ret.Instance = sInstance
	ret.SolutionName = solution.ObjectMeta.Name

	assignments, err := api_utils.AssignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}
	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}
	return ret, nil
}

// func Get(scope string, name string, targets map[string]symphony.TargetSpec) ([]symphony.ComponentSpec, error) {
// 	payload, _ := json.Marshal(targets)
// 	ret, err := callRestAPI("solution/instances?scope="+scope+"&name="+name, "GET", payload) //TODO: URL enchode params
// 	if err != nil {
// 		return nil, err
// 	}
// 	var components []symphony.ComponentSpec
// 	err = json.Unmarshal(ret, &components)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return components, nil
// }

// func NeedsUpdate(deployment symphony.DeploymentSpec, components []symphony.ComponentSpec) bool {
// 	return !symphony.SlicesEqual(deployment.Solution.Components, components)
// }
// func NeedsRemove(deployment symphony.DeploymentSpec, components []symphony.ComponentSpec) bool {
// 	return symphony.SlicesCover(deployment.Solution.Components, components)
// }

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

type ParsedAPIError struct {
	Code string               `json:"code"`
	Spec symphony.SummarySpec `json:"spec"`
}
