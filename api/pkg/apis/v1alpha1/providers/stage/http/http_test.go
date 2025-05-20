/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package http

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

func TestHttpInitWithMap(t *testing.T) {
	provider := HttpStageProvider{}
	input := map[string]string{
		"name":          "test",
		"url":           "https://www.bing.com",
		"method":        "GET",
		"successCodes":  "200, 202",
		"wait.start":    "200",
		"wait.fail":     "404",
		"wait.success":  "200",
		"wait.url":      "https://www.bing.com",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	assert.Equal(t, []int{200, 202}, provider.Config.SuccessCodes)
}

func TestHttpProcessOverrideConfig(t *testing.T) {
	provider := HttpStageProvider{}
	input := map[string]string{
		"name":          "test",
		"url":           "https://www.bing.com",
		"method":        "GET",
		"successCodes":  "200, 202",
		"wait.start":    "200",
		"wait.fail":     "404",
		"wait.success":  "200",
		"wait.url":      "https://www.bing.com",
		"wait.interval": "1",
		"wait.count":    "3",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"wait.fail": []int{500},
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.NotNil(t, outputs["body"])
	assert.Equal(t, []int{500}, provider.Config.WaitFailedCodes)
	assert.Equal(t, []int{200}, provider.Config.WaitStartCodes)
}
func TestPingBing(t *testing.T) {
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method: "GET",
		Url:    "https://www.bing.com",
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.NotNil(t, outputs["body"])
}
func TestPostRequestWithJson(t *testing.T) {
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method: "POST",
		Url:    "https://jsonplaceholder.typicode.com/posts",
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"body": map[string]interface{}{
			"title":  "foo",
			"body":   "bar",
			"userId": 1,
		},
	})
	// refer to https://jsonplaceholder.typicode.com/guide/
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.NotNil(t, outputs["body"])
	var respBody map[string]interface{}
	_, ok := outputs["body"].(string)
	assert.True(t, ok)
	bodyBytes := []byte(outputs["body"].(string))
	err = json.Unmarshal(bodyBytes, &respBody)
	assert.Nil(t, err)
	assert.Equal(t, "foo", respBody["title"])
	assert.Equal(t, "bar", respBody["body"])
	assert.EqualValues(t, 1, respBody["userId"])
	assert.EqualValues(t, 101, respBody["id"])
}
func TestCallLogicApp(t *testing.T) {
	testLogicApp := os.Getenv("TEST_HTTP_PROCESS_LOGICAPP")
	if testLogicApp != "yes" {
		t.Skip("Skipping becasue TEST_HTTP_PROCESS_LOGICAPP is missing or not set to 'yes'")
	}
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method: "POST",
		Url:    "<put your Logic App activation URL here>",
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"body": map[string]interface{}{ // this is a sample request body
			"solution": "solution1",
			"instance": "instance1",
			"target":   "target1",
			"id":       "instance1-solution1-target1",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.NotNil(t, outputs["body"])
}
func TestGitHubAction(t *testing.T) {
	testGitHubAction := os.Getenv("GET_GITHUB_ACTION")
	if testGitHubAction != "yes" {
		t.Skip("Skipping becasue GET_GITHUB_ACTION is missing or not set to 'yes'")
	}
	// sample GitHub Action
	// name: Manual workflow
	// on:
	//   repository_dispatch:
	// 	types: [hello]
	// 	inputs:
	// 	  name:
	// 		description: 'Person to greet'
	// 		required: true
	// 		type: string

	// jobs:
	//   greet:
	// 	runs-on: ubuntu-latest
	// 	steps:
	// 	- name: Send greeting
	// 	  run: echo "Hello, ${{ github.event.client_payload.name }}"

	// TODO: Waitting for GitHub action turns out to be quite complicated, as you have no straightforward way to get the run ID
	// 	 and the run ID is required to get the status of the run.
	// 	 So we are skipping this pat of the test for now (by setting WaitUrl to empty string).
	// 	 We may need to create a GitHub specific provider to handle this.
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method:           "POST",
		Url:              "https://api.github.com/repos/<user>/<repo>/dispatches",
		WaitStartCodes:   []int{204},
		WaitUrl:          "", // https://api.github.com/repos/<user>/<repo>/actions/runs"
		WaitSuccessCodes: []int{200},
		WaitFailedCodes:  []int{404},
		WaitInterval:     5,
		WaitCount:        10,
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"header.Authorization": "Bearer <GitHub token with repo scope and workflow scope>",
		"header.Content-Type":  "application/json",
		"body": map[string]interface{}{ // this is a sample request body
			"event_type": "hello",
			"client_payload": map[string]interface{}{
				"name": "bob", // how to pass input parameters to GitHub Action
			},
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 204, outputs["status"])
	assert.Equal(t, "", outputs["body"])
}

func TestFakeAPIWithJsonPathFloatResult(t *testing.T) {
	// This is what's returned from the jsonplacehoder URL:
	// {
	// 	"userId": 1,
	// 	"id": 1,
	// 	"title": "delectus aut autem",
	// 	"completed": false
	// }
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method:             "GET",
		Url:                "https://www.bing.com",
		WaitStartCodes:     []int{200},
		WaitUrl:            "https://jsonplaceholder.typicode.com/todos/1",
		WaitCount:          1,
		WaitInterval:       1,
		WaitExpression:     "$[?(@.title==\"delectus aut autem\")].id",
		WaitExpressionType: "jsonpath",
		WaitSuccessCodes:   []int{200},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.Equal(t, float64(1), outputs["waitResult"])
}

func TestFakeAPIWithJsonPathArrayResult(t *testing.T) {
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method:             "GET",
		Url:                "https://www.bing.com",
		WaitStartCodes:     []int{200},
		WaitUrl:            "https://jsonplaceholder.typicode.com/users",
		WaitCount:          1,
		WaitInterval:       1,
		WaitExpression:     "$[:3].id",
		WaitExpressionType: "jsonpath",
		WaitSuccessCodes:   []int{200},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.EqualValues(t, []interface{}{float64(1), float64(2), float64(3)}, outputs["waitResult"])
}

func TestFakeAPIWithJsonPathMapResult(t *testing.T) {
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method:             "GET",
		Url:                "https://www.bing.com",
		WaitStartCodes:     []int{200},
		WaitUrl:            "https://jsonplaceholder.typicode.com/users/1",
		WaitCount:          1,
		WaitInterval:       1,
		WaitExpression:     "$.address",
		WaitExpressionType: "jsonpath",
		WaitSuccessCodes:   []int{200},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.EqualValues(
		t,
		map[string]interface{}{
			"street":  "Kulas Light",
			"suite":   "Apt. 556",
			"city":    "Gwenborough",
			"zipcode": "92998-3874",
			"geo":     map[string]interface{}{"lat": "-37.3159", "lng": "81.1496"}},
		outputs["waitResult"])
}

func TestFakeAPIWithSymphonyExpression(t *testing.T) {
	// This is what's returned from the jsonplacehoder URL:
	// {
	// 	"userId": 1,
	// 	"id": 1,
	// 	"title": "delectus aut autem",
	// 	"completed": false
	// }
	provider := HttpStageProvider{}
	err := provider.Init(HttpStageProviderConfig{
		Method:           "GET",
		Url:              "https://www.bing.com",
		WaitStartCodes:   []int{200},
		WaitUrl:          "https://jsonplaceholder.typicode.com/todos/1",
		WaitCount:        1,
		WaitInterval:     1,
		WaitExpression:   "${{$equal($val('$.title'), 'delectus aut autem')}}",
		WaitSuccessCodes: []int{200},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, 200, outputs["status"])
	assert.Equal(t, true, outputs["waitResult"])
}
