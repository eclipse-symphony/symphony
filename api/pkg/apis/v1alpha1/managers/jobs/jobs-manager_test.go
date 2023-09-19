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

package jobs

import (
	"context"
	"os"
	"testing"

	"github.com/azure/symphony/coa/pkg/apis/v1alpha2"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/azure/symphony/coa/pkg/apis/v1alpha2/providers/states/memorystate"
	"github.com/stretchr/testify/assert"
)

func TestHandleEvent(t *testing.T) {
	testInstanceId := os.Getenv("TEST_INSTANCE_ID")
	if testInstanceId == "" {
		t.Skip("Skipping becasue TEST_INSTANCE_ID is missing")
	}
	stateProvider := &memorystate.MemoryStateProvider{}
	stateProvider.Init(memorystate.MemoryStateProviderConfig{})
	manager := JobsManager{}
	err := manager.Init(nil, managers.ManagerConfig{
		Properties: map[string]string{
			"providers.state": "state",
			"baseUrl":         "http://localhost:8082/v1alpha2/",
			"password":        "",
			"user":            "admin",
			"interval":        "#15",
		},
	}, map[string]providers.IProvider{
		"state": stateProvider,
	})
	assert.Nil(t, err)
	errs := manager.HandleJobEvent(context.Background(), v1alpha2.Event{
		Metadata: map[string]string{
			"objectType": "instance",
		},
		Body: v1alpha2.JobData{
			Id:     testInstanceId,
			Action: "UPDATE",
		},
	})
	assert.Nil(t, errs)
}
