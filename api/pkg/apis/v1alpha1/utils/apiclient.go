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

	TokenProvider func(ctx context.Context, baseUrl string, client *http.Client) (string, error)

	SummaryGetter interface {
		GetSummary(ctx context.Context, id string, namespace string) (*model.SummaryResult, error)
	}

	Dispatcher interface {
		QueueJob(ctx context.Context, id string, namespace string, isDelete bool, isTarget bool) error
		QueueDeploymentJob(ctx context.Context, namespace string, isDelete bool, deployment model.DeploymentSpec) error
	}

	ApiClient interface {
		SummaryGetter
		Dispatcher
		GetInstancesForAllNamespaces(ctx context.Context) ([]model.InstanceState, error)
		GetInstances(ctx context.Context, namespace string) ([]model.InstanceState, error)
		GetInstance(ctx context.Context, instance string, namespace string) (model.InstanceState, error)
		CreateInstance(ctx context.Context, instance string, payload []byte, namespace string) error
		DeleteInstance(ctx context.Context, instance string, namespace string) error
		DeleteTarget(ctx context.Context, target string, namespace string) error
		GetSolutions(ctx context.Context, namespace string) ([]model.SolutionState, error)
		GetSolution(ctx context.Context, solution string, namespace string) (model.SolutionState, error)
		CreateSolution(ctx context.Context, solution string, payload []byte, namespace string) error
		DeleteSolution(ctx context.Context, solution string, namespace string) error
		GetTargetsForAllNamespaces(ctx context.Context) ([]model.TargetState, error)
		GetTarget(ctx context.Context, target string, namespace string) (model.TargetState, error)
		GetTargets(ctx context.Context, namespace string) ([]model.TargetState, error)
		CreateTarget(ctx context.Context, target string, payload []byte, namespace string) error
		Reconcile(ctx context.Context, deployment model.DeploymentSpec, isDelete bool, namespace string) (model.SummarySpec, error)
	}
)

func noTokenProvider(ctx context.Context, baseUrl string, client *http.Client) (string, error) {
	return "", nil
}

func WithUserPassword(ctx context.Context, user string, password string) ApiClientOption {
	return func(a *apiClient) {
		a.tokenProvider = func(ctx context.Context, baseUrl string, _ *http.Client) (string, error) {
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
		a.tokenProvider = func(ctx context.Context, _ string, _ *http.Client) (string, error) {
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

func NewAPIClient(ctx context.Context, baseUrl string, opts ...ApiClientOption) (*apiClient, error) {
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

func (a *apiClient) GetInstances(ctx context.Context, namespace string) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) GetInstancesForAllNamespaces(ctx context.Context) ([]model.InstanceState, error) {
	ret := make([]model.InstanceState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) GetInstance(ctx context.Context, instance string, namespace string) (model.InstanceState, error) {
	ret := model.InstanceState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)

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

func (a *apiClient) CreateInstance(ctx context.Context, instance string, payload []byte, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) DeleteInstance(ctx context.Context, instance string, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "instances/"+url.QueryEscape(instance)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteTarget(ctx context.Context, target string, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targets/registry/"+url.QueryEscape(target)+"?direct=true&namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetSolutions(ctx context.Context, namespace string) ([]model.SolutionState, error) {
	ret := make([]model.SolutionState, 0)
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) GetSolution(ctx context.Context, solution string, namespace string) (model.SolutionState, error) {
	ret := model.SolutionState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) CreateSolution(ctx context.Context, solution string, payload []byte, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutions/"+url.QueryEscape(solution)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) DeleteSolution(ctx context.Context, solution string, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "solutions/"+url.QueryEscape(solution)+"?namespace="+url.QueryEscape(namespace), "DELETE", nil, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetTarget(ctx context.Context, target string, namespace string) (model.TargetState, error) {
	ret := model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) GetTargets(ctx context.Context, namespace string) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) GetTargetsForAllNamespaces(ctx context.Context) ([]model.TargetState, error) {
	ret := []model.TargetState{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) CreateTarget(ctx context.Context, target string, payload []byte, namespace string) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return err
	}

	_, err = a.callRestAPI(ctx, "targets/registry/"+url.QueryEscape(target)+"?namespace="+url.QueryEscape(namespace), "POST", payload, token)
	if err != nil {
		return err
	}

	return nil
}

func (a *apiClient) GetSummary(ctx context.Context, id string, namespace string) (*model.SummaryResult, error) {
	result := model.SummaryResult{}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
	if err != nil {
		return nil, err
	}

	log.Infof("apiClient.GetSummary: id: %s, namespace: %s", id, namespace)
	ret, err := a.callRestAPI(ctx, "solution/queue?instance="+url.QueryEscape(id)+"&namespace="+url.QueryEscape(namespace), "GET", nil, token)
	// callRestApi Does a weird thing where it returns nil if the status code is 404 so we'll recreate the error here
	if err == nil && ret == nil {
		log.Infof("apiClient.GetSummary: Not found")
		return nil, v1alpha2.NewCOAError(nil, "Not found", v1alpha2.NotFound)
	}

	if err != nil {
		return nil, err
	}
	if ret != nil {
		log.Infof("apiClient.GetSummary: ret: %s", string(ret))
		err = json.Unmarshal(ret, &result)
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func (a *apiClient) QueueDeploymentJob(ctx context.Context, namespace string, isDelete bool, deployment model.DeploymentSpec) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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
	log.Infof("apiClient.QueueDeploymentJob: Deployment payload: %s", string(payload))
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
func (a *apiClient) QueueJob(ctx context.Context, id string, namespace string, isDelete bool, isTarget bool) error {
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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

func (a *apiClient) Reconcile(ctx context.Context, deployment model.DeploymentSpec, isDelete bool, namespace string) (model.SummarySpec, error) {
	summary := model.SummarySpec{}
	payload, _ := json.Marshal(deployment)

	path := "solution/reconcile" + "?namespace=" + namespace
	if isDelete {
		path = path + "&delete=true"
	}
	token, err := a.tokenProvider(ctx, a.baseUrl, a.client)
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
