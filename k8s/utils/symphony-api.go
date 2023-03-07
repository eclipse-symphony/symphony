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
	"bytes"
	"encoding/json"
	"fmt"
	fabricv1 "gopls-workspace/apis/fabric/v1"
	solutionv1 "gopls-workspace/apis/solution/v1"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	symphony "github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

const (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
)

func MatchTargets(instance solutionv1.Instance, targets fabricv1.TargetList) []fabricv1.Target {
	ret := make(map[string]fabricv1.Target)
	if instance.Spec.Stages[0].Target.Name != "" {
		for _, t := range targets.Items {

			if matchString(instance.Spec.Stages[0].Target.Name, t.ObjectMeta.Name) {
				ret[t.ObjectMeta.Name] = t
			}
		}
	}
	if len(instance.Spec.Stages[0].Target.Selector) > 0 {
		for _, t := range targets.Items {
			fullMatch := true
			for k, v := range instance.Spec.Stages[0].Target.Selector {
				if tv, ok := t.Spec.Properties[k]; !ok || !matchString(v, tv) {
					fullMatch = false
				}
			}
			if fullMatch {
				ret[t.ObjectMeta.Name] = t
			}
		}
	}
	slice := make([]fabricv1.Target, 0, len(ret))
	for _, v := range ret {
		slice = append(slice, v)
	}
	return slice
}

func CreateSymphonyDeploymentFromTarget(target fabricv1.Target) (symphony.DeploymentSpec, error) {
	ret := symphony.DeploymentSpec{}
	// create solution
	solution := symphony.SolutionSpec{
		DisplayName: "target-runtime",
		Scope:       "default",
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
		Scope:       "default",
		Stages: []symphony.StageSpec{
			{
				Solution: "target-runtime",
				Target: symphony.TargetRefSpec{
					Name: target.ObjectMeta.Name,
				},
			},
		},
	}

	ret.Solution = solution
	ret.Instance = instance
	ret.Targets = targets
	ret.SolutionName = "target-runtime"
	assignments, err := assignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}
	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}
	return ret, nil
}

func CreateSymphonyDeployment(instance solutionv1.Instance, solution solutionv1.Solution, targets []fabricv1.Target, devices []fabricv1.Device) (symphony.DeploymentSpec, error) {

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

	assignments, err := assignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}
	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}
	return ret, nil
}

func assignComponentsToTargets(components []symphony.ComponentSpec, targets map[string]symphony.TargetSpec) (map[string]string, error) {
	//TODO: evaluate constraints
	ret := make(map[string]string)
	for key, target := range targets {
		ret[key] = ""
		for _, component := range components {
			match := true
			for _, s := range component.Constraints {
				if !s.Match(target.Properties) {
					match = false
				}
			}
			if match {
				ret[key] += "{" + component.Name + "}"
			}
		}
	}
	return ret, nil
}
func Deploy(deployment symphony.DeploymentSpec) (symphony.SummarySpec, error) {
	summary := symphony.SummarySpec{}
	payload, _ := json.Marshal(deployment)
	ret, err := callRestAPI("solution/instances", "POST", payload)
	if err != nil {
		return summary, err
	}
	if ret != nil {
		err = json.Unmarshal(ret, &summary)
		if err != nil {
			return summary, err
		}
	}
	return summary, nil
}

func Remove(deployment symphony.DeploymentSpec) (symphony.SummarySpec, error) {
	summary := symphony.SummarySpec{}
	payload, _ := json.Marshal(deployment)
	ret, err := callRestAPI("solution/instances", "DELETE", payload)
	if err != nil {
		return summary, err
	}
	if ret != nil {
		err = json.Unmarshal(ret, &summary)
		if err != nil {
			return summary, err
		}
	}
	return summary, nil
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

func callRestAPI(route string, method string, payload []byte) ([]byte, error) {
	client := &http.Client{}
	rUrl := SymphonyAPIAddressBase + route
	req, err := http.NewRequest(method, rUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		if resp.StatusCode == 404 { // API service is already gone
			return nil, nil
		}
		return nil, fmt.Errorf("failed to invoke Symphony API: [%d] - %v", resp.StatusCode, string(bodyBytes))
	}
	return bodyBytes, nil
}
