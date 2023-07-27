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
	"io/ioutil"
	"net/http"

	"github.com/azure/symphony/api/pkg/apis/v1alpha1/model"
)

const (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type authResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

// SummarySpecError represents an error that includes a SummarySpec in its message
// field.
type SummarySpecError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *SummarySpecError) Error() string {
	return fmt.Sprintf(
		"failed to invoke Symphony API: [%s] - %s",
		e.Code,
		e.Message,
	)
}

func GetInstances(baseUrl string, user string, password string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "instances", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func GetInstance(baseUrl string, instance string, user string, password string) (model.InstanceState, error) {
	ret := model.InstanceState{}
	token, err := auth(baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "instances/"+instance, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func CreateInstance(baseUrl string, instance string, user string, password string, payload []byte) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "instances/"+instance, "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func DeleteInstance(baseUrl string, instance string, user string, password string) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "instances/"+instance+"?direct=true", "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTarget(baseUrl string, target string, user string, password string) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "targets/registry/"+target+"?direct=true", "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func GetSolutions(baseUrl string, user string, password string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "solutions", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func GetSolution(baseUrl string, solution string, user string, password string) (model.SolutionState, error) {
	ret := model.SolutionState{}
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "solutions/"+solution, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func CreateSolution(baseUrl string, solution string, user string, password string, payload []byte) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "solutions/"+solution, "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func DeleteSolution(baseUrl string, solution string, user string, password string) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "solutions/"+solution, "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func GetTarget(baseUrl string, target string, user string, password string) (model.TargetState, error) {
	ret := model.TargetState{}
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "targets/registry/"+target, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func GetTargets(baseUrl string, user string, password string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(baseUrl, "targets/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func CreateTarget(baseUrl string, target string, user string, password string, payload []byte) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(baseUrl, "targets/registry/"+target, "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func MatchTargets(instance model.InstanceState, targets []model.TargetState) []model.TargetState {
	ret := make(map[string]model.TargetState)
	if instance.Spec.Target.Name != "" {
		for _, t := range targets {
			if matchString(instance.Spec.Target.Name, t.Id) {
				ret[t.Id] = t
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
				ret[t.Id] = t
			}
		}
	}

	slice := make([]model.TargetState, 0, len(ret))
	for _, v := range ret {
		slice = append(slice, v)
	}

	return slice
}

func CreateSymphonyDeploymentFromTarget(target model.TargetState) (model.DeploymentSpec, error) {
	key := fmt.Sprintf("%s-%s", "target-runtime", target.Id)
	scope := target.Spec.Scope
	if scope == "" {
		scope = "default"
	}
	ret := model.DeploymentSpec{}
	solution := model.SolutionSpec{
		DisplayName: key,
		Scope:       scope,
		Components:  make([]model.ComponentSpec, 0),
		Metadata:    make(map[string]string, 0),
	}
	for k, v := range target.Spec.Metadata {
		solution.Metadata[k] = v
	}

	for _, component := range target.Spec.Components {
		var c model.ComponentSpec
		data, _ := json.Marshal(component)
		err := json.Unmarshal(data, &c)

		if err != nil {
			return ret, err
		}
		solution.Components = append(solution.Components, c)
	}

	targets := make(map[string]model.TargetSpec)
	var t model.TargetSpec
	data, _ := json.Marshal(target.Spec)
	err := json.Unmarshal(data, &t)
	if err != nil {
		return ret, err
	}

	targets[target.Id] = t

	instance := model.InstanceSpec{
		Name:        key,
		DisplayName: key,
		Scope:       scope,
		Solution:    key,
		Target: model.TargetSelector{
			Name: target.Id,
		},
	}

	ret.Solution = solution
	ret.Instance = instance
	ret.Targets = targets
	ret.SolutionName = key
	assignments, err := AssignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}

	return ret, nil
}

func CreateSymphonyDeployment(instance model.InstanceState, solution model.SolutionState, targets []model.TargetState, devices []model.DeviceState) (model.DeploymentSpec, error) {
	ret := model.DeploymentSpec{}
	ret.Generation = instance.Spec.Generation
	// convert instance
	sInstance := instance.Spec

	sInstance.Name = instance.Id
	sInstance.Scope = instance.Spec.Scope
	if sInstance.Scope == "" {
		sInstance.Scope = "default"
	}

	// convert solution
	sSolution := solution.Spec

	sSolution.DisplayName = solution.Spec.DisplayName
	sSolution.Scope = solution.Spec.Scope

	// convert targets
	sTargets := make(map[string]model.TargetSpec)
	for _, t := range targets {
		sTargets[t.Id] = *t.Spec
	}

	//TODO: handle devices
	ret.Solution = *sSolution
	ret.Targets = sTargets
	ret.Instance = *sInstance
	ret.SolutionName = solution.Id

	assignments, err := AssignComponentsToTargets(ret.Solution.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}

	return ret, nil
}

func AssignComponentsToTargets(components []model.ComponentSpec, targets map[string]model.TargetSpec) (map[string]string, error) {
	//TODO: evaluate constraints
	ret := make(map[string]string)
	for key, target := range targets {
		ret[key] = ""
		for _, component := range components {
			match := true
			if component.Constraints != "" {
				parser := NewParser(component.Constraints)
				val, err := parser.Eval(EvaluationContext{Properties: target.Properties})
				if err != nil {
					return ret, err
				}
				match = val == "true"
			}
			if match {
				ret[key] += "{" + component.Name + "}"
			}
		}
	}

	return ret, nil
}
func GetSummary(baseUrl string, user string, password string, id string) (model.SummaryResult, error) {
	result := model.SummaryResult{}
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return result, err
	}
	ret, err := callRestAPI(baseUrl, "solution/queue?instance="+id, "GET", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
	if err != nil {
		return result, err
	}
	if ret != nil {
		err = json.Unmarshal(ret, &result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
func QueueJob(baseUrl string, user string, password string, id string, isDelete bool, isTarget bool) error {
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "solution/queue?instance=" + id
	if isDelete {
		path += "&delete=true"
	}
	if isTarget {
		path += "&target=true"
	}
	_, err = callRestAPI(baseUrl, path, "POST", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
	if err != nil {
		return err
	}
	return nil
}
func Reconcile(baseUrl string, user string, password string, deployment model.DeploymentSpec, isDelete bool) (model.SummarySpec, error) {
	summary := model.SummarySpec{}
	payload, _ := json.Marshal(deployment)

	path := "solution/reconcile"
	if isDelete {
		path = "solution/reconcile?delete=true"
	}
	token, err := auth(baseUrl, user, password)
	if err != nil {
		return summary, err
	}
	ret, err := callRestAPI(baseUrl, path, "POST", payload, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
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
func auth(baseUrl string, user string, password string) (string, error) {
	request := authRequest{Username: user, Password: password}
	requestData, _ := json.Marshal(request)
	ret, err := callRestAPI(baseUrl, "users/auth", "POST", requestData, "")
	if err != nil {
		return "", err
	}

	var response authResponse
	err = json.Unmarshal(ret, &response)
	if err != nil {
		return "", err
	}

	return response.AccessToken, nil
}
func callRestAPI(baseUrl string, route string, method string, payload []byte, token string) ([]byte, error) {
	client := &http.Client{}
	rUrl := baseUrl + route
	var req *http.Request
	var err error
	if payload != nil {
		req, err = http.NewRequest(method, rUrl, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, rUrl, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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
		object := &SummarySpecError{
			Code:    fmt.Sprintf("Symphony API: [%d]", resp.StatusCode),
			Message: string(bodyBytes),
		}
		return nil, object
	}

	return bodyBytes, nil
}
