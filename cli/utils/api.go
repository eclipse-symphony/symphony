/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"sigs.k8s.io/yaml"
)

type authRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

func Remove(url string, username string, password string, objType string, objName string) error {
	token, err := Login(url, username, password)
	if err != nil {
		return err
	}
	route := ""
	switch objType {
	case "target", "targets":
		route = "/targets/registry"
	case "solution", "solutions":
		route = "/solutions"
	case "instance", "instances":
		route = "/instances"
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
	token, err := Login(url, username, password)
	if err != nil {
		return err
	}
	route := ""
	switch objType {
	case "target", "targets":
		route = "/targets/registry"
	case "solution", "solutions":
		route = "/solutions"
	case "instance", "instances":
		route = "/instances"
	}
	if objName == "" {
		return errors.New("object name is missing")
	}
	route += "/" + objName
	payload, err = yamlToJson(payload)
	if err != nil {
		return err
	}
	_, err = callRestAPI(url, route, "POST", payload, token, nil)
	if err != nil {
		return err
	}
	return nil
}

type YamlArtifact struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Spec       interface{}            `json:"spec"`
}

func yamlToJson(payload []byte) ([]byte, error) {
	var o YamlArtifact
	err := yaml.Unmarshal(payload, &o)
	if err != nil {
		return nil, err
	}
	return json.Marshal(o)
}

func Get(url string, username string, password string, objType string, path string, docType string, objName string) ([]interface{}, error) {
	token, err := Login(url, username, password)
	if err != nil {
		return nil, err
	}
	route := ""
	switch objType {
	case "target", "targets":
		route = "/targets/registry"
	case "device", "devices":
		route = "/devices"
	case "solution", "solutions":
		route = "/solutions"
	case "instance", "instances":
		route = "/instances"
	}
	if objName != "" {
		route += "/" + objName
	}
	params := make(map[string]string)
	if path != "" {
		params["path"] = path
	}
	if docType != "" {
		params["doc-type"] = docType
	}
	resp, err := callRestAPI(url, route, "GET", nil, token, params)
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
	data, _ := json.Marshal(authRequest{
		UserName: username,
		Password: password,
	})
	resp, err := callRestAPI(url, "/users/auth", "POST", data, "", nil)
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
	client := &http.Client{}
	rUrl := url + route
	req, err := http.NewRequest(method, rUrl, bytes.NewBuffer(payload))
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
		if resp.StatusCode == 404 { // API service is already gone
			return nil, nil
		}
		return nil, fmt.Errorf("failed to invoke Symphony API: [%d] - %v", resp.StatusCode, string(bodyBytes))
	}
	return bodyBytes, nil
}
