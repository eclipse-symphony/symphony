/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/eclipse-symphony/symphony/api/constants"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/fsnotify/fsnotify"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
)

type (
	apiClient struct {
		baseUrl       string
		tokenProvider TokenProvider
		client        *http.Client
		caCertPath    string
	}

	ApiClientOption func(*apiClient)

	TokenProvider func(ctx context.Context, baseUrl string, client *http.Client, user string, password string) (string, error)

	SummaryGetter interface {
		GetSummary(ctx context.Context, id string, namespace string, user string, password string) (*model.SummaryResult, error)
	}

	Dispatcher interface {
		QueueJob(ctx context.Context, id string, namespace string, isDelete bool, isTarget bool, user string, password string) error
		QueueDeploymentJob(ctx context.Context, namespace string, isDelete bool, deployment model.DeploymentSpec, user string, password string) error
	}

	ApiClient interface {
		SummaryGetter
		Dispatcher
		GetInstancesForAllNamespaces(ctx context.Context, user string, password string) ([]model.InstanceState, error)
		GetInstances(ctx context.Context, namespace string, user string, password string) ([]model.InstanceState, error)
		GetInstance(ctx context.Context, instance string, namespace string, user string, password string) (model.InstanceState, error)
		CreateInstance(ctx context.Context, instance string, payload []byte, namespace string, user string, password string) error
		DeleteInstance(ctx context.Context, instance string, namespace string, user string, password string) error
		DeleteTarget(ctx context.Context, target string, namespace string, user string, password string) error
		GetSolutions(ctx context.Context, namespace string, user string, password string) ([]model.SolutionState, error)
		GetSolution(ctx context.Context, solution string, namespace string, user string, password string) (model.SolutionState, error)
		CreateSolution(ctx context.Context, solution string, payload []byte, namespace string, user string, password string) error
		DeleteSolution(ctx context.Context, solution string, namespace string, user string, password string) error
		GetTargetsForAllNamespaces(ctx context.Context, user string, password string) ([]model.TargetState, error)
		GetTarget(ctx context.Context, target string, namespace string, user string, password string) (model.TargetState, error)
		GetTargets(ctx context.Context, namespace string, user string, password string) ([]model.TargetState, error)
		CreateTarget(ctx context.Context, target string, payload []byte, namespace string, user string, password string) error
		Reconcile(ctx context.Context, deployment model.DeploymentSpec, isDelete bool, namespace string, user string, password string) (model.SummarySpec, error)
		CatalogHook(ctx context.Context, payload []byte, user string, password string) error
		PublishActivationEvent(ctx context.Context, event v1alpha2.ActivationData, user string, password string) error
		GetCatalog(ctx context.Context, catalog string, namespace string, user string, password string) (model.CatalogState, error)
		UpsertCatalog(ctx context.Context, catalog string, payload []byte, user string, password string) error
		DeleteCatalog(ctx context.Context, catalog string, user string, password string) error
		UpsertSolution(ctx context.Context, solution string, payload []byte, namespace string, user string, password string) error
		GetSites(ctx context.Context, user string, password string) ([]model.SiteState, error)
		GetCatalogs(ctx context.Context, namespace string, user string, password string) ([]model.CatalogState, error)
		GetCatalogsWithFilter(ctx context.Context, namespace string, filterType string, filterValue string, user string, password string) ([]model.CatalogState, error)
		UpdateSite(ctx context.Context, site string, payload []byte, user string, password string) error
		GetABatchForSite(ctx context.Context, site string, user string, password string) (model.SyncPackage, error)
		SyncActivationStatus(ctx context.Context, status model.ActivationStatus, user string, password string) error
		SendVisualizationPacket(ctx context.Context, payload []byte, user string, password string) error
		ReportCatalogs(ctx context.Context, instance string, components []model.ComponentSpec, user string, password string) error
		CreateInstanceContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error
		DeleteInstanceContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error
		GetInstanceContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.InstanceContainerState, error)
		CreateSolutionContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error
		DeleteSolutionContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error
		GetSolutionContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.SolutionContainerState, error)
		CreateTargetContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error
		DeleteTargetContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error
		GetTargetContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.TargetContainerState, error)
		CreateCatalogContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error
		DeleteCatalogContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error
		GetCatalogContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.CatalogContainerState, error)
		CreateCampaignContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error
		DeleteCampaignContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error
		GetCampaignContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.CampaignContainerState, error)
	}
)

func noTokenProvider(ctx context.Context, baseUrl string, client *http.Client, user string, passowrd string) (string, error) {
	return "", nil
}

func WithUserPassword(ctx context.Context, user string, password string) ApiClientOption {
	return func(a *apiClient) {
		a.tokenProvider = func(ctx context.Context, baseUrl string, _ *http.Client, user string, password string) (string, error) {
			request := authRequest{Username: user, Password: password}
			requestData, _ := json.Marshal(request)
			ret, err := a.callRestAPI(ctx, "users/auth", "POST", requestData, "")
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
	}
}

func WithServiceAccountToken() ApiClientOption {
	return func(a *apiClient) {
		a.tokenProvider = func(ctx context.Context, _ string, _ *http.Client, _ string, _ string) (string, error) {
			path := os.Getenv(constants.SATokenPathName)
			if path == "" {
				path = constants.SATokenPath
			}
			token, err := os.ReadFile(path)
			if err != nil {
				return "", v1alpha2.NewCOAError(nil, "Token creation error: unable to read from volume.", v1alpha2.InternalError)
			}
			return string(token), nil
		}
	}
}

func WithCertAuth(caCertPath string) ApiClientOption {
	return func(a *apiClient) {
		a.caCertPath = caCertPath
	}
}

func NewApiClient(ctx context.Context, baseUrl string, opts ...ApiClientOption) (*apiClient, error) {
	rUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	isSecure := rUrl.Scheme == "https"

	client, err := newHttpClient(ctx, isSecure)
	if err != nil {
		return nil, err
	}

	a := &apiClient{
		baseUrl:       baseUrl,
		tokenProvider: noTokenProvider,
		client:        client,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a, nil
}

func (a *apiClient) GetInstances(ctx context.Context, namespace string, user string, password string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}
	response, err := a.callRestAPI(ctx, "instances?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) GetInstancesForAllNamespaces(ctx context.Context, user string, password string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "instances", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) GetInstance(ctx context.Context, instance string, namespace string, user string, password string) (model.InstanceState, error) {
	ret := model.InstanceState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "instances/"+url.QueryEscape(instance)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateInstance(ctx context.Context, instance string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	//use proper url encoding in the following statement
	_, err = a.callRestAPI(ctx, "instances/"+url.QueryEscape(instance)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteInstance(ctx context.Context, instance string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "instances/"+url.QueryEscape(instance)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteTarget(ctx context.Context, target string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targets/registry/"+url.QueryEscape(target)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetSolutions(ctx context.Context, namespace string, user string, password string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "solutions?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) GetSolution(ctx context.Context, solution string, namespace string, user string, password string) (model.SolutionState, error) {
	ret := model.SolutionState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "solutions/"+url.QueryEscape(solution)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateSolution(ctx context.Context, solution string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutions/"+url.QueryEscape(solution)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteSolution(ctx context.Context, solution string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutions/"+url.QueryEscape(solution)+"?namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetTarget(ctx context.Context, target string, namespace string, user string, password string) (model.TargetState, error) {
	ret := model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "targets/registry/"+url.QueryEscape(target)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) GetTargets(ctx context.Context, namespace string, user string, password string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "targets/registry?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) GetTargetsForAllNamespaces(ctx context.Context, user string, password string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "targets/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateTarget(ctx context.Context, target string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targets/registry/"+url.QueryEscape(target)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetSummary(ctx context.Context, id string, namespace string, user string, password string) (*model.SummaryResult, error) {
	result := model.SummaryResult{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return nil, err
	}

	log.Debugf("apiClient.GetSummary: id: %s, namespace: %s", id, namespace)
	ret, err := a.callRestAPI(ctx, "solution/queue?instance="+url.QueryEscape(id)+"&namespace="+url.QueryEscape(namespace), "GET", nil, token)
	// callRestApi Does a weird thing where it returns nil if the status code is 404 so we'll recreate the error here
	if err == nil && ret == nil {
		log.Debugf("apiClient.GetSummary: Not found")
		return nil, v1alpha2.NewCOAError(nil, "Not found", v1alpha2.NotFound)
	}

	if err != nil {
		return nil, err
	}
	if ret != nil {
		log.Debugf("apiClient.GetSummary: ret: %s", string(ret))
		err = json.Unmarshal(ret, &result)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func (a *apiClient) QueueDeploymentJob(ctx context.Context, namespace string, isDelete bool, deployment model.DeploymentSpec, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	path := "solution/queue"
	query := url.Values{
		"namespace":  []string{namespace},
		"delete":     []string{fmt.Sprintf("%t", isDelete)},
		"objectType": []string{"deployment"},
	}
	var payload []byte
	if err != nil {
		return err
	}
	payload, err = json.Marshal(deployment)
	log.Debugf("apiClient.QueueDeploymentJob: Deployment payload: %s", string(payload))
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, fmt.Sprintf("%s?%s", path, query.Encode()), "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated: Use QueueDeploymentJob instead
func (a *apiClient) QueueJob(ctx context.Context, id string, namespace string, isDelete bool, isTarget bool, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	path := "solution/queue"
	query := url.Values{
		"instance":   []string{id},
		"namespace":  []string{namespace},
		"delete":     []string{fmt.Sprintf("%t", isDelete)},
		"objectType": []string{"instance"},
	}

	if isTarget {
		query.Set("objectType", "target")
	}

	_, err = a.callRestAPI(ctx, fmt.Sprintf("%s?%s", path, query.Encode()), "POST", nil, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) Reconcile(ctx context.Context, deployment model.DeploymentSpec, isDelete bool, namespace string, user string, password string) (model.SummarySpec, error) {
	summary := model.SummarySpec{}
	payload, _ := json.Marshal(deployment)

	path := "solution/reconcile" + "?namespace=" + namespace
	if isDelete {
		path = path + "&delete=true"
	}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return summary, err
	}
	ret, err := a.callRestAPI(ctx, path, "POST", payload, token) // TODO: We can pass empty token now because is path is a "back-door", as it was designed to be invoked from a trusted environment, which should be also protected with auth
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

func (a *apiClient) CatalogHook(ctx context.Context, payload []byte, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	path := "federation/k8shook?objectType=catalog"
	_, err = a.callRestAPI(ctx, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) PublishActivationEvent(ctx context.Context, event v1alpha2.ActivationData, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return err
	}
	jData, _ := json.Marshal(event)
	_, err = a.callRestAPI(ctx, "jobs", "POST", jData, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetCatalog(ctx context.Context, catalog string, namespace string, user string, password string) (model.CatalogState, error) {
	ret := model.CatalogState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	catalogName := catalog
	if strings.HasPrefix(catalogName, "<") && strings.HasSuffix(catalogName, ">") {
		catalogName = catalogName[1 : len(catalogName)-1]
	}

	path := "catalogs/registry/" + url.QueryEscape(catalogName)
	if namespace != "" {
		path = path + "?namespace=" + url.QueryEscape(namespace)
	}
	response, err := a.callRestAPI(ctx, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (a *apiClient) GetCatalogsWithFilter(ctx context.Context, namespace string, filterType string, filterValue string, user string, password string) ([]model.CatalogState, error) {
	ret := make([]model.CatalogState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
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
	response, err := a.callRestAPI(ctx, path, "GET", nil, token)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}
func (a *apiClient) GetCatalogs(ctx context.Context, namespace string, user string, password string) ([]model.CatalogState, error) {
	return a.GetCatalogsWithFilter(ctx, namespace, "", "", user, password)
}

func (a *apiClient) UpsertCatalog(ctx context.Context, catalog string, payload []byte, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "catalogs/registry/"+url.QueryEscape(catalog), "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) DeleteCatalog(ctx context.Context, catalog string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "catalogs/registry/"+url.QueryEscape(catalog), "DELETE", nil, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) ReportCatalogs(ctx context.Context, instance string, components []model.ComponentSpec, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	path := "catalogs/status/" + url.QueryEscape(instance)
	jData, _ := json.Marshal(components)
	_, err = a.callRestAPI(ctx, path, "POST", jData, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) UpsertSolution(ctx context.Context, solution string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	path := "solutions/" + url.QueryEscape(solution)
	path = path + "?namespace=" + url.QueryEscape(namespace)
	_, err = a.callRestAPI(ctx, path, "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) GetSites(ctx context.Context, user string, password string) ([]model.SiteState, error) {
	ret := make([]model.SiteState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "federation/registry", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) UpdateSite(ctx context.Context, site string, payload []byte, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "federation/status/"+url.QueryEscape(site), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetABatchForSite(ctx context.Context, site string, user string, password string) (model.SyncPackage, error) {
	ret := model.SyncPackage{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "federation/sync/"+url.QueryEscape(site)+"?count=10", "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (a *apiClient) SyncActivationStatus(ctx context.Context, status model.ActivationStatus, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return err
	}
	jData, _ := json.Marshal(status)
	_, err = a.callRestAPI(ctx, "federation/sync", "POST", jData, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) SendVisualizationPacket(ctx context.Context, payload []byte, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	_, err = a.callRestAPI(ctx, "visualization", "POST", payload, token)
	if err != nil {
		return err
	}
	return nil
}

func (a *apiClient) callRestAPI(ctx context.Context, route string, method string, payload []byte, token string) ([]byte, error) {
	urlString := fmt.Sprintf("%s%s", a.baseUrl, path.Clean(route))
	ctx, span := observability.StartSpan("Symphony-API-Client", ctx, &map[string]string{
		"method":      "callRestAPI",
		"http.method": method,
		"http.url":    urlString,
	})
	var err error = nil
	defer observ_utils.CloseSpanWithError(span, &err)

	var rUrl *url.URL
	rUrl, err = url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	var req *http.Request
	var reqBody io.Reader
	if payload != nil {
		reqBody = bytes.NewBuffer(payload)
	}

	req, err = http.NewRequestWithContext(ctx, method, rUrl.String(), reqBody)
	observ_utils.PropagateSpanContextToHttpRequestHeader(req)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	var resp *http.Response
	resp, err = a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
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

func (a *apiClient) CreateInstanceContainer(ctx context.Context, instanceContainer string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}
	//use proper url encoding in the following statement
	_, err = a.callRestAPI(ctx, "instancecontainers/"+url.QueryEscape(instanceContainer)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteInstanceContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "instancecontainers/"+url.QueryEscape(instanceContainer)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetInstanceContainer(ctx context.Context, instanceContainer string, namespace string, user string, password string) (model.InstanceContainerState, error) {
	ret := model.InstanceContainerState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "instancecontainers/"+url.QueryEscape(instanceContainer)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateSolutionContainer(ctx context.Context, solutionContainer string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutioncontainers/"+url.QueryEscape(solutionContainer)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteSolutionContainer(ctx context.Context, solutionContainer string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutioncontainers/"+url.QueryEscape(solutionContainer)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetSolutionContainer(ctx context.Context, solutionContainer string, namespace string, user string, password string) (model.SolutionContainerState, error) {
	ret := model.SolutionContainerState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "solutioncontainers/"+url.QueryEscape(solutionContainer)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateTargetContainer(ctx context.Context, targetContainer string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targetcontainers/"+url.QueryEscape(targetContainer)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteTargetContainer(ctx context.Context, targetContainer string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targetcontainers/"+url.QueryEscape(targetContainer)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetTargetContainer(ctx context.Context, targetContainer string, namespace string, user string, password string) (model.TargetContainerState, error) {
	ret := model.TargetContainerState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "targetcontainers/"+url.QueryEscape(targetContainer)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateCatalogContainer(ctx context.Context, catalogContainer string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "catalogcontainers/"+url.QueryEscape(catalogContainer)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteCatalogContainer(ctx context.Context, catalogContainer string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "catalogcontainers/"+url.QueryEscape(catalogContainer)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetCatalogContainer(ctx context.Context, catalogContainer string, namespace string, user string, password string) (model.CatalogContainerState, error) {
	ret := model.CatalogContainerState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "catalogcontainers/"+url.QueryEscape(catalogContainer)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func (a *apiClient) CreateCampaignContainer(ctx context.Context, campaignContainer string, payload []byte, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "campaigncontainers/"+url.QueryEscape(campaignContainer)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteCampaignContainer(ctx context.Context, campaignContainer string, namespace string, user string, password string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "campaigncontainers/"+url.QueryEscape(campaignContainer)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetCampaignContainer(ctx context.Context, campaignContainer string, namespace string, user string, password string) (model.CampaignContainerState, error) {
	ret := model.CampaignContainerState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client, user, password)

	if err != nil {
		return ret, err
	}

	response, err := a.callRestAPI(ctx, "campaigncontainers/"+url.QueryEscape(campaignContainer)+"?namespace="+url.QueryEscape(namespace), "GET", nil, token)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(response, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func newHttpClient(ctx context.Context, secure bool) (*http.Client, error) {
	client := &http.Client{}
	if !secure {
		return client, nil
	}

	certBytes, err := os.ReadFile(apiCertPath)
	if err != nil {
		return nil, err
	}

	updateTransport := func(certBytes []byte) {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(certBytes)
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: false,
			},
		}
	}

	updateTransport(certBytes)

	// setup a file watcher to reload the cert pool when the symphony cert changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// watch for cert changes
	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					newCertBytes, readErr := os.ReadFile(apiCertPath)
					if readErr != nil {
						continue
					}
					updateTransport(newCertBytes)
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	err = watcher.Add(apiCertPath)
	if err != nil {
		return nil, err
	}

	return client, nil
}
