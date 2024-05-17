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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
)

var (
	SymphonyAPIAddressBase = "http://symphony-service:8080/v1alpha2/"
	useSAToken             = os.Getenv(constants.UseServiceAccountTokenEnvName)
	apiCertPath            = os.Getenv(constants.ApiCertEnvName)
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

func GetSymphonyAPIAddressBase() string {
	if os.Getenv(constants.SymphonyAPIUrlEnvName) == "" {
		return SymphonyAPIAddressBase
	}
	return os.Getenv(constants.SymphonyAPIUrlEnvName)
}

var symphonyApiClients sync.Map

func GetApiClient() (*apiClient, error) {
	symphonyBaseUrl := os.Getenv(constants.SymphonyAPIUrlEnvName)
	if value, ok := symphonyApiClients.Load(symphonyBaseUrl); ok {
		client, ok := value.(*apiClient)
		if !ok {
			log.Infof("Symphony base url apiclient is broken. Recreating it.")
		} else {
			return client, nil
		}
	}
	log.Infof("Creating the symphony base url apiclient.")
	client, err := getApiClient()
	if err != nil {
		log.Errorf("Failed to create the apiclient: %+v", err.Error())
		return nil, err
	}
	symphonyApiClients.Store(symphonyBaseUrl, client)
	return client, nil
}

func getApiClient() (*apiClient, error) {
	clientOptions := make([]ApiClientOption, 0)
	baseUrl := GetSymphonyAPIAddressBase()
	if caCert, ok := os.LookupEnv(constants.ApiCertEnvName); ok {
		clientOptions = append(clientOptions, WithCertAuth(caCert))
	}

	if ShouldUseSATokens() {
		clientOptions = append(clientOptions, WithServiceAccountToken())
	} else {
		clientOptions = append(clientOptions, WithUserPassword(context.TODO(), "", ""))
	}

	client, err := NewApiClient(context.Background(), baseUrl, clientOptions...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetParentApiClient(baseUrl string) (*apiClient, error) {
	clientOptions := make([]ApiClientOption, 0)

	if caCert, ok := os.LookupEnv(constants.ApiCertEnvName); ok {
		clientOptions = append(clientOptions, WithCertAuth(caCert))
	}

	clientOptions = append(clientOptions, WithUserPassword(context.TODO(), "", ""))
	client, err := NewApiClient(context.Background(), baseUrl, clientOptions...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ShouldUseSATokens() bool {
	return useSAToken == "true"
}

func ShouldUseUserCreds() bool {
	return useSAToken == "false"
}

var log = logger.NewLogger("coa.runtime")

func GetInstancesForAllNamespaces(context context.Context, baseUrl string, user string, password string) ([]model.InstanceState, error) {
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

func GetInstances(context context.Context, baseUrl string, user string, password string, namespace string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "instances?namespace=" + url.QueryEscape(namespace)
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
func ReportCatalogs(context context.Context, baseUrl string, user string, password string, instance string, components []model.ComponentSpec) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "catalogs/status/" + url.QueryEscape(instance)
	jData, _ := json.Marshal(components)
	_, err = callRestAPI(context, baseUrl, path, "POST", jData, token)
	if err != nil {
		return err
	}
	return nil
}

func GetCatalogsWithFilter(context context.Context, baseUrl string, user string, password string, namespace string, filterType string, filterValue string) ([]model.CatalogState, error) {
	ret := make([]model.CatalogState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	path := "catalogs/registry"
	if filterType != "" && filterValue != "" {
		path = path + "?filterType=" + url.QueryEscape(filterType) + "&filterValue=" + url.QueryEscape(filterValue)
		if namespace != "" {
			path = path + "&namespace=" + url.QueryEscape(namespace)
		}
	} else if namespace != "" {
		path = path + "?namespace=" + url.QueryEscape(namespace)
	}
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
func GetCatalogs(context context.Context, baseUrl string, user string, password string, namespace string) ([]model.CatalogState, error) {
	return GetCatalogsWithFilter(context, baseUrl, user, password, namespace, "", "")
}
func GetCatalog(context context.Context, baseUrl string, catalog string, user string, password string, namespace string) (model.CatalogState, error) {
	ret := model.CatalogState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	catalogName := catalog
	if strings.HasPrefix(catalogName, "<") && strings.HasSuffix(catalogName, ">") {
		catalogName = catalogName[1 : len(catalogName)-1]
	}

	var name string
	var version string
	parts := strings.Split(catalog, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return ret, errors.New("invalid catalog name")
	}

	path := "catalogs/registry/" + url.QueryEscape(name) + "/" + url.QueryEscape(version) + url.QueryEscape(catalogName)
	if namespace != "" {
		path = path + "?namespace=" + url.QueryEscape(namespace)
	}
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
func GetCampaign(context context.Context, baseUrl string, campaign string, user string, password string, namespace string) (model.CampaignState, error) {
	ret := model.CampaignState{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	var name string
	var version string
	parts := strings.Split(campaign, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return ret, errors.New("invalid campaign name")
	}

	path := "campaigns/" + url.QueryEscape(name) + "/" + url.QueryEscape(version) + url.QueryEscape(campaign)
	if namespace != "" {
		path = path + "?namespace=" + url.QueryEscape(namespace)

	}
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

	response, err := callRestAPI(context, baseUrl, "federation/sync/"+url.QueryEscape(site)+"?count=10", "GET", nil, token)
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

	response, err := callRestAPI(context, baseUrl, "activations/registry/"+url.QueryEscape(activation), "GET", nil, token)
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
	_, err = callRestAPI(context, baseUrl, "activations/status/"+url.QueryEscape(name), "POST", jData, token)
	if err != nil {
		return err
	}
	return nil
}
func GetInstance(context context.Context, baseUrl string, instance string, user string, password string, namespace string) (model.InstanceState, error) {
	ret := model.InstanceState{}
	token, err := auth(context, baseUrl, user, password)

	if err != nil {
		return ret, err
	}

	path := "instances/" + url.QueryEscape(instance)
	path = path + "?namespace=" + url.QueryEscape(namespace)
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

	var name string
	var version string
	parts := strings.Split(catalog, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid catalog name")
	}

	_, err = callRestAPI(context, baseUrl, "catalogs/registry/"+url.QueryEscape(name)+"/"+url.QueryEscape(version)+url.QueryEscape(catalog), "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func CreateInstance(context context.Context, baseUrl string, instance string, user string, password string, payload []byte, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	var name string
	var version string
	parts := strings.Split(instance, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid instance name")
	}

	path := "instances/" + url.QueryEscape(name) + "/" + url.QueryEscape(version) + url.QueryEscape(instance)
	path = path + "?namespace=" + url.QueryEscape(namespace)
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
	var name string
	var version string
	parts := strings.Split(catalog, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid catalog name")
	}

	_, err = callRestAPI(context, baseUrl, "catalogs/registry/"+url.QueryEscape(name)+"/"+url.QueryEscape(version)+url.QueryEscape(catalog), "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteInstance(context context.Context, baseUrl string, instance string, user string, password string, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	var name string
	var version string
	parts := strings.Split(instance, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid instance name")
	}

	path := "instances/" + url.QueryEscape(name) + "/" + url.QueryEscape(version) + url.QueryEscape(instance)
	path = path + "?direct=true&namespace=" + url.QueryEscape(namespace)
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTarget(context context.Context, baseUrl string, target string, user string, password string, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	var name string
	var version string
	parts := strings.Split(target, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid target name")
	}

	path := "targets/registry/" + url.QueryEscape(name) + "/" + url.QueryEscape(version) + url.QueryEscape(target)
	path = path + "?direct=true&namespace=" + url.QueryEscape(namespace)
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func GetSolutionsForAllNamespaces(context context.Context, baseUrl string, user string, password string) ([]model.SolutionState, error) {
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

func GetSolutions(context context.Context, baseUrl string, user string, password string, namespace string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "solution" + "?namespace=" + url.QueryEscape(namespace)
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

func GetSolution(context context.Context, baseUrl string, solution string, user string, password string, namespace string) (model.SolutionState, error) {
	ret := model.SolutionState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}

	var name string
	var version string
	parts := strings.Split(solution, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return ret, errors.New("invalid solution name")
	}

	path := "solutions/" + url.QueryEscape(name) + "/" + url.QueryEscape(version)
	path = path + "?namespace=" + url.QueryEscape(namespace)
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

func UpsertSolution(context context.Context, baseUrl string, solution string, user string, password string, payload []byte, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	var name string
	var version string

	log.Infof("Symphony API UpsertSolution, solution: %s namespace: %s", solution, namespace)

	parts := strings.Split(solution, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid solution name")
	}

	log.Infof("Symphony API UpsertSolution, parts: %s, %s", parts[0], parts[1])

	path := "solutions/" + url.QueryEscape(name) + "/" + url.QueryEscape(version)
	path = path + "?namespace=" + url.QueryEscape(namespace)
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteSolution(context context.Context, baseUrl string, solution string, user string, password string, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}

	var name string
	var version string

	parts := strings.Split(solution, ":")
	if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		return errors.New("invalid solution name")
	}

	path := "solutions/" + url.QueryEscape(name) + "/" + url.QueryEscape(version)
	path = path + "?namespace=" + url.QueryEscape(namespace)
	_, err = callRestAPI(context, baseUrl, path, "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func GetTarget(context context.Context, baseUrl string, target string, user string, password string, namespace string) (model.TargetState, error) {
	ret := model.TargetState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "targets/registry/" + url.QueryEscape(target)
	path = path + "?namespace=" + url.QueryEscape(namespace)
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

func GetTargetsForAllNamespaces(context context.Context, baseUrl string, user string, password string) ([]model.TargetState, error) {
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

func GetTargets(context context.Context, baseUrl string, user string, password string, namespace string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return ret, err
	}
	path := "targets/registry"
	path = path + "?namespace=" + url.QueryEscape(namespace)
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

	_, err = callRestAPI(context, baseUrl, "federation/status/"+url.QueryEscape(site), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func CreateTarget(context context.Context, baseUrl string, target string, user string, password string, payload []byte, namespace string) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "targets/registry/" + url.QueryEscape(target)
	path = path + "?namespace=" + url.QueryEscape(namespace)
	_, err = callRestAPI(context, baseUrl, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func MatchTargets(instance model.InstanceState, targets []model.TargetState) []model.TargetState {
	ret := make(map[string]model.TargetState)
	if instance.Spec.Target.Name != "" {
		for _, t := range targets {
			targetName := instance.Spec.Target.Name
			if strings.Contains(targetName, ":") {
				targetName = strings.ReplaceAll(targetName, ":", "-")
			}
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

func CreateSymphonyDeploymentFromTarget(target model.TargetState, namespace string) (model.DeploymentSpec, error) {
	key := fmt.Sprintf("%s-%s", "target-runtime", target.ObjectMeta.Name)
	scope := target.Spec.Scope
	if scope == "" {
		scope = constants.DefaultScope
	}

	ret := model.DeploymentSpec{
		ObjectNamespace: namespace,
	}
	solution := model.SolutionState{
		ObjectMeta: model.ObjectMeta{
			Name:      key,
			Namespace: target.ObjectMeta.Namespace,
		},
		Spec: &model.SolutionSpec{
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
		solution.Spec.Components = append(solution.Spec.Components, c)
	}

	targets := make(map[string]model.TargetState)
	targets[target.ObjectMeta.Name] = target

	instance := model.InstanceState{
		ObjectMeta: model.ObjectMeta{
			Name:      key,
			Namespace: target.ObjectMeta.Namespace,
		},
		Spec: &model.InstanceSpec{
			Scope:       scope,
			DisplayName: key,
			Solution:    key,
			Target: model.TargetSelector{
				Name: target.ObjectMeta.Name,
			},
		},
	}

	ret.Solution = solution
	ret.Instance = instance
	ret.Targets = targets
	ret.SolutionName = key
	// set the target generation to the deployment
	ret.Generation = target.Spec.Generation
	assignments, err := AssignComponentsToTargets(ret.Solution.Spec.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}

	return ret, nil
}

func CreateSymphonyDeployment(instance model.InstanceState, solution model.SolutionState, targets []model.TargetState, devices []model.DeviceState, namespace string) (model.DeploymentSpec, error) {
	ret := model.DeploymentSpec{
		ObjectNamespace: namespace,
	}
	ret.Generation = instance.Spec.Generation

	// convert targets
	sTargets := make(map[string]model.TargetState)
	for _, t := range targets {
		sTargets[t.ObjectMeta.Name] = t
	}

	if instance.Spec.Scope == "" {
		instance.Spec.Scope = constants.DefaultScope
	}

	//TODO: handle devices
	ret.Solution = solution
	ret.Targets = sTargets
	ret.Instance = instance
	ret.SolutionName = solution.ObjectMeta.Name
	ret.Instance.ObjectMeta.Name = instance.ObjectMeta.Name

	assignments, err := AssignComponentsToTargets(ret.Solution.Spec.Components, ret.Targets)
	if err != nil {
		return ret, err
	}

	ret.Assignments = make(map[string]string)
	for k, v := range assignments {
		ret.Assignments[k] = v
	}

	return ret, nil
}

func AssignComponentsToTargets(components []model.ComponentSpec, targets map[string]model.TargetState) (map[string]string, error) {
	//TODO: evaluate constraints
	ret := make(map[string]string)
	for key, target := range targets {
		ret[key] = ""
		for _, component := range components {
			match := true
			if component.Constraints != "" {
				parser := NewParser(component.Constraints)
				val, err := parser.Eval(utils.EvaluationContext{Properties: target.Spec.Properties})
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
func GetSummary(context context.Context, baseUrl string, user string, password string, id string, namespace string) (model.SummaryResult, error) {
	result := model.SummaryResult{}
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return result, err
	}
	path := "solution/queue"
	path = path + "?instance=" + url.QueryEscape(id) + "&namespace=" + url.QueryEscape(namespace)
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

	log.Infof("Summary result: %s", string(ret))

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

func QueueJob(context context.Context, baseUrl string, user string, password string, id string, namespace string, isDelete bool, isTarget bool) error {
	token, err := auth(context, baseUrl, user, password)
	if err != nil {
		return err
	}
	path := "solution/queue?instance=" + url.QueryEscape(id)
	if isDelete {
		path += "&delete=true"
	}
	if isTarget {
		path += "&target=true"
	}
	path = path + "&namespace=" + namespace
	_, err = callRestAPI(context, baseUrl, path, "POST", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
	if err != nil {
		return err
	}
	return nil
}
func Reconcile(context context.Context, baseUrl string, user string, password string, deployment model.DeploymentSpec, namespace string, isDelete bool) (model.SummarySpec, error) {
	summary := model.SummarySpec{}
	payload, _ := json.Marshal(deployment)

	path := "solution/reconcile" + "?namespace=" + url.QueryEscape(namespace)
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
