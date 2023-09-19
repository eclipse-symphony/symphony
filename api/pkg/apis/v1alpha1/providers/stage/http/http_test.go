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

package http

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/stretchr/testify/assert"
)

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
