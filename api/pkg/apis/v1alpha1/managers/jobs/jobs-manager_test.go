/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
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
