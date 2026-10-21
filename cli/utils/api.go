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
	"io"
	"net/http"

	"sigs.k8s.io/yaml"
)

type authRequest struct {
	UserName  string `json:"username"`
	Password  string `json:"password"`
	CsrfToken string `json:"csrfToken,omitempty"`
}

type authResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

// ChatMessage represents a single message in an OpenAI-compatible chat exchange.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// ChatCompletion sends an OpenAI-compatible chat completion request to the
// Symphony API's model router endpoint and returns the assistant's reply.
// When endpoint is empty, the server's default model router endpoint is used.
func ChatCompletion(url string, username string, password string, endpoint string, model string, messages []ChatMessage) (string, error) {
	token, err := Login(url, username, password)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(chatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	params := make(map[string]string)
	if endpoint != "" {
		params["endpoint"] = endpoint
	}
	resp, err := callRestAPI(url, "/modelrouter/chat/completions", "POST", payload, token, params)
	if err != nil {
		return "", err
	}
	var chatResp chatCompletionResponse
	if err := json.Unmarshal(resp, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse chat response: %v", err)
	}
	if len(chatResp.Choices) == 0 {
		return "", errors.New("no response returned by the model")
	}
	return chatResp.Choices[0].Message.Content, nil
}

func Remove(url string, username string, password string, objType string, objName string) error {
	token, err := Login(url, username, password)
	if err != nil {
		return err
	}
	route := ""
	switch objType {
	case "solution", "solutions":
		route = "/solutions"
	case "activation", "activations":
		route = "/activations/registry"
	case "target", "targets":
		route = "/targets/registry"
	case "solutionversion", "solutionversions":
		route = "/solutionversions"
	case "instance", "instances":
		route = "/instances"
	case "campaign", "campaigns":
		route = "/campaigns"
	case "campaignversion", "campaignversions":
		route = "/campaignversions"
	}
	if objName == "" {
		return errors.New("object name is missing")
	}
	route += "/" + objName
	_, err = callRestAPI(url, route, "DELETE", nil, token, nil)
	if err != nil {
		return err
	}
	return nil
}
func Upsert(url string, username string, password string, objType string, objName string, payload []byte) error {
	return UpsertInNamespace(context.Background(), url, username, password, objType, objName, payload, "")
}

func UpsertInNamespace(ctx context.Context, url string, username string, password string, objType string, objName string, payload []byte, namespace string) error {
	token, err := loginWithContext(ctx, url, username, password)
	if err != nil {
		return err
	}
	route := ""
	switch objType {
	case "solution", "solutions":
		route = "/solutions"
	case "activation", "activations":
		route = "/activations/registry"
	case "target", "targets":
		route = "/targets/registry"
	case "solutionversion", "solutionversions":
		route = "/solutionversions"
	case "instance", "instances":
		route = "/instances"
	case "campaign", "campaigns":
		route = "/campaigns"
	case "campaignversion", "campaignversions":
		route = "/campaignversions"
	}
	if objName == "" {
		return errors.New("object name is missing")
	}
	route += "/" + objName
	payload, err = yamlToJson(payload)
	if err != nil {
		return err
	}
	params := make(map[string]string)
	if namespace != "" {
		params["namespace"] = namespace
	}
	_, err = callRestAPIWithContext(ctx, url, route, "POST", payload, token, params)
	if err != nil {
		return err
	}
	return nil
}

func yamlToJson(payload []byte) ([]byte, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(payload, &m); err != nil {
		return nil, err
	}
	delete(m, "apiVersion")
	delete(m, "kind")
	return json.Marshal(m)
}

func Get(url string, username string, password string, objType string, path string, docType string, objName string) ([]interface{}, error) {
	return GetInNamespace(context.Background(), url, username, password, objType, path, docType, objName, "")
}

func GetInNamespace(ctx context.Context, url string, username string, password string, objType string, path string, docType string, objName string, namespace string) ([]interface{}, error) {
	token, err := loginWithContext(ctx, url, username, password)
	if err != nil {
		return nil, err
	}
	route := ""
	switch objType {
	case "activation", "activations":
		route = "/activations/registry"
	case "target", "targets":
		route = "/targets/registry"
	case "device", "devices":
		route = "/devices"
	case "solution", "solutions":
		route = "/solutions"
	case "campaign", "campaigns":
		route = "/campaigns"
	case "solutionversion", "solutionversions":
		route = "/solutionversions"
	case "instance", "instances":
		route = "/instances"
	case "catalog", "catalogs":
		route = "/catalogs"
	case "catalogversion", "catalogversions":
		route = "/catalogversions/registry"
	case "campaignversion", "campaignversions":
		route = "/campaignversions"
	default:
		return nil, fmt.Errorf("unsupported object type: %s", objType)
	}
	if objName != "" {
		route += "/" + objName
	}
	params := make(map[string]string)
	if namespace != "" {
		params["namespace"] = namespace
	}
	if path != "" {
		params["path"] = path
	}
	if docType != "" {
		params["doc-type"] = docType
	}
	resp, err := callRestAPIWithContext(ctx, url, route, "GET", nil, token, params)
	if err != nil {
		return nil, err
	}
	var ret []interface{}

	if objName != "" {
		var obj interface{}
		err = json.Unmarshal(resp, &obj)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	} else {
		err = json.Unmarshal(resp, &ret)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func Login(url string, username string, password string) (string, error) {
	return loginWithContext(context.Background(), url, username, password)
}

func loginWithContext(ctx context.Context, url string, username string, password string) (string, error) {
	data, _ := json.Marshal(authRequest{
		UserName: username,
		Password: password,
	})
	resp, err := callRestAPIWithContext(ctx, url, "/users/auth", "POST", data, "", nil)
	if err != nil {
		return "", err
	}
	var authResp authResponse
	err = json.Unmarshal(resp, &authResp)
	if err != nil {
		return "", err
	}
	return "Bearer " + authResp.AccessToken, nil
}

func callRestAPI(url string, route string, method string, payload []byte, token string, parameters map[string]string) ([]byte, error) {
	return callRestAPIWithContext(context.Background(), url, route, method, payload, token, parameters)
}

func callRestAPIWithContext(ctx context.Context, url string, route string, method string, payload []byte, token string, parameters map[string]string) ([]byte, error) {
	client := &http.Client{}
	rUrl := url + route
	req, err := http.NewRequestWithContext(ctx, method, rUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", token)
	}

	if parameters != nil {
		query := req.URL.Query()
		for k, v := range parameters {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		if resp.StatusCode == 404 && method == http.MethodDelete {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to invoke Symphony API: [%d] - %v", resp.StatusCode, string(bodyBytes))
	}
	return bodyBytes, nil
}
