/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

const (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type authResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

// We shouldn't use specific error types
// SummarySpecError represents an error that includes a SummarySpec in its message
// field.
// type SummarySpecError struct {
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// }

// func (e *SummarySpecError) Error() string {
// 	return fmt.Sprintf(
// 		"failed to invoke Symphony API: [%s] - %s",
// 		e.Code,
// 		e.Message,
// 	)
// }

var log = logger.NewLogger("coa.runtime")

func GetInstancesForAllScope(context context.Context, baseUrl string, user string, password string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "instances", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetInstances(context context.Context, baseUrl string, user string, password string, scope string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "instances?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func GetSites(context context.Context, baseUrl string, user string, password string) ([]model.SiteState, error) {
	ret := make([]model.SiteState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "federation/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}
func SyncActivationStatus(context context.Context, baseUrl string, user string, password string, status model.ActivationStatus) error {
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return err
	}
	jData, _ := json.Marshal(status)
	_, err = callRestAPI(context, baseUrl, "federation/sync", "POST", jData, token)
	if err != nil {
		return err
	}

	return nil
}
func GetCatalogs(context context.Context, baseUrl string, user string, password string) ([]model.CatalogState, error) {
	ret := make([]model.CatalogState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "catalogs/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}
func GetCatalog(context context.Context, baseUrl string, catalog string, user string, password string) (model.CatalogState, error) {
	ret := model.CatalogState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	catalogName := catalog
	if strings.HasPrefix(catalogName, "<") && strings.HasSuffix(catalogName, ">") {
		catalogName = catalogName[1 : len(catalogName)-1]
	}

	response, err := callRestAPI(context, baseUrl, "catalogs/registry/"+catalogName, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func GetCampaign(context context.Context, baseUrl string, campaign string, user string, password string) (model.CampaignState, error) {
	ret := model.CampaignState{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "campaigns/"+campaign, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func PublishActivationEvent(context context.Context, baseUrl string, user string, password string, event v1alpha2.ActivationData) error {
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return err
	}
	jData, _ := json.Marshal(event)
	_, err = callRestAPI(context, baseUrl, "jobs", "POST", jData, token)
	if err != nil {
		return err
	}

	return nil
}
func GetABatchForSite(context context.Context, baseUrl string, site string, user string, password string) (model.SyncPackage, error) {
	ret := model.SyncPackage{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "federation/sync/"+site+"?count=10", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func GetActivation(context context.Context, baseUrl string, activation string, user string, password string) (model.ActivationState, error) {
	ret := model.ActivationState{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "activations/registry/"+activation, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func ReportActivationStatus(context context.Context, baseUrl string, name string, user string, password string, activation model.ActivationStatus) error {
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return err
	}

	jData, _ := json.Marshal(activation)
	_, err = callRestAPI(context, baseUrl, "activations/status/"+name, "POST", jData, token)
	if err != nil {
		return err
	}
	return nil
}
func GetInstance(context context.Context, baseUrl string, instance string, user string, password string, scope string) (model.InstanceState, error) {
	ret := model.InstanceState{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	path := "instances/" + instance
	path = path + "?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func UpsertCatalog(context context.Context, baseUrl string, catalog string, user string, password string, payload []byte) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(context, baseUrl, "catalogs/registry/"+catalog, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func CreateInstance(context context.Context, baseUrl string, instance string, user string, password string, payload []byte, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	path := "instances/" + instance
	path = path + "?scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteCatalog(context context.Context, baseUrl string, catalog string, user string, password string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(context, baseUrl, "catalogs/registry/"+catalog, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteInstance(context context.Context, baseUrl string, instance string, user string, password string, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "instances/" + instance
	path = path + "?direct=true&scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTarget(context context.Context, baseUrl string, target string, user string, password string, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "targets/registry/" + target
	path = path + "?direct=true&scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func GetSolutionsForAllScope(context context.Context, baseUrl string, user string, password string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "solutions", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetSolutions(context context.Context, baseUrl string, user string, password string, scope string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "solution" + "?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetSolution(context context.Context, baseUrl string, solution string, user string, password string, scope string) (model.SolutionState, error) {
	ret := model.SolutionState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "solutions/" + solution
	path = path + "?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func UpsertTarget(context context.Context, baseUrl string, solution string, user string, password string, payload []byte, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "targets/registry/" + solution
	path = path + "?scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func UpsertSolution(context context.Context, baseUrl string, solution string, user string, password string, payload []byte, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "solutions/" + solution
	path = path + "?scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteSolution(context context.Context, baseUrl string, solution string, user string, password string, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "solutions/" + solution
	path = path + "?scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func GetTarget(context context.Context, baseUrl string, target string, user string, password string, scope string) (model.TargetState, error) {
	ret := model.TargetState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "targets/registry/" + target
	path = path + "?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetTargetsForAllScope(context context.Context, baseUrl string, user string, password string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	response, err := callRestAPI(context, baseUrl, "targets/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func GetTargets(context context.Context, baseUrl string, user string, password string, scope string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "targets/registry"
	path = path + "?scope=" + scope
	response, err := callRestAPI(context, baseUrl, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func SendVisualizationPacket(context context.Context, baseUrl string, user string, password string, payload []byte) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	_, err = callRestAPI(context, baseUrl, "visualization", "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSite(context context.Context, baseUrl string, site string, user string, password string, payload []byte) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	_, err = callRestAPI(context, baseUrl, "federation/status/"+site, "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func CreateTarget(context context.Context, baseUrl string, target string, user string, password string, payload []byte, scope string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "targets/registry/" + target
	path = path + "?scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	log.Info(">>>>>CreateTarget Succeed: " + target + " " + scope)
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
				val, err := parser.Eval(utils.EvaluationContext{Properties: target.Properties})
				if err != nil {
					return ret, err
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
func GetSummary(context context.Context, baseUrl string, user string, password string, id string, scope string) (model.SummaryResult, error) {
	result := model.SummaryResult{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return result, err
	}
	path := "solution/queue"
	path = path + "?instance=" + id + "&scope=" + scope
	ret, err := callRestAPI(context, baseUrl, path, "GET", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
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
func CatalogHook(context context.Context, baseUrl string, user string, password string, payload []byte) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "federation/k8shook?objectType=catalog"
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func QueueJob(context context.Context, baseUrl string, user string, password string, id string, scope string, isDelete bool, isTarget bool) error {
	token, err := auth(context, baseUrl, user, password)
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
	path = path + "&scope=" + scope
	_, err = callRestAPI(context, baseUrl, path, "POST", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
	if err != nil {
		return err
	}
	return nil
}
func Reconcile(context context.Context, baseUrl string, user string, password string, deployment model.DeploymentSpec, scope string, isDelete bool) (model.SummarySpec, error) {
	summary := model.SummarySpec{}
	payload, _ := json.Marshal(deployment)

	path := "solution/reconcile" + "?scope=" + scope
	if isDelete {
		path = path + "&delete=true"
	}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return summary, err
	}
	ret, err := callRestAPI(context, baseUrl, path, "POST", payload, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
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
func auth(context context.Context, baseUrl string, user string, password string) (string, error) {
	request := authRequest{Username: user, Password: password}
	requestData, _ := json.Marshal(request)
	ret, err := callRestAPI(context, baseUrl, "users/auth", "POST", requestData, "")
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
func callRestAPI(context context.Context, baseUrl string, route string, method string, payload []byte, token string) ([]byte, error) {
	context, span := observability.StartSpan("Symphony-API-Client", context, &map[string]string{
		"method":      "callRestAPI",
		"http.method": method,
		"http.url":    baseUrl + route,
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	log.Infof("Calling Symphony API: %s %s, spanId: %s, traceId: %s", method, baseUrl+route, span.SpanContext().SpanID().String(), span.SpanContext().TraceID().String())

	client := &http.Client{}
	rUrl := baseUrl + route
	var req *http.Request
	if payload != nil {
		req, err = http.NewRequestWithContext(context, method, rUrl, bytes.NewBuffer(payload))
		observ_utils.PropagateSpanContextToHttpRequestHeader(req)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(context, method, rUrl, nil)
		observ_utils.PropagateSpanContextToHttpRequestHeader(req)
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
		// TODO: Can we remove the following? It doesn't seem right.
		// I'm afraid some downstream logic is expecting this behavior, though.
		// if resp.StatusCode == 404 { // API service is already gone
		// 	return nil, nil
		// }
		err = v1alpha2.FromHTTPResponseCode(resp.StatusCode, bodyBytes)
		return nil, err
	}
	err = nil
	log.Infof("Symphony API succeeded: %s %s, spanId: %s, traceId: %s", method, baseUrl+route, span.SpanContext().SpanID().String(), span.SpanContext().TraceID().String())

	return bodyBytes, nil
}
